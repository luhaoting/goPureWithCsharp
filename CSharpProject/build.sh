#!/bin/bash

# C# 编译脚本 - 编译为原生 .so 库

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

print_header "C# 编译为本地库 (.so)"

# 获取输出目录（如果没有提供，使用默认值）
OUTPUT_DIR="${1:-.}"

# 检查 dotnet 是否安装
if ! command -v dotnet &> /dev/null; then
    print_error "dotnet 未安装"
    print_info "请安装 .NET SDK"
    exit 1
fi

print_info ".NET 版本: $(dotnet --version)"

# 显示编译配置
print_info "项目文件: $(pwd)/TestExport.csproj"
print_info "输出目录: $OUTPUT_DIR"
print_info ""

# 清理旧的编译输出
print_info "清理旧的编译产物..."
rm -rf bin obj

print_info ""
print_header "编译 Release 版本"

# 编译 Release 版本
dotnet publish -c Release -r linux-x64 \
    -p:PublishAot=true \
    -p:NativeLib=Shared \
    -p:SelfContained=true

if [ $? -eq 0 ]; then
    print_info "✓ Release 编译成功"
    
    # 复制 .so 文件
    SO_FILE="bin/Release/net8.0/linux-x64/publish/TestExport.so"
    if [ -f "$SO_FILE" ]; then
        install -v "$SO_FILE" "$OUTPUT_DIR/TestExport_Release.so"
        print_info "✓ 复制 Release .so: $OUTPUT_DIR/TestExport_Release.so"
        ls -lh "$OUTPUT_DIR/TestExport_Release.so"
        
        # ✅ 复制到 csharp/lib 目录
        CSHARP_LIB_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)/csharp/lib"
        mkdir -p "$CSHARP_LIB_DIR"
        cp -v "$SO_FILE" "$CSHARP_LIB_DIR/TestExport_Release.so"
        print_info "✓ 复制到 csharp/lib: $CSHARP_LIB_DIR/TestExport_Release.so"
        ls -lh "$CSHARP_LIB_DIR/TestExport_Release.so"
    else
        print_error "未找到 Release .so 文件: $SO_FILE"
        find bin -name "*.so" -type f
        exit 1
    fi
else
    print_error "Release 编译失败"
    exit 1
fi

print_info ""
print_header "编译 Debug 版本"

# 清理旧的编译输出
rm -rf bin obj

# 编译 Debug 版本
dotnet publish -c Debug -r linux-x64 \
    -p:PublishAot=true \
    -p:NativeLib=Shared \
    -p:SelfContained=true

if [ $? -eq 0 ]; then
    print_info "✓ Debug 编译成功"
    
    # 复制 .so 文件
    SO_FILE="bin/Debug/net8.0/linux-x64/publish/TestExport.so"
    if [ -f "$SO_FILE" ]; then
        install -v "$SO_FILE" "$OUTPUT_DIR/TestExport_Debug.so"
        print_info "✓ 复制 Debug .so: $OUTPUT_DIR/TestExport_Debug.so"
        ls -lh "$OUTPUT_DIR/TestExport_Debug.so"
        
        # ✅ 复制到 csharp/lib 目录
        CSHARP_LIB_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)/csharp/lib"
        mkdir -p "$CSHARP_LIB_DIR"
        cp -v "$SO_FILE" "$CSHARP_LIB_DIR/TestExport_Debug.so"
        print_info "✓ 复制到 csharp/lib: $CSHARP_LIB_DIR/TestExport_Debug.so"
        ls -lh "$CSHARP_LIB_DIR/TestExport_Debug.so"
    else
        print_error "未找到 Debug .so 文件: $SO_FILE"
        find bin -name "*.so" -type f
        exit 1
    fi
else
    print_error "Debug 编译失败"
    exit 1
fi

print_info ""
print_header "编译完成"

print_info "导出函数验证:"
nm -D "$OUTPUT_DIR/TestExport_Release.so" | grep " T "

print_info ""
print_info "编译产物位置:"
ls -lh "$OUTPUT_DIR"/TestExport*.so

print_info ""
print_header "✓ 编译全部完成！"
