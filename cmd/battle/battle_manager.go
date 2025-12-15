package main

import (
	"fmt"
	"sync"
	"time"

	"goPureWithCsharp/csharp"
	pb "goPureWithCsharp/csharp/proto"

	"google.golang.org/protobuf/proto"
)

// CreateBattleCommand 创建战斗命令
type CreateBattleCommand struct {
	BattleID  uint64
	AtkTeamID uint32
	DefTeamID uint32
	ResChan   chan error // 返回错误结果
}

// BattleOutputEvent 战斗输出事件
type BattleOutputEvent struct {
	Timestamp time.Time
	// 可以添加更多字段用于存储输出数据
}

// ============================================================================
// BattleManager 状态定义
// ============================================================================
type BattleManagerState int

const (
	StateCreated BattleManagerState = iota
	StateRunning
	StateStopped
)

// ============================================================================
type EventBus interface {
	SubscribeCtx() <-chan *pb.BattleContext
	Publish(outputEv *pb.BattleContext)
}

// EventBusImpl 事件总线实现
type EventBusImpl struct {
	eventChan chan *pb.BattleContext
	mu        sync.RWMutex
	closed    bool
}

func NewEventBus(bufferSize int) *EventBusImpl {
	return &EventBusImpl{
		eventChan: make(chan *pb.BattleContext, bufferSize),
	}
}

func (eb *EventBusImpl) SubscribeCtx() <-chan *pb.BattleContext {
	return eb.eventChan
}

func (eb *EventBusImpl) Publish(event *pb.BattleContext) {
	eb.mu.RLock()
	closed := eb.closed
	eb.mu.RUnlock()
	if closed {
		return
	}
	select {
	case eb.eventChan <- event:
	default:
		switch event.Option.(type) {
		case *pb.BattleContext_BattleInput:
			fmt.Printf("[EventBus] !!!!!!!!!!!!!!!事件队列已满， battId %d 丢弃输入\n", event.BattleId)
		case *pb.BattleContext_BattleOutput:
			fmt.Printf("[EventBus] !!!!!!!!!!!!!!!事件队列已满， battId %d 丢弃输出\n", event.BattleId)
		}
	}
}

func (eb *EventBusImpl) Close() {
	eb.mu.Lock()
	defer eb.mu.Unlock()
	if !eb.closed {
		close(eb.eventChan)
		eb.closed = true
	}
}

// ============================================================================
// BattleDispatcher - 战斗调度接口
// ============================================================================
type BattleDisptcher interface {
	CreateBattle(battleID uint64, env *pb.BattleEnv) error
	InputBattle(battleID uint64, inputData proto.Message) error // inputData proto.Message 是 oneof pb.BattleInput
	DestroyBattle(battleID uint64) error
	DisptcherShutDown() error
}

// Proxy 战斗调度代理
type Proxy struct {
	mu                sync.RWMutex
	bcMap             map[uint64]*BattleController
	frameSeqGenerator FrameSeqProvider
}

func NewProxy(frameSeqGenerator FrameSeqProvider) *Proxy {
	return &Proxy{
		bcMap:             make(map[uint64]*BattleController),
		frameSeqGenerator: frameSeqGenerator,
	}
}

func (p *Proxy) CreateBattle(battleID uint64, env *pb.BattleEnv) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if _, exists := p.bcMap[battleID]; exists {
		return fmt.Errorf("战斗 %d 已存在", battleID)
	}

	bc := NewBattleController(p.frameSeqGenerator, p)

	var atkTeamID, defTeamID uint32
	if env.Atk != nil {
		atkTeamID = env.Atk.TeamId
	}
	if env.Def != nil {
		defTeamID = env.Def.TeamId
	}

	if err := csharp.CreateBattle(uint32(battleID), atkTeamID, defTeamID); err != nil {
		return fmt.Errorf("C# 创建战斗失败: %w", err)
	}

	p.bcMap[battleID] = bc

	return nil
}

func (p *Proxy) InputBattle(battleID uint64, inputData proto.Message) error {
	p.mu.RLock()
	bc, exists := p.bcMap[battleID]
	p.mu.RUnlock()

	if !exists {
		return fmt.Errorf("战斗 %d 不存在", battleID)
	}

	return bc.BattleInput(battleID, inputData)
}

func (p *Proxy) DestroyBattle(battleID uint64) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	_, exists := p.bcMap[battleID]
	if !exists {
		return fmt.Errorf("战斗 %d 不存在", battleID)
	}

	if err := csharp.DestroyBattle(battleID); err != nil {
		return fmt.Errorf("C# 销毁战斗失败: %w", err)
	}

	delete(p.bcMap, battleID)

	return nil
}

