package main

import (
	"fmt"
	"goPureWithCsharp/csharp"
	pb "goPureWithCsharp/csharp/proto"
	"unsafe"

	"google.golang.org/protobuf/proto"
)

var configDir = "../../config"

var INPUT_BUFFER_SIZE = 512
var OUTPUT_BUFFER_SIZE = 1024

type BattleInfo struct {
	ID  int64
	Env *pb.BattleEnv
}

func CreateBattleInfo(env *pb.BattleEnv) *BattleInfo {
	return &BattleInfo{
		Env: env,
	}
}

type FrameSeqProvider interface {
	GetCurrentFrame() uint64
}

type BattleOutput interface {
	OutPutResult(*pb.BattleResult) error
	OutPutReply(*pb.BattleReplay) error
}

type BattleController struct {
	FrameSeqProvider
	BattleOutput

	BattleMsgContextBuilder

	battleMap        map[uint64]*BattleInfo
	inputBuffHander  []byte
	outputBuffHander []byte
}

func (bc *BattleController) GetInputBuffer() (buff []byte, maxLen int) {
	return bc.inputBuffHander, len(bc.inputBuffHander)
}

func NewBattleController(p FrameSeqProvider, o BattleOutput) *BattleController {
	ctrl := &BattleController{
		FrameSeqProvider: p,
		BattleOutput:     o,
		battleMap:        make(map[uint64]*BattleInfo),
		inputBuffHander:  make([]byte, INPUT_BUFFER_SIZE),
		outputBuffHander: make([]byte, OUTPUT_BUFFER_SIZE),
	}

	ctrl.BattleMsgContextBuilder = *NewBattleContextBuilder(ctrl)

	return ctrl
}

func (bc *BattleController) CreateBattle(battleId uint64, env *pb.BattleEnv) error {
	battleInfo := CreateBattleInfo(env)
	bc.battleMap[battleId] = battleInfo
	return nil
}

func (bc *BattleController) BattleInput(battleId uint64, input proto.Message) error {
	if input == nil {
		return fmt.Errorf("input cannot be nil")
	}

	inputBuffLen, err := bc.InjectInput(uint32(battleId), input)
	if err != nil {
		fmt.Printf("[Battle] 构建输入消息失败: %v\n", err)
		return err
	}

	err = csharp.PrcessBattleContextInput(unsafe.Pointer(&bc.inputBuffHander[0]), uint32(inputBuffLen))
	if err != nil {
		return err
	}
	return nil
}

func (bc *BattleController) OnTick(logicFrameSeq uint64) {
	// TODO : 调用 C# 的 OnTick 函数 传入逻辑帧数
	csharp.OnTick()
}

func (bc *BattleController) DestroyBattle(battleId uint64) {
	csharp.DestroyBattle(battleId)
	bc.battleMap[battleId] = nil
}
