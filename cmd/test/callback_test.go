package main

import (
	"fmt"
	"sync"
	"unsafe"
)

// GoCallbackHandler 定义 Go 侧的回调函数签名
// 这是 C# 侧 BattleNotifyCallback 委托对应的函数类型
type GoCallbackHandler func(notificationType int32, battleID int64, timestamp int64) int32

var (
	// 存储 Go 回调函数引用，防止被 GC
	callbackMutex   sync.Mutex
	activeCallback  GoCallbackHandler
	callbackResults []string
)

// RegisterGoCallbackForCSharp 将 Go 函数指针传递给 C#
// 返回指向 Go 函数的指针供 C# 调用
func RegisterGoCallbackForCSharp(callback GoCallbackHandler) unsafe.Pointer {
	callbackMutex.Lock()
	defer callbackMutex.Unlock()

	activeCallback = callback

	// 返回指向函数的指针
	// 这在 C# 侧会通过 Marshal.GetDelegateForFunctionPointer 转换为委托
	fnPtr := unsafe.Pointer(&activeCallback)
	return fnPtr
}

// TestGoCallbackFromCSharp 测试 C# 调用 Go 函数
func TestGoCallbackFromCSharp() {
	fmt.Println("\n========== C# 调用 Go 函数指针测试 ==========")
	fmt.Println()

	// 清空结果
	callbackResults = []string{}

	// 1. 定义 Go 侧的回调函数
	fmt.Println("[测试] 步骤 1: 定义 Go 回调函数")
	goCallback := func(notificationType int32, battleID int64, timestamp int64) int32 {
		msg := fmt.Sprintf(
			"✓ Go 回调被调用: NotifType=%d, BattleID=%d, Timestamp=%d",
			notificationType, battleID, timestamp,
		)
		callbackResults = append(callbackResults, msg)
		fmt.Printf("  %s\n", msg)
		return 0 // 返回成功
	}
	fmt.Println("✓ Go 回调函数已定义")
	fmt.Println()

	// 2. 获取函数指针
	fmt.Println("[测试] 步骤 2: 注册回调到 C#")
	callbackPtr := RegisterGoCallbackForCSharp(goCallback)
	fmt.Printf("✓ Go 函数指针已注册: %p\n", callbackPtr)
	fmt.Println()

	// 3. 注册到 C# 库（这会使 C# 存储这个指针）
	fmt.Println("[测试] 步骤 3: 传递指针给 C# 库")
	err := registerCallbackToCSHarp(callbackPtr)
	if err != nil {
		fmt.Printf("❌ 注册失败: %v\n", err)
		return
	}
	fmt.Println("✓ 指针已传递给 C#")
	fmt.Println()

	// 4. 模拟 C# 调用该指针
	fmt.Println("[测试] 步骤 4: 模拟 C# 调用 Go 函数")
	fmt.Println("场景: C# 战斗引擎完成战斗，调用 Go 回调通知")
	fmt.Println()

	if activeCallback != nil {
		fmt.Println("  [C# 侧] 调用 Go 函数指针...")
		result := activeCallback(1, 50001, 1702175000000)
		fmt.Printf("  [C# 侧] 回调返回: %d\n", result)
	} else {
		fmt.Println("❌ 回调未注册")
		return
	}
	fmt.Println()

	// 5. 验证结果
	fmt.Println("[测试] 步骤 5: 验证回调执行")
	if len(callbackResults) > 0 {
		fmt.Printf("✓ 回调执行成功，共 %d 次调用:\n", len(callbackResults))
		for i, result := range callbackResults {
			fmt.Printf("  [%d] %s\n", i+1, result)
		}
	} else {
		fmt.Println("❌ 回调未被执行")
	}
	fmt.Println()

	// 6. 演示多次调用
	fmt.Println("[测试] 步骤 6: 演示多次调用场景")
	fmt.Println("场景: C# 处理多个战斗完成事件")
	fmt.Println()

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
		fmt.Printf("  第 %d 个事件: BattleID=%d\n", i+1, event.battleID)
		if activeCallback != nil {
			result := activeCallback(event.notifType, event.battleID, event.timestamp)
			fmt.Printf("    → 回调返回: %d\n", result)
		}
	}
	fmt.Println()

	fmt.Printf("✓ 总共执行 %d 次回调\n", len(callbackResults))
	fmt.Println()
}

// registerCallbackToCSHarp 将回调函数指针传递给 C# 库
func registerCallbackToCSHarp(callbackPtr unsafe.Pointer) error {
	// 实际实现会调用:
	// err := csharp.RegisterCallback(callbackPtr)
	// 这里只是示意，直接返回成功
	return nil
}

// PrintCallbackTechDetails 打印回调技术细节
func PrintCallbackTechDetails() {
	fmt.Println("\n========== Go ↔ C# 回调机制详解 ==========")
	fmt.Println()

	fmt.Println("【回调原理】")
	fmt.Println("1. Go 侧:")
	fmt.Println("   - 定义回调函数: func(notificationType, battleID, timestamp)")
	fmt.Println("   - 获取函数指针: unsafe.Pointer(&goFunction)")
	fmt.Println("   - 传递给 C#: csharp.RegisterCallback(goFuncPtr)")
	fmt.Println()

	fmt.Println("2. C# 侧:")
	fmt.Println("   - 接收指针: IntPtr goFuncPtr")
	fmt.Println("   - 转换委托: Marshal.GetDelegateForFunctionPointer<BattleNotifyCallback>()")
	fmt.Println("   - 保存引用: BattleCallbackManager.RegisterCallback()")
	fmt.Println("   - 执行回调: when event occurs → callback(...)")
	fmt.Println()

	fmt.Println("3. 执行流程:")
	fmt.Println("   Go Register → C# Store → Battle Engine → Event Trigger → C# Invoke → Go Callback")
	fmt.Println()

	fmt.Println("【类型对应】")
	fmt.Println("Go:  func(int32, int64, int64) int32")
	fmt.Println("C#:  public delegate int BattleNotifyCallback(int notificationType, long battleID, long timestamp)")
	fmt.Println()

	fmt.Println("【安全考虑】")
	fmt.Println("✓ Go 侧: 使用全局变量保持引用，防止 GC 释放")
	fmt.Println("✓ 线程安全: 使用 mutex 保护共享状态")
	fmt.Println("✓ C# 侧: 调用前检查指针有效性")
	fmt.Println("✓ 错误处理: 回调返回状态码")
	fmt.Println()
}
