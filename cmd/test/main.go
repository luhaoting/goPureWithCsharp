package main

import (
	"fmt"
	"log"

	"goPureWithCsharp/csharp"
)

func main() {
	fmt.Println("========== Go ↔ C# 双向调用集成测试 ==========")
	fmt.Println()

	// ========== 初始化 ==========
	fmt.Println("[TEST] 步骤 1: 初始化 C# 库")
	err := csharp.InitCSharpLib("Release")
	if err != nil {
		log.Fatalf("初始化失败: %v", err)
	}
	defer csharp.CloseCSharpLib()
	fmt.Println("✓ C# 库已初始化")
	fmt.Println()

	// ========== 运行各个测试 ==========
	// 这里可以调用测试函数，但主要是为了支持集成测试框架
	// 使用 go test 命令来运行所有 *_test.go 文件中的测试

	fmt.Println("提示: 使用 'go test -v' 来运行所有测试")
	fmt.Println("或者使用 'go test -run TestXxx' 来运行指定测试")
	fmt.Println()
	fmt.Println("========== 程序就绪 ==========")
}
