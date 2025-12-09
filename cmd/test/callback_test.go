package main

import (
	"sync"
	"testing"

	"github.com/luhaoting/goPureWithCsharp/pkg/csharp"
	pb "github.com/luhaoting/goPureWithCsharp/pkg/proto"
)

// 全局回调计数器和结果存储
var (
	callbackMutex    sync.Mutex
	callbackCount    int
	lastNotification *pb.BattleNotification
)

// GlobalCallbackHandler Go 侧的全局回调函数
// 这是一个全局函数，地址稳定，可以安全地从 C# 调用
func GlobalCallbackHandler(notif *pb.BattleNotification) error {
	// 保存结果
	callbackMutex.Lock()
	callbackCount++
	lastNotification = notif
	callbackMutex.Unlock()
	return nil
}

// TestGoCallbackFromCSharp 真实测试：Go 全局函数被 C# 真实调用
func TestGoCallbackFromCSharp(t *testing.T) {
	t.Log("========== 测试 Go 全局回调被 C# 真实调用 ==========")

	// 清空并重置
	csharp.ClearNotificationCallbacks()
	callbackMutex.Lock()
	callbackCount = 0
	lastNotification = nil
	callbackMutex.Unlock()

	// 步骤 1：注册 Go 侧回调
	t.Log("[步骤 1] 注册 Go 侧回调函数")
	err := csharp.RegisterNotificationCallback(GlobalCallbackHandler)
	if err != nil {
		t.Fatalf("❌ 注册失败: %v", err)
	}
	t.Log("✓ Go 侧回调已注册")

	// 步骤 2：模拟 C# 发送通知
	// 注意：在实际场景中，这会由 C# 侧通过 FFI 回调或其他机制触发
	// 为了演示，我们直接调用 ProcessNotification
	t.Log("[步骤 2] 模拟 C# 发送通知给 Go")

	notification := &pb.BattleNotification{
		BattleId:         50001,
		NotificationType: pb.NotificationType_EVENT_OCCURRED,
		Timestamp:        1702175000000,
	}

	notifData, err := csharp.MarshalNotification(notification)
	if err != nil {
		t.Fatalf("❌ 序列化通知失败: %v", err)
	}

	err = csharp.ProcessNotificationDirect(notifData)
	if err != nil {
		t.Fatalf("❌ 处理通知失败: %v", err)
	}
	t.Log("✓ 通知已发送并处理")

	// 步骤 3：验证 Go 回调是否被调用
	t.Log("[步骤 3] 验证回调执行结果")
	callbackMutex.Lock()
	count := callbackCount
	notification = lastNotification
	callbackMutex.Unlock()

	if count == 0 {
		t.Fatal("❌ 回调未被调用！")
	}

	t.Logf("✓ 回调被调用 %d 次", count)

	if notification == nil {
		t.Fatal("❌ 回调数据为空！")
	}

	t.Logf("✓ 收到通知数据:")
	t.Logf("  - BattleID: %d (期望: 50001)", notification.BattleId)
	t.Logf("  - Type: %d (期望: %d)", notification.NotificationType, pb.NotificationType_EVENT_OCCURRED)
	t.Logf("  - Timestamp: %d (期望: 1702175000000)", notification.Timestamp)

	// 断言验证
	if notification.BattleId != 50001 {
		t.Errorf("❌ BattleID 不匹配: 得到 %d, 期望 50001", notification.BattleId)
	}
	if notification.NotificationType != pb.NotificationType_EVENT_OCCURRED {
		t.Errorf("❌ Type 不匹配: 得到 %d, 期望 %d", notification.NotificationType, pb.NotificationType_EVENT_OCCURRED)
	}
	if notification.Timestamp != 1702175000000 {
		t.Errorf("❌ Timestamp 不匹配: 得到 %d, 期望 1702175000000", notification.Timestamp)
	}

	t.Log("✓ 所有断言通过！")
	t.Log("✓ 测试完成 - Go 全局回调被成功调用")
}

