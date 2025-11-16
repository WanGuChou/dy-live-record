@echo off
REM ========================================
REM 完整编译脚本（自动设置 SDK 路径）
REM ========================================

echo.
echo ========================================
echo 抖音直播监控系统 - 完整编译
echo ========================================
echo.

REM 1. 检测并设置 Windows SDK
echo [步骤 1/5] 检测 Windows SDK...
set SDK_BASE=C:\Program Files (x86)\Windows Kits\10\Include

for /f "delims=" %%i in ('dir /b /ad /o-n "%SDK_BASE%" 2^>nul') do (
    set SDK_VERSION=%%i
    goto :sdk_found
)

:sdk_found
if "%SDK_VERSION%"=="" (
    echo [警告] 未找到 Windows SDK，将编译无 WebView2 版本
    set USE_WEBVIEW=0
    goto :skip_sdk
)

echo [找到] Windows SDK 版本: %SDK_VERSION%
set "INCLUDE=%INCLUDE%;%SDK_BASE%\%SDK_VERSION%\um"
set "INCLUDE=%INCLUDE%;%SDK_BASE%\%SDK_VERSION%\shared"
set "INCLUDE=%INCLUDE%;%SDK_BASE%\%SDK_VERSION%\winrt"

set SDK_LIB=C:\Program Files (x86)\Windows Kits\10\Lib\%SDK_VERSION%
set "LIB=%LIB%;%SDK_LIB%\um\x64"
set "LIB=%LIB%;%SDK_LIB%\ucrt\x64"

set USE_WEBVIEW=1
echo.

:skip_sdk

REM 2. 打包浏览器插件
echo [步骤 2/5] 打包浏览器插件...
cd browser-monitor
call pack.bat
if %ERRORLEVEL% NEQ 0 (
    echo [错误] 插件打包失败
    cd ..
    goto :error
)
cd ..
echo.

REM 3. 编译 server-go
echo [步骤 3/5] 编译 server-go...
cd server-go

if "%USE_WEBVIEW%"=="1" (
    echo [信息] 编译完整版本（包含 WebView2）
    REM 先恢复 WebView2 依赖
    go mod edit -require=github.com/webview/webview_go@v0.0.0-20240831120633-6173450d4dd6
    go mod tidy
) else (
    echo [信息] 编译无 WebView2 版本
)

set CGO_ENABLED=1
go build -v -ldflags="-H windowsgui" -o dy-live-monitor.exe .

if %ERRORLEVEL% NEQ 0 (
    echo [错误] server-go 编译失败
    cd ..
    goto :error
)

echo [成功] server-go 编译完成
cd ..
echo.

REM 4. 编译 server-active
echo [步骤 4/5] 编译 server-active...
cd server-active
go mod tidy
go build -v -o dy-live-license.exe .

if %ERRORLEVEL% NEQ 0 (
    echo [错误] server-active 编译失败
    cd ..
    goto :error
)

echo [成功] server-active 编译完成
cd ..
echo.

REM 5. 检查生成的文件
echo [步骤 5/5] 验证编译结果...
echo.

if exist "server-go\dy-live-monitor.exe" (
    echo [✓] server-go\dy-live-monitor.exe
) else (
    echo [✗] server-go\dy-live-monitor.exe
    set HAS_ERROR=1
)

if exist "server-active\dy-live-license.exe" (
    echo [✓] server-active\dy-live-license.exe
) else (
    echo [✗] server-active\dy-live-license.exe
    set HAS_ERROR=1
)

if exist "server-go\assets\browser-monitor.zip" (
    echo [✓] server-go\assets\browser-monitor.zip
) else (
    echo [✗] server-go\assets\browser-monitor.zip
    set HAS_ERROR=1
)

echo.
if "%HAS_ERROR%"=="1" goto :error

REM 成功
echo ========================================
echo 🎉 编译成功！
echo ========================================
echo.
echo 生成的文件:
echo   - server-go\dy-live-monitor.exe      (主程序)
echo   - server-active\dy-live-license.exe  (授权服务)
echo   - server-go\assets\browser-monitor.zip (浏览器插件)
echo.

if "%USE_WEBVIEW%"=="1" (
    echo 功能: 完整版本 (包含图形界面)
) else (
    echo 功能: 系统托盘版本 (无图形界面)
)

echo.
echo 运行主程序:
echo   cd server-go
echo   .\dy-live-monitor.exe
echo.
pause
exit /b 0

:error
echo.
echo ========================================
echo ❌ 编译失败
echo ========================================
echo.
echo 请检查上面的错误信息
echo 或查看 WEBVIEW2_FIX.md 获取帮助
echo.
pause
exit /b 1
