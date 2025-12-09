# 快速参考指南

## 编译和运行

### 编译 C# 库

```bash
cd /home/vagrant/workspace
bash build_csharp_so.sh
```

输出:
```
✓ 已保存: lib/TestExport_Release.so (3.7MB)
✓ 已保存: lib/TestExport_Debug.so (6.5MB)
```

### 编译 Go 程序

```bash
# 编译例子
go build -o example cmd/example/main.go

# 编译测试
go build -o test_battle cmd/test/main.go
```

### 运行程序

```bash
# 运行测试
./test_battle

# 预期输出
✓ 战斗执行成功
✓ 批量战斗执行成功
✓ 回调已注册
✓ 库已恢复
========== 所有测试完成 ==========
```

## 基本 API

### 初始化和清理

```go
import "github.com/luhaoting/goPureWithCsharp/pkg/csharp"

err := csharp.InitCSharpLib("Release")  // 或 "Debug"
if err != nil {
    log.Fatal(err)
}
defer csharp.CloseCSharpLib()
```

### 执行战斗

```go
import "github.com/luhaoting/goPureWithCsharp/pkg/proto"

battleReq := &proto.StartBattle{
    Atk: &proto.Team{
        TeamId:   1001,
        TeamName: "Red",
        Lineup:   []uint32{101, 102, 103},
    },
    Def: &proto.Team{
        TeamId:   1002,
        TeamName: "Blue",
        Lineup:   []uint32{201, 202, 203},
    },
    BattleId: 50001,
}

result, err := csharp.ExecBattle(battleReq)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Winner: %d, Score: %d\n", result.Winner, result.BattleScore)
```

### 执行批量战斗

```go
battles := []*proto.StartBattle{...}

batchReq := &proto.BatchBattleRequest{
    BatchId: "batch_001",
    Battles: battles,
}

batchResult, err := csharp.ExecBatchBattle(batchReq)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Success: %d, Failed: %d\n", 
    batchResult.SuccessCount, batchResult.FailureCount)
```

### 注册回调

```go
csharp.RegisterNotificationCallback(func(notif *proto.BattleNotification) error {
    fmt.Printf("Notification: Type=%d, BattleID=%d\n",
        notif.NotificationType, notif.BattleId)
    return nil
})
```

## 数据类型

### 请求消息

```protobuf
message StartBattle {
  Team atk = 1;          // 攻击队伍
  Team def = 2;          // 防守队伍
  uint32 battle_id = 3;  // 战斗 ID
  int64 timestamp = 4;   // 时间戳
}

message Team {
  repeated uint32 lineup = 1;  // 阵容
  uint32 team_id = 2;          // 队伍 ID
  string team_name = 3;        // 队伍名称
}
```

### 响应消息

```protobuf
message BattleResult {
  uint32 winner = 1;           // 胜方
  uint32 loser = 2;            // 败方
  int32 atk_damage = 3;        // 攻击伤害
  int32 def_damage = 4;        // 防守伤害
  repeated uint32 kills = 5;   // 击杀列表
  int64 duration = 6;          // 持续时间(ms)
  int32 battle_score = 7;      // 积分
}

message BattleResponse {
  int32 code = 1;      // 错误码
  string message = 2;  // 错误信息
  bytes result = 3;    // 结果数据
  int64 timestamp = 4; // 时间戳
}
```

### 错误码

```
0  - SUCCESS
1  - INVALID_REQUEST
2  - TEAM_NOT_FOUND
3  - INVALID_TEAM_SIZE
4  - BATTLE_NOT_FOUND
5  - DUPLICATE_BATTLE
6  - INTERNAL_ERROR
7  - TIMEOUT
8  - INVALID_PROTO_FORMAT
```

## 文件位置

```
项目根目录: /home/vagrant/workspace

核心代码:
  pkg/csharp/caller_purego.go       - Go → C# 包装
  pkg/proto/battle.pb.go             - Protobuf 消息
  CSharpProject/ExportedFunctions.cs - C# 导出函数
  CSharpProject/SimpleBattleEngine.cs - C# 业务逻辑
  CSharpProject/BattleCallback.cs    - C# 回调管理

测试:
  cmd/test/main.go                   - 集成测试
  cmd/example/main.go                - 使用示例

构建:
  build_csharp_so.sh                 - C# 编译脚本
  gen_proto.sh                       - Proto 生成脚本
  protos/battle.proto                - Proto 定义

库文件:
  lib/TestExport_Release.so          - Release 库 (3.7MB)
  lib/TestExport_Debug.so            - Debug 库 (6.5MB)

文档:
  docs/PUREGO_GUIDE.md               - Purego 完整指南
  docs/INTEGRATION_TEST.md           - 集成测试文档
  PROJECT_SUMMARY.md                 - 项目总结
  QUICK_REFERENCE.md                 - 本文件
```

## 常见问题

**Q: 如何调试?**
```bash
go build -o test_battle cmd/test/main.go
./test_battle
# 查看详细的日志输出
```

**Q: 如何处理错误?**
```go
result, err := csharp.ExecBattle(battleReq)
if err != nil {
    fmt.Printf("Error: %v\n", err)
    return
}
```

**Q: 支持并发调用吗?**
```go
// 是的，Go 侧使用 RWMutex 确保线程安全
var wg sync.WaitGroup
for i := 0; i < 100; i++ {
    wg.Add(1)
    go func() {
        defer wg.Done()
        csharp.ExecBattle(battleReq)
    }()
}
wg.Wait()
```

**Q: 如何获取更多消息?**
查看 `protos/battle.proto` 了解所有消息定义。

**Q: 如何扩展功能?**
1. 修改 `protos/battle.proto`
2. 运行 `bash gen_proto.sh` 生成代码
3. 在 C# 侧实现业务逻辑
4. 在 Go 侧添加包装函数

## 性能提示

- 使用 Release 版本库 (vs Debug: 快 2x)
- 批量处理 (vs 单个: 快 10x)
- 避免频繁初始化/关闭库
- 在需要时使用 goroutine 并发调用

## 下一步

1. 查看 `docs/INTEGRATION_TEST.md` 了解完整测试
2. 查看 `docs/PUREGO_GUIDE.md` 了解高级特性
3. 查看 `PROJECT_SUMMARY.md` 了解项目架构

---

**最后更新**: 2025-12-09
**版本**: 1.0
