#!/bin/bash

# 获取脚本所在目录
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

# 进入脚本目录
cd "$SCRIPT_DIR"

# 输出目录 (绝对路径)
OUTPUT_DIR="$SCRIPT_DIR/lib"

# 确保输出目录存在
mkdir -p "$OUTPUT_DIR"

# 同步目录 (用于 csharp 包测试)
CSHARP_LIB_DIR="$SCRIPT_DIR/csharp/lib"
mkdir -p "$CSHARP_LIB_DIR"

# 运行 C# 编译脚本，传入绝对路径
bash "$SCRIPT_DIR/CSharpProject/build.sh" "$OUTPUT_DIR"

# 同步 SO 文件到 csharp/lib
if [ -f "$OUTPUT_DIR/TestExport_Release.so" ]; then
    cp "$OUTPUT_DIR/TestExport_Release.so" "$CSHARP_LIB_DIR/"
    echo "[✓] 已同步 TestExport_Release.so 到 $CSHARP_LIB_DIR"
fi

if [ -f "$OUTPUT_DIR/TestExport_Debug.so" ]; then
    cp "$OUTPUT_DIR/TestExport_Debug.so" "$CSHARP_LIB_DIR/"
    echo "[✓] 已同步 TestExport_Debug.so 到 $CSHARP_LIB_DIR"
fi