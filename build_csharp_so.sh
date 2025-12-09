#!/bin/bash

# 获取脚本所在目录
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

# 进入脚本目录
cd "$SCRIPT_DIR"

# 输出目录 (绝对路径)
OUTPUT_DIR="$SCRIPT_DIR/lib"

# 确保输出目录存在
mkdir -p "$OUTPUT_DIR"

# 运行 C# 编译脚本，传入绝对路径
bash "$SCRIPT_DIR/CSharpProject/build.sh" "$OUTPUT_DIR"