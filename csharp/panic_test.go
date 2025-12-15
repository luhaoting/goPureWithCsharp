package csharp

import (
	"testing"
)

func Test_CSharpPainc(t *testing.T) {

	// f, err := os.OpenFile("crash.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	// if err != nil {
	// 	t.Fatalf("无法创建 crash.log 文件: %v", err)
	// 	t.Fail()
	// }

	// debug.SetCrashOutput(f, debug.CrashOptions{})

	// // CANT CATCH HHHHHA
	// defer func() {
	// 	if r := recover(); r != nil {
	// 		t.Logf("✓ 捕获到 Panic: %v", r)
	// 	} else {
	// 		t.Errorf("❌ 未捕获到 Panic")
	// 	}
	// }()

	// cleanup := setupConfigLoaderTest(t)
	// defer cleanup()

	// t.Log("[步骤 1] 调用 C# 侧触发 Panic 的函数")
	// CallCSharpPainc()
}
