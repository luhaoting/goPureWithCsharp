#!/bin/bash

# Proto 文件编译脚本
# 功能: 编译 .proto 文件为 Go 和 C# 代码

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

print_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

print_header() {
    echo -e "${BLUE}========================================${NC}"
    echo -e "${BLUE}$1${NC}"
    echo -e "${BLUE}========================================${NC}"
}

# 获取脚本目录
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
cd "$SCRIPT_DIR"

print_header "Proto 文件编译器"

# 检查 protoc 是否安装
if ! command -v protoc &> /dev/null; then
    print_error "protoc 未安装"
    print_info "请运行: sudo apt install -y protobuf-compiler"
    exit 1
fi

print_info "protoc 版本: $(protoc --version)"

# 定义输出目录
PROTO_DIR="protos"
GO_OUT_DIR="csharp/proto"
CSHARP_OUT_DIR="CSharpProject/Proto"

# 创建输出目录
mkdir -p "$GO_OUT_DIR" "$CSHARP_OUT_DIR"

print_info "Proto 源目录: $PROTO_DIR"
print_info "Go 输出目录: $GO_OUT_DIR"
print_info "C# 输出目录: $CSHARP_OUT_DIR"

echo ""

# 编译 Go 代码
print_header "编译 Go Proto 代码"
protoc \
    --go_out="$GO_OUT_DIR" \
    --go_opt=paths=source_relative \
    -I="$PROTO_DIR" \
    "$PROTO_DIR"/*.proto

if [ $? -eq 0 ]; then
    print_info "✓ Go 代码编译成功"
    ls -lh "$GO_OUT_DIR"/*.pb.go 2>/dev/null || print_error "未找到生成的 Go 文件"
else
    print_error "Go 代码编译失败"
    exit 1
fi

echo ""

# 编译 C# 代码
print_header "编译 C# Proto 代码"
protoc \
    --csharp_out="$CSHARP_OUT_DIR" \
    --csharp_opt=file_extension=.g.cs \
    -I="$PROTO_DIR" \
    "$PROTO_DIR"/*.proto

if [ $? -eq 0 ]; then
    print_info "✓ C# 代码编译成功"
    ls -lh "$CSHARP_OUT_DIR"/*.g.cs 2>/dev/null || print_error "未找到生成的 C# 文件"
else
    print_error "C# 代码编译失败"
    exit 1
fi

echo ""
print_header "编译完成！"
print_info "Go 生成文件:"
find "$GO_OUT_DIR" -name "*.pb.go" -exec ls -lh {} \;

echo ""
print_info "C# 生成文件:"
find "$CSHARP_OUT_DIR" -name "*.g.cs" -exec ls -lh {} \;
