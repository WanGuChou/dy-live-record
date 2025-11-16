@echo off
REM ========================================
REM 修复 WebView2 编译问题
REM ========================================

echo.
echo ========================================
echo 修复 WebView2 编译环境
echo ========================================
echo.

REM 设置 Windows SDK 版本
set SDK_VERSION=10.0.26100.0
set SDK_BASE=C:\Program Files (x86)\Windows Kits\10

echo [1/4] 设置 SDK 路径...
echo SDK 版本: %SDK_VERSION%
echo SDK 基础路径: %SDK_BASE%
echo.

REM 设置 Include 路径（必须包含 winrt）
set "INCLUDE=%INCLUDE%;%SDK_BASE%\Include\%SDK_VERSION%\um"
set "INCLUDE=%INCLUDE%;%SDK_BASE%\Include\%SDK_VERSION%\shared"
set "INCLUDE=%INCLUDE%;%SDK_BASE%\Include\%SDK_VERSION%\winrt"
set "INCLUDE=%INCLUDE%;%SDK_BASE%\Include\%SDK_VERSION%\ucrt"
set "INCLUDE=%INCLUDE%;%SDK_BASE%\Include\%SDK_VERSION%\cppwinrt"

REM 设置 Lib 路径
set "LIB=%LIB%;%SDK_BASE%\Lib\%SDK_VERSION%\um\x64"
set "LIB=%LIB%;%SDK_BASE%\Lib\%SDK_VERSION%\ucrt\x64"

echo [2/4] 设置 CGO 编译选项...

REM 设置 CGO 标志（关键！）
set CGO_ENABLED=1
set "CGO_CFLAGS=-I%SDK_BASE%\Include\%SDK_VERSION%\um"
set "CGO_CFLAGS=%CGO_CFLAGS% -I%SDK_BASE%\Include\%SDK_VERSION%\shared"
set "CGO_CFLAGS=%CGO_CFLAGS% -I%SDK_BASE%\Include\%SDK_VERSION%\winrt"
set "CGO_CFLAGS=%CGO_CFLAGS% -I%SDK_BASE%\Include\%SDK_VERSION%\ucrt"
set "CGO_CFLAGS=%CGO_CFLAGS% -I%SDK_BASE%\Include\%SDK_VERSION%\cppwinrt"

set "CGO_LDFLAGS=-L%SDK_BASE%\Lib\%SDK_VERSION%\um\x64"
set "CGO_LDFLAGS=%CGO_LDFLAGS% -L%SDK_BASE%\Lib\%SDK_VERSION%\ucrt\x64"

echo CGO_ENABLED: %CGO_ENABLED%
echo CGO_CFLAGS: %CGO_CFLAGS%
echo CGO_LDFLAGS: %CGO_LDFLAGS%
echo.

echo [3/4] 验证关键文件...

REM 检查 EventToken.h 是否存在
if exist "%SDK_BASE%\Include\%SDK_VERSION%\winrt\EventToken.h" (
    echo [✓] 找到 EventToken.h
) else (
    echo [✗] 未找到 EventToken.h
    echo 路径: %SDK_BASE%\Include\%SDK_VERSION%\winrt\EventToken.h
    echo.
    echo 请确认 Windows SDK 安装完整
    pause
    exit /b 1
)

REM 检查 WebView2.h
if exist "%SDK_BASE%\Include\%SDK_VERSION%\um\WebView2.h" (
    echo [✓] 找到 WebView2.h
) else (
    echo [⚠] 未找到 WebView2.h（可能在其他位置）
)

echo.
echo [4/4] 开始编译 server-go...
echo.

cd server-go

REM 清理之前的编译缓存
echo [清理] 删除旧的编译文件...
if exist dy-live-monitor.exe del /f dy-live-monitor.exe
go clean -cache
echo.

REM 编译（显示详细输出）
echo [编译] 编译中，请稍候...
go build -v -x -o dy-live-monitor.exe . 2>&1 | findstr /v "pkg-config"

if %ERRORLEVEL% NEQ 0 (
    echo.
    echo ========================================
    echo ❌ 编译失败
    echo ========================================
    echo.
    echo 可能的原因:
    echo 1. MinGW-w64 版本过旧
    echo 2. Windows SDK 安装不完整
    echo 3. CGO 编译器配置问题
    echo.
    echo 建议:
    echo 1. 尝试更新 MinGW-w64 到最新版本
    echo 2. 或者使用无 WebView2 版本（参考 QUICK_BUILD.md 方案 3）
    echo.
    cd ..
    pause
    exit /b 1
)

echo.
echo ========================================
echo 🎉 编译成功！
echo ========================================
echo.
echo 生成文件: server-go\dy-live-monitor.exe
echo.

cd ..

echo 运行程序:
echo   cd server-go
echo   .\dy-live-monitor.exe
echo.
pause
