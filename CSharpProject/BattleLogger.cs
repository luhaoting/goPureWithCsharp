using System;

namespace GoPureWithCsharp
{
    /// <summary>
    /// 日志级别
    /// </summary>
    public enum LogLevel
    {
        Debug = 0,
        Info = 1,
        Warn = 2,
        Error = 3,
        None = 4  // 禁用所有日志
    }

    /// <summary>
    /// 战斗日志管理器 - 控制所有战斗相关的打印输出
    /// </summary>
    public static class BattleLogger
    {
        private static LogLevel _currentLevel = LogLevel.Info;

        /// <summary>
        /// 设置日志级别
        /// </summary>
        public static void SetLogLevel(LogLevel level)
        {
            _currentLevel = level;
        }

        /// <summary>
        /// 获取当前日志级别
        /// </summary>
        public static LogLevel GetLogLevel()
        {
            return _currentLevel;
        }

        /// <summary>
        /// 启用所有日志
        /// </summary>
        public static void EnableAll()
        {
            _currentLevel = LogLevel.Debug;
        }

        /// <summary>
        /// 禁用所有日志
        /// </summary>
        public static void DisableAll()
        {
            _currentLevel = LogLevel.None;
        }

        /// <summary>
        /// Debug 级别日志
        /// </summary>
        public static void Debug(string message)
        {
            if (_currentLevel <= LogLevel.Debug)
            {
                Console.WriteLine($"C#[DEBUG] {message}");
            }
        }

        /// <summary>
        /// Info 级别日志
        /// </summary>
        public static void Info(string message)
        {
            if (_currentLevel <= LogLevel.Info)
            {
                Console.WriteLine($"C#[INFO] {message}");
            }
        }

        /// <summary>
        /// Warn 级别日志
        /// </summary>
        public static void Warn(string message)
        {
            if (_currentLevel <= LogLevel.Warn)
            {
                Console.WriteLine($"C#[WARN] {message}");
            }
        }

        /// <summary>
        /// Error 级别日志
        /// </summary>
        public static void Error(string message)
        {
            if (_currentLevel <= LogLevel.Error)
            {
                Console.WriteLine($"C#[ERROR] {message}");
            }
        }
    }
}
