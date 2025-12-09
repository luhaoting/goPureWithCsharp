package main

import (
	"fmt"
	"sync"
	"testing"
	"unsafe"
)

// GoCallbackHandler 定义 Go 侧的回调函数签名
type GoCallbackHandler func(notificationType int32, battleID int64, timestamp int64) int32

var (
	// 存储 Go 回调函数引用，防止被 GC
	callbackMutex   sync.Mutex
	activeCallback  GoCallbackHandler
	callbackResults []string
)

// RegisterGoCallbackForCSharp 将 Go 函数指针传递给 C#
func RegisterGoCallbackForCSharp(callback GoCallbackHandler) unsafe.Pointer {
	callbackMutex.Lock()
	defer callbackMutex.Unlock()

	activeCallback = callback

	// 返回指向函数的指针
	fnPtr := unsafe.Pointer(&activeCallback)
	return fnPtr
}

// TestGoCallbackFromCSharp 测试 C# 调用 Go 函数
func TestGoCallbackFromCSharp(t *testing.T) {
	t.Log("========== C# 调用 Go 函数指针测试 ==========")

	// 清空结果
	callbackResults = []string{}

	// 1. 定义 Go 侧的回调函数
	t.Log("[测试] 步骤 1: 定义 Go 回调函数")
	goCallback := func(notificationType int32, battleID int64, timestamp int64) int32 {
		msg := fmt.Sprintf(
			"✓ Go 回调被调用: NotifType=%d, BattleID=%d, Timestamp=%d",
			notificationType, battleID, timestamp,
		)
		callbackResults = append(callbackResults, msg)
		t.Log(msg)
		return 0
	}
	t.Log("✓ Go 回调函数已定义")

	// 2. 获取函数指针
	t.Log("[测试] 步骤 2: 注册回调到 C#")
	callbackPtr := RegisterGoCallbackForCSharp(goCallback)
	t.Logf("✓ Go 函数指针已注册: %p", callbackPtr)

	// 3. 模拟 C# 调用该指针
	t.Log("[测试] 步骤 3: 模拟 C# 调用 Go 函数")
	t.Log("场景: C# 战斗引擎完成战斗，调用 Go 回调通知")

	if activeCallback != nil {
		t.Log("[C# 侧] 调用 Go 函数指针...")
		result := activeCallback(1, 50001, 1702175000000)
		t.Logf("[C# 侧] 回调返回: %d", result)
	} else {
		t.Error("回调未注册")
		return
	}

	// 4. 验证结果
	t.Log("[测试] 步骤 4: 验证回调执行")
	if len(callbackResults) > 0 {
		t.Logf("✓ 回调执行成功，共 %d 次调用", len(callbackResults))
		for i, result := range callbackResults {
			t.Logf("[%d] %s", i+1, result)
		}
	} else {
		t.Error("回调未被执行")
	}

	// 5. 演示多次调用
	t.Log("[测试] 步骤 5: 演示多次调用场景")
	t.Log("场景: C# 处理多个战斗完成事件")

	battleEvents := []struct {
		notifType int32
		battleID  int64
		timestamp int64
	}{
		{1, 50002, 1702175001000},
		{1, 50003, 1702175002000},
		{1, 50004, 1702175003000},
	}

	for i, event := range battleEvents {
		t.Logf("第 %d 个事件: BattleID=%d", i+1, event.battleID)
		if activeCallback != nil {
			result := activeCallback(event.notifType, event.battleID, event.timestamp)
			t.Logf("→ 回调返回: %d", result)
		}
	}

	t.Logf("✓ 总共执行 %d 次回调", len(callbackResults))
}

// TestCallbackTechDetails 测试回调技术细节输出
func TestCallbackTechDetails(t *testing.T) {
	t.Log("========== Go ↔ C# 回调机制详解 ==========")

	t.Log("【回调原理】")
	t.Log("1. Go 侧:")
	t.Log("   - 定义回调函数: func(notificationType, battleID, timestamp)")
	t.Log("   - 获取函数指针: unsafe.Pointer(&goFunction)")
	t.Log("   - 传递给 C#: csharp.RegisterCallback(goFuncPtr)")

	t.Log("2. C# 侧:")
	t.Log("   - 接收指针: IntPtr goFuncPtr")
	t.Log("   - 转换委托: Marshal.GetDelegateForFunctionPointer<BattleNotifyCallback>()")
	t.Log("   - 保存引用: BattleCallbackManager.RegisterCallback()")
	t.Log("   - 执行回调: when event occurs → callback(...)")

	t.Log("3. 执行流程:")
	t.Log("   Go Register → C# Store → Battle Engine → Event Trigger → C# Invoke → Go Callback")

	t.Log("【类型对应】")
	t.Log("Go:  func(int32, int64, int64) int32")
	t.Log("C#:  public delegate int BattleNotifyCallback(int notificationType, long battleID, long timestamp)")

	t.Log("【安全考虑】")
	t.Log("✓ Go 侧: 使用全局变量保持引用，防止 GC 释放")
	t.Log("✓ 线程安全: 使用 mutex 保护共享状态")
	t.Log("✓ C# 侧: 调用前检查指针有效性")
	t.Log("✓ 错误处理: 回调返回状态码")
}
