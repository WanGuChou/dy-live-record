@echo off
REM ========================================
REM 修复 CGO 路径空格问题
REM ========================================

echo.
echo ========================================
echo 修复 CGO 编译路径问题
echo ========================================
echo.

set SDK_VERSION=10.0.26100.0

echo [方案] 使用 Windows 短路径名（8.3 格式）
echo.

REM Program Files (x86) 的短路径是 PROGRA~2
set SDK_BASE=C:\PROGRA~2\Windows Kits\10

echo [1/3] 设置 CGO 编译标志...

REM 使用短路径名，避免空格问题
set CGO_ENABLED=1
set "CGO_CFLAGS=-I%SDK_BASE%\Include\%SDK_VERSION%\winrt -I%SDK_BASE%\Include\%SDK_VERSION%\um -I%SDK_BASE%\Include\%SDK_VERSION%\shared -I%SDK_BASE%\Include\%SDK_VERSION%\ucrt"
set "CGO_LDFLAGS=-L%SDK_BASE%\Lib\%SDK_VERSION%\um\x64 -L%SDK_BASE%\Lib\%SDK_VERSION%\ucrt\x64"

echo CGO_ENABLED: %CGO_ENABLED%
echo CGO_CFLAGS: %CGO_CFLAGS%
echo.

echo [2/3] 验证路径...
if exist "%SDK_BASE%\Include\%SDK_VERSION%\winrt\EventToken.h" (
    echo [✓] EventToken.h 找到
) else (
    echo [✗] EventToken.h 未找到
    echo 完整路径: C:\Program Files (x86)\Windows Kits\10\Include\%SDK_VERSION%\winrt\EventToken.h
    pause
    exit /b 1
)

echo.
echo [3/3] 开始编译...
cd server-go

REM 清理缓存
go clean -cache

echo.
echo 编译中，请稍候（首次编译约需 5-10 分钟）...
echo.

go build -v -o dy-live-monitor.exe .

if %ERRORLEVEL% NEQ 0 (
    echo.
    echo ========================================
    echo ❌ 编译失败
    echo ========================================
    echo.
    echo 如果仍然有路径错误，尝试备选方案:
    echo   1. 使用 Visual Studio 命令提示符
    echo   2. 使用无 WebView2 版本（见下方）
    echo.
    echo 无 WebView2 版本编译（30秒完成）:
    echo   go mod edit -droprequire=github.com/webview/webview_go
    echo   go mod tidy
    echo   go build -o dy-live-monitor.exe .
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
echo 大小: 
dir dy-live-monitor.exe | findstr "dy-live-monitor"
echo.

cd ..

echo 运行程序:
echo   cd server-go
echo   .\dy-live-monitor.exe
echo.
pause
