package main

import (
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/luhaoting/goPureWithCsharp/pkg/csharp"
	"github.com/luhaoting/goPureWithCsharp/pkg/proto"
)

// TestMain 初始化测试环境
func TestMain(m *testing.M) {
	fmt.Println("========== Go ↔ C# 双向调用集成测试 ==========")
	fmt.Println()

	// ========== 准备阶段：编译 C# 库 ==========
	if err := SetupCSharpLibrary(); err != nil {
		log.Fatalf("C# 库准备失败: %v", err)
	}

	// ========== 验证阶段：检查库文件 ==========
	if err := VerifyCSharpLibrary(); err != nil {
		log.Fatalf("C# 库验证失败: %v", err)
	}

	// ========== 初始化阶段 ==========
	fmt.Println("[初始化] 加载 C# 库")
	err := csharp.InitCSharpLib("Release")
	if err != nil {
		log.Fatalf("库初始化失败: %v", err)
	}
	
	// defer 必须在函数返回前执行，确保库在所有测试之后关闭
	defer func() {
		csharp.CloseCSharpLib()
	}()
	
	fmt.Println("✓ C# 库已初始化")
	fmt.Println()

	// 运行所有测试
	code := m.Run()

	fmt.Println()
	fmt.Println("========== 所有测试完成 ==========")

	os.Exit(code)
}

// TestSingleBattle 测试单场战斗
func TestSingleBattle(t *testing.T) {
	t.Log("[TEST] 测试单场战斗")

	battleReq := &proto.StartBattle{
		Atk: &proto.Team{
			TeamId:   1001,
			TeamName: "Red Team",
			Lineup:   []uint32{101, 102, 103},
		},
		Def: &proto.Team{
			TeamId:   1002,
			TeamName: "Blue Team",
			Lineup:   []uint32{201, 202, 203},
		},
		BattleId:  50001,
		Timestamp: 1702175000000,
	}

	t.Logf("ATK: Team %d (%s)\n", battleReq.Atk.TeamId, battleReq.Atk.TeamName)
	t.Logf("DEF: Team %d (%s)\n", battleReq.Def.TeamId, battleReq.Def.TeamName)

	result, err := csharp.ExecBattle(battleReq)
	if err != nil {
		t.Fatalf("战斗执行失败: %v", err)
	}

	if result.Winner == 0 {
		t.Error("战斗结果异常: 胜方为 0")
	}

	t.Logf("✓ 战斗执行成功 - 胜方: Team %d, 积分: %d", result.Winner, result.BattleScore)
}

// TestBatchBattle 测试批量战斗
func TestBatchBattle(t *testing.T) {
	t.Log("[TEST] 测试批量战斗")

	battles := []*proto.StartBattle{}

	for i := 0; i < 2; i++ {
		bid := uint32(50010 + i)
		tid1 := uint32(1001 + i)
		tid2 := uint32(1002 + i)

		battles = append(battles, &proto.StartBattle{
			Atk: &proto.Team{
				TeamId:   tid1,
				TeamName: fmt.Sprintf("Team_%d", tid1),
				Lineup:   []uint32{1, 2, 3},
			},
			Def: &proto.Team{
				TeamId:   tid2,
				TeamName: fmt.Sprintf("Team_%d", tid2),
				Lineup:   []uint32{4, 5, 6},
			},
			BattleId:  bid,
			Timestamp: 1702175000000 + int64(i)*1000,
		})
	}

	batchReq := &proto.BatchBattleRequest{
		BatchId:  "batch_001",
		Battles:  battles,
		Parallel: 1,
	}

	t.Logf("批次ID: %s, 战斗数: %d", batchReq.BatchId, len(batchReq.Battles))

	batchResult, err := csharp.ExecBatchBattle(batchReq)
	if err != nil {
		t.Fatalf("批量战斗执行失败: %v", err)
	}

	if batchResult.SuccessCount != int32(len(battles)) {
		t.Errorf("期望成功数 %d, 实际 %d", len(battles), batchResult.SuccessCount)
	}

	t.Logf("✓ 批量战斗执行成功 - 成功数: %d, 失败数: %d", batchResult.SuccessCount, batchResult.FailureCount)
}

// TestCallbackRegistration 测试回调注册
func TestCallbackRegistration(t *testing.T) {
	t.Log("[TEST] 测试回调注册")

	err := csharp.RegisterNotificationCallback(func(notif *proto.BattleNotification) error {
		t.Logf("[回调] 收到通知: Type=%d, BattleID=%d", notif.NotificationType, notif.BattleId)
		return nil
	})
	if err != nil {
		t.Errorf("回调注册失败: %v", err)
	}

	t.Log("✓ 回调已注册")
}

// TestErrorHandling 测试错误处理
func TestErrorHandling(t *testing.T) {
	t.Log("[TEST] 测试错误处理")

	validReq := &proto.StartBattle{
		Atk: &proto.Team{
			TeamId:   2001,
			TeamName: "Recovery Test Team 1",
			Lineup:   []uint32{1},
		},
		Def: &proto.Team{
			TeamId:   2002,
			TeamName: "Recovery Test Team 2",
			Lineup:   []uint32{2},
		},
		BattleId:  60000,
		Timestamp: 1702175000000,
	}

	result, err := csharp.ExecBattle(validReq)
	if err != nil {
		t.Fatalf("错误恢复测试失败: %v", err)
	}

	t.Logf("✓ 库已恢复，战斗结果: 胜方=%d, 积分=%d", result.Winner, result.BattleScore)
}
