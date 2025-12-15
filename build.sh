#!/bin/bash

# Unified Build Script: Proto Generation, C# Compilation, and Go Build
# Usage:
#   ./build.sh [target] [options]
#   ./build.sh proto              - Generate proto files
#   ./build.sh csharp             - Build C# SO library
#   ./build.sh test               - Build Go test binary
#   ./build.sh all                - Generate proto, build C#, then build all Go commands

set -e

# Color codes for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Default values
TARGET=""
CMD_NAME=""
BUILD_FLAGS="-gcflags=all=-d=checkptr"
VERBOSE=false
SHOW_HELP=false

# Parse arguments
while [ $# -gt 0 ]; do
    case $1 in
        -h|--help)
            SHOW_HELP=true
            shift
            ;;
        -v|--verbose)
            VERBOSE=true
            shift
            ;;
        -gcflags=*|--ldflags=*|--tags=*)
            BUILD_FLAGS="$BUILD_FLAGS $1"
            shift
            ;;
        -*)
            echo "${RED}❌ Unknown option: $1${NC}"
            SHOW_HELP=true
            shift
            ;;
        *)
            if [ -z "$TARGET" ]; then
                TARGET="$1"
            else
                BUILD_FLAGS="$BUILD_FLAGS $1"
            fi
            shift
            ;;
    esac
done

# ============================================================================
# Helper Functions
# ============================================================================

print_header() {
    echo ""
    echo "${BLUE}╔════════════════════════════════════════════════════════════════╗${NC}"
    echo "${BLUE}║ $1${NC}"
    echo "${BLUE}╚════════════════════════════════════════════════════════════════╝${NC}"
    echo ""
}

print_info() {
    echo -e "${GREEN}[✓]${NC} $1"
}

print_warn() {
    echo -e "${YELLOW}[!]${NC} $1"
}

print_error() {
    echo -e "${RED}[✗]${NC} $1"
}

