package csharp

import (
	"os"
	"runtime/debug"
	"testing"
)

var globalBuff = make([]byte, 1024*10)

func ExecptHandler() {
	exceptmsg := string(globalBuff)
	println("[Go] 捕获到 C# 异常:", exceptmsg)
}

func Test_CSharpPainc(t *testing.T) {

	cleanup := setupConfigLoaderTest(t)
	defer cleanup()

	f, err := os.OpenFile("crash.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		t.Fatalf("无法创建 crash.log 文件: %v", err)
		t.Fail()
	}

	debug.SetCrashOutput(f, debug.CrashOptions{})

	// CANT CATCH HHHHHA
	defer func() {
		if r := recover(); r != nil {
			t.Logf("✓ 捕获到 Panic: %v", r)
		} else {
			t.Errorf("❌ 未捕获到 Panic")
		}
	}()

	exCtx := CSharpExceptionContext{
		CallbackPtr:           ExecptHandler,
		NotifyExceptionBuffer: globalBuff,
	}

	InjectCSharpExceptionCallback(exCtx)

	CallCSharpPainc()
}