// GetBattleController 获取战斗控制器（内部使用）
func (p *Proxy) GetBattleController(battleID uint64) (*BattleController, bool) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	bc, exists := p.bcMap[battleID]
	return bc, exists
}

// DisptcherShutDown 关闭所有战斗
func (p *Proxy) DisptcherShutDown() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	// 销毁所有战斗
	for battleID := range p.bcMap {
		if err := csharp.DestroyBattle(battleID); err != nil {
			fmt.Printf("[Proxy] 销毁战斗 %d 失败: %v\n", battleID, err)
		}
	}
	p.bcMap = make(map[uint64]*BattleController)

	return nil
}

// OutPutResult 实现 BattleOutput 接口
func (p *Proxy) OutPutResult(result *pb.BattleResult) error {
	fmt.Printf("[Proxy] 接收到战斗结果: Winner=%d, Loser=%d\n", result.GetWinner(), result.GetLoser())
	return nil
}

// OutPutReply 实现 BattleOutput 接口
func (p *Proxy) OutPutReply(replay *pb.BattleReplay) error {
	fmt.Printf("[Proxy] 接收到战斗回放: BattleId=%d, Events=%d\n", replay.GetBattleId(), len(replay.GetEvents()))
	return nil
}

// ============================================================================
// BattleManager 单例管理
// ============================================================================

var (
	globalBattleManagerInstance *BattleManager
	globalBattleManagerMutex    sync.Once
)

// GetBattleManager 获取 BattleManager 全局单例
// 线程安全，只初始化一次
func GetBattleManager() *BattleManager {
	globalBattleManagerMutex.Do(func() {
		globalBattleManagerInstance = NewBattleManager(30) // 默认 30fps
	})
	return globalBattleManagerInstance
}

// GetBattleManagerWithFPS 获取 BattleManager 全局单例（指定FPS）
// 只在第一次调用时有效，后续调用返回已存在的实例
func GetBattleManagerWithFPS(fps int64) *BattleManager {
	globalBattleManagerMutex.Do(func() {
		globalBattleManagerInstance = NewBattleManager(fps)
	})
	return globalBattleManagerInstance
}

// InitBattleManager 初始化 BattleManager 全局单例
// 必须在 GetBattleManager 之前调用，否则无效
func InitBattleManager(builder *BattleManagerBuilder) *BattleManager {
	globalBattleManagerMutex.Do(func() {
		globalBattleManagerInstance = builder.Build()
	})
	return globalBattleManagerInstance
}

// ResetBattleManager 重置 BattleManager 单例（仅用于测试）
// 注意：线程不安全，只应在测试中使用
func ResetBattleManager() {
	globalBattleManagerInstance = nil
	globalBattleManagerMutex = sync.Once{}
}

// ============================================================================
// BattleManager - 战斗管理器
// ============================================================================
type BattleManager struct {
	EventBus // 嵌入 EventBus 接口

	fpsProvider FrameSeqProvider
	createChan  chan *pb.BattleEnv // 创建战斗命令通道
	outPutChan  chan *pb.BattleContext

	battleCtrls BattleDisptcher

	// 状态管理
	mu       sync.RWMutex
	state    BattleManagerState
	stopChan chan struct{}
}

func NewBattleManager(fps int64) *BattleManager {
	return NewBattleManagerBuilder().
		WithFPS(fps).
		Build()
}

func (bm *BattleManager) Init() error {

	err := csharp.InitCSharpLib("Release")
	if err != nil {
		fmt.Printf("[Battle] ✗ C# 库加载失败: %v\n", err)
		return err
	}

	err = bm.prepareCallback()
	if err != nil {
		fmt.Printf("[Battle] ✗ 回调准备失败: %v\n", err)
		return err
	}

	return nil
}

func (bm *BattleManager) Dispose() error {
	return csharp.CloseCSharpLib()
}

func (bm *BattleManager) prepareCallback() error {

	err := csharp.RegisterConfigLoader(loadConfig)
	if err != nil {
		return err
	}

	// TODO : 注入CrashHandler
	// [UnmanagedCallersOnly(EntryPoint = "InitLibrary")]
	// public static void InitLibrary()
	// {
	//     // 捕获未处理的托管异常
	//     AppDomain.CurrentDomain.UnhandledException += (sender, e) =>
	//     {
	//         string log = $"Crash detected: {e.ExceptionObject}";
	//         File.AppendAllText("native_crash.log", log);
	//         // 注意：这里执行完后，进程依然会终止
	//     };
	// }
	// err = csharp.RegisterPanicOutput(Painic)
	// if err != nil {
	// 	return err
	// }

	err = csharp.RegisterBattleEndNotify(battleOutput)
	if err != nil {
		return err
	}
	return nil
}

