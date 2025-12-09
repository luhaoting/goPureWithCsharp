using System;
using System.Runtime.InteropServices;
using Google.Protobuf;
using GoPureWithCsharp.Battle;

namespace GoPureWithCsharp
{
    /// <summary>
    /// C# 导出函数 - 供 Go 通过 Purego 调用
    /// </summary>
    public class ExportedFunctions
    {
        // 响应缓冲区大小
        private const int SINGLE_RESPONSE_BUFFER_SIZE = 10240;
        private const int BATCH_RESPONSE_BUFFER_SIZE = 102400;

        /// <summary>
        /// 处理单个 Protobuf 消息
        /// 
        /// 函数签名 (C 风格):
        /// void ProcessProtoMessage(
        ///     const uint8_t* request_data,
        ///     int32_t request_len,
        ///     uint8_t* response_buffer,
        ///     int32_t* response_len
        /// );
        /// </summary>
        [UnmanagedCallersOnly(EntryPoint = "ProcessProtoMessage")]
        public static void ProcessProtoMessage(IntPtr requestDataPtr, int requestLen, IntPtr responseBufferPtr, IntPtr responseLenPtr)
        {
            try
            {
                Console.WriteLine($"[Export] ProcessProtoMessage 被调用, 请求长度={requestLen}");

                // 读取请求数据
                byte[] requestData = new byte[requestLen];
                Marshal.Copy(requestDataPtr, requestData, 0, requestLen);

                // 尝试解析为 StartBattle 请求
                BattleResponse response;

                try
                {
                    var battleRequest = StartBattle.Parser.ParseFrom(requestData);
                    Console.WriteLine($"[Export] 收到战斗请求: BattleID={battleRequest.BattleId}");

                    // 执行战斗
                    var battleResult = SimpleBattleEngine.ExecuteBattle(battleRequest);

                    // 构建响应
                    response = new BattleResponse
                    {
                        Code = (int)BattleErrorCode.Success,
                        Message = "战斗执行成功",
                        Result = battleResult.ToByteString(),
                        Timestamp = DateTimeOffset.Now.ToUnixTimeMilliseconds(),
                    };
                }
                catch (InvalidProtocolBufferException ex)
                {
                    Console.WriteLine($"[Export] Protobuf 解析错误: {ex.Message}");
                    response = new BattleResponse
                    {
                        Code = (int)BattleErrorCode.InvalidProtoFormat,
                        Message = $"Protobuf 格式错误: {ex.Message}",
                        Timestamp = DateTimeOffset.Now.ToUnixTimeMilliseconds(),
                    };
                }
                catch (Exception ex)
                {
                    Console.WriteLine($"[Export] 内部错误: {ex.Message}");
                    response = new BattleResponse
                    {
                        Code = (int)BattleErrorCode.InternalError,
                        Message = $"内部错误: {ex.Message}",
                        Timestamp = DateTimeOffset.Now.ToUnixTimeMilliseconds(),
                    };
                }

                // 序列化响应
                byte[] responseData = response.ToByteArray();

                // 检查缓冲区大小
                if (responseData.Length > SINGLE_RESPONSE_BUFFER_SIZE)
                {
                    Console.WriteLine($"[Export] 警告: 响应大小 {responseData.Length} 超过缓冲区 {SINGLE_RESPONSE_BUFFER_SIZE}");
                    // 在实际应用中，这里应该返回错误或使用动态分配
                    Array.Resize(ref responseData, SINGLE_RESPONSE_BUFFER_SIZE);
                }

                // 写入响应数据
                Marshal.Copy(responseData, 0, responseBufferPtr, responseData.Length);

                // 设置响应长度
                Marshal.WriteInt32(responseLenPtr, responseData.Length);
                Console.WriteLine($"[Export] 响应已发送, 长度={responseData.Length}");
            }
            catch (Exception ex)
            {
                Console.WriteLine($"[Export] 异常: {ex}");
                // 写入错误响应
                var errorResponse = new BattleResponse
                {
                    Code = (int)BattleErrorCode.InternalError,
                    Message = $"处理异常: {ex.Message}",
                    Timestamp = DateTimeOffset.Now.ToUnixTimeMilliseconds(),
                }.ToByteArray();

                if (errorResponse.Length <= SINGLE_RESPONSE_BUFFER_SIZE)
                {
                    Marshal.Copy(errorResponse, 0, responseBufferPtr, errorResponse.Length);
                    Marshal.WriteInt32(responseLenPtr, errorResponse.Length);
                }
            }
        }

