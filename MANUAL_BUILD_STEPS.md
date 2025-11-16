# 手动编译步骤（WebView2 版本）

## 🔧 问题分析

你遇到的错误：
```
fatal error: EventToken.h: No such file or directory
```

**原因**：`EventToken.h` 在 `winrt` 目录，你只设置了 `um` 和 `shared`。

---

## ✅ 完整解决方案

### 方法 1: 使用自动脚本（推荐）

在 **CMD**（不是 PowerShell）中运行：

```cmd
cd C:\Users\AHS\Documents\code\dy\dy-live-record
.\FIX_WEBVIEW2_BUILD.bat
```

**这个脚本会**：
- ✅ 自动设置所有必需的路径（包括 winrt）
- ✅ 设置正确的 CGO 编译选项
- ✅ 验证关键文件存在
- ✅ 清理缓存并重新编译

---

### 方法 2: 手动设置（完整版）

在 **CMD** 中，**按顺序执行**：

```cmd
REM 1. 设置 SDK 路径
set SDK_BASE=C:\Program Files (x86)\Windows Kits\10
set SDK_VERSION=10.0.26100.0

REM 2. 设置 INCLUDE（注意：必须包含 winrt）
set "INCLUDE=%INCLUDE%;%SDK_BASE%\Include\%SDK_VERSION%\um"
set "INCLUDE=%INCLUDE%;%SDK_BASE%\Include\%SDK_VERSION%\shared"
set "INCLUDE=%INCLUDE%;%SDK_BASE%\Include\%SDK_VERSION%\winrt"
set "INCLUDE=%INCLUDE%;%SDK_BASE%\Include\%SDK_VERSION%\ucrt"
set "INCLUDE=%INCLUDE%;%SDK_BASE%\Include\%SDK_VERSION%\cppwinrt"

REM 3. 设置 LIB
set "LIB=%LIB%;%SDK_BASE%\Lib\%SDK_VERSION%\um\x64"
set "LIB=%LIB%;%SDK_BASE%\Lib\%SDK_VERSION%\ucrt\x64"

REM 4. 设置 CGO（关键步骤！）
set CGO_ENABLED=1
set "CGO_CFLAGS=-I%SDK_BASE%\Include\%SDK_VERSION%\um -I%SDK_BASE%\Include\%SDK_VERSION%\shared -I%SDK_BASE%\Include\%SDK_VERSION%\winrt -I%SDK_BASE%\Include\%SDK_VERSION%\ucrt"
set "CGO_LDFLAGS=-L%SDK_BASE%\Lib\%SDK_VERSION%\um\x64 -L%SDK_BASE%\Lib\%SDK_VERSION%\ucrt\x64"

REM 5. 验证 EventToken.h 存在
dir "%SDK_BASE%\Include\%SDK_VERSION%\winrt\EventToken.h"

REM 6. 编译
cd server-go
go clean -cache
go build -v -o dy-live-monitor.exe .
```

---

### 方法 3: 使用 Visual Studio 命令提示符（最稳定）

#### 步骤 1: 打开 VS Developer Command Prompt

```
开始菜单 → Visual Studio 2022 → Developer Command Prompt for VS 2022
```

或者安装 **Build Tools for Visual Studio 2022**：
- https://visualstudio.microsoft.com/downloads/#build-tools-for-visual-studio-2022

#### 步骤 2: 编译

```cmd
cd C:\Users\AHS\Documents\code\dy\dy-live-record\server-go
set CGO_ENABLED=1
go build -v -o dy-live-monitor.exe .
```

**优点**：VS 命令提示符会自动设置所有 SDK 路径。

---

## 🐛 为什么 MinGW 找不到头文件？

### 问题根源

MinGW-w64 的 GCC 编译器使用的是 **Unix 风格的路径**，而 Windows SDK 在 `C:\Program Files (x86)\`。

```
Windows 路径:  C:\Program Files (x86)\Windows Kits\10\...
MinGW 看到的:  /c/Program Files (x86)/Windows Kits/10/...
```

### 解决方法

**设置 CGO_CFLAGS**，让 Go 的 CGO 告诉 GCC 去哪里找头文件：

```cmd
set "CGO_CFLAGS=-IC:\Program Files (x86)\Windows Kits\10\Include\10.0.26100.0\winrt"
```

---

## 📁 EventToken.h 的实际位置

```
C:\Program Files (x86)\Windows Kits\10\Include\10.0.26100.0\
├── shared/          ← 你设置了 ✅
├── um/              ← 你设置了 ✅
├── winrt/           ← 你缺少了 ❌（EventToken.h 在这里！）
├── ucrt/            ← 建议添加 ⚠️
└── cppwinrt/        ← 建议添加 ⚠️
```

---

## 🔍 验证设置是否正确

### 1. 检查文件是否存在

```cmd
dir "C:\Program Files (x86)\Windows Kits\10\Include\10.0.26100.0\winrt\EventToken.h"
```

**预期输出**：
```
 驱动器 C 中的卷是 Windows
 卷的序列号是 XXXX-XXXX

 C:\Program Files (x86)\Windows Kits\10\Include\10.0.26100.0\winrt 的目录

