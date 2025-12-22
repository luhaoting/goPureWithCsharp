using GoPureWithCsharp;
using System;
using System.Runtime.InteropServices;
using System.Text;


namespace GoPureWithCsharp
{

    // 异常处理工具类：适配Go侧调用的统一逻辑

    [UnmanagedFunctionPointer(CallingConvention.Cdecl)]
    public delegate void ExNotifyCallback();

    public static class NativeAOTExceptionInjector
    {
        // 错误码枚举（统一管理）
        public enum NativeErrorCode
        {
            Success = 0,
            ParameterError = -1,    // 参数异常
            BusinessError = -2,     // 业务异常
            SystemError = -3,       // 系统异常
            UnknownError = -999     // 未知异常
        }

        /// <summary>
        /// Go侧注入的错误栈写入区指针（缓冲区首地址）
        /// </summary>
        private static IntPtr exBufferPtr;

        /// <summary>
        /// Go侧注入的缓冲区大小（字节）
        /// </summary>
        private static Int32 exBufferSize;


        private static ExNotifyCallback? exNotifyCallback;

        [UnmanagedCallersOnly(CallConvs = new[] { typeof(System.Runtime.CompilerServices.CallConvCdecl) }, EntryPoint = "InjectedExceptionContext")]
        public static void InjectedExceptionContext(IntPtr callbackPtr, IntPtr goBuffPtr, Int32 BufferSize )
        {
            Console.WriteLine($"C# [InjectedExceptionBuff] Go 开始注入的异常信息写入栈");
            NativeAOTExceptionInjector.exBufferPtr = goBuffPtr;
            NativeAOTExceptionInjector.exBufferSize = BufferSize;
            if (callbackPtr == IntPtr.Zero)
            {
                exNotifyCallback = null;
                Console.WriteLine("[exNotifyCallback] 回调已注销");
                return;
            }

            exNotifyCallback = Marshal.GetDelegateForFunctionPointer<ExNotifyCallback>(callbackPtr);


            Console.WriteLine($"C# [InjectedExceptionBuff] Go 注入的异常信息写入栈 地址0x{exBufferPtr:X} 长度 {BufferSize}  通知地址0x{exNotifyCallback:X}");
        }

        // ========== 核心：捕获异常后通知Go ==========
        // 所有导出函数的统一包装（捕获异常+通知Go）
        public static T WrapExportFunction<T>(Func<T> func)
        {
            try
            {
                Console.WriteLine($"C# WrapExportFunction Called");
                return func();
            }
            catch (Exception ex)
            {
                Console.WriteLine($"[NativeAOTExceptionInjector] 捕获异常: {ex}");
                // 1. 处理异常并通知Go
                NativeAOTExceptionInjector.HandleException(ex);
                // 2. 返回默认值（适配FFI调用）
                // 显式处理引用类型的 default 返回
                if (default(T) is null && typeof(T).IsClass)
                {
                    return (T)(object)null!;
                }
                return default!;
            }
        }

        /// <summary>
        /// 统一处理异常，写入Go侧缓冲区并返回错误码
        /// </summary>
        /// <param name="ex">捕获的异常</param>
        /// <param name="errorMsgBuffer">Go侧传入的缓冲区指针</param>
        /// <param name="bufferSize">缓冲区大小</param>
        /// <returns>统一错误码</returns>
        public static NativeErrorCode HandleException(Exception ex)
        {


            // 1. 构建结构化的错误信息（JSON格式，方便Go侧解析）
            var errorBuilder = new StringBuilder();
            errorBuilder.Append("{");

            // 根据异常类型分类处理
            switch (ex)
            {
                //case ParameterValidationException paramEx:
                //    errorBuilder.Append($"\"type\":\"ParameterValidationException\",")
                //                .Append($"\"errorCode\":{paramEx.ErrorCode},")
                //                .Append($"\"parameterName\":\"{paramEx.ParameterName}\",")
                //                .Append($"\"message\":\"{EscapeJson(paramEx.Message)}\",")
                //                .Append($"\"stackTrace\":\"{EscapeJson(paramEx.StackTrace)}\"");
                //    break;

                //case BusinessRuleException businessEx:
                //    errorBuilder.Append($"\"type\":\"BusinessRuleException\",")
                //                .Append($"\"businessCode\":{businessEx.BusinessCode},")
                //                .Append($"\"message\":\"{EscapeJson(businessEx.Message)}\"");
                //    break;

                default:
                    // 系统异常（如NullReferenceException）
                    errorBuilder.Append($"\"type\":\"{ex.GetType().Name}\",")
                                .Append($"\"message\":\"{EscapeJson(ex.Message ?? string.Empty)}\",")
                                .Append($"\"stackTrace\":\"{EscapeJson(ex.StackTrace ?? string.Empty)}\"");
                    break;
            }

            errorBuilder.Append("}");
            var errorJson = errorBuilder.ToString();

            // 2. 将JSON写入Go侧缓冲区（注意缓冲区大小限制）
            var errorBytes = Encoding.UTF8.GetBytes(errorJson);
            var copyLength = Math.Min(errorBytes.Length, exBufferSize - 1);
            Marshal.Copy(errorBytes, 0, exBufferPtr, copyLength);
            Marshal.WriteByte(exBufferPtr, copyLength, 0);
            exNotifyCallback?.Invoke();
            // 3. 返回统一错误码（供Go侧快速判断）
            return ex switch
            {
                //ParameterValidationException => NativeErrorCode.ParameterError,
                //BusinessRuleException => NativeErrorCode.BusinessError,
                _ => NativeErrorCode.SystemError
            };
        }

        // 辅助方法：JSON字符串转义（避免双引号、换行符等破坏JSON格式）
        private static string EscapeJson(string input)
        {
            if (string.IsNullOrEmpty(input)) return "";
            return input.Replace("\"", "\\\"")
                        .Replace("\n", "\\n")
                        .Replace("\r", "\\r");
        }
    }
}