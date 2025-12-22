using GoPureWithCsharp;
using System.Runtime.InteropServices;


class Battle
{
    int ConfigValue = 42;
    public void Start()
    {
        Console.WriteLine("Battle started!");
    }

    public void input()
    {
        Console.WriteLine("Input received!");
    }

    public void End()
    {
        Console.WriteLine("Battle ended!");
    }

    public void ResultReplay()
    {
        Console.WriteLine("Battle ended!");
    }
}

public class BattleDemo
{
    [UnmanagedCallersOnly(CallConvs = new[] { typeof(System.Runtime.CompilerServices.CallConvCdecl) }, EntryPoint = "PrintCS")]
    public static int PrintCS(int value)
    {
        Console.WriteLine($"C# Output: {value}");
        return value * 2;
    }

    [UnmanagedCallersOnly(CallConvs = new[] { typeof(System.Runtime.CompilerServices.CallConvCdecl) }, EntryPoint = "CsharpPanic")]
    public static int CsharpPanic()
    {
        return NativeAOTExceptionInjector.WrapExportFunction(() => { 
            // 无法捕获段错误 基于系统的原生崩溃
            IntPtr nullPtr = IntPtr.Zero;
            Marshal.ReadInt32(nullPtr); // 崩溃点
            Console.WriteLine($"C# Output: Triggering IndexOutOfRangeException");
            int[] arr = new int[3];
            return arr[5]; // 托管数组越界 → 触发 IndexOutOfRangeException
        });
    }
   
}

