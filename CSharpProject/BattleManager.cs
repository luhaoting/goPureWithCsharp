using System;
using System.Collections.Generic;
using System.Runtime.InteropServices;
using GoPureWithCsharp.Battle;

namespace GoPureWithCsharp
{
    /// <summary>
    /// 配置加载器回调委托 - Go 侧实现
    /// 参数:
    ///   - configNamePtr: 配置名称数据指针
    ///   - configNameLen: 配置名称长度（字节）
    ///   - outDataPtrPtr: [out] 指向配置数据指针的指针
    ///   - outDataLenPtr: [out] 指向配置数据长度的指针
    /// 返回值: 0 表示成功, -1 表示失败
    /// </summary>
    [UnmanagedFunctionPointer(CallingConvention.Cdecl)]
    public delegate int ConfigLoaderCallback(
        IntPtr configNamePtr,
        int configNameLen,
        IntPtr outDataPtrPtr,
        IntPtr outDataLenPtr);

    /// <summary>
    /// 战斗结果回调委托 - Go 侧实现
    /// 输出:
    ///   - outDataPtr: [out] BattleContext 数据指针
    ///   - outDataLen: [out] BattleContext 数据长度（字节）
    /// 返回值: 0 表示成功, -1 表示失败
    /// </summary>
    [UnmanagedFunctionPointer(CallingConvention.Cdecl)]
    public delegate int BattleResultCallback(
        out IntPtr outDataPtr,
        out int outDataLen);

    /// <summary>
    /// 战斗管理器 - 存储和管理所有战斗实例
    /// </summary>
    public static class BattleManager
    {
        private static readonly Dictionary<uint, BattleInstance> _battles = new Dictionary<uint, BattleInstance>();
        private static readonly object _lockObj = new object();
        private static BattleConfig? _config;

        /// <summary>
        /// 委托：由 Go 侧实现，用于获取配置数据
        /// </summary>
        private static ConfigLoaderCallback? _configLoader;

        /// <summary>
        /// 委托：由 Go 侧实现，用于接收战斗结果
        /// </summary>
        private static BattleResultCallback? _resultCallback;

        /// <summary>
        /// 注册配置加载器 (由 Go 调用)
        /// </summary>
        public static void RegisterConfigLoader(ConfigLoaderCallback configLoader)
        {
            lock (_lockObj)
            {
                _configLoader = configLoader;
                BattleLogger.Info("配置加载器已注册");
            }
        }

        /// <summary>
        /// 加载配置 (由 Go 调用)
        /// </summary>
        public static int LoadConfig(string configName)
        {
            lock (_lockObj)
            {
                if (_configLoader == null)
                {
                    BattleLogger.Error("配置加载器未注册");
                    return -1;
                }

                // 将配置名转换为字节数组
                byte[] configNameBytes = System.Text.Encoding.UTF8.GetBytes(configName);
                IntPtr namePtr = Marshal.AllocHGlobal(configNameBytes.Length);
                Marshal.Copy(configNameBytes, 0, namePtr, configNameBytes.Length);

                // 准备输出参数指针
                IntPtr outDataPtr = IntPtr.Zero;
                int outDataLen = 0;
                IntPtr outDataPtrPtr = Marshal.AllocHGlobal(IntPtr.Size);
                IntPtr outDataLenPtr = Marshal.AllocHGlobal(sizeof(int));

                // 调用 Go 侧加载器
                int result = _configLoader(namePtr, configNameBytes.Length, outDataPtrPtr, outDataLenPtr);

                // 从输出参数中读取数据指针和长度
                outDataPtr = Marshal.ReadIntPtr(outDataPtrPtr);
                outDataLen = Marshal.ReadInt32(outDataLenPtr);

                // 释放临时指针
                Marshal.FreeHGlobal(namePtr);
                Marshal.FreeHGlobal(outDataPtrPtr);
                Marshal.FreeHGlobal(outDataLenPtr);

                if (result != 0)
                {
                    BattleLogger.Error($"加载配置失败: {configName}");
                    return -1;
                }

                // 从返回的指针读取配置数据
                byte[] configData = new byte[outDataLen];
                if (outDataLen > 0)
                {
                    Marshal.Copy(outDataPtr, configData, 0, outDataLen);
                }

                // 保存配置
                _config = new BattleConfig
                {
                    Name = configName,
                    Data = configData
                };

                BattleLogger.Info($"配置已加载: {configName} ({outDataLen} 字节)");
                return 0;
            }
        }

