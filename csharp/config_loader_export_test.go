package csharp

import (
	"testing"
	"unsafe"
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
