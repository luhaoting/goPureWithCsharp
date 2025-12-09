#include <dlfcn.h>
#include <iostream>

using PrintCSFunc = int (*)(int);

int main()
{
	const char* soPath = "./TestExport.so"; // 或者相对/绝对路径，例如 "./bin/Release/net8.0/linux-x64/publish/TestExport.so"

	void* handle = dlopen(soPath, RTLD_NOW);
	if (!handle)
	{
		std::cerr << "dlopen failed: " << dlerror() << std::endl;
		return 1;
	}

	dlerror(); // 清空之前的错误

	// C# 端通过 [UnmanagedCallersOnly(EntryPoint = "PrintCS")] 导出，符号名就是 "PrintCS"
	void* sym = dlsym(handle, "PrintCS");
	const char* err = dlerror();
	if (err != nullptr)
	{
		std::cerr << "dlsym failed: " << err << std::endl;
		//dlclose(handle);
		return 1;
	}

	auto printCs = reinterpret_cast<PrintCSFunc>(sym);

	// 调用 C# 导出的函数
	int result = printCs(123);
	std::cout << "Result from C#: " << result << std::endl;

	result = printCs(76);
	std::cout << "Result from C#: " << result << std::endl;

	//dlclose(handle);
	return 0;
}
