package csharp

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"unsafe"

	proto_pb "goPureWithCsharp/csharp/proto"

	"github.com/ebitengine/purego"
	"google.golang.org/protobuf/proto"
)

var (
	libHandle uintptr
	libMutex  sync.RWMutex

	// Go 侧日志级别控制
	goLogLevel = LogLevelInfo
	goLogMutex sync.RWMutex

	// 函数指针缓存 - 加速 SyscallN 调用
	fnCache sync.Map

	requiredFuncs = []string{
		// 低级 API
		"ProcessProtoMessage",
		"ProcessBatchProtoMessage",
		"RegisterCallback",
		"TestNotifyCallback",

		// 高级 API - 战斗管理
		"CreateBattle",
		"DestroyBattle",
		"OnTick",
		"GetBattleCount",
		"ProcessBattleInput",

		// 高级 API - 日志控制
		"SetBattleLogLevel",
		"GetBattleLogLevel",
		"EnableBattleLogging",
		"DisableBattleLogging",

		// 其他 API
		"CsharpPanic",
		"RegisterBattleResultCallback",
	}
)

// goLog 根据日志级别输出日志
func goLog(level int, format string, args ...interface{}) {
	goLogMutex.RLock()
	defer goLogMutex.RUnlock()

	if level >= goLogLevel {
		fmt.Printf(format, args...)
	}
}

// ============================================================================
// 函数指针缓存和加速机制
// ============================================================================

// getCachedFunction 从缓存获取函数指针，如果不存在则加载并缓存
// 这个函数加快了重复调用的速度，避免每次都调用 Dlsym
// 使用 sync.Map 自动处理并发，无需手动管理锁
func getCachedFunction(libHandle uintptr, funcName string) (uintptr, error) {
	// 先尝试从缓存读取
	if fnPtr, ok := fnCache.Load(funcName); ok {
		return fnPtr.(uintptr), nil
	}

	// 缓存未命中，从库中加载函数指针
	fnPtr, err := purego.Dlsym(libHandle, funcName)
	if err != nil {
		return 0, fmt.Errorf("找不到函数: %s - %w", funcName, err)
	}

	// 写入缓存（sync.Map 内部自动处理并发）
	fnCache.Store(funcName, fnPtr)
	return fnPtr, nil
}

// validateLibrary 验证 SO 文件是否包含所有必需的导出函数
// 在库初始化时调用，确保 SO 文件完整
func validateLibrary(libHandle uintptr) error {
	missingFuncs := []string{}

	for _, funcName := range requiredFuncs {
		if _, err := getCachedFunction(libHandle, funcName); err != nil {
			missingFuncs = append(missingFuncs, funcName)
		}
	}

	if len(missingFuncs) > 0 {
		return fmt.Errorf("SO 文件缺少以下导出函数: %v", missingFuncs)
	}

	fmt.Println("[Go] SO 文件验证成功，所有必需函数已找到")
	return nil
}

// clearFunctionCache 清空函数指针缓存
// 在库卸载时调用
func clearFunctionCache() {
	// sync.Map 没有直接的清空方法
	// 使用 Range + Delete 遍历清空
	fnCache.Range(func(key, value interface{}) bool {
		fnCache.Delete(key)
		return true
	})
}

// ============================================================================
// 低级 API (与 C# 直接交互)
// ============================================================================

// ProcessProtoMessage 处理单个 Protobuf 消息 (低级 API)
// 使用缓存的函数指针加速调用
func ProcessProtoMessage(requestData []byte) ([]byte, error) {
	libMutex.RLock()
	defer libMutex.RUnlock()

	if libHandle == 0 {
		return nil, fmt.Errorf("C# 库未初始化")
	}

	// 获取缓存的函数指针
	fnPtr, err := getCachedFunction(libHandle, "ProcessProtoMessage")
	if err != nil {
		return nil, err
	}

	// 准备响应缓冲区
	respBuffer := make([]byte, 10240)
	respLen := int32(len(respBuffer))

	// 构建函数调用
	fn := unsafe.Pointer(fnPtr)
	purego.SyscallN(
		uintptr(fn),
		uintptr(unsafe.Pointer(&requestData[0])),
		uintptr(len(requestData)),
		uintptr(unsafe.Pointer(&respBuffer[0])),
		uintptr(unsafe.Pointer(&respLen)),
	)

	return respBuffer[:respLen], nil
}

