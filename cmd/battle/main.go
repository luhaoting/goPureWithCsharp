package main

import (
	"fmt"
	pb "goPureWithCsharp/csharp/proto"
)

func main() {
	fmt.Println("========== BattleManager 单例系统演示 ==========")
	fmt.Println()

	outChan := make(chan *pb.BattleContext, OUTPUT_BUFFER_SIZE)
	// 使用 Builder 模式创建并初始化为单例
	bm := NewBattleManagerBuilder().
		WithBattleOutputChan(outChan).
		WithFPS(30).
		BuildAsSingleton()

	fmt.Printf("[Main] BattleManager 单例已初始化: %p\n", bm)
	fmt.Println()

	fmt.Println("[BattleManager-Goroutine] 启动 BattleManager...")
	if err := bm.Start(); err != nil {
		fmt.Printf("[BattleManager-Goroutine] ✗ 启动失败: %v\n", err)
		return
	}

	fmt.Println()
	fmt.Println("========== 战斗系统运行演示 ==========")
	fmt.Println()

	// ========== 创建战斗 ==========
	battleID := GenerateBattleID()
	atkTeamID := uint32(100)
	defTeamID := uint32(101)

	fmt.Printf("[Main] 创建战斗 - ID: %d, 攻击方: %d, 防守方: %d\n", battleID, atkTeamID, defTeamID)

	// 获取创建通道
	createCh := bm.GetCreateChannel()
	// 构建 BattleEnv
	env := &pb.BattleEnv{
		BattleId: uint32(battleID),
		Atk:      &pb.Team{TeamId: atkTeamID},
		Def:      &pb.Team{TeamId: defTeamID},
	}

	endCh := make(chan struct{})
	go func() {
		for outChanput := range outChan {
			fmt.Printf("[BattleOutput] 战斗输出: %+v\n", outChanput)
			endCh <- struct{}{}
		}
	}()

	createCh <- env
	<-endCh
}
