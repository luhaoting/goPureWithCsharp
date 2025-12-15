package csharp

import (
	"fmt"
	"unsafe"

	"github.com/ebitengine/purego"
)

// 全局保存所有回调指针，防止被 GC 回收
var savedCallbacks []uintptr

// RegisterConfigLoaderFunc Go 侧的配置加载器回调函数签名
// 参数:
//
//	configNamePtr - 配置文件名称指针
//	configNameLen - 配置文件名称长度（字节）
//	outDataPtrPtr - 指向数据指针的指针（输出参数）
//	outDataLenPtr - 指向数据长度的指针（输出参数）
//
// 返回值:
//
//	0  - 成功，数据已写入输出参数
//	-1 - 失败
//
// 设计原因:
// purego.NewCallback 只支持 0-1 个返回值，为了传递多个值（数据指针、长度、错误码）
// 采用输出参数指针的 C 传统模式：通过指针参数返回多个值
type RegisterConfigLoaderFunc func(
	configNamePtr unsafe.Pointer,
	configNameLen int32,
	outDataPtrPtr unsafe.Pointer,
	outDataLenPtr unsafe.Pointer,
) int32

// RegisterConfigLoader 向 C# 注册 Go 侧的配置加载器函数
//
// 完整流程（双向调用）：
// ════════════════════════════════════════════════════════════════
// 1. Go 侧定义全局函数：globalTestConfigReader(configNamePtr, configNameLen, outDataPtrPtr, outDataLenPtr) int32
// 2. Go 侧调用此函数注册：RegisterConfigLoader(globalTestConfigReader)
// 3. 此函数使用 purego.NewCallback 包装 Go 函数为 C 可调用的函数指针
// 4. 将函数指针通过 purego.SyscallN 传给 C# 的 RegisterConfigLoader 导出函数
// 5. C# 侧使用 Marshal.GetDelegateForFunctionPointer 解析并存储这个回调
// 6. 后续 Go 调用 C# 的 LoadConfig 时，C# 会调用这个回调来获取配置数据
func RegisterConfigLoader(fn RegisterConfigLoaderFunc) error {

	libMutex.RLock()
	defer libMutex.RUnlock()

	if libHandle == 0 {
		return fmt.Errorf("C# 库未初始化")
	}

	// 获取 C# 侧的 RegisterConfigLoader 导出函数
	rgPtr, err := purego.Dlsym(libHandle, "RegisterConfigLoader")
	if err != nil {
		return fmt.Errorf("找不到函数: RegisterConfigLoader - %w", err)
	}
	// rgPtr, err := getCachedFunction(libHandle, "RegisterConfigLoader")
	// if err != nil {
	// 	return fmt.Errorf("找不到函数: RegisterConfigLoader - %w", err)
	// }
	callbackPtr := purego.NewCallback(fn)
	// 调用 C# 的 RegisterConfigLoader，将回调指针传过去
	// 注意：现在返回类型是 void，所以只调用，不处理返回值
	purego.SyscallN(
		uintptr(rgPtr),
		callbackPtr,
	)
	fmt.Println("[ConfigLoader] SyscallN 调用完成")

	fmt.Println("[Go] 配置加载器已注册给 C#")
	return nil
}

// LoadConfig 加载配置
func LoadConfig(configName string) error {
	// 首先检查是否有已注册的配置加载器（在 Go 侧）

	libMutex.RLock()
	defer libMutex.RUnlock()

	if libHandle == 0 {
		return fmt.Errorf("C# 库未初始化")
	}

	fnPtr, err := purego.Dlsym(libHandle, "LoadConfig")
	if err != nil {
		return fmt.Errorf("找不到函数: LoadConfig - %w", err)
	}

	nameBytes := []byte(configName)
	namePtr := unsafe.Pointer(&nameBytes[0])
	nameLen := int32(len(nameBytes))

	fn := unsafe.Pointer(fnPtr)
	result, _, _ := purego.SyscallN(
		uintptr(fn),
		uintptr(namePtr),
		uintptr(nameLen),
	)

	if result != 0 {
		return fmt.Errorf("LoadConfig 返回错误: %d", result)
	}
	fmt.Printf("[Go] 配置已加载: %s\n", configName)
	return nil
}
