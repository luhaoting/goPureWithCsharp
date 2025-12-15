package csharp

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

// envStep 是一个可选的测试环境设置步骤
// 如果测试需要在运行前进行某些初始化，可以使用这种模式
type envStep struct {
	name string
	fn   func(*testing.T) error
}

// setupConfigLoaderTest 为配置加载器测试设置环境
// 返回一个清理函数，测试完毕后应该调用它
func setupConfigLoaderTest(t *testing.T) func() {
	t.Helper()
	steps := []envStep{
		{
			name: "加载 C# SO 库",
			fn:   ensureLibraryLoaded,
		},
		{
			name: "清空配置缓存",
			fn:   clearTestConfigCache,
		},
	}

	// 执行所有环境步骤
	for _, step := range steps {
		if err := step.fn(t); err != nil {
			t.Fail()
		}
	}

	// 返回清理函数
	return func() {
		globalConfigDataCache = make(map[string][]byte)
		globalTestConfigLoaderCallCount = 0
	}
}

// ensureLibraryLoaded 环境步骤：确保 C# SO 库已加载
func ensureLibraryLoaded(t *testing.T) error {
	// 检查库是否已初始化（libHandle > 0 表示已加载）
	// 由于 libHandle 是包私有的，我们通过尝试调用一个函数来判断
	// 如果返回"未初始化"错误，说明库未加载；否则库已加载

	// 简单的方式：检查库文件是否存在，如果存在就初始化
	wd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("获取工作目录失败: %w", err)
	}

	// 从当前目录向上查找项目根目录
	projectRoot := wd
	for i := 0; i < 3; i++ {
		libPath := filepath.Join(projectRoot, "lib", "TestExport_Release.so")
		if _, err := os.Stat(libPath); err == nil {
			// 库文件存在，初始化它
			if err := InitCSharpLib("Release"); err != nil {
				return fmt.Errorf("初始化 C# 库失败: %w", err)
			}
			return nil
		}
		projectRoot = filepath.Dir(projectRoot)
	}

	return fmt.Errorf("找不到 C# 库文件 TestExport_Release.so")
}

// clearTestConfigCache 环境步骤：清空测试配置缓存
func clearTestConfigCache(t *testing.T) error {
	globalConfigDataCache = make(map[string][]byte)
	globalTestConfigLoaderCallCount = 0
	return nil
}
