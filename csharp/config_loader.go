package csharp

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

// ConfigLoader 配置文件加载器
type ConfigLoader struct {
	configDir string
	mutex     sync.RWMutex
	cache     map[string][]byte
}

// globalConfigFileLoader 全局配置文件加载器实例
var globalConfigFileLoader *ConfigLoader

// init 初始化全局配置文件加载器
func init() {
	// 查找 config 目录
	configDir := findConfigDir()

	globalConfigFileLoader = &ConfigLoader{
		configDir: configDir,
		cache:     make(map[string][]byte),
	}

	fmt.Printf("[ConfigLoader] 初始化完成，配置目录: %s\n", configDir)
}

// findConfigDir 查找项目中的 config 目录
func findConfigDir() string {
	possiblePaths := []string{
		"./config",
		"../config",
		"../../config",
		"/home/vagrant/workspace/config",
	}

	for _, path := range possiblePaths {
		absPath, err := filepath.Abs(path)
		if err != nil {
			continue
		}

		if info, err := os.Stat(absPath); err == nil && info.IsDir() {
			return absPath
		}
	}

	// 创建默认目录
	defaultPath := "./config"
	_ = os.MkdirAll(defaultPath, 0755)
	absPath, _ := filepath.Abs(defaultPath)
	return absPath
}

// LoadConfigFile 从 config 目录加载配置文件
// 这是一个真实的全局函数，可以从 C# 通过配置加载器回调调用
func LoadConfigFile(configName string) ([]byte, error) {
	if globalConfigFileLoader == nil {
		return nil, fmt.Errorf("配置加载器未初始化")
	}

	return globalConfigFileLoader.LoadFile(configName)
}

// LoadFile 从配置目录加载文件
func (cl *ConfigLoader) LoadFile(filename string) ([]byte, error) {
	if filename == "" {
		return nil, fmt.Errorf("配置文件名不能为空")
	}

	// 检查缓存
	cl.mutex.RLock()
	if cached, ok := cl.cache[filename]; ok {
		cl.mutex.RUnlock()
		fmt.Printf("[ConfigLoader] 从缓存加载: %s (%d 字节)\n", filename, len(cached))
		return cached, nil
	}
	cl.mutex.RUnlock()

	// 构建文件路径
	filePath := filepath.Join(cl.configDir, filename)

	// 安全检查
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return nil, fmt.Errorf("文件路径错误: %w", err)
	}

	absConfigDir, _ := filepath.Abs(cl.configDir)
	if !isPathInside(absPath, absConfigDir) {
		return nil, fmt.Errorf("不允许访问配置目录外的文件: %s", filePath)
	}

	// 读取文件
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("读取配置文件失败 %s: %w", filePath, err)
	}

	fmt.Printf("[ConfigLoader] 已加载: %s (%d 字节)\n", filename, len(data))

	// 缓存文件内容
	cl.mutex.Lock()
	cl.cache[filename] = data
	cl.mutex.Unlock()

	return data, nil
}

// isPathInside 检查 path 是否在 basePath 内部
func isPathInside(path, basePath string) bool {
	rel, err := filepath.Rel(basePath, path)
	if err != nil {
		return false
	}

	if filepath.IsAbs(rel) || rel == ".." || len(rel) > 2 && rel[:3] == ".."+string(filepath.Separator) {
		return false
	}

	return true
}

// ClearConfigCache 清空配置文件缓存
func ClearConfigCache() {
	if globalConfigFileLoader == nil {
		return
	}

	globalConfigFileLoader.mutex.Lock()
	defer globalConfigFileLoader.mutex.Unlock()

	globalConfigFileLoader.cache = make(map[string][]byte)
	fmt.Println("[ConfigLoader] 缓存已清空")
}

// GetConfigDir 获取配置目录路径
func GetConfigDir() string {
	if globalConfigFileLoader == nil {
		return ""
	}
	return globalConfigFileLoader.configDir
}

// GetConfigFileLoader 获取全局配置文件加载器实例
func GetConfigFileLoader() *ConfigLoader {
	return globalConfigFileLoader
}

// GetCacheStats 获取缓存统计信息
func GetCacheStats() map[string]interface{} {
	if globalConfigFileLoader == nil {
		return nil
	}

	globalConfigFileLoader.mutex.RLock()
	defer globalConfigFileLoader.mutex.RUnlock()

	cacheSize := 0
	for _, data := range globalConfigFileLoader.cache {
		cacheSize += len(data)
	}

	return map[string]interface{}{
		"cached_files": len(globalConfigFileLoader.cache),
		"cache_size":   cacheSize,
		"config_dir":   globalConfigFileLoader.configDir,
	}
}
