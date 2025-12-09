using System;
using System.Runtime.InteropServices;

namespace GoPureWithCsharp
{
    /// <summary>
    /// C# -> Go 回调委托定义
    /// 用于 C# 通知 Go 战斗事件
    /// </summary>
    [UnmanagedFunctionPointer(CallingConvention.Cdecl)]
    public delegate void BattleNotifyCallback(IntPtr notificationData, int notificationLen);

    /// <summary>
    /// 回调管理器 - 管理 C# 和 Go 之间的事件通知
    /// </summary>
    public static class BattleCallbackManager
    {
        private static BattleNotifyCallback? _battleNotifyCallback;
        private static readonly object _lockObj = new object();

        /// <summary>
        /// 注册 Go 提供的回调函数
        /// </summary>
        /// <param name="callbackPtr">Go 函数指针</param>
        public static void RegisterCallback(IntPtr callbackPtr)
        {
            lock (_lockObj)
            {
                if (callbackPtr == IntPtr.Zero)
                {
                    _battleNotifyCallback = null;
                    Console.WriteLine("[BattleCallback] 回调已注销");
                    return;
                }

                _battleNotifyCallback = Marshal.GetDelegateForFunctionPointer<BattleNotifyCallback>(callbackPtr);
                Console.WriteLine("[BattleCallback] Go 回调已注册");
            }
        }

        /// <summary>
        /// 触发战斗通知（发送给 Go）
        /// </summary>
        /// <param name="notificationData">通知数据字节数组</param>
        public static void NotifyBattle(byte[] notificationData)
        {
            lock (_lockObj)
            {
                if (_battleNotifyCallback == null)
                {
                    Console.WriteLine("[BattleCallback] 没有注册回调函数，忽略通知");
                    return;
                }

                try
                {
                    // 分配非托管内存存储数据
                    IntPtr dataPtr = Marshal.AllocHGlobal(notificationData.Length);
                    Marshal.Copy(notificationData, 0, dataPtr, notificationData.Length);

                    // 调用 Go 回调
                    _battleNotifyCallback(dataPtr, notificationData.Length);

                    // 释放内存
                    Marshal.FreeHGlobal(dataPtr);
                }
                catch (Exception ex)
                {
                    Console.WriteLine($"[BattleCallback] 调用回调时出错: {ex.Message}");
                }
            }
        }

        /// <summary>
        /// 是否已注册回调
        /// </summary>
        public static bool IsCallbackRegistered
        {
            get
            {
                lock (_lockObj)
                {
                    return _battleNotifyCallback != null;
                }
            }
        }
    }
}