// ProcessBatchProtoMessage 批量处理 Protobuf 消息 (低级 API)
// 使用缓存的函数指针加速调用
func ProcessBatchProtoMessage(requestData []byte) ([]byte, error) {
	libMutex.RLock()
	defer libMutex.RUnlock()

	if libHandle == 0 {
		return nil, fmt.Errorf("C# 库未初始化")
	}

	// 获取缓存的函数指针
	fnPtr, err := getCachedFunction(libHandle, "ProcessBatchProtoMessage")
	if err != nil {
		return nil, err
	}

	// 准备响应缓冲区
	respBuffer := make([]byte, 102400)
	respLen := int32(len(respBuffer))

	// 构建函数调用
	fn := unsafe.Pointer(fnPtr)
	purego.SyscallN(
		uintptr(fn),
		uintptr(unsafe.Pointer(&requestData[0])),
		uintptr(len(requestData)),
		uintptr(unsafe.Pointer(&respBuffer[0])),
		uintptr(unsafe.Pointer(&respLen)),
	)

	return respBuffer[:respLen], nil
}

// RegisterCallback 注册 Go 回调函数到 C#
func RegisterCallback(callbackPtr unsafe.Pointer) error {
	libMutex.RLock()
	defer libMutex.RUnlock()

	if libHandle == 0 {
		return fmt.Errorf("C# 库未初始化")
	}

	fnPtr, err := getCachedFunction(libHandle, "RegisterCallback")
	if err != nil {
		return fmt.Errorf("找不到函数: RegisterCallback - %w", err)
	}

	fn := unsafe.Pointer(fnPtr)
	purego.SyscallN(
		uintptr(fn),
		uintptr(callbackPtr),
	)

	fmt.Println("[Go] 回调函数已注册给 C#")
	return nil
}

// TestNotifyCallback 测试 C# 侧触发回调
// 用于验证 Go 回调是否正确工作
func TestNotifyCallback(notificationType int32, battleID int64, timestamp int64) (int32, error) {
	libMutex.RLock()
	defer libMutex.RUnlock()

	if libHandle == 0 {
		return -1, fmt.Errorf("C# 库未初始化")
	}

	fnPtr, err := getCachedFunction(libHandle, "TestNotifyCallback")
	if err != nil {
		return -1, fmt.Errorf("找不到函数: TestNotifyCallback - %w", err)
	}

	fn := unsafe.Pointer(fnPtr)
	result, _, _ := purego.SyscallN(
		uintptr(fn),
		uintptr(notificationType),
		uintptr(battleID),
		uintptr(timestamp),
	)

	fmt.Printf("[Go] C# 回调测试结果: %d\n", result)
	return int32(result), nil
}

// ============================================================================
// 库生命周期管理
// ============================================================================

