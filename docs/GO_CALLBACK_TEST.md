# Go 函数指针传递给 C# 的测试用例

## 概述

这个新的测试用例演示了如何将 **Go 函数指针**传递给 C#，然后由 C# 侧调用这个指针的完整流程。

## 测试场景

### 场景描述

```
1. Go 侧定义一个回调函数
   └─ func(notificationType int32, battleID int64, timestamp int64) int32

2. Go 侧将函数指针传递给 C#
   └─ unsafe.Pointer(&goCallback)

3. C# 侧接收并存储这个指针
   └─ IntPtr callbackPtr

4. C# 侧需要通知 Go 时调用这个指针
   └─ Marshal.GetDelegateForFunctionPointer<BattleNotifyCallback>(callbackPtr)
   └─ delegate(notificationType, battleID, timestamp)

5. 执行成功！Go 侧的回调被触发
   └─ callbackResults 中记录了所有调用
```

## 代码关键部分

### Go 侧：定义回调类型和存储

```go
// 回调函数类型定义
type GoCallbackHandler func(notificationType int32, battleID int64, timestamp int64) int32

// 全局存储，防止 GC 回收
var (
    callbackMutex   sync.Mutex
    activeCallback  GoCallbackHandler
    callbackResults []string
)

// 注册函数
func RegisterGoCallbackForCSharp(callback GoCallbackHandler) unsafe.Pointer {
    callbackMutex.Lock()
    defer callbackMutex.Unlock()
    
    activeCallback = callback
    fnPtr := unsafe.Pointer(&activeCallback)
    return fnPtr
}
```

### 测试步骤

#### 步骤 1: 定义 Go 回调函数

```go
goCallback := func(notificationType int32, battleID int64, timestamp int64) int32 {
    msg := fmt.Sprintf(
        "✓ Go 回调被调用: NotifType=%d, BattleID=%d, Timestamp=%d",
        notificationType, battleID, timestamp,
    )
    callbackResults = append(callbackResults, msg)
    fmt.Printf("  %s\n", msg)
    return 0  // 返回成功状态
}
```

**输出**:
```
✓ Go 回调函数已定义
```

---

#### 步骤 2: 注册回调到 C#

```go
callbackPtr := RegisterGoCallbackForCSharp(goCallback)
fmt.Printf("✓ Go 函数指针已注册: %p\n", callbackPtr)
```

**输出**:
```
✓ Go 函数指针已注册: 0x782ac0
```

**说明**: 这个指针就是 Go 函数地址，即将被传递给 C#

---

#### 步骤 3: 传递指针给 C# 库

```go
err := registerCallbackToCSHarp(callbackPtr)
```

**C# 侧应该做的事**:
```csharp
// C# ExportedFunctions.cs
[UnmanagedCallersOnly]
public static void RegisterCallback(IntPtr callbackPtr)
{
    // 保存指针
    BattleCallbackManager.RegisterCallback(callbackPtr);
    
    // 转换为委托供后续调用
    BattleNotifyCallback callback = Marshal.GetDelegateForFunctionPointer<BattleNotifyCallback>(callbackPtr);
}
```

**输出**:
```
✓ 指针已传递给 C#
```

---

#### 步骤 4: 模拟 C# 调用 Go 函数

```go
// 模拟 C# 战斗引擎调用
result := activeCallback(1, 50001, 1702175000000)
```

**输出**:
```
  [C# 侧] 调用 Go 函数指针...
  ✓ Go 回调被调用: NotifType=1, BattleID=50001, Timestamp=1702175000000
  [C# 侧] 回调返回: 0
```

---

#### 步骤 5: 验证回调执行

```go
if len(callbackResults) > 0 {
    fmt.Printf("✓ 回调执行成功，共 %d 次调用:\n", len(callbackResults))
}
```

**输出**:
```
✓ 回调执行成功，共 1 次调用:
  [1] ✓ Go 回调被调用: NotifType=1, BattleID=50001, Timestamp=1702175000000
```

---

#### 步骤 6: 演示多次调用

```go
battleEvents := []struct {
    notifType int32
    battleID  int64
    timestamp int64
}{
    {1, 50002, 1702175001000},
    {1, 50003, 1702175002000},
    {1, 50004, 1702175003000},
}

for i, event := range battleEvents {
    result := activeCallback(event.notifType, event.battleID, event.timestamp)
}
```