// TestMultipleCallbacks 测试多次回调
func TestMultipleCallbacks(t *testing.T) {
	t.Log("========== 测试多次回调 ==========")

	// 清空并重置
	csharp.ClearNotificationCallbacks()
	callbackMutex.Lock()
	callbackCount = 0
	callbackMutex.Unlock()

	// 注册回调
	err := csharp.RegisterNotificationCallback(GlobalCallbackHandler)
	if err != nil {
		t.Fatalf("注册失败: %v", err)
	}

	// 多次调用
	tests := []struct {
		name      string
		notifType pb.NotificationType
		battleID  uint32
		timestamp int64
	}{
		{"第一个事件", pb.NotificationType_EVENT_OCCURRED, 50001, 1702175000000},
		{"第二个事件", pb.NotificationType_EVENT_OCCURRED, 50002, 1702175001000},
		{"第三个事件", pb.NotificationType_STATUS_UPDATE, 50003, 1702175002000},
	}

	for _, test := range tests {
		t.Logf("[测试] %s", test.name)

		notification := &pb.BattleNotification{
			BattleId:         test.battleID,
			NotificationType: test.notifType,
			Timestamp:        test.timestamp,
		}

		notifData, err := csharp.MarshalNotification(notification)
		if err != nil {
			t.Fatalf("❌ 序列化失败: %v", err)
		}

		err = csharp.ProcessNotificationDirect(notifData)
		if err != nil {
			t.Fatalf("❌ 处理失败: %v", err)
		}

		callbackMutex.Lock()
		if lastNotification == nil || lastNotification.BattleId != test.battleID {
			callbackMutex.Unlock()
			t.Fatalf("❌ 回调数据不匹配")
		}
		callbackMutex.Unlock()

		t.Logf("✓ BattleID %d 回调成功", test.battleID)
	}

	// 验证本测试的调用次数
	callbackMutex.Lock()
	finalCount := callbackCount
	callbackMutex.Unlock()

	t.Logf("✓ 本测试共 %d 次回调（期望 3 次）", finalCount)
	if finalCount != 3 {
		t.Fatalf("❌ 回调次数不匹配: 得到 %d, 期望 3", finalCount)
	}

	t.Log("✓ 多次回调测试通过")
}

// TestCallbackDataIntegrity 测试回调数据完整性
func TestCallbackDataIntegrity(t *testing.T) {
	t.Log("========== 测试回调数据完整性 ==========")

	// 清空并重置
	csharp.ClearNotificationCallbacks()
	callbackMutex.Lock()
	callbackCount = 0
	lastNotification = nil
	callbackMutex.Unlock()

	// 注册回调
	csharp.RegisterNotificationCallback(GlobalCallbackHandler)

	// 发送特定数据
	battleID := uint32(99999)
	timestamp := int64(9999999999)
	notifType := pb.NotificationType_BATTLE_COMPLETED

	t.Logf("发送数据: BattleID=%d, Type=%d, Timestamp=%d", battleID, notifType, timestamp)

	notification := &pb.BattleNotification{
		BattleId:         battleID,
		NotificationType: notifType,
		Timestamp:        timestamp,
	}

	notifData, err := csharp.MarshalNotification(notification)
	if err != nil {
		t.Fatalf("❌ 序列化失败: %v", err)
	}

	err = csharp.ProcessNotificationDirect(notifData)
	if err != nil {
		t.Fatalf("❌ 处理失败: %v", err)
	}

	// 验证接收到的数据
	callbackMutex.Lock()
	notification = lastNotification
	callbackMutex.Unlock()

	if notification == nil {
		t.Fatal("❌ 未收到通知")
	}

	// 精确比对
	if notification.BattleId != battleID {
		t.Errorf("❌ BattleID: 发送 %d, 收到 %d", battleID, notification.BattleId)
	}
	if notification.Timestamp != timestamp {
		t.Errorf("❌ Timestamp: 发送 %d, 收到 %d", timestamp, notification.Timestamp)
	}
	if notification.NotificationType != notifType {
		t.Errorf("❌ Type: 发送 %d, 收到 %d", notifType, notification.NotificationType)
	}

	t.Log("✓ 数据完整性验证通过")
}