// InitCSharpLib 初始化 C# 动态库
func InitCSharpLib(version string) error {
	libMutex.Lock()
	defer libMutex.Unlock()

	// 如果已经初始化，先关闭旧的库句柄
	if libHandle != 0 {
		fmt.Printf("[InitCSharpLib] 检测到旧库句柄 %d，正在关闭...\n", libHandle)
		// 尝试关闭旧的句柄，但不失败
		_ = purego.Dlclose(libHandle)
		libHandle = 0
		fmt.Println("[InitCSharpLib] 旧库句柄已关闭")
	}

	if version == "" {
		version = "Release"
	}

	// 先尝试相对路径（从项目根目录运行）
	libPath := filepath.Join("lib", fmt.Sprintf("TestExport_%s.so", version))

	// 检查文件是否存在
	if _, err := os.Stat(libPath); err != nil {
		// 如果相对路径不存在，尝试从当前工作目录向上查找项目根目录
		// 这适用于从 cmd/test 等子目录运行测试的情况
		wd, err2 := os.Getwd()
		if err2 == nil {
			// 尝试向上一级
			projectRoot := filepath.Dir(wd)
			altPath := filepath.Join(projectRoot, "lib", fmt.Sprintf("TestExport_%s.so", version))
			if _, err3 := os.Stat(altPath); err3 == nil {
				libPath = altPath
			} else {
				// 再尝试向上一级（从 cmd/test 到项目根目录）
				projectRoot = filepath.Dir(projectRoot)
				altPath = filepath.Join(projectRoot, "lib", fmt.Sprintf("TestExport_%s.so", version))
				if _, err4 := os.Stat(altPath); err4 == nil {
					libPath = altPath
				} else {
					return fmt.Errorf("库文件不存在: %s - %w", libPath, err)
				}
			}
		} else {
			return fmt.Errorf("库文件不存在: %s - %w", libPath, err)
		}
	}

	// 打开动态库
	handle, err := purego.Dlopen(libPath, purego.RTLD_NOW|purego.RTLD_GLOBAL)
	if err != nil {
		return fmt.Errorf("打开库失败: %s - %w", libPath, err)
	}

	libHandle = handle
	fmt.Printf("[Go] C# 库已加载: %s (handle=%d)\n", libPath, libHandle)

	// 验证库中所有必需的导出函数
	if err := validateLibrary(libHandle); err != nil {
		// 关闭库并清空句柄
		_ = purego.Dlclose(libHandle)
		libHandle = 0
		clearFunctionCache()
		return err
	}

	return nil
}

// CloseCSharpLib 关闭 C# 动态库
func CloseCSharpLib() error {
	libMutex.Lock()
	defer libMutex.Unlock()

	if libHandle == 0 {
		// 库未初始化，不是错误
		return nil
	}

	// 关闭库
	err := purego.Dlclose(libHandle)
	libHandle = 0

	// 清空函数指针缓存
	clearFunctionCache()

	if err != nil {
		// 记录错误但继续，因为库句柄已经清空
		fmt.Printf("[Go] 关闭库时出错: %v\n", err)
		return nil
	}

	fmt.Println("[Go] C# 库已关闭")
	return nil
}

func CallCSharpPainc() error {
	libMutex.RLock()
	defer libMutex.RUnlock()

	if libHandle == 0 {
		return fmt.Errorf("C# 库未初始化")
	}

	fnPtr, err := getCachedFunction(libHandle, "CsharpPanic")
	if err != nil {
		return fmt.Errorf("找不到函数: CsharpPanic - %w", err)
	}

	result, _, _ := purego.SyscallN(
		uintptr(fnPtr),
	)

	if result != 0 {
		return fmt.Errorf("CsharpPanic 返回错误: %d", result)
	}
	return nil
}

// ============================================================================
// 战斗管理 API (BattleManager 导出)
// ============================================================================

// BattleResultCallbackFunc Go 侧的战斗结果回调签名
// 返回: 返回值码 (0=成功, -1=失败), BattleContext 数据指针, 数据长度
type BattleResultCallbackFunc func() (int32, unsafe.Pointer, int32)

// RegisterBattleResultCallback 注册战斗结果回调
func RegisterBattleResultCallback(fn BattleResultCallbackFunc) error {
	libMutex.RLock()
	defer libMutex.RUnlock()

	if libHandle == 0 {
		return fmt.Errorf("C# 库未初始化")
	}

	fnPtr, err := getCachedFunction(libHandle, "RegisterBattleResultCallback")
	if err != nil {
		return fmt.Errorf("找不到函数: RegisterBattleResultCallback - %w", err)
	}

	callbackPtr := unsafe.Pointer(&fn)

	fn2 := unsafe.Pointer(fnPtr)
	result, _, _ := purego.SyscallN(
		uintptr(fn2),
		uintptr(callbackPtr),
	)

	if result != 0 {
		return fmt.Errorf("RegisterBattleResultCallback 返回错误: %d", result)
	}
	fmt.Println("[Go] 战斗结果回调已注册")
	return nil
}

