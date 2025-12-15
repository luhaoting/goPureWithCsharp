package main

import (
	pb "goPureWithCsharp/csharp/proto"
)

type BattleManagerBuilder struct {
	eventBus   EventBus
	dispatcher BattleDisptcher
	fps        int64
	bufferSize int
	outPutChan chan *pb.BattleContext
}

func NewBattleManagerBuilder() *BattleManagerBuilder {
	return &BattleManagerBuilder{
		fps:        30,
		bufferSize: 100,
	}
}

func (b *BattleManagerBuilder) WithEventBus(eb EventBus) *BattleManagerBuilder {
	b.eventBus = eb
	return b
}

func (b *BattleManagerBuilder) WithDispatcher(d BattleDisptcher) *BattleManagerBuilder {
	b.dispatcher = d
	return b
}

func (b *BattleManagerBuilder) WithBattleOutputChan(ch chan *pb.BattleContext) *BattleManagerBuilder {
	b.outPutChan = ch
	return b
}

func (b *BattleManagerBuilder) WithFPS(fps int64) *BattleManagerBuilder {
	b.fps = fps
	return b
}

func (b *BattleManagerBuilder) WithBufferSize(size int) *BattleManagerBuilder {
	b.bufferSize = size
	return b
}

func (b *BattleManagerBuilder) Build() *BattleManager {

	if b.eventBus == nil {
		b.eventBus = NewEventBus(b.bufferSize)
	}

	// 创建 FrameSeqGenerator
	fpsProvider := NewFrameSeqGenerator(b.fps)

	if b.dispatcher == nil {
		b.dispatcher = NewProxy(fpsProvider)
	}

	// 创建命令通道
	createChan := make(chan *pb.BattleEnv, b.bufferSize)

	return &BattleManager{
		EventBus:    b.eventBus,
		fpsProvider: fpsProvider,
		createChan:  createChan,
		outPutChan:  b.outPutChan,
		battleCtrls: b.dispatcher,
		state:       StateCreated,
		stopChan:    make(chan struct{}),
	}
} // BuildAsSingleton 构建并初始化为全局单例
// 如果单例已存在，直接返回现有实例，不会再次构建
func (b *BattleManagerBuilder) BuildAsSingleton() *BattleManager {
	return InitBattleManager(b)
}
