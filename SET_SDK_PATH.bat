@echo off
REM ========================================
REM 设置 Windows SDK 环境变量
REM ========================================

echo.
echo [1/3] 检测 Windows SDK 安装路径...

REM 常见的 SDK 版本路径
set SDK_BASE=C:\Program Files (x86)\Windows Kits\10\Include

REM 自动检测最新版本
for /f "delims=" %%i in ('dir /b /ad /o-n "%SDK_BASE%" 2^>nul') do (
    set SDK_VERSION=%%i
    goto :found
)

:found
if "%SDK_VERSION%"=="" (
    echo [错误] 未找到 Windows 10 SDK
    echo 请先安装: https://developer.microsoft.com/en-us/windows/downloads/windows-sdk/
    pause
    exit /b 1
)

echo [找到] Windows SDK 版本: %SDK_VERSION%
echo.

echo [2/3] 设置环境变量...
set "INCLUDE=%INCLUDE%;%SDK_BASE%\%SDK_VERSION%\um"
set "INCLUDE=%INCLUDE%;%SDK_BASE%\%SDK_VERSION%\shared"
set "INCLUDE=%INCLUDE%;%SDK_BASE%\%SDK_VERSION%\winrt"

set SDK_LIB=C:\Program Files (x86)\Windows Kits\10\Lib\%SDK_VERSION%
set "LIB=%LIB%;%SDK_LIB%\um\x64"
set "LIB=%LIB%;%SDK_LIB%\ucrt\x64"

echo [完成] 环境变量已设置
echo.

echo [3/3] 当前 SDK 路径:
echo INCLUDE: %SDK_BASE%\%SDK_VERSION%
echo LIB: %SDK_LIB%
echo.

echo ========================================
echo 环境变量设置成功！
echo 现在可以编译 server-go 了
echo ========================================
echo.
echo 运行编译命令:
echo   cd server-go
echo   go build -o dy-live-monitor.exe .
echo.
pause
