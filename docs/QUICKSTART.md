# 快速开始指南

## 5 分钟快速开始

### 前置条件

✅ .NET 8.0 SDK  
✅ Go 1.23+  
✅ Linux x86-64  

### 步骤 1: 初始化（30秒）

```bash
cd /home/vagrant/workspace

# 确保已编译 C# 库
bash build_csharp_so.sh

# 验证库文件
ls -lh lib/TestExport_*.so
```

### 步骤 2: 编译 Go 程序（20秒）

```bash
go build -o test_battle cmd/test/main.go
```

### 步骤 3: 运行（2秒）

```bash
./test_battle
```

### 预期输出

```
========== Go ↔ C# 双向调用集成测试 ==========

[TEST] 步骤 1: 初始化 C# 库
✓ C# 库已初始化

[TEST] 步骤 2: 测试单场战斗 (同步调用)
✓ 战斗执行成功
  胜方: Team 1002
  败方: Team 1001

[TEST] 步骤 3: 测试批量战斗 (同步调用)
✓ 批量战斗执行成功

========== 所有测试完成 ==========
```

## 在自己的代码中使用

### 导入包

```go
import (
    "github.com/luhaoting/goPureWithCsharp/pkg/csharp"
    "github.com/luhaoting/goPureWithCsharp/pkg/proto"
)
```

### 初始化和调用

```go
package main

import (
    "fmt"
    "github.com/luhaoting/goPureWithCsharp/pkg/csharp"
    "github.com/luhaoting/goPureWithCsharp/pkg/proto"
)

func main() {
    // 1. 初始化
    err := csharp.InitCSharpLib("Release")
    if err != nil {
        panic(err)
    }
    defer csharp.CloseCSharpLib()

    // 2. 创建请求
    battleReq := &proto.StartBattle{
        Atk: &proto.Team{
            TeamId:   1001,
            TeamName: "Team A",
            Lineup:   []uint32{101, 102, 103},
        },
        Def: &proto.Team{
            TeamId:   1002,
            TeamName: "Team B",
            Lineup:   []uint32{201, 202, 203},
        },
        BattleId:  50001,
    }

    // 3. 执行战斗
    result, err := csharp.ExecBattle(battleReq)
    if err != nil {
        fmt.Printf("战斗执行失败: %v\n", err)
        return
    }

    // 4. 处理结果
    fmt.Printf("胜方: %d, 败方: %d\n", result.Winner, result.Loser)
    fmt.Printf("战斗积分: %d\n", result.BattleScore)
}
```

## 常见问题

### Q: 如何修改战斗逻辑？

A: 编辑 `CSharpProject/SimpleBattleEngine.cs`：

```csharp
// 修改初始血量
int atkHealth = 500;  // 从 300 改为 500
int defHealth = 500;

// 修改伤害范围
int atkDamage = _random.Next(30, 60);  // 改为 30-60
```

然后重新编译：
```bash
bash build_csharp_so.sh
```

### Q: 如何添加新的消息类型？

A: 编辑 `protos/battle.proto`，添加新消息：

```proto
message NewMessage {
  uint32 id = 1;
  string name = 2;
}
```

然后重新生成代码：
```bash
bash gen_proto.sh
bash build_csharp_so.sh
```

### Q: 如何调试？

A: 使用 Debug 版本的库：

```go
err := csharp.InitCSharpLib("Debug")
```

Debug 库（6.5MB）包含调试符号和详细日志。

### Q: 支持哪些平台？

A: 当前支持 **Linux x86-64**。

扩展到其他平台需要：
1. 修改 `CSharpProject/TestExport.csproj` 中的 `RuntimeIdentifier`
2. 在对应平台编译 C# 库
3. 调整 Go 侧的库加载路径

## 项目结构

```
/home/vagrant/workspace/
├── protos/
│   └── battle.proto           # Protobuf 消息定义
├── pkg/
│   ├── proto/
│   │   ├── battle.pb.go       # Go 生成的代码
│   │   └── init.go
│   └── csharp/
│       └── caller_purego.go   # Purego 包装
├── CSharpProject/
│   ├── ExportedFunctions.cs   # C# 导出函数
│   ├── SimpleBattleEngine.cs  # 战斗引擎
│   ├── BattleCallback.cs      # 回调管理
│   ├── Proto/
│   │   └── Battle.g.cs        # C# 生成的代码
│   └── build.sh               # C# 构建脚本
├── cmd/
│   ├── example/
│   │   └── main.go            # 简单示例
│   └── test/
│       └── main.go            # 集成测试
├── lib/
│   ├── TestExport_Release.so  # Release 库
│   └── TestExport_Debug.so    # Debug 库
├── docs/
│   ├── PUREGO_GUIDE.md        # 完整指南
│   ├── INTEGRATION_TEST.md    # 测试文档
│   └── QUICKSTART.md          # 本文件
└── build_csharp_so.sh         # 编译脚本
```

## 下一步

1. **阅读完整文档**：`docs/PUREGO_GUIDE.md`
2. **查看测试用例**：`cmd/test/main.go`
3. **修改战斗逻辑**：`CSharpProject/SimpleBattleEngine.cs`
4. **添加自定义消息**：`protos/battle.proto`

## 获取帮助

查看相关文档：
- 📖 完整指南：`docs/PUREGO_GUIDE.md`
- 🧪 测试文档：`docs/INTEGRATION_TEST.md`
- 📚 Purego 官网：https://github.com/ebitengine/purego
- 📚 Protobuf 官网：https://protobuf.dev

## 许可证

MIT License - 自由使用和修改
