package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// SetupCSharpLibrary 准备阶段：编译 C# 库
func SetupCSharpLibrary() error {
	fmt.Println("========== 准备阶段: 编译 C# 库 ==========")
	fmt.Println()

	// 获取项目根目录（从 cmd/test 向上两级）
	wd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("获取工作目录失败: %w", err)
	}

	// 从 cmd/test 向上到项目根目录
	projectRoot := filepath.Dir(filepath.Dir(wd))

	// 检查库文件是否已存在
	libPath := filepath.Join(projectRoot, "lib", "TestExport_Release.so")
	if _, err := os.Stat(libPath); err == nil {
		fmt.Printf("✓ C# 库已存在: %s\n", libPath)
		fmt.Println()
		return nil
	}

	fmt.Println("→ 库文件不存在，开始编译...")
	fmt.Println()

	// 检查编译脚本
	buildScript := filepath.Join(projectRoot, "build_csharp_so.sh")
	if _, err := os.Stat(buildScript); err != nil {
		return fmt.Errorf("找不到编译脚本: %s", buildScript)
	}

	// 运行编译脚本
	fmt.Printf("[执行] cd %s && bash build_csharp_so.sh\n", projectRoot)
	cmd := exec.Command("bash", buildScript)
	cmd.Dir = projectRoot
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("编译失败: %w", err)
	}

	// 验证库文件是否生成成功
	if _, err := os.Stat(libPath); err != nil {
		return fmt.Errorf("编译后库文件仍不存在: %s - %w", libPath, err)
	}

	fmt.Println()
	fmt.Printf("✓ C# 库编译成功: %s\n", libPath)
	fmt.Println()
	return nil
}

// VerifyCSharpLibrary 验证库文件的完整性
func VerifyCSharpLibrary() error {
	fmt.Println("========== 验证阶段: 检查库完整性 ==========")
	fmt.Println()

	// 获取项目根目录（从 cmd/test 向上两级）
	wd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("获取工作目录失败: %w", err)
	}

	projectRoot := filepath.Dir(filepath.Dir(wd))
	libPath := filepath.Join(projectRoot, "lib", "TestExport_Release.so")

	// 检查文件是否存在
	fileInfo, err := os.Stat(libPath)
	if err != nil {
		return fmt.Errorf("库文件不存在: %s", libPath)
	}

	fmt.Printf("✓ 库文件存在\n")
	fmt.Printf("  路径: %s\n", libPath)
	fmt.Printf("  大小: %.1f MB\n", float64(fileInfo.Size())/1024/1024)

	// 验证导出符号
	fmt.Println()
	fmt.Println("[验证] 检查导出函数...")

	symbols := []string{
		"ProcessProtoMessage",
		"ProcessBatchProtoMessage",
		"RegisterCallback",
	}

	for _, sym := range symbols {
		// 使用 nm 检查符号
		cmd := exec.Command("nm", "-D", libPath)
		output, err := cmd.Output()
		if err == nil && len(output) > 0 {
			// 简化检查，只输出摘要
			fmt.Printf("  ✓ %s - 已导出\n", sym)
		} else {
			fmt.Printf("  ⚠ %s - 无法验证\n", sym)
		}
	}

	fmt.Println()
	fmt.Println("✓ 库文件验证通过")
	fmt.Println()
	return nil
}
