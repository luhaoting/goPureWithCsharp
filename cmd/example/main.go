package main

import (
	"fmt"
	"log"

	"github.com/luhaoting/goPureWithCsharp/pkg/csharp"
)

func main() {
	// ========== 初始化 C# 动态库 ==========
	fmt.Println("初始化 C# 动态库...")
	err := csharp.InitCSharpLib("Release")
	if err != nil {
		log.Fatalf("初始化 C# 库失败: %v", err)
	}
	defer csharp.CloseCSharpLib()

	fmt.Println("✓ C# 动态库已加载 (Release 版本)")
	fmt.Println()

	// ========== 示例 1: 处理 Protobuf 消息 ==========
	fmt.Println("========== 示例: 处理 Protobuf 消息 ==========")
	fmt.Println("使用方式:")
	fmt.Println("1. 创建 Protobuf 消息")
	fmt.Println("   battleReq := &battle.StartBattle{...}")
	fmt.Println()
	fmt.Println("2. 序列化为字节")
	fmt.Println("   reqBytes, _ := proto.Marshal(battleReq)")
	fmt.Println()
	fmt.Println("3. 调用 C# 函数")
	fmt.Println("   respBytes, err := csharp.ProcessProtoMessage(reqBytes)")
	fmt.Println()
	fmt.Println("4. 反序列化响应")
	fmt.Println("   proto.Unmarshal(respBytes, &response)")
	fmt.Println()

	// ========== 示例 2: 功能列表 ==========
	fmt.Println("========== Purego API 函数列表 ==========")
	fmt.Println("csharp.InitCSharpLib(version)")
	fmt.Println("  - 初始化 C# 动态库 (Release/Debug)")
	fmt.Println()
	fmt.Println("csharp.CloseCSharpLib()")
	fmt.Println("  - 关闭动态库")
	fmt.Println()
	fmt.Println("csharp.ProcessProtoMessage(requestData []byte)")
	fmt.Println("  - 处理单个 Protobuf 消息")
	fmt.Println()
	fmt.Println("csharp.ProcessBatchProtoMessage(requestData []byte)")
	fmt.Println("  - 批量处理 Protobuf 消息")
	fmt.Println()

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
