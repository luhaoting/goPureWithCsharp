package main

import (
	"fmt"
	"log"

	"goPureWithCsharp/csharp"
)

func main() {
	// ========== 初始化 C# 动态库 ==========
	fmt.Println("初始化 C# 动态库...")
	err := csharp.InitCSharpLib("Release")
	if err != nil {
		log.Fatalf("初始化 C# 库失败: %v", err)
	}
	defer csharp.CloseCSharpLib()

	// ========== 示例 3: 切换版本 ==========
	fmt.Println("========== 切换到 Debug 版本 ==========")
	csharp.CloseCSharpLib()
	err = csharp.InitCSharpLib("Debug")
	if err != nil {
		log.Printf("切换失败: %v", err)
	} else {
		fmt.Println("✓ Debug 版本已加载")
	}
	fmt.Println()
}
