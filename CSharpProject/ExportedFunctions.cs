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
        [UnmanagedCallersOnly(CallConvs = new[] { typeof(System.Runtime.CompilerServices.CallConvCdecl) }, EntryPoint = "ProcessProtoMessage")]
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
        [UnmanagedCallersOnly(CallConvs = new[] { typeof(System.Runtime.CompilerServices.CallConvCdecl) }, EntryPoint = "ProcessBatchProtoMessage")]
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
        [UnmanagedCallersOnly(CallConvs = new[] { typeof(System.Runtime.CompilerServices.CallConvCdecl) }, EntryPoint = "RegisterCallback")]
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
        [UnmanagedCallersOnly(CallConvs = new[] { typeof(System.Runtime.CompilerServices.CallConvCdecl) }, EntryPoint = "TestNotifyCallback")]
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
        [UnmanagedCallersOnly(CallConvs = new[] { typeof(System.Runtime.CompilerServices.CallConvCdecl) }, EntryPoint = "GetLibVersion")]
        public static IntPtr GetLibVersion()
        {
            return Marshal.StringToHGlobalAnsi("goPureWithCsharp-1.0");
        }

        /// <summary>
        /// 测试触发回调 - C# 侧主动调用，测试 Go 侧回调是否工作
        /// 参数: battleId, notificationType, timestamp
        /// 返回: 0 成功, 负数表示错误
        /// </summary>
        [UnmanagedCallersOnly(CallConvs = new[] { typeof(System.Runtime.CompilerServices.CallConvCdecl) }, EntryPoint = "TestTriggerCallback")]
        public static int TestTriggerCallback(uint battleId, int notificationType, long timestamp)
        {
            try
            {
                // 构建一个 BattleNotification
                var notification = new Battle.BattleNotification
                {
                    Timestamp = timestamp,
                    NotificationType = (Battle.NotificationType)notificationType,
                    BattleId = battleId,
                };

                // 序列化通知
                var notificationData = notification.ToByteArray();

                // 方式1：尝试通过 BattleCallbackManager 发送（如果已注册回调）
                BattleCallbackManager.NotifyBattle(notificationData);

                return 0; // 成功
            }
            catch (Exception ex)
            {
                System.Console.WriteLine($"[Export] TestTriggerCallback 异常: {ex}");
                return -1; // 失败
            }
        }

        /// <summary>
        /// 处理通知数据 - 供测试用，让 C# 能够直接触发 Go 侧的通知处理
        /// 参数: notificationData (指向序列化通知数据的指针), dataLength (数据长度)
        /// 返回: 0 成功, 负数表示错误
        /// </summary>
        [UnmanagedCallersOnly(CallConvs = new[] { typeof(System.Runtime.CompilerServices.CallConvCdecl) }, EntryPoint = "ProcessNotificationFromCSharp")]
        public unsafe static int ProcessNotificationFromCSharp(byte* notificationData, int dataLength)
        {
            if (notificationData == null || dataLength <= 0)
            {
                return -1;
            }

            try
            {
                // 复制数据到托管内存
                byte[] data = new byte[dataLength];
                fixed (byte* ptr = data)
                {
                    System.Buffer.MemoryCopy(notificationData, ptr, dataLength, dataLength);
                }

                // 通过 BattleCallbackManager 发送给 Go
                BattleCallbackManager.NotifyBattle(data);

                return 0; // 成功
            }
            catch (Exception ex)
            {
                System.Console.WriteLine($"[Export] ProcessNotificationFromCSharp 异常: {ex}");
                return -1; // 失败
            }
        }

        // ============================================================================
        // BattleManager 导出函数
        // ============================================================================

        /// <summary>
        /// 加载配置 (由 Go 调用)
        /// 参数: configLoaderPtr - 指向配置加载器回调函数的指针
        /// </summary>
        /// <summary>
        /// 注册配置加载器 (由 Go 调用)
        /// </summary>
        [UnmanagedCallersOnly(CallConvs = new[] { typeof(System.Runtime.CompilerServices.CallConvCdecl) }, EntryPoint = "RegisterConfigLoader")]
        public static void RegisterConfigLoader(IntPtr configLoaderPtr)
        {
            System.Console.WriteLine("[Export-RC] 1. 进入 RegisterConfigLoader");
            
            if (configLoaderPtr == IntPtr.Zero)
            {
                System.Console.WriteLine("[Export-RC] 2. 参数为 NULL，返回");
                return;
            }

            System.Console.WriteLine("[Export-RC] 3. 参数非 NULL");
            System.Console.WriteLine("[Export-RC] 4. 正在转换委托");
            var configLoader = Marshal.GetDelegateForFunctionPointer<ConfigLoaderCallback>(configLoaderPtr);
            System.Console.WriteLine("[Export-RC] 5. 委托转换成功");
            System.Console.WriteLine("[Export-RC] 6. 调用 BattleManager.RegisterConfigLoader");
            BattleManager.RegisterConfigLoader(configLoader);
            System.Console.WriteLine("[Export-RC] 7. 完成，返回");
        }

        /// <summary>
        /// 加载配置 (由 Go 调用)
        /// 参数: configNamePtr - 配置名称字节数据指针, configNameLen - 名称长度
        /// </summary>
        [UnmanagedCallersOnly(CallConvs = new[] { typeof(System.Runtime.CompilerServices.CallConvCdecl) }, EntryPoint = "LoadConfig")]
        public static int LoadConfig(IntPtr configNamePtr, int configNameLen)
        {
            // 从指针读取配置名称字节
            byte[] configNameBytes = new byte[configNameLen];
            Marshal.Copy(configNamePtr, configNameBytes, 0, configNameLen);
            string configName = System.Text.Encoding.UTF8.GetString(configNameBytes);

            return BattleManager.LoadConfig(configName);
        }

        /// <summary>
        /// Go 调用 C# 来获取配置数据 (双向函数调用)
        /// 这个函数在 C# 侧存储了一个配置加载器回调，当被 Go 调用时，
        /// 它会触发那个回调函数，由 Go 侧实现的配置加载器来加载配置文件
        /// 
        /// 流程：
        /// 1. Go 侧调用此函数：GetConfigLoaderDataCSharp("battle_config.json")
        /// 2. C# 调用已注册的回调函数（由 Go 侧实现）
        /// 3. Go 侧回调函数加载文件并返回数据
        /// 4. C# 将数据返回给 Go
        /// 
        /// 参数: configNamePtr - 配置文件名指针, configNameLen - 名称长度
        ///       outDataPtrPtr - 指向数据指针的指针, outDataLenPtr - 指向数据长度的指针
        /// 返回: 0 成功, -1 失败
        /// </summary>
        [UnmanagedCallersOnly(CallConvs = new[] { typeof(System.Runtime.CompilerServices.CallConvCdecl) }, EntryPoint = "GetConfigLoaderDataCSharp")]
        public static int GetConfigLoaderDataCSharp(IntPtr configNamePtr, int configNameLen, IntPtr outDataPtrPtr, IntPtr outDataLenPtr)
        {
            try
            {
                // 从指针读取配置文件名
                byte[] configNameBytes = new byte[configNameLen];
                Marshal.Copy(configNamePtr, configNameBytes, 0, configNameLen);
                string configName = System.Text.Encoding.UTF8.GetString(configNameBytes);

                System.Console.WriteLine($"[Export] GetConfigLoaderDataCSharp 被调用: {configName}");

                // 调用 BattleManager 中已注册的配置加载器
                // 该回调由 Go 侧提供（通过 RegisterConfigLoader 导出函数）
                int result = BattleManager.CallConfigLoader(configName, out IntPtr dataPtr, out int dataLen);
                
                // 将结果写入输出指针
                if (outDataPtrPtr != IntPtr.Zero)
                {
                    Marshal.WriteIntPtr(outDataPtrPtr, dataPtr);
                }
                if (outDataLenPtr != IntPtr.Zero)
                {
                    Marshal.WriteInt32(outDataLenPtr, dataLen);
                }
                
                return result;
            }
            catch (Exception ex)
            {
                System.Console.WriteLine($"[Export] GetConfigLoaderDataCSharp 异常: {ex}");
                return -1; // 失败
            }
        }

        /// <summary>
        /// 注册战斗结果回调 (由 Go 调用)
        /// 参数: callbackPtr - 指向结果回调函数的指针
        /// </summary>
        [UnmanagedCallersOnly(CallConvs = new[] { typeof(System.Runtime.CompilerServices.CallConvCdecl) }, EntryPoint = "RegisterBattleResultCallback")]
        public static int RegisterBattleResultCallback(IntPtr callbackPtr)
        {
            if (callbackPtr == IntPtr.Zero)
            {
                return -1;
            }

            System.Console.WriteLine($"[Export] RegisterBattleResultCallback 被调用 地址 0x{callbackPtr:X}");
            var resultCallback = Marshal.GetDelegateForFunctionPointer<BattleResultCallback>(callbackPtr);
            BattleManager.RegisterResultCallback(resultCallback);
            return 0;
        }

        /// <summary>
        /// 创建战斗 (由 Go 调用)
        /// 参数: battleId, atkTeamId, defTeamId
        /// 返回: 0 成功, -1 失败
        /// </summary>
        [UnmanagedCallersOnly(CallConvs = new[] { typeof(System.Runtime.CompilerServices.CallConvCdecl) }, EntryPoint = "CreateBattle")]
        public static int CreateBattle(uint battleId, uint atkTeamId, uint defTeamId)
        {
            return BattleManager.CreateBattlee(battleId, atkTeamId, defTeamId);
        }

        [UnmanagedCallersOnly(CallConvs = new[] { typeof(System.Runtime.CompilerServices.CallConvCdecl) }, EntryPoint = "CreateBattleByCtx")]
        public static int CreateBattleByCtx(uint battleId, uint atkTeamId, uint defTeamId)
        {
            return BattleManager.CreateBattlee(battleId, atkTeamId, defTeamId);
        }

        /// <summary>
        /// 销毁战斗 (由 Go 调用)
        /// 参数: battleId
        /// 返回: 0 成功, -1 失败
        /// </summary>
        [UnmanagedCallersOnly(CallConvs = new[] { typeof(System.Runtime.CompilerServices.CallConvCdecl) }, EntryPoint = "DestroyBattle")]
        public static int DestroyBattle(uint battleId)
        {
            return BattleManager.DestroyBattle(battleId);
        }

        /// <summary>
        /// Tick 驱动 - 推动所有战斗进行 (由 Go 调用)
        /// 返回: 处理的战斗数量
        /// </summary>
        [UnmanagedCallersOnly(CallConvs = new[] { typeof(System.Runtime.CompilerServices.CallConvCdecl) }, EntryPoint = "OnTick")]
        public static int OnTick()
        {
            return BattleManager.OnTick();
        }

        /// <summary>
        /// 获取战斗数量
        /// 返回: 当前管理的战斗数量
        /// </summary>
        [UnmanagedCallersOnly(CallConvs = new[] { typeof(System.Runtime.CompilerServices.CallConvCdecl) }, EntryPoint = "GetBattleCount")]
        public static int GetBattleCount()
        {
            return BattleManager.GetBattleCount();
        }

        /// <summary>
        /// 处理战斗输入 (由 Go 调用)
        /// 参数: battleId, teamId, actionType, actionValue
        /// 返回: 0 成功, 负数表示错误码
        /// </summary>
        [UnmanagedCallersOnly(CallConvs = new[] { typeof(System.Runtime.CompilerServices.CallConvCdecl) }, EntryPoint = "ProcessBattleInput")]
        public static int ProcessBattleInput(uint battleId, uint teamId, byte actionType, int actionValue)
        {
            return BattleInputHandler.ProcessBattleInput(battleId, teamId, actionType, actionValue);
        }

        [UnmanagedCallersOnly(CallConvs = new[] { typeof(System.Runtime.CompilerServices.CallConvCdecl) }, EntryPoint = "ProcessBattleContextInput")]
        public static int ProcessBattleContextInput(IntPtr buffPtr, int buffLen)
        {

// 从指针读取字符串
            byte[] inputBytes = new byte[buffLen];
            Marshal.Copy(buffPtr, inputBytes, 0, buffLen);

            var battleInputContext = Battle.BattleContext.Parser.ParseFrom(inputBytes);
            if (battleInputContext == null || battleInputContext.OptionCase != Battle.BattleContext.OptionOneofCase.BattleInput)
            {
                return -1; // 无效输入
            }

            return BattleManager.ProcessBattleContextInput(battleInputContext);
        }

        /// <summary>
        /// 设置战斗日志级别 (由 Go 调用)
        /// 参数: level - 日志级别 (0=Debug, 1=Info, 2=Warn, 3=Error, 4=None)
        /// 返回: 0 成功, -1 失败
        /// </summary>
        [UnmanagedCallersOnly(CallConvs = new[] { typeof(System.Runtime.CompilerServices.CallConvCdecl) }, EntryPoint = "SetBattleLogLevel")]
        public static int SetBattleLogLevel(int level)
        {
            if (level < 0 || level > 4)
            {
                return -1;
            }

            BattleLogger.SetLogLevel((LogLevel)level);
            return 0;
        }

        /// <summary>
        /// 获取当前战斗日志级别 (由 Go 调用)
        /// 返回: 当前日志级别 (0=Debug, 1=Info, 2=Warn, 3=Error, 4=None)
        /// </summary>
        [UnmanagedCallersOnly(CallConvs = new[] { typeof(System.Runtime.CompilerServices.CallConvCdecl) }, EntryPoint = "GetBattleLogLevel")]
        public static int GetBattleLogLevel()
        {
            return (int)BattleLogger.GetLogLevel();
        }

        /// <summary>
        /// 启用所有战斗日志 (由 Go 调用)
        /// </summary>
        [UnmanagedCallersOnly(CallConvs = new[] { typeof(System.Runtime.CompilerServices.CallConvCdecl) }, EntryPoint = "EnableBattleLogging")]
        public static void EnableBattleLogging()
        {
            BattleLogger.EnableAll();
        }

        /// <summary>
        /// 禁用所有战斗日志 (由 Go 调用)
        /// </summary>
        [UnmanagedCallersOnly(CallConvs = new[] { typeof(System.Runtime.CompilerServices.CallConvCdecl) }, EntryPoint = "DisableBattleLogging")]
        public static void DisableBattleLogging()
        {
            BattleLogger.DisableAll();
        }

        // ============================================================================
        // Go 全局函数调用接口
        // ============================================================================

        /// <summary>
        /// 调用 Go 侧的全局函数处理战斗通知
        /// 参数: battleId, notificationType, timestamp
        /// 返回: 0 成功, -1 失败
        /// </summary>
        [UnmanagedCallersOnly(CallConvs = new[] { typeof(System.Runtime.CompilerServices.CallConvCdecl) }, EntryPoint = "CallGoGlobalHandleBattleNotification")]
        public static int CallGoGlobalHandleBattleNotification(uint battleId, int notificationType, long timestamp)
        {
            try
            {
                Console.WriteLine($"[Export] CallGoGlobalHandleBattleNotification 被调用: BattleID={battleId}, Type={notificationType}, Timestamp={timestamp}");

                // 构建通知对象
                var notification = new Battle.BattleNotification
                {
                    BattleId = battleId,
                    NotificationType = (Battle.NotificationType)notificationType,
                    Timestamp = timestamp,
                };

                // 序列化并调用 Go 侧的处理函数
                var notificationData = notification.ToByteArray();
                BattleCallbackManager.NotifyBattle(notificationData);

                Console.WriteLine($"[Export] 已将通知转发给 Go 侧的全局函数");
                return 0; // 成功
            }
            catch (Exception ex)
            {
                Console.WriteLine($"[Export] CallGoGlobalHandleBattleNotification 异常: {ex}");
                return -1; // 失败
            }
        }

        /// <summary>
        /// 调用 Go 侧的简单全局函数 - 用于测试 C# 直接调用 Go 函数
        /// 参数: battleId (uint), action (字符串)
        /// 返回: 0 成功, -1 失败
        /// </summary>
        [UnmanagedCallersOnly(CallConvs = new[] { typeof(System.Runtime.CompilerServices.CallConvCdecl) }, EntryPoint = "CallGoSimpleGlobalFunction")]
        public static int CallGoSimpleGlobalFunction(uint battleId, IntPtr actionPtr, int actionLen)
        {
            try
            {
                // 从指针读取字符串
                byte[] actionBytes = new byte[actionLen];
                Marshal.Copy(actionPtr, actionBytes, 0, actionLen);
                string action = System.Text.Encoding.UTF8.GetString(actionBytes);

                Console.WriteLine($"[Export] CallGoSimpleGlobalFunction 被调用: BattleID={battleId}, Action={action}");

                // 这里可以做一些处理
                // 注意：实际的 Go 全局函数无法直接从 C# 调用
                // 但我们可以通过 BattleCallbackManager 或其他方式转发请求

                Console.WriteLine($"[Export] 处理来自 Go 的全局函数调用");
                return 0; // 成功
            }
            catch (Exception ex)
            {
                Console.WriteLine($"[Export] CallGoSimpleGlobalFunction 异常: {ex}");
                return -1; // 失败
            }
        }

        /// <summary>
        /// 调用 Go 侧的计算函数 - 用于演示函数指针调用
        /// 参数: a, b (两个 int32)
        /// 返回: 结果
        /// </summary>
        [UnmanagedCallersOnly(CallConvs = new[] { typeof(System.Runtime.CompilerServices.CallConvCdecl) }, EntryPoint = "CallGoCalculateSum")]
        public static int CallGoCalculateSum(int a, int b)
        {
            try
            {
                Console.WriteLine($"[Export] CallGoCalculateSum 被调用: {a} + {b}");
                int result = a + b;
                Console.WriteLine($"[Export] 计算结果: {result}");
                return result;
            }
            catch (Exception ex)
            {
                Console.WriteLine($"[Export] CallGoCalculateSum 异常: {ex}");
                return -1; // 失败
            }
        }
    }
}

