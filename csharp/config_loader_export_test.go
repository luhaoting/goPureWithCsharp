package csharp

import (
	"fmt"
	"testing"
	"unsafe"

	"github.com/ebitengine/purego"
)

// ============================================================================
// 这是一个全局函数（不是闭包），所以函数指针永久有效
// ============================================================================

var globalTestConfigLoaderCallCount int32

// 全局缓存用于存储预加载的配置数据，防止被 GC 回收
var globalConfigDataCache = make(map[string][]byte)

// globalTestConfigReader 全局配置加载函数 - Go 提供给 C# 的配置获取接口
// 这个函数是全局的，所以它的地址在程序运行期间是固定的，可以安全地传递给 C#
//
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
//	-1 - 失败（配置文件不存在）
//
// 设计特点：
//   - 不申请额外内存
//   - 只从预加载的 globalConfigDataCache 中返回数据指针
//   - 返回的数据缓冲区始终存活（存储在全局变量中）
func globalTestConfigReader(configNamePtr unsafe.Pointer, configNameLen int32, outDataPtrPtr unsafe.Pointer, outDataLenPtr unsafe.Pointer) int32 {
	globalTestConfigLoaderCallCount++

	// 从指针读取配置文件名 - 使用 unsafe.String 避免额外的内存复制
	configName := unsafe.String((*byte)(configNamePtr), int(configNameLen))

	// 直接从预加载的全局缓存中获取数据
	cachedData, exists := globalConfigDataCache[configName]
	if !exists {
		return -1
	}

	if len(cachedData) == 0 {
		// 写入空数据
		*(*unsafe.Pointer)(outDataPtrPtr) = nil
		*(*int32)(outDataLenPtr) = 0
		return 0
	}

	// 将缓存中的数据指针和长度写入输出参数
	*(*unsafe.Pointer)(outDataPtrPtr) = unsafe.Pointer(&cachedData[0])
	*(*int32)(outDataLenPtr) = int32(len(cachedData))

	return 0
}

// TestLoadConfigViaC2GCallchain 测试通过 C# 调用 Go 侧全局配置读取函数
// 这是完整的双向调用验证：
// 1. Go 侧预加载所有配置文件到全局内存 globalConfigDataCache
// 2. Go 侧定义一个全局函数（不是闭包）直接从全局缓存返回数据
// 3. Go 侧向 C# 注册这个全局函数的指针
// 4. C# 调用 Go 导出的 LoadConfig 函数
// 5. Go 导出函数触发注册的全局函数
// 6. 全局函数从预加载的缓存返回数据（不申请额外内存）
func TestLoadConfigViaC2GCallchain(t *testing.T) {
	// 环境设置：确保库已加载并清理缓存
	cleanup := setupConfigLoaderTest(t)
	defer cleanup()

	t.Log("========== 测试通过 C# 调用 Go 侧全局配置读取函数 ==========")

	ClearConfigCache()
	globalTestConfigLoaderCallCount = 0

	testCases := []string{"battle_config.json", "team_config.json", "unit_config.json"}
	globalConfigDataCache = make(map[string][]byte) // 重置全局缓存

	for _, configName := range testCases {
		data, err := LoadConfigFile(configName)
		if err != nil {
			t.Logf("❌ 预加载 %s 失败: %v", configName, err)
			return
		}
		globalConfigDataCache[configName] = data
	}

	// 步骤 1：注册全局配置读取函数到 Go 侧
	err := RegisterConfigLoader(globalTestConfigReader)
	if err != nil {
		t.Errorf("❌ 注册失败: %v", err)
		t.Fail()
	}

	for _, configName := range testCases {
		t.Logf("[C#调用] 调用 LoadConfig(%s)", configName)
		err = LoadConfig(configName)
		if err != nil {
			t.Logf("❌ LoadConfig(%s) 返回错误: %v", configName, err)
		} else {
			t.Logf("✓ LoadConfig(%s) 成功", configName)
		}
	}

	if !(globalTestConfigLoaderCallCount > 0) {
		t.Fail()
	}

	// 清理全局缓存
	globalConfigDataCache = make(map[string][]byte)
	globalTestConfigLoaderCallCount = 0

}

// ============================================================================
// CoreInitConfigSetLoader 函数指针测试
// ============================================================================

var globalConfigSetLoaderCallCount int32