        /// <summary>
        /// 处理批量 Protobuf 消息
        /// 
        /// 函数签名 (C 风格):
        /// void ProcessBatchProtoMessage(
        ///     const uint8_t* request_data,
        ///     int32_t request_len,
        ///     uint8_t* response_buffer,
        ///     int32_t* response_len
        /// );
        /// </summary>
        [UnmanagedCallersOnly(EntryPoint = "ProcessBatchProtoMessage")]
        public static void ProcessBatchProtoMessage(IntPtr requestDataPtr, int requestLen, IntPtr responseBufferPtr, IntPtr responseLenPtr)
        {
            try
            {
                Console.WriteLine($"[Export] ProcessBatchProtoMessage 被调用, 请求长度={requestLen}");

                // 读取请求数据
                byte[] requestData = new byte[requestLen];
                Marshal.Copy(requestDataPtr, requestData, 0, requestLen);

                BattleResponse response;

                try
                {
                    var batchRequest = BatchBattleRequest.Parser.ParseFrom(requestData);
                    Console.WriteLine($"[Export] 收到批量战斗请求: BatchID={batchRequest.BatchId}, 数量={batchRequest.Battles.Count}");

                    // 执行批量战斗
                    var batchResult = SimpleBattleEngine.ExecuteBatchBattle(batchRequest);

                    // 构建响应
                    response = new BattleResponse
                    {
                        Code = (int)BattleErrorCode.Success,
                        Message = "批量战斗执行成功",
                        Result = batchResult.ToByteString(),
                        Timestamp = DateTimeOffset.Now.ToUnixTimeMilliseconds(),
                    };
                }
                catch (InvalidProtocolBufferException ex)
                {
                    Console.WriteLine($"[Export] Protobuf 解析错误: {ex.Message}");
                    response = new BattleResponse
                    {
                        Code = (int)BattleErrorCode.InvalidProtoFormat,
                        Message = $"Protobuf 格式错误: {ex.Message}",
                        Timestamp = DateTimeOffset.Now.ToUnixTimeMilliseconds(),
                    };
                }
                catch (Exception ex)
                {
                    Console.WriteLine($"[Export] 内部错误: {ex.Message}");
                    response = new BattleResponse
                    {
                        Code = (int)BattleErrorCode.InternalError,
                        Message = $"内部错误: {ex.Message}",
                        Timestamp = DateTimeOffset.Now.ToUnixTimeMilliseconds(),
                    };
                }

                // 序列化响应
                byte[] responseData = response.ToByteArray();

                // 检查缓冲区大小
                if (responseData.Length > BATCH_RESPONSE_BUFFER_SIZE)
                {
                    Console.WriteLine($"[Export] 警告: 响应大小 {responseData.Length} 超过缓冲区 {BATCH_RESPONSE_BUFFER_SIZE}");
                    Array.Resize(ref responseData, BATCH_RESPONSE_BUFFER_SIZE);
                }

                // 写入响应数据
                Marshal.Copy(responseData, 0, responseBufferPtr, responseData.Length);

                // 设置响应长度
                Marshal.WriteInt32(responseLenPtr, responseData.Length);
                Console.WriteLine($"[Export] 批量响应已发送, 长度={responseData.Length}");
            }
            catch (Exception ex)
            {
                Console.WriteLine($"[Export] 异常: {ex}");
                var errorResponse = new BattleResponse
                {
                    Code = (int)BattleErrorCode.InternalError,
                    Message = $"处理异常: {ex.Message}",
                    Timestamp = DateTimeOffset.Now.ToUnixTimeMilliseconds(),
                }.ToByteArray();

                if (errorResponse.Length <= BATCH_RESPONSE_BUFFER_SIZE)
                {
                    Marshal.Copy(errorResponse, 0, responseBufferPtr, errorResponse.Length);
                    Marshal.WriteInt32(responseLenPtr, errorResponse.Length);
                }
            }
        }

        /// <summary>
        /// 注册 Go 提供的回调函数
        /// 
        /// 函数签名 (C 风格):
        /// void RegisterCallback(void (*callback)(const uint8_t* data, int32_t len));
        /// </summary>
        [UnmanagedCallersOnly(EntryPoint = "RegisterCallback")]
        public static void RegisterCallback(IntPtr callbackPtr)
        {
            try
            {
                Console.WriteLine($"[Export] RegisterCallback 被调用, CallbackPtr={callbackPtr}");
                BattleCallbackManager.RegisterCallback(callbackPtr);
            }
            catch (Exception ex)
            {
                Console.WriteLine($"[Export] 注册回调失败: {ex}");
            }
        }

        /// <summary>
        /// 测试回调 - 用于测试 Go 侧的回调是否正确工作
        /// Go 调用此函数，C# 会创建一个 BattleNotification 并通过回调发送给 Go
        /// </summary>
        [UnmanagedCallersOnly(EntryPoint = "TestNotifyCallback")]
        public static int TestNotifyCallback(int notificationType, long battleID, long timestamp)
        {
            try
            {
                Console.WriteLine($"[Export] TestNotifyCallback 被调用: Type={notificationType}, BattleID={battleID}, Timestamp={timestamp}");
                
                // 构建一个 BattleNotification
                var notification = new Battle.BattleNotification
                {
                    Timestamp = timestamp,
                    NotificationType = (Battle.NotificationType)notificationType,
                    BattleId = (uint)battleID,
                };

                // 序列化通知
                var notificationData = notification.ToByteArray();
                
                // 通过注册的回调发送给 Go
                Console.WriteLine($"[Export] 通过回调发送通知数据，长度={notificationData.Length}");
                BattleCallbackManager.NotifyBattle(notificationData);
                
                return 0; // 成功
            }
            catch (Exception ex)
            {
                Console.WriteLine($"[Export] TestNotifyCallback 异常: {ex}");
                return -1; // 失败
            }
        }

        /// <summary>
        /// 获取库版本
        /// </summary>
        [UnmanagedCallersOnly(EntryPoint = "GetLibVersion")]
        public static IntPtr GetLibVersion()
        {
            return Marshal.StringToHGlobalAnsi("goPureWithCsharp-1.0");
        }
    }
}