        /// <summary>
        /// 调用已注册的配置加载器获取配置数据
        /// 这是 Go 侧 GetConfigLoaderDataCSharp 导出函数的实现支持
        /// </summary>
        public static int CallConfigLoader(string configName, out IntPtr outDataPtr, out int outDataLen)
        {
            outDataPtr = IntPtr.Zero;
            outDataLen = 0;

            lock (_lockObj)
            {
                if (_configLoader == null)
                {
                    BattleLogger.Error($"配置加载器未注册，无法加载: {configName}");
                    return -1;
                }

                // 将配置名转换为字节数组
                byte[] configNameBytes = System.Text.Encoding.UTF8.GetBytes(configName);
                IntPtr namePtr = Marshal.AllocHGlobal(configNameBytes.Length);
                Marshal.Copy(configNameBytes, 0, namePtr, configNameBytes.Length);

                // 准备输出参数指针
                IntPtr outDataPtrPtr = Marshal.AllocHGlobal(IntPtr.Size);
                IntPtr outDataLenPtr = Marshal.AllocHGlobal(sizeof(int));

                try
                {
                    // 调用 Go 侧加载器 - 这会触发 Go 侧的回调函数
                    int result = _configLoader(namePtr, configNameBytes.Length, outDataPtrPtr, outDataLenPtr);
                    
                    // 从输出参数中读取数据指针和长度
                    outDataPtr = Marshal.ReadIntPtr(outDataPtrPtr);
                    outDataLen = Marshal.ReadInt32(outDataLenPtr);
                    
                    if (result == 0)
                    {
                        BattleLogger.Info($"配置加载器返回数据: {configName} ({outDataLen} 字节)");
                    }
                    else
                    {
                        BattleLogger.Error($"配置加载器返回错误: {configName}, 错误码={result}");
                    }

                    return result;
                }
                finally
                {
                    // 释放临时指针
                    Marshal.FreeHGlobal(namePtr);
                    Marshal.FreeHGlobal(outDataPtrPtr);
                    Marshal.FreeHGlobal(outDataLenPtr);
                }
            }
        }

        /// <summary>
        /// 注册战斗结果回调 (由 Go 调用)
        /// </summary>
        public static void RegisterResultCallback(BattleResultCallback resultCallback)
        {
            lock (_lockObj)
            {
                _resultCallback = resultCallback;
                BattleLogger.Info("战斗结果回调已注册");
            }
        }

        /// <summary>
        /// 创建战斗 (由 Go 调用)
        /// </summary>
        public static int CreateBattle(uint battleId, uint atkTeamId, uint defTeamId)
        {
            lock (_lockObj)
            {
                if (_battles.ContainsKey(battleId))
                {
                    BattleLogger.Error($"战斗 ID={battleId} 已存在");
                    return -1;
                }

                // 从配置获取初始血量
                int initialHealth = 300; // 默认值
                if (_config != null)
                {
                    // 可以从配置中解析初始血量
                }

                BattleInstance battle = new BattleInstance(battleId, atkTeamId, defTeamId, initialHealth);
                _battles[battleId] = battle;

                BattleLogger.Info($"战斗已创建: ID={battleId}, ATK={atkTeamId}, DEF={defTeamId}");
                return 0; // 成功
            }
        }

        /// <summary>
        /// 销毁战斗 (由 Go 调用)
        /// </summary>
        public static int DestroyBattle(uint battleId)
        {
            lock (_lockObj)
            {
                if (!_battles.ContainsKey(battleId))
                {
                    BattleLogger.Error($"战斗 ID={battleId} 不存在");
                    return -1;
                }

                _battles.Remove(battleId);
                BattleLogger.Info($"战斗已销毁: ID={battleId}");
                return 0; // 成功
            }
        }

        /// <summary>
        /// Tick 驱动 - 执行所有进行中的战斗 (由 Go 调用)
        /// </summary>
        public static int OnTick()
        {
            lock (_lockObj)
            {
                int battleCount = 0;
                List<uint> finishedBattles = new List<uint>();

                foreach (var kvp in _battles)
                {
                    uint battleId = kvp.Key;
                    BattleInstance battle = kvp.Value;

                    if (!battle.IsFinished)
                    {
                        battle.ExecuteRound(20, 50); // minDamage=20, maxDamage=50

                        if (battle.IsFinished)
                        {
                            finishedBattles.Add(battleId);

                            // 通知 Go 战斗结果并获取 BattleContext 输出
                            if (_resultCallback != null)
                            {
                                int result = _resultCallback(out IntPtr dataPtr, out int dataLen);
                                if (result == 0 && dataPtr != IntPtr.Zero)
                                {
                                    BattleLogger.Debug($"战斗结果已处理: ID={battleId}, 输出长度={dataLen} 字节");
                                }
                            }
                        }

                        battleCount++;
                    }
                }

                if (battleCount > 0)
                {
                    BattleLogger.Debug($"Tick: 处理 {battleCount} 场战斗, 完成 {finishedBattles.Count} 场");
                }

                return battleCount; // 返回处理的战斗数
            }
        }

        /// <summary>
        /// 获取战斗状态 (内部使用)
        /// </summary>
        public static BattleInstance? GetBattle(uint battleId)
        {
            lock (_lockObj)
            {
                _battles.TryGetValue(battleId, out var battle);
                return battle;
            }
        }

        /// <summary>
        /// 获取所有战斗数量
        /// </summary>
        public static int GetBattleCount()
        {
            lock (_lockObj)
            {
                return _battles.Count;
            }
        }
    }
}