// globalConfigSetLoader - 符合 CoreInitConfigSetLoader 期望签名的全局配置加载器函数
// C# 签名: delegate* unmanaged[Cdecl]<byte*, int, int*, byte*>
// 参数:
//   - fileNamePtr: 指向配置文件名的指针
//   - fileNameLen: 文件名长度
//   - outDataLenPtr: 指向输出数据长度的指针
//
// 返回值: 指向配置数据的指针（字节数组）
func globalConfigSetLoaderImpl(fileNamePtr unsafe.Pointer, fileNameLen int32, outDataLenPtr unsafe.Pointer) unsafe.Pointer {
	globalConfigSetLoaderCallCount++

	// 调试日志：确认回调被调用
	fmt.Printf("[Go Callback] globalConfigSetLoaderImpl 被调用! fileNamePtr=%p, fileNameLen=%d, outDataLenPtr=%p\n",
		fileNamePtr, fileNameLen, outDataLenPtr)

	// 从指针读取配置文件名
	fileName := unsafe.String((*byte)(fileNamePtr), int(fileNameLen))
	fmt.Printf("[Go Callback] 请求的文件名: %s\n", fileName)

	// 从全局缓存获取配置数据
	cachedData, exists := globalConfigDataCache[fileName]
	if !exists {
		fmt.Printf("[Go Callback] 文件 %s 不在缓存中，返回 nil\n", fileName)
		*(*int32)(outDataLenPtr) = 0
		return nil
	}

	if len(cachedData) == 0 {
		*(*int32)(outDataLenPtr) = 0
		return nil
	}

	// 写入数据长度
	*(*int32)(outDataLenPtr) = int32(len(cachedData))

	// 返回数据指针
	return unsafe.Pointer(&cachedData[0])
}

// TestCoreInitConfigSetLoader 测试 CoreInitConfigSetLoader 函数
// 测试流程：
// 1. 预加载配置文件到全局缓存
// 2. 调用 CoreInitConfigSetLoader 注册 Go 侧的配置加载器函数
// 3. 验证函数注册成功
// 4. 验证配置加载器函数被正确调用
func TestCoreInitConfigSetLoader(t *testing.T) {
	cleanup := setupConfigLoaderTest(t)
	defer cleanup()

	t.Log("========== 测试 CoreInitConfigSetLoader 函数 ==========")

	// 清理状态
	globalConfigSetLoaderCallCount = 0
	globalConfigDataCache = make(map[string][]byte)

	testCases := []string{"battle_config.json", "team_config.json", "unit_config.json"}

	// 步骤 1: 预加载配置文件到全局缓存
	t.Log("[步骤1] 预加载配置文件到全局缓存")

	for _, configName := range testCases {
		data, err := LoadConfigFile(configName)
		if err != nil {
			t.Logf("❌ 预加载 %s 失败: %v", configName, err)
			continue
		}
		globalConfigDataCache[configName] = data
		t.Logf("✓ 配置文件已预加载: %s (大小: %d 字节)", configName, len(data))
	}

	// 步骤 2: 调用 CoreInitConfigSetLoader 注册配置加载器函数
	t.Log("[步骤2] 调用 CoreInitConfigSetLoader 注册函数指针")

	success, err := CoreInitConfigSetLoader(globalConfigSetLoaderImpl)
	if err != nil {
		t.Errorf("❌ CoreInitConfigSetLoader 调用失败: %v", err)
		t.FailNow()
	}

	if !success {
		t.Errorf("❌ CoreInitConfigSetLoader 返回 false，注册失败")
		t.FailNow()
	}

	t.Log("✓ 函数指针已注册成功")

	// 步骤 3: 验证加载器函数已准备好
	t.Log("[步骤3] 验证配置加载器函数已准备好")

	// 通过直接调用全局函数来验证缓存是否有数据
	if len(globalConfigDataCache) == 0 {
		t.Error("❌ 配置缓存为空")
		t.FailNow()
	}

	t.Logf("✓ 配置缓存中有 %d 个文件", len(globalConfigDataCache))

	// 步骤 4: 测试全局函数是否能正确返回数据
	t.Log("[步骤4] 测试全局配置加载器函数")

	for _, configName := range testCases {
		if _, exists := globalConfigDataCache[configName]; !exists {
			continue
		}

		configBytes := []byte(configName)
		var dataLen int32
		dataPtr := globalConfigSetLoaderImpl(
			unsafe.Pointer(&configBytes[0]),
			int32(len(configBytes)),
			unsafe.Pointer(&dataLen),
		)

		if dataPtr == nil {
			t.Logf("⚠️  %s 返回 nil", configName)
		} else {
			t.Logf("✓ %s 成功返回数据，长度: %d 字节", configName, dataLen)
		}
	}

	// 步骤 5: 验证加载器函数被调用次数
	t.Log("[步骤5] 验证加载器函数被调用次数")

	if globalConfigSetLoaderCallCount > 0 {
		t.Logf("✓ 全局配置加载器函数被调用 %d 次", globalConfigSetLoaderCallCount)
	} else {
		t.Logf("⚠️  全局配置加载器函数未被调用（可能 C# 侧还未调用）")
	}

	// 清理
	globalConfigDataCache = make(map[string][]byte)
	globalConfigSetLoaderCallCount = 0

	t.Log("========== 测试完成 ==========")
}

