package csharp

import (
	"fmt"
	"unsafe"

	"github.com/ebitengine/purego"
)

type RegisterNotifyCb func(
	outDataPtrPtr unsafe.Pointer,
	DataLen int32) int

func RegisterBattleEndNotify(fn RegisterNotifyCb) error {

	libMutex.RLock()
	defer libMutex.RUnlock()

	if libHandle == 0 {
		return fmt.Errorf("c# 库未初始化")
	}

	// 获取 C# 侧的 RegisterConfigLoader 导出函数
	rgPtr, err := purego.Dlsym(libHandle, "RegisterBattleResultCallback")
	if err != nil {
		return fmt.Errorf("找不到函数: RegisterBattleResultCallback - %w", err)
	}

	// 使用 purego.NewCallback 将 Go 函数转换为 C 可调用的函数指针
	callbackPtr := purego.NewCallback(fn)

	// 调用 C# 的 RegisterConfigLoader，将回调指针传过去
	result, _, _ := purego.SyscallN(
		uintptr(rgPtr),
		callbackPtr,
	)
	// ox 打印 callbackPtr
	if result != 0 {
		return fmt.Errorf("RegisterBattleEndNotify 返回错误: %d", result)
	}

	fmt.Println("[Go] 战斗结束通知已注册给 C# 回调地址 : %p", callbackPtr)
	return nil
}
