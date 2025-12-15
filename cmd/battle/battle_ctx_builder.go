package main

import (
	"fmt"

	pb "goPureWithCsharp/csharp/proto"

	"google.golang.org/protobuf/proto"
)

// 这里是一个 proto协议 build构建器，用于构建 BattleContext 协议对象  可以穿越 ctx input的oneof 结构，自动生成ctx结构，可以直接返回[]byte 数据,但是我智能在外部提供的 byte buffer 写数据， 返回一个 byte头指针和数据长度。
// 这个byte buff 由外部提供，外部还提供 battleid 注入，和 frameseq 的获得，帮我写一个这样的宿主interface
// 这个 结构需要  injectInput 方法， 以及 praseOutput  这样的方法，input返回数据是buff， prase output 是  byte头指针 和 数据长度指针

// BattleContextHost BattleContext 宿主接口
// 提供 BattleContextBuilder 所需的外部依赖：BattleID、FrameSeq、字节缓冲
type BattleContextHost interface {
	GetCurrentFrame() uint64

	// GetInputBuffer 获取用于接收输入数据的字节缓冲
	// 返回字节切片，外部使用此缓冲存储 BattleInput 数据
	GetInputBuffer() (buff []byte, maxLen int)
}

// BattleMsgContextBuilder BattleContext 构建器
// 根据不同的操作类型自动构建 BattleContext，并序列化为 protobuf 数据
type BattleMsgContextBuilder struct {
	host BattleContextHost
}

// NewBattleContextBuilder 创建新的 BattleContext 构建器
func NewBattleContextBuilder(host BattleContextHost) *BattleMsgContextBuilder {
	if host == nil {
		panic("BattleContextHost cannot be nil")
	}

	return &BattleMsgContextBuilder{
		host: host,
	}
}

// InjectInput 注入战斗输入
// inputType: 输入类型（对应 BattleInputOperation）
// inputData: 输入数据（会写入外部提供的缓冲）
// 返回: 写入的字节数，错误信息
func (bcb *BattleMsgContextBuilder) InjectInput(battleID uint32, inputData proto.Message) (int, error) {
	if inputData == nil {
		return 0, fmt.Errorf("inputData cannot be nil")
	}

	// 根据不同的操作类型设置 oneof 字段
	battleInput := &pb.BattleInput{}
	switch input := inputData.(type) {
	case *pb.BattleUseItem:
		battleInput.Input = &pb.BattleInput_Use{Use: input}

	case *pb.BattlePause:
		battleInput.Input = &pb.BattleInput_Pause{Pause: input}

	case *pb.BattleResume:
		battleInput.Input = &pb.BattleInput_Resume{Resume: input}

	case *pb.BattleUserOp:
		battleInput.Input = &pb.BattleInput_UserOp{UserOp: input}

	default:
		return 0, fmt.Errorf("unsupported input operation: %v", inputData)
	}

	// 根据输入类型构建对应的 BattleInput 对象
	battleInputCtx := &pb.BattleContext{
		BattleId: battleID,
		Tick:     bcb.host.GetCurrentFrame(),
		Option:   &pb.BattleContext_BattleInput{BattleInput: battleInput},
	}

	// 获取外部提供的输入缓冲
	inputBuf, maxLen := bcb.host.GetInputBuffer()
	if maxLen == 0 {
		return 0, fmt.Errorf("input buffer size is 0")
	}
	// 使用 MarshalAppend 直接在已存在的 buffer 上追加序列化数据（无需额外内存分配）
	// inputBuf[:0] 表示从 buffer 的起始位置开始写入
	opts := proto.MarshalOptions{}
	result, err := opts.MarshalAppend(inputBuf[:0], battleInputCtx)
	if err != nil {
		return 0, fmt.Errorf("failed to marshal BattleInput: %w", err)
	}

	// 检查序列化后的数据是否超过缓冲大小
	if len(result) > maxLen {
		return 0, fmt.Errorf("marshaled data size (%d) exceeds buffer size (%d)", len(result), maxLen)
	}

	return len(result), nil
}

// ParseOutput 解析战斗输出
// outputData: 外部提供的输出数据（从 C# 返回）
// 返回: 指向缓冲的指针、数据长度、错误信息
func (bcb *BattleMsgContextBuilder) ParseOutput(outputData []byte) (*pb.BattleContext, error) {
	if len(outputData) == 0 {
		return nil, fmt.Errorf("output data is empty")
	}

	outPutCtx := &pb.BattleContext{}
	err := proto.Unmarshal(outputData, outPutCtx)
	if err != nil {
		return nil, err
	}

	return outPutCtx, nil
}