2021/11/16  15:30             1,234 EventToken.h
               1 个文件          1,234 字节
```

### 2. 测试 CGO 能否找到头文件

```cmd
echo #include "EventToken.h" > test.c
gcc -v -E test.c 2>&1 | findstr EventToken
```

如果成功，会显示找到的头文件路径。

---

## 🚀 推荐方案（按优先级）

### 🥇 方案 A: 使用自动脚本
```cmd
.\FIX_WEBVIEW2_BUILD.bat
```
**时间**: 2 分钟  
**难度**: ⭐ (最简单)

---

### 🥈 方案 B: Visual Studio Command Prompt
```cmd
# 在 VS 命令提示符中
cd server-go
go build -o dy-live-monitor.exe .
```
**时间**: 3 分钟  
**难度**: ⭐⭐

---

### 🥉 方案 C: 手动设置所有路径
```cmd
# 按照上面"方法 2"的完整步骤
```
**时间**: 5 分钟  
**难度**: ⭐⭐⭐

---

### 🏅 方案 D: 放弃 WebView2（最快）

**如果上述方案都不行**，使用无 WebView2 版本：

```cmd
cd server-go
go mod edit -droprequire=github.com/webview/webview_go
go mod tidy
go build -ldflags="-H windowsgui" -o dy-live-monitor.exe .
```

**功能影响**: 只是没有图形界面，核心功能 100% 相同。

---

## 📊 MinGW-w64 版本要求

**推荐版本**: MinGW-w64 8.1.0 或更高

### 检查当前版本
```cmd
gcc --version
```

### 如果版本过旧
下载最新版本：
- https://www.mingw-w64.org/downloads/
- 或使用 MSYS2: https://www.msys2.org/

---

## 🔄 完整编译流程（复制粘贴）

**在 CMD 中，一次性执行**：

```cmd
REM ========== 设置环境 ==========
set SDK_BASE=C:\Program Files (x86)\Windows Kits\10
set SDK_VERSION=10.0.26100.0
set "INCLUDE=%INCLUDE%;%SDK_BASE%\Include\%SDK_VERSION%\um;%SDK_BASE%\Include\%SDK_VERSION%\shared;%SDK_BASE%\Include\%SDK_VERSION%\winrt;%SDK_BASE%\Include\%SDK_VERSION%\ucrt"
set "LIB=%LIB%;%SDK_BASE%\Lib\%SDK_VERSION%\um\x64;%SDK_BASE%\Lib\%SDK_VERSION%\ucrt\x64"
set CGO_ENABLED=1
set "CGO_CFLAGS=-I%SDK_BASE%\Include\%SDK_VERSION%\winrt -I%SDK_BASE%\Include\%SDK_VERSION%\um -I%SDK_BASE%\Include\%SDK_VERSION%\shared"

REM ========== 验证 ==========
dir "%SDK_BASE%\Include\%SDK_VERSION%\winrt\EventToken.h"

REM ========== 编译 ==========
cd server-go
go clean -cache
go build -v -o dy-live-monitor.exe .
cd ..

echo ✅ 编译完成！
```

---

## 📝 常见问题

### Q1: 仍然找不到 EventToken.h
**A**: 
1. 确认文件存在：`dir "C:\Program Files (x86)\Windows Kits\10\Include\10.0.26100.0\winrt\EventToken.h"`
2. 如果不存在，重新安装 Windows SDK，**确保勾选 "Windows SDK for UWP C++ Apps"**

### Q2: 编译很慢
**A**: 
- WebView2 编译需要 5-10 分钟是正常的
- 后续编译会更快（有缓存）

### Q3: 其他编译错误
**A**: 
- 尝试使用 VS 2022 命令提示符
- 或者使用无 WebView2 版本（方案 D）

---

## 📞 需要帮助？

如果以上方案都不行：
1. 运行 `.\FIX_WEBVIEW2_BUILD.bat`
2. 截图完整的错误信息
3. 提供 `gcc --version` 和 `go version` 输出

或者直接使用**系统托盘版本**（无 WebView2），功能完全相同！

---

**最后更新**: 2025-11-15  
**适用版本**: Windows 10 SDK 10.0.26100.0  
**测试环境**: Windows 10/11 + MinGW-w64
