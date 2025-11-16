@echo off
chcp 65001 > nul
echo ========================================
echo   抖音直播监控系统 - 编译脚本
echo ========================================
echo.

echo [1/3] 清理旧的编译文件...
if exist dy-live-monitor.exe del dy-live-monitor.exe
if exist dy-live-monitor-debug.exe del dy-live-monitor-debug.exe

echo [2/3] 编译 Release 版本...
go build -ldflags "-H windowsgui -s -w" -o dy-live-monitor.exe main.go
if errorlevel 1 (
    echo ❌ Release 编译失败！
    pause
    exit /b 1
)
echo ✅ Release 版本编译成功: dy-live-monitor.exe

echo [3/3] 编译 Debug 版本（带控制台）...
go build -o dy-live-monitor-debug.exe main.go
if errorlevel 1 (
    echo ❌ Debug 编译失败！
    pause
    exit /b 1
)
echo ✅ Debug 版本编译成功: dy-live-monitor-debug.exe

echo.
echo ========================================
echo   编译完成！
echo ========================================
echo Release 版本（无控制台）: dy-live-monitor.exe
echo Debug 版本（有控制台）: dy-live-monitor-debug.exe
echo.
echo 提示：首次运行需要安装 WebView2 Runtime
echo      下载地址: https://developer.microsoft.com/microsoft-edge/webview2/
echo.
pause
