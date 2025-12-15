package main

import (
	"fmt"
	"os"
	"path/filepath"
	"unsafe"

	pb "goPureWithCsharp/csharp/proto"

	"google.golang.org/protobuf/proto"
)

// 当前文件下所有的函数 都要求可重入 (need Reentrant)

// 提交给C#调用 加载配置
func loadConfig(
	configNamePtr unsafe.Pointer, // C# config name
	configNameLen int32,
	outDataPtrPtr unsafe.Pointer, // GO wirte here, buff hander in C# config blob pointer
	outDataLen unsafe.Pointer) int32 {

	configName := unsafe.String((*byte)(configNamePtr), int(configNameLen))
	filePath := filepath.Join(configDir, configName)
	file, err := os.Open(filePath)
	if err != nil {
		return -1
	}
	defer file.Close()

	// 获取文件大小
	fileInfo, err := file.Stat()
	if err != nil {
		return -1
	}
	fileSize := fileInfo.Size()
	if fileSize > 20480 { //TODO 统一到 pb const 约定配置缓冲区
		return -1
	}

	// 获取 C# 提供的缓冲区指针和大小
	if outDataPtrPtr == nil || outDataLen == nil {
		return -1
	}

	bufferPtr := *(*unsafe.Pointer)(outDataPtrPtr)
	bufferSize := *(*int32)(outDataLen)

	// 直接读取到 C# 提供的缓冲区
	buffer := unsafe.Slice((*byte)(bufferPtr), bufferSize)
	n, err := file.Read(buffer)
	if err != nil {
		return -1
	}

	// 更新实际读取的数据长度
	*(*int32)(outDataLen) = int32(n)

	return 0
}

// 需要函数可重入
func battleOutput(
	outDataPtrPtr unsafe.Pointer, // C# battle output
	len int32) int {
	fmt.Printf(" battleOutput 数据地址 %p, 数据长度 %d\n", outDataPtrPtr, len)
	outPutCtx := &pb.BattleContext{}
	// 从指针读取结果数据
	if outDataPtrPtr != nil && len != 0 {
		dataLen := len
		dataSlice := unsafe.Slice((*byte)(outDataPtrPtr), dataLen)

		err := proto.Unmarshal(dataSlice, outPutCtx)
		if err != nil {
			fmt.Printf("[Battle] ✗ 反序列化战斗结果失败: %v\n", err)
			return -1
		}
	}

	GetBattleManager().Publish(outPutCtx)
	return 0
}