// CreateBattle 创建战斗
func CreateBattle(battleId, atkTeamId, defTeamId uint32) error {
	libMutex.RLock()
	defer libMutex.RUnlock()

	if libHandle == 0 {
		return fmt.Errorf("C# 库未初始化")
	}

	fnPtr, err := getCachedFunction(libHandle, "CreateBattle")
	if err != nil {
		return fmt.Errorf("找不到函数: CreateBattle - %w", err)
	}

	fn := unsafe.Pointer(fnPtr)
	result, _, _ := purego.SyscallN(
		uintptr(fn),
		uintptr(battleId),
		uintptr(atkTeamId),
		uintptr(defTeamId),
	)

	if result != 0 {
		return fmt.Errorf("CreateBattle 返回错误: %d", result)
	}
	goLog(LogLevelInfo, "[Go] 战斗已创建: ID=%d, ATK=%d, DEF=%d\n", battleId, atkTeamId, defTeamId)
	return nil
}

// DestroyBattle 销毁战斗
func DestroyBattle(battleId uint64) error {
	libMutex.RLock()
	defer libMutex.RUnlock()

	if libHandle == 0 {
		return fmt.Errorf("C# 库未初始化")
	}

	fnPtr, err := getCachedFunction(libHandle, "DestroyBattle")
	if err != nil {
		return fmt.Errorf("找不到函数: DestroyBattle - %w", err)
	}

	fn := unsafe.Pointer(fnPtr)
	result, _, _ := purego.SyscallN(
		uintptr(fn),
		uintptr(battleId),
	)

	if result != 0 {
		return fmt.Errorf("DestroyBattle 返回错误: %d", result)
	}
	goLog(LogLevelInfo, "[Go] 战斗已销毁: ID=%d\n", battleId)
	return nil
}

// OnTick 推动战斗进行一个 Tick，返回处理的战斗数量
func OnTick() (int32, error) {
	libMutex.RLock()
	defer libMutex.RUnlock()

	if libHandle == 0 {
		return -1, fmt.Errorf("C# 库未初始化")
	}

	fnPtr, err := getCachedFunction(libHandle, "OnTick")
	goLog(LogLevelInfo, "[Go] 战斗已创建: OnTick 函数指针=%d\n", fnPtr)
	if err != nil {
		return -1, fmt.Errorf("找不到函数: OnTick - %w", err)
	}

	fn := unsafe.Pointer(fnPtr)
	result, _, _ := purego.SyscallN(uintptr(fn))

	return int32(result), nil
}

// GetBattleCount 获取当前战斗数量
func GetBattleCount() (int32, error) {
	libMutex.RLock()
	defer libMutex.RUnlock()

	if libHandle == 0 {
		return -1, fmt.Errorf("C# 库未初始化")
	}

	fnPtr, err := getCachedFunction(libHandle, "GetBattleCount")
	if err != nil {
		return -1, fmt.Errorf("找不到函数: GetBattleCount - %w", err)
	}

	fn := unsafe.Pointer(fnPtr)
	result, _, _ := purego.SyscallN(uintptr(fn))

	return int32(result), nil
}

// ProcessBattleInput 处理战斗输入
func ProcessBattleInput(battleId uint32, teamId uint32, actionType byte, actionValue int32) error {
	libMutex.RLock()
	defer libMutex.RUnlock()

	if libHandle == 0 {
		return fmt.Errorf("C# 库未初始化")
	}

	fnPtr, err := getCachedFunction(libHandle, "ProcessBattleInput")
	if err != nil {
		return fmt.Errorf("找不到函数: ProcessBattleInput - %w", err)
	}

	fn := unsafe.Pointer(fnPtr)
	result, _, _ := purego.SyscallN(
		uintptr(fn),
		uintptr(battleId),
		uintptr(teamId),
		uintptr(actionType),
		uintptr(actionValue),
	)

	if result != 0 {
		return fmt.Errorf("ProcessBattleInput 返回错误: %d", result)
	}
	return nil
}