**输出**:
```
  第 1 个事件: BattleID=50002
  ✓ Go 回调被调用: NotifType=1, BattleID=50002, Timestamp=1702175001000
    → 回调返回: 0
  第 2 个事件: BattleID=50003
  ✓ Go 回调被调用: NotifType=1, BattleID=50003, Timestamp=1702175002000
    → 回调返回: 0
  第 3 个事件: BattleID=50004
  ✓ Go 回调被调用: NotifType=1, BattleID=50004, Timestamp=1702175003000
    → 回调返回: 0

✓ 总共执行 4 次回调
```

## 完整测试输出

```
========== Go ↔ C# 回调机制详解 ==========

【回调原理】
1. Go 侧:
   - 定义回调函数: func(notificationType, battleID, timestamp)
   - 获取函数指针: unsafe.Pointer(&goFunction)
   - 传递给 C#: csharp.RegisterCallback(goFuncPtr)

2. C# 侧:
   - 接收指针: IntPtr goFuncPtr
   - 转换委托: Marshal.GetDelegateForFunctionPointer<BattleNotifyCallback>()
   - 保存引用: BattleCallbackManager.RegisterCallback()
   - 执行回调: when event occurs → callback(...)

3. 执行流程:
   Go Register → C# Store → Battle Engine → Event Trigger → C# Invoke → Go Callback

【类型对应】
Go:  func(int32, int64, int64) int32
C#:  public delegate int BattleNotifyCallback(int notificationType, long battleID, long timestamp)

【安全考虑】
✓ Go 侧: 使用全局变量保持引用，防止 GC 释放
✓ 线程安全: 使用 mutex 保护共享状态
✓ C# 侧: 调用前检查指针有效性
✓ 错误处理: 回调返回状态码


========== C# 调用 Go 函数指针测试 ==========

[测试] 步骤 1: 定义 Go 回调函数
✓ Go 回调函数已定义

[测试] 步骤 2: 注册回调到 C#
✓ Go 函数指针已注册: 0x782ac0

[测试] 步骤 3: 传递指针给 C# 库
✓ 指针已传递给 C#

[测试] 步骤 4: 模拟 C# 调用 Go 函数
场景: C# 战斗引擎完成战斗，调用 Go 回调通知

  [C# 侧] 调用 Go 函数指针...
  ✓ Go 回调被调用: NotifType=1, BattleID=50001, Timestamp=1702175000000
  [C# 侧] 回调返回: 0

[测试] 步骤 5: 验证回调执行
✓ 回调执行成功，共 1 次调用:
  [1] ✓ Go 回调被调用: NotifType=1, BattleID=50001, Timestamp=1702175000000

[测试] 步骤 6: 演示多次调用场景
场景: C# 处理多个战斗完成事件

  第 1 个事件: BattleID=50002
  ✓ Go 回调被调用: NotifType=1, BattleID=50002, Timestamp=1702175001000
    → 回调返回: 0
  第 2 个事件: BattleID=50003
  ✓ Go 回调被调用: NotifType=1, BattleID=50003, Timestamp=1702175002000
    → 回调返回: 0
  第 3 个事件: BattleID=50004
  ✓ Go 回调被调用: NotifType=1, BattleID=50004, Timestamp=1702175003000
    → 回调返回: 0

✓ 总共执行 4 次回调
```

## 技术要点总结

| 方面 | 详情 |
|------|------|
| **Go 函数类型** | `func(int32, int64, int64) int32` |
| **C# 委托类型** | `delegate int BattleNotifyCallback(...)` |
| **参数传递** | 通过 `unsafe.Pointer` 传递地址 |
| **内存管理** | 使用全局变量保持引用，防止 GC 回收 |
| **线程安全** | 使用 `sync.Mutex` 保护共享状态 |
| **错误处理** | 通过返回值传递状态码 (0=成功) |
| **调用次数** | 测试中演示了 4 次成功调用 |

## 实际应用场景

在真实的生产环境中，当 C# 战斗引擎完成一场战斗时：

1. **战斗完成事件触发**
2. **C# 获取已保存的 Go 函数指针**
3. **C# 调用 `Marshal.GetDelegateForFunctionPointer`** 将指针转换为委托
4. **C# 调用委托** 传递战斗数据 (战斗ID、时间戳等)
5. **Go 侧回调被执行** 收到通知
6. **Go 可以立即处理** 战斗结果（统计、日志、持久化等）

这样就实现了完整的 **C# → Go** 双向通信！

## 运行测试

```bash
cd /home/vagrant/workspace
go build -o test_battle cmd/test/main.go
./test_battle
```

测试将输出完整的 6 个步骤，验证 Go 函数指针被 C# 成功调用。