show_help() {
    cat << 'EOF'
╔══════════════════════════════════════════════════════════════════════════════╗
║                    Unified Build Script v2.0                                 ║
║   Proto Generation + C# Compilation + Go Build                              ║
╚══════════════════════════════════════════════════════════════════════════════╝

Usage:
  ./build.sh [target] [options]

Targets:
  proto               生成 Proto 文件 (Go + C#)
  csharp              编译 C# SO 库
  test                编译 test 命令
  example             编译 example 命令
  battle              编译 battle 命令
  demo_client         编译 demo_client 命令
  all                 全部构建 (proto → csharp → go commands)
  <cmd_name>          编译特定命令 (cmd 目录下的任意子目录)

Options:
  -gcflags=FLAGS      Go 编译器标志 (默认: -gcflags=all=-d=checkptr)
  --ldflags=FLAGS     链接器标志
  --tags=TAGS         构建标签
  -v, --verbose       详细输出
  -h, --help          显示此帮助信息

Examples:
  # 生成 proto 文件
  ./build.sh proto

  # 编译 C# SO 库
  ./build.sh csharp

  # 编译 test 命令
  ./build.sh test

  # 全部构建（推荐第一次运行）
  ./build.sh all

  # 编译 test 并显示详细信息
  ./build.sh test -v

  # 编译多个命令
  ./build.sh battle
  ./build.sh demo_client

Available Commands in cmd/:
EOF
    
    # List available cmd directories
    if [ -d "./cmd" ]; then
        echo ""
        for dir in ./cmd/*/; do
            if [ -d "$dir" ]; then
                cmd_name=$(basename "$dir")
                if [ -f "$dir/main.go" ]; then
                    echo "  ✓ $cmd_name"
                else
                    echo "  - $cmd_name (no main.go)"
                fi
            fi
        done
    fi
    
    echo ""
}

# ============================================================================
# Proto Generation
# ============================================================================

build_proto() {
    print_header "Proto Generation (Go + C#)"
    
    # Check if protoc is installed
    if ! command -v protoc &> /dev/null; then
        print_error "protoc not installed"
        print_info "Install with: sudo apt install -y protobuf-compiler"
        return 1
    fi
    
    print_info "protoc version: $(protoc --version)"
    echo ""
    
    # Define directories
    PROTO_DIR="protos"
    GO_OUT_DIR="csharp/proto"
    CSHARP_OUT_DIR="CSharpProject/Proto"
    
    # Create output directories
    mkdir -p "$GO_OUT_DIR" "$CSHARP_OUT_DIR"
    
    print_info "Proto source: $PROTO_DIR"
    print_info "Go output: $GO_OUT_DIR"
    print_info "C# output: $CSHARP_OUT_DIR"
    echo ""
    
    # Compile Go proto code
    echo "${YELLOW}Compiling Go proto code...${NC}"
    if protoc \
        --go_out="$GO_OUT_DIR" \
        --go_opt=paths=source_relative \
        -I="$PROTO_DIR" \
        "$PROTO_DIR"/*.proto; then
        print_info "Go proto compilation successful"
        ls -lh "$GO_OUT_DIR"/*.pb.go 2>/dev/null || print_warn "No Go files found"
    else
        print_error "Go proto compilation failed"
        return 1
    fi
    
    echo ""
    
    # Compile C# proto code
    echo "${YELLOW}Compiling C# proto code...${NC}"
    if protoc \
        --csharp_out="$CSHARP_OUT_DIR" \
        --csharp_opt=file_extension=.g.cs \
        -I="$PROTO_DIR" \
        "$PROTO_DIR"/*.proto; then
        print_info "C# proto compilation successful"
        ls -lh "$CSHARP_OUT_DIR"/*.g.cs 2>/dev/null || print_warn "No C# files found"
    else
        print_error "C# proto compilation failed"
        return 1
    fi
    
    echo ""
    print_info "Proto generation completed!"
    return 0
}

# ============================================================================
# C# SO Library Build
# ============================================================================

build_csharp() {
    print_header "Building C# SO Library"
    
    SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
    OUTPUT_DIR="$SCRIPT_DIR/lib"
    
    # Create output directory
    mkdir -p "$OUTPUT_DIR"
    
    print_info "Output directory: $OUTPUT_DIR"
    print_info "Calling CSharpProject/build.sh..."
    echo ""
    
    # Run C# build script
    if bash "$SCRIPT_DIR/CSharpProject/build.sh" "$OUTPUT_DIR"; then
        print_info "C# SO library built successfully"
        ls -lh "$OUTPUT_DIR"/*.so 2>/dev/null || print_warn "No SO files found"
        return 0
    else
        print_error "C# build failed"
        return 1
    fi
}

# ============================================================================
# Go Command Build
# ============================================================================

build_go_command() {
    local CMD_NAME="$1"
    
    # Check if cmd directory exists
    if [ ! -d "./cmd/$CMD_NAME" ]; then
        print_error "cmd/$CMD_NAME not found"
        return 1
    fi
    
    # Check if main.go exists
    if [ ! -f "./cmd/$CMD_NAME/main.go" ]; then
        print_error "./cmd/$CMD_NAME/main.go not found"
        return 1
    fi
    
    # Prepare build path
    BUILD_PATH="./cmd/$CMD_NAME"
    
    # Output info
    echo "${YELLOW}Build Configuration:${NC}"
    echo "  Command: $CMD_NAME"
    echo "  Path: $BUILD_PATH"
    echo "  Build Flags: ${BUILD_FLAGS:-(none)}"
    echo ""
    
    # Set verbose flag
    if [ "$VERBOSE" = true ]; then
        echo "${YELLOW}Build Output:${NC}"
        echo ""
        if ! go build -v $BUILD_FLAGS "$BUILD_PATH"; then
            print_error "Build failed!"
            return 1
        fi
    else
        echo "${YELLOW}Building...${NC}"
        if ! go build $BUILD_FLAGS "$BUILD_PATH" 2>&1; then
            print_error "Build failed!"
            return 1
        fi
    fi
    
    echo ""
    print_info "Build completed successfully!"
    
    # Show build info
    BINARY_NAME=$(basename "$BUILD_PATH")
    BUILD_TIME=$(date '+%Y-%m-%d %H:%M:%S')
    echo "${BLUE}Build Info:${NC}"
    echo "  Command: $BINARY_NAME"
    echo "  Time: $BUILD_TIME"
    echo ""
    
    return 0
}

# ============================================================================
# Main Logic
# ============================================================================

# Show help if requested or no target specified
if [ "$SHOW_HELP" = true ] || [ -z "$TARGET" ]; then
    show_help
    exit 0
fi

# Handle different targets
case "$TARGET" in
    proto)
        build_proto
        exit $?
        ;;
    csharp)
        build_csharp
        exit $?
        ;;
    all)
        print_header "Full Build (Proto + C# + Go)"
        
        # Build proto
        if ! build_proto; then
            print_error "Proto generation failed"
            exit 1
        fi
        
        echo ""
        
        # Build C#
        if ! build_csharp; then
            print_error "C# build failed"
            exit 1
        fi
        
        echo ""
        
        # Build all Go commands
        print_header "Building Go Commands"
        
        if [ -d "./cmd" ]; then
            FAILED_COMMANDS=""
            for dir in ./cmd/*/; do
                if [ -d "$dir" ]; then
                    cmd_name=$(basename "$dir")
                    if [ -f "$dir/main.go" ]; then
                        echo ""
                        echo "${BLUE}Building: $cmd_name${NC}"
                        if ! build_go_command "$cmd_name"; then
                            FAILED_COMMANDS="$FAILED_COMMANDS $cmd_name"
                        fi
                    fi
                fi
            done
            
            if [ -n "$FAILED_COMMANDS" ]; then
                echo ""
                print_error "Failed commands:$FAILED_COMMANDS"
                exit 1
            fi
        fi
        
        echo ""
        print_header "Full Build Completed!"
        exit 0
        ;;
    *)
        # Build specific Go command
        print_header "Building Go Command: $TARGET"
        build_go_command "$TARGET"
        exit $?
        ;;
esac
