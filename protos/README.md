# Proto 文件使用指南

## 项目结构

```
protos/
├── battle.proto          # Proto 定义文件
└── README.md            # 本文件

pkg/
└── proto/
    └── battle.pb.go     # 生成的 Go 代码

CSharpProject/
└── Proto/
    └── Battle.g.cs      # 生成的 C# 代码
```

## Proto 文件说明

### battle.proto 包含的消息

#### 基础结构
- **Team** - 队伍信息（阵容ID、队伍ID、队伍名称）

#### 请求消息
- **StartBattle** - 开始战斗请求（攻防双方队伍）
- **BattleInput** - 通用战斗输入（操作ID、原始数据、参数）
- **BattleUseItem** - 使用道具请求（道具ID、使用者、数量）

#### 响应消息
- **BattleResult** - 战斗结果（胜负方、伤害、击杀数等）
- **BattleStatus** - 战斗状态（回合、生命值、战斗状态）
- **BattleResponse** - 通用响应（错误码、消息、结果）

#### 批量操作
- **BatchBattleRequest** - 批量战斗请求
- **BatchBattleResponse** - 批量战斗响应

## 编译 Proto 文件

### 前置条件

```bash
# 安装 protobuf 编译器
sudo apt install -y protobuf-compiler

# 安装 Go protobuf 插件
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
```

### 编译命令

```bash
# 编译所有 proto 文件（生成 Go 和 C# 代码）
bash gen_proto.sh

# 或手动编译
protoc --go_out=. --go_opt=paths=source_relative protos/*.proto
protoc --csharp_out=CSharpProject/Proto --csharp_opt=file_extension=.g.cs protos/*.proto
```

## Go 中使用

### 1. 创建消息

```go
package main

import (
    "github.com/luhaoting/goPureWithCsharp/pkg/proto"
)

func main() {
    // 创建队伍
    team := &proto.Team{
        Lineup:   []uint32{1, 2, 3, 4, 5},
        TeamId:   1001,
        TeamName: "Dragon Team",
    }

    // 创建战斗请求
    battleReq := &proto.StartBattle{
        Atk:       team,
        Def:       &proto.Team{TeamId: 1002, TeamName: "Tiger Team"},
        BattleId:  50001,
        Timestamp: time.Now().UnixMilli(),
    }
    
    // 序列化
    data, _ := proto.Marshal(battleReq)
    
    // 反序列化
    received := &proto.StartBattle{}
    proto.Unmarshal(data, received)
}
```

### 2. 批量操作

```go
func main() {
    batchReq := &proto.BatchBattleRequest{
        BatchId: "batch_001",
        Battles: []*proto.StartBattle{
            // ... 多个战斗请求
        },
        Parallel: 1, // 是否并行
    }
    
    data, _ := proto.Marshal(batchReq)
    // 发送给 C# 处理
}
```

## C# 中使用

### 1. 创建消息

```csharp
using GoPureWithCsharp.Battle;

var team = new Team
{
    Lineup = { 1, 2, 3, 4, 5 },
    TeamId = 1001,
    TeamName = "Dragon Team"
};

var battleReq = new StartBattle
{
    Atk = team,
    Def = new Team { TeamId = 1002, TeamName = "Tiger Team" },
    BattleId = 50001,
    Timestamp = DateTimeOffset.UtcNow.ToUnixTimeMilliseconds()
};

// 序列化
byte[] data = battleReq.ToByteArray();

// 反序列化
var received = StartBattle.Parser.ParseFrom(data);
```

## Go 和 C# 通信示例

### C# 导出函数

```csharp
[DllExport]
public static int ProcessBattle(
    IntPtr requestBytes, int requestLen,
    IntPtr responsePtr, ref int responseLen)
{
    try
    {
        // 反序列化请求
        byte[] reqData = new byte[requestLen];
        Marshal.Copy(requestBytes, reqData, 0, requestLen);
        var request = StartBattle.Parser.ParseFrom(reqData);

        // 处理战斗逻辑
        var result = new BattleResult
        {
            Winner = request.Atk.TeamId,
            Loser = request.Def.TeamId,
            AtkDamage = 100,
            DefDamage = 50,
            Duration = 5000,
            BattleScore = 1000
        };

        // 序列化响应
        byte[] respData = result.ToByteArray();
        if (respData.Length > responseLen)
        {
            responseLen = respData.Length;
            return -1;
        }

        Marshal.Copy(respData, 0, responsePtr, respData.Length);
        responseLen = respData.Length;
        return 0;
    }
    catch (Exception ex)
    {
        return -1;
    }
}
```

### Go 调用代码

```go
package main

import (
    "unsafe"
    pb "github.com/luhaoting/goPureWithCsharp/pkg/proto"
)

// #cgo LDFLAGS: -L./lib -lTestExport_Release
// #include <stdint.h>
// int ProcessBattle(void* request, int request_len, void* response, int* response_len);
import "C"

func CallCSharpBattle(battleReq *pb.StartBattle) (*pb.BattleResult, error) {
    // 序列化请求
    reqData, _ := proto.Marshal(battleReq)
    
    // 准备响应缓冲区
    respBuffer := make([]byte, 10240)
    respLen := len(respBuffer)
    
    // 调用 C# 函数
    retCode := C.ProcessBattle(
        unsafe.Pointer(&reqData[0]),
        C.int(len(reqData)),
        unsafe.Pointer(&respBuffer[0]),
        (*C.int)(unsafe.Pointer(&respLen)),
    )
    
    if retCode != 0 {
        return nil, fmt.Errorf("C# 函数返回错误: %d", retCode)
    }
    
    // 反序列化响应
    result := &pb.BattleResult{}
    proto.Unmarshal(respBuffer[:respLen], result)
    
    return result, nil
}
```

## 修改 Proto 文件

1. 编辑 `protos/battle.proto`
2. 运行 `bash gen_proto.sh` 重新生成
3. 在 Go/C# 代码中使用新的消息定义

## 常见错误

| 错误 | 原因 | 解决方案 |
|------|------|--------|
| `protoc: program not found` | protoc 未安装 | `sudo apt install protobuf-compiler` |
| `protoc-gen-go: not found` | Go 插件未安装 | `go install google.golang.org/protobuf/cmd/protoc-gen-go@latest` |
| 字段标号重复 | 同一消息中字段号相同 | 检查 proto 文件中 `= N` 编号 |
| `syntax error` | Proto 语法错误 | 检查是否是 `proto3` 语法 |

## 性能对比

| 指标 | JSON | **Protobuf** |
|------|------|-----------|
| 消息大小 | ~1KB | ~100B |
| 序列化时间 | 10ms | 1ms |
| 反序列化时间 | 10ms | 1ms |
| 跨版本兼容 | 较好 | 最好 |

## 更多资源

- [Protobuf 官方文档](https://developers.google.com/protocol-buffers)
- [Go protobuf 包](https://pkg.go.dev/google.golang.org/protobuf)
- [C# protobuf 包](https://github.com/protocolbuffers/protobuf/tree/main/csharp)