func PrcessBattleContextInput(inputBuff unsafe.Pointer, bufflen uint32) error {
	libMutex.RLock()
	defer libMutex.RUnlock()

	if libHandle == 0 {
		return fmt.Errorf("C# 库未初始化")
	}

	fnPtr, err := getCachedFunction(libHandle, "ProcessBattleContextInput")
	if err != nil {
		return fmt.Errorf("找不到函数: ProcessBattleContextInput - %w", err)
	}

	fn := unsafe.Pointer(fnPtr)
	result, _, _ := purego.SyscallN(
		uintptr(fn),
		uintptr(inputBuff),
		uintptr(bufflen),
	)

	if result != 0 {
		return fmt.Errorf("ProcessBattleInput 返回错误: %d", result)
	}
	return nil
}

// ============================================================================
// 日志控制 API
// ============================================================================

// LogLevel 日志级别常量
const (
	LogLevelDebug = 0
	LogLevelInfo  = 1
	LogLevelWarn  = 2
	LogLevelError = 3
	LogLevelNone  = 4
)

// SetBattleLogLevel 设置战斗日志级别 (由 Go 调用)
// level: 0=Debug, 1=Info, 2=Warn, 3=Error, 4=None
func SetBattleLogLevel(level int) error {
	// 同时设置 Go 侧日志级别
	if level >= 0 && level <= 4 {
		goLogMutex.Lock()
		goLogLevel = level
		goLogMutex.Unlock()
	}

	libMutex.RLock()
	defer libMutex.RUnlock()

	if libHandle == 0 {
		return fmt.Errorf("C# 库未初始化")
	}

	fnPtr, err := getCachedFunction(libHandle, "SetBattleLogLevel")
	if err != nil {
		return fmt.Errorf("找不到函数: SetBattleLogLevel - %w", err)
	}

	fn := unsafe.Pointer(fnPtr)
	result, _, _ := purego.SyscallN(
		uintptr(fn),
		uintptr(level),
	)

	if result != 0 {
		return fmt.Errorf("SetBattleLogLevel 返回错误: %d", result)
	}
	return nil
}

// GetBattleLogLevel 获取当前战斗日志级别
func GetBattleLogLevel() (int, error) {
	libMutex.RLock()
	defer libMutex.RUnlock()

	if libHandle == 0 {
		return -1, fmt.Errorf("C# 库未初始化")
	}

	fnPtr, err := getCachedFunction(libHandle, "GetBattleLogLevel")
	if err != nil {
		return -1, fmt.Errorf("找不到函数: GetBattleLogLevel - %w", err)
	}

	fn := unsafe.Pointer(fnPtr)
	result, _, _ := purego.SyscallN(uintptr(fn))

	return int(result), nil
}

// EnableBattleLogging 启用所有战斗日志
func EnableBattleLogging() error {
	// 同时启用 Go 侧日志
	goLogMutex.Lock()
	goLogLevel = LogLevelDebug
	goLogMutex.Unlock()

	libMutex.RLock()
	defer libMutex.RUnlock()

	if libHandle == 0 {
		return fmt.Errorf("C# 库未初始化")
	}

	fnPtr, err := getCachedFunction(libHandle, "EnableBattleLogging")
	if err != nil {
		return fmt.Errorf("找不到函数: EnableBattleLogging - %w", err)
	}

	fn := unsafe.Pointer(fnPtr)
	purego.SyscallN(uintptr(fn))
	return nil
}

// DisableBattleLogging 禁用所有战斗日志
func DisableBattleLogging() error {
	// 同时禁用 Go 侧日志
	goLogMutex.Lock()
	goLogLevel = LogLevelNone
	goLogMutex.Unlock()

	libMutex.RLock()
	defer libMutex.RUnlock()

	if libHandle == 0 {
		return fmt.Errorf("C# 库未初始化")
	}

	fnPtr, err := getCachedFunction(libHandle, "DisableBattleLogging")
	if err != nil {
		return fmt.Errorf("找不到函数: DisableBattleLogging - %w", err)
	}

	fn := unsafe.Pointer(fnPtr)
	purego.SyscallN(uintptr(fn))
	return nil
}

