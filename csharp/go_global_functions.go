package csharp

import (
	"fmt"
	"sync"

	"google.golang.org/protobuf/proto"

	pb "goPureWithCsharp/csharp/proto"
)

// ============================================================================
// Go 侧全局可调用函数
// 这些函数可以通过导出函数注册到 C#，C# 然后可以存储引用并调用
// ============================================================================

// BattleNotificationCallback 战斗通知回调函数类型
type BattleNotificationCallback func(notification *pb.BattleNotification) error

// 全局变量
var (
	callbackMutex         sync.RWMutex
	notificationCallbacks []BattleNotificationCallback
	processNotification   func(data []byte) error
)

// GoGlobalFunctions 存储所有可从 C# 调用的 Go 全局函数
type GoGlobalFunctions struct {
	mutex sync.RWMutex

	// HandleBattleNotification 处理战斗通知的全局函数
	// 这个函数可以从 C# 调用
	HandleBattleNotification func(notification *pb.BattleNotification) error

	// ProcessBattleNotificationData 处理二进制通知数据的全局函数
	ProcessBattleNotificationData func(data []byte) error
}

// globalGoFunctions 全局函数管理器
var globalGoFunctions *GoGlobalFunctions

// init 初始化全局函数管理器
func init() {
	globalGoFunctions = &GoGlobalFunctions{
		// 默认的通知处理函数
		HandleBattleNotification:      defaultHandleBattleNotification,
		ProcessBattleNotificationData: defaultProcessBattleNotificationData,
	}
}

// defaultHandleBattleNotification 默认的战斗通知处理器
func defaultHandleBattleNotification(notification *pb.BattleNotification) error {
	// 调用已注册的回调函数
	callbackMutex.RLock()
	callbacks := make([]BattleNotificationCallback, len(notificationCallbacks))
	copy(callbacks, notificationCallbacks)
	callbackMutex.RUnlock()

	for _, callback := range callbacks {
		if callback != nil {
			if err := callback(notification); err != nil {
				return err
			}
		}
	}

	return nil
}

// defaultProcessBattleNotificationData 默认的二进制数据处理器
func defaultProcessBattleNotificationData(data []byte) error {
	return ProcessNotification(data)
}

// ProcessNotification 处理二进制通知数据
func ProcessNotification(data []byte) error {
	var notification pb.BattleNotification
	err := proto.Unmarshal(data, &notification)
	if err != nil {
		return fmt.Errorf("反序列化战斗通知失败: %w", err)
	}
	return globalGoFunctions.HandleBattleNotification(&notification)
}

// GetGoGlobalFunctions 返回全局函数管理器
// C# 可以通过这个来获取所有可调用的 Go 函数
func GetGoGlobalFunctions() *GoGlobalFunctions {
	return globalGoFunctions
}

// SetBattleNotificationHandler 设置战斗通知处理函数
// 这允许 Go 侧动态修改通知处理逻辑
func SetBattleNotificationHandler(handler func(notification *pb.BattleNotification) error) {
	if globalGoFunctions == nil {
		return
	}

	globalGoFunctions.mutex.Lock()
	defer globalGoFunctions.mutex.Unlock()

	globalGoFunctions.HandleBattleNotification = handler
	fmt.Println("[Go] 战斗通知处理器已更新")
}

// SetProcessNotificationDataHandler 设置二进制数据处理函数
func SetProcessNotificationDataHandler(handler func(data []byte) error) {
	if globalGoFunctions == nil {
		return
	}

	globalGoFunctions.mutex.Lock()
	defer globalGoFunctions.mutex.Unlock()

	globalGoFunctions.ProcessBattleNotificationData = handler
	fmt.Println("[Go] 二进制数据处理器已更新")
}

// ============================================================================
// 导出给 C# 的函数 - 用于 C# 调用 Go 的全局函数
// ============================================================================

// CallGoHandleBattleNotification C# 调用此函数，然后会调用 Go 的全局函数
// 参数: battleId, notificationType, timestamp
// 返回: 0 成功, -1 失败
func CallGoHandleBattleNotification(battleID uint32, notificationType int32, timestamp int64) error {
	if globalGoFunctions == nil {
		return fmt.Errorf("全局函数未初始化")
	}

	notification := &pb.BattleNotification{
		BattleId:         battleID,
		NotificationType: pb.NotificationType(notificationType),
		Timestamp:        timestamp,
	}

	globalGoFunctions.mutex.RLock()
	handler := globalGoFunctions.HandleBattleNotification
	globalGoFunctions.mutex.RUnlock()

	if handler == nil {
		return fmt.Errorf("通知处理器未设置")
	}

	return handler(notification)
}

// CallGoProcessNotificationData C# 调用此函数，处理二进制通知数据
// 参数: 二进制数据
// 返回: 错误信息，nil 表示成功
func CallGoProcessNotificationData(data []byte) error {
	if globalGoFunctions == nil {
		return fmt.Errorf("全局函数未初始化")
	}

	globalGoFunctions.mutex.RLock()
	handler := globalGoFunctions.ProcessBattleNotificationData
	globalGoFunctions.mutex.RUnlock()

	if handler == nil {
		return fmt.Errorf("数据处理器未设置")
	}

	return handler(data)
}

// ============================================================================
// 简单示例：Go 侧定义的全局函数，可以直接给 C# 使用
// ============================================================================

// GoSimpleGlobalFunction 这是一个简单的全局 Go 函数
// C# 可以通过获取这个函数的引用并调用它
func GoSimpleGlobalFunction(battleID uint32, action string) string {
	fmt.Printf("[Go] GoSimpleGlobalFunction 被调用: BattleID=%d, Action=%s\n", battleID, action)
	return fmt.Sprintf("Go处理完成: %s for Battle %d", action, battleID)
}

// GoCalculateSum 计算两个数之和的全局函数
// 这个函数也可以从 C# 调用
func GoCalculateSum(a int32, b int32) int32 {
	fmt.Printf("[Go] GoCalculateSum 被调用: %d + %d\n", a, b)
	return a + b
}

// GoStringHandler 处理字符串的全局函数
func GoStringHandler(input string) string {
	fmt.Printf("[Go] GoStringHandler 被调用: %s\n", input)
	return fmt.Sprintf("[GO处理] %s", input)
}