// Start 启动 BattleManager
func (bm *BattleManager) Start() error {

	if bm.state != StateCreated {
		return fmt.Errorf("无法启动，当前状态: %v", bm.state)
	}
	bm.state = StateRunning
	err := bm.Init() // 库加载和回调注册
	if err != nil {
		return err

	}

	go bm.run()
	return nil
}

// Stop 停止 BattleManager
func (bm *BattleManager) Stop() {
	bm.mu.Lock()
	if bm.state != StateRunning {
		bm.mu.Unlock()
		return
	}
	bm.state = StateStopped
	bm.mu.Unlock()

	fmt.Println("[BattleManager] 停止运行...")

	// 销毁所有战斗
	bm.battleCtrls.DisptcherShutDown()

	// 关闭通道
	close(bm.createChan)
	if eb, ok := bm.EventBus.(*EventBusImpl); ok {
		eb.Close()
	}

	// 等待事件循环退出
	close(bm.stopChan)

	bm.Dispose()
	fmt.Println("[BattleManager] ✓ 已停止")
}

// IsRunning 返回是否正在运行
func (bm *BattleManager) IsRunning() bool {
	bm.mu.RLock()
	defer bm.mu.RUnlock()
	return bm.state == StateRunning
}

// 便捷访问方法
func (bm *BattleManager) GetCreateChannel() chan<- *pb.BattleEnv {
	return bm.createChan
}

// run 主事件循环
func (bm *BattleManager) run() {
	fmt.Println("[BattleManager] 启动事件循环")
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("[BattleManager] 事件循环崩溃: %v\n", r)
		}
		bm.mu.Lock()
		bm.state = StateStopped
		bm.mu.Unlock()
		fmt.Println("[BattleManager] 事件循环已退出")
	}()

	ticker := time.NewTicker(time.Second * 1) // 每5秒处理一次逻辑帧 以免业务堆积
	defer ticker.Stop()
	for {
		select {
		case cmd := <-bm.createChan:
			bm.handleCreateBattle(cmd)
		case ctx := <-bm.SubscribeCtx():
			bm.handleProcessBattleCtx(ctx)
		case <-ticker.C:
			bm.processTick()
		case <-bm.stopChan:
			fmt.Println("[BattleManager] 收到停止信号，退出事件循环")
			return
		}
	}
}

// handleCreateBattle 处理创建战斗命令
func (bm *BattleManager) handleCreateBattle(e *pb.BattleEnv) error {
	bId := uint64(e.BattleId)
	if bId == 0 {
		bId = uint64(GenerateBattleID())
	}
	fmt.Printf("[BattleManager] 创建战斗命令 - ID: %d\n", bId)
	err := bm.battleCtrls.CreateBattle(bId, e)
	if err != nil {
		fmt.Printf("[BattleManager] 创建战斗失败: %v\n", err)
		return err
	}
	return nil
}

// processTick 处理逻辑帧事件
func (bm *BattleManager) processTick() {
	processed, err := csharp.OnTick()
	if err != nil {
		fmt.Printf("[BattleManager] OnTick 失败: %v\n", err)
		return
	}

	if processed > 0 {
		fmt.Printf("[BattleManager] 处理了 %d 场战斗\n", processed)
	}
}

func (bm *BattleManager) handleProcessBattleCtx(e *pb.BattleContext) error {
	e.Tick = e.GetTick()

	switch e.Option.(type) {
	case *pb.BattleContext_BattleInput:
		err := bm.battleCtrls.InputBattle(uint64(e.GetBattleId()), e)
		if err != nil {
			fmt.Printf("[BattleManager] 输入战斗失败: %v\n", err)
			return err
		}
	case *pb.BattleContext_BattleOutput:

		switch output := e.GetBattleOutput().Output.(type) {
		case *pb.BattleOutput_Result:
			fmt.Printf("[BattleManager] 战斗输出 - 结果: BattleID=%d, Winner=%d, Loser=%d\n", e.GetBattleId(),
				output.Result.GetWinner(), output.Result.GetLoser())
			//TODO 定时删除 结束的战斗
		case *pb.BattleOutput_Replay:
			fmt.Printf("[BattleManager] 战斗输出 - 回放: BattleID=%d, Events=%d\n", e.GetBattleId(),
				len(output.Replay.GetEvents()))
		}
		bm.outPutChan <- e // 透传
	default:
		fmt.Printf("[BattleManager] 未知的 BattleContext 类型 - BattleID: %d, Tick: %d\n", e.GetBattleId(), e.GetTick())
		return fmt.Errorf("未知的 BattleContext 类型")

	}
	return nil
}