// ============================================================================
// 高级 API (强类型，自动序列化/反序列化)
// ============================================================================

// ExecBattle 执行单场战斗
func ExecBattle(battleReq *proto_pb.StartBattle) (*proto_pb.BattleResult, error) {
	// 序列化请求
	reqData, err := proto.Marshal(battleReq)
	if err != nil {
		return nil, fmt.Errorf("请求序列化失败: %w", err)
	}

	// 调用 C# 函数
	respData, err := ProcessProtoMessage(reqData)
	if err != nil {
		return nil, fmt.Errorf("C# 调用失败: %w", err)
	}

	// 反序列化响应
	resp := &proto_pb.BattleResponse{}
	if err := proto.Unmarshal(respData, resp); err != nil {
		return nil, fmt.Errorf("响应反序列化失败: %w", err)
	}

	// 检查错误码
	if resp.Code != 0 {
		return nil, fmt.Errorf("C# 返回错误 (Code=%d): %s", resp.Code, resp.Message)
	}

	// 解析战斗结果
	result := &proto_pb.BattleResult{}
	if err := proto.Unmarshal(resp.Result, result); err != nil {
		return nil, fmt.Errorf("战斗结果反序列化失败: %w", err)
	}

	return result, nil
}

// ExecBatchBattle 执行批量战斗
func ExecBatchBattle(batchReq *proto_pb.BatchBattleRequest) (*proto_pb.BatchBattleResponse, error) {
	// 序列化请求
	reqData, err := proto.Marshal(batchReq)
	if err != nil {
		return nil, fmt.Errorf("请求序列化失败: %w", err)
	}

	// 调用 C# 函数
	respData, err := ProcessBatchProtoMessage(reqData)
	if err != nil {
		return nil, fmt.Errorf("C# 调用失败: %w", err)
	}

	// 反序列化响应
	resp := &proto_pb.BattleResponse{}
	if err := proto.Unmarshal(respData, resp); err != nil {
		return nil, fmt.Errorf("响应反序列化失败: %w", err)
	}

	// 检查错误码
	if resp.Code != 0 {
		return nil, fmt.Errorf("C# 返回错误 (Code=%d): %s", resp.Code, resp.Message)
	}

	// 解析批量战斗结果
	batchResult := &proto_pb.BatchBattleResponse{}
	if err := proto.Unmarshal(resp.Result, batchResult); err != nil {
		return nil, fmt.Errorf("批量战斗结果反序列化失败: %w", err)
	}

	return batchResult, nil
}

// Export2CSharpBattleEndNotify 导出战斗结束通知由c#调用
func Export2CSharpBattleEndNotify(battleNotification *proto_pb.BattleNotification) error {
	// 序列化请求
	notifyData, err := proto.Marshal(battleNotification)
	if err != nil {
		return fmt.Errorf("战斗结束通知序列化失败: %w", err)
	}

	// TODO: 调用 C# 侧的处理函数或回调
	_ = notifyData // 移除未使用警告

	return nil
}

// ============================================================================
// 调用 C# 侧全局导出函数的包装器
// ============================================================================