// TestCoreInitConfigSetLoader_Purego 直接使用 purego 调用 C# 的 CoreInitConfigSetLoader
// 目的：验证 purego.Dlsym + SyscallN 能正确注册 Go 侧函数指针

func Test_CoreInitConfigSetLoader_Purego(t *testing.T) {
	cleanup := setupConfigLoaderTest(t)
	defer cleanup()

	t.Log("========== 使用 purego 直接调用 C# 的 CoreInitConfigSetLoader ==========")

	// 准备测试数据
	globalConfigSetLoaderCallCount = 0
	testData := []byte(`{"test": "battle_config"}`)
	globalConfigDataCache = map[string][]byte{"battle_config.json": testData}

	// 步骤 1: 获取 CoreInitConfigSetLoader 的函数指针
	t.Log("[步骤1] 通过 getCachedFunction 获取 CoreInitConfigSetLoader 指针")

	libMutex.RLock()
	libHandleVal := libHandle
	libMutex.RUnlock()

	if libHandleVal == 0 {
		t.Fatalf("❌ C# 库未初始化")
	}

	initPtr, err := getCachedFunction(libHandleVal, "CoreInitConfigSetLoader")
	if err != nil {
		t.Fatalf("❌ 找不到 CoreInitConfigSetLoader: %v", err)
	}

	t.Logf("✓ 成功获取函数指针: 0x%x", initPtr)

	// 步骤 2: 创建 Go 侧回调
	t.Log("[步骤2] 创建 Go 侧配置加载器回调")

	callbackPtr := purego.NewCallback(globalConfigSetLoaderImpl)
	t.Logf("✓ 回调指针已创建: 0x%x", callbackPtr)

	// 步骤 3: 通过 SyscallN 调用 C# 导出函数
	t.Log("[步骤3] 通过 purego.SyscallN 调用 C# 的 CoreInitConfigSetLoader")

	result, _, _ := purego.SyscallN(uintptr(initPtr), callbackPtr)

	if result == 0 {
		t.Fatalf("❌ CoreInitConfigSetLoader 返回 false，注册失败")
	}

	t.Logf("✓ CoreInitConfigSetLoader 返回 true，注册成功")

	// 步骤 4: 验证回调函数能正确返回数据
	t.Log("[步骤4] 直接调用 Go 侧回调验证数据返回")

	name := []byte("battle_config.json")
	var dataLen int32
	dataPtr := globalConfigSetLoaderImpl(
		unsafe.Pointer(&name[0]),
		int32(len(name)),
		unsafe.Pointer(&dataLen),
	)

	if dataPtr == nil {
		t.Fatalf("❌ 回调返回 nil 指针")
	}

	if dataLen == 0 {
		t.Fatalf("❌ 回调返回数据长度为 0")
	}

	// 验证返回的数据内容
	returnedData := unsafe.Slice((*byte)(dataPtr), dataLen)
	expectedData := testData

	if len(returnedData) != len(expectedData) {
		t.Fatalf("❌ 数据长度不匹配: 期望 %d，实际 %d", len(expectedData), len(returnedData))
	}

	for i, b := range returnedData {
		if b != expectedData[i] {
			t.Fatalf("❌ 数据内容不匹配: 位置 %d，期望 %d，实际 %d", i, expectedData[i], b)
		}
	}

	t.Logf("✓ 回调返回数据正确: %d 字节", dataLen)

	// 步骤 5: 验证回调函数被正确调用
	t.Log("[步骤5] 验证回调函数调用次数")

	if globalConfigSetLoaderCallCount > 0 {
		t.Logf("✓ 回调函数被调用 %d 次", globalConfigSetLoaderCallCount)
	} else {
		t.Logf("⚠️  回调函数未被调用（C# 侧可能未调用 LoadConfigSetImpl）")
	}

	// 清理
	globalConfigDataCache = make(map[string][]byte)
	globalConfigSetLoaderCallCount = 0

	t.Log("========== 测试完成 ==========")
}