// CallCSharpGlobalFunction 调用 C# 侧导出的全局函数
// 这个函数演示如何从 Go 侧调用 C# 的导出函数
func CallCSharpGlobalFunction(functionName string, battleID uint32, notificationType int, timestamp int64) (int32, error) {
	libMutex.RLock()
	defer libMutex.RUnlock()

	if libHandle == 0 {
		return -1, fmt.Errorf("C# 库未初始化")
	}

	switch functionName {
	case "CallGoGlobalHandleBattleNotification":
		fnPtr, err := getCachedFunction(libHandle, "CallGoGlobalHandleBattleNotification")
		if err != nil {
			return -1, fmt.Errorf("找不到函数: %s - %w", functionName, err)
		}

		fn := unsafe.Pointer(fnPtr)
		result, _, _ := purego.SyscallN(
			uintptr(fn),
			uintptr(battleID),
			uintptr(notificationType),
			uintptr(timestamp),
		)

		fmt.Printf("[Go] 调用 C# 全局函数 %s 完成，结果=%d\n", functionName, result)
		return int32(result), nil

	case "CallGoCalculateSum":
		fnPtr, err := getCachedFunction(libHandle, "CallGoCalculateSum")
		if err != nil {
			return -1, fmt.Errorf("找不到函数: %s - %w", functionName, err)
		}

		// 这个函数接收两个 int32，返回结果
		fn := unsafe.Pointer(fnPtr)
		result, _, _ := purego.SyscallN(
			uintptr(fn),
			uintptr(int32(battleID)),         // 作为第一个整数
			uintptr(int32(notificationType)), // 作为第二个整数
		)

		fmt.Printf("[Go] 调用 C# 计算函数完成，结果=%d\n", result)
		return int32(result), nil

	default:
		return -1, fmt.Errorf("未知的 C# 全局函数: %s", functionName)
	}
}

// CallCSharpSimpleGlobalFunction 调用 C# 侧的简单全局函数
func CallCSharpSimpleGlobalFunction(battleID uint32, action string) (int32, error) {
	libMutex.RLock()
	defer libMutex.RUnlock()

	if libHandle == 0 {
		return -1, fmt.Errorf("C# 库未初始化")
	}

	fnPtr, err := purego.Dlsym(libHandle, "CallGoSimpleGlobalFunction")
	if err != nil {
		return -1, fmt.Errorf("找不到函数: CallGoSimpleGlobalFunction - %w", err)
	}

	// 将字符串转为字节数组
	actionBytes := []byte(action)

	fn := unsafe.Pointer(fnPtr)
	result, _, _ := purego.SyscallN(
		uintptr(fn),
		uintptr(battleID),
		uintptr(unsafe.Pointer(&actionBytes[0])),
		uintptr(len(actionBytes)),
	)

	fmt.Printf("[Go] 调用 C# 简单全局函数完成，结果=%d\n", result)
	return int32(result), nil
}

// CallCSharpGetConfigLoaderData 通过 C# 获取配置数据（间接双向调用）
// 这个函数演示完整的双向调用链：
// 1. Go 侧调用 C# 导出的 GetConfigLoaderDataCSharp 函数
// 2. C# 侧调用已注册的配置加载器回调（由 Go 侧实现）
// 3. Go 侧回调加载文件并返回数据给 C#
// 4. C# 将数据返回给 Go
func CallCSharpGetConfigLoaderData(configName string) ([]byte, error) {
	libMutex.RLock()
	defer libMutex.RUnlock()

	if libHandle == 0 {
		return nil, fmt.Errorf("C# 库未初始化")
	}

	fnPtr, err := getCachedFunction(libHandle, "GetConfigLoaderDataCSharp")
	if err != nil {
		return nil, fmt.Errorf("找不到函数: GetConfigLoaderDataCSharp - %w", err)
	}

	// 转换配置名称为字节数组
	nameBytes := []byte(configName)

	// 准备输出参数的指针
	var outDataPtr uintptr
	var outDataLen int32

	fn := unsafe.Pointer(fnPtr)
	result, _, _ := purego.SyscallN(
		uintptr(fn),
		uintptr(unsafe.Pointer(&nameBytes[0])),
		uintptr(len(nameBytes)),
		uintptr(unsafe.Pointer(&outDataPtr)),
		uintptr(unsafe.Pointer(&outDataLen)),
	)

	if result != 0 {
		return nil, fmt.Errorf("GetConfigLoaderDataCSharp 返回错误代码: %d", result)
	}

	// 从指针复制数据到 Go 切片
	if outDataPtr == 0 || outDataLen <= 0 {
		return nil, nil
	}

	data := make([]byte, outDataLen)
	copy(data, (*[1 << 30]byte)(unsafe.Pointer(outDataPtr))[:outDataLen:outDataLen])

	fmt.Printf("[Go] 通过 C# 获取配置数据: %s (%d 字节)\n", configName, outDataLen)

	return data, nil
}
