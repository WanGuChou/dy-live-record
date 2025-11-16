@echo off
chcp 65001 >nul 2>&1
cls
echo ========================================
echo 抖音直播监控系统 - 调试启动
echo ========================================
echo.
echo 版本: v3.2.1
echo 模式: 调试模式（跳过 License）
echo.
echo ========================================
echo.

REM 进入目录
cd server-go

REM 检查配置
if not exist config.json (
    echo [1/3] 配置调试模式...
    copy config.debug.json config.json >nul
    echo ✓ 调试配置已应用
) else (
    echo [1/3] 使用现有配置
)

REM 检查文件
echo [2/3] 检查程序文件...
if not exist dy-live-monitor.exe (
    echo ✓ 将运行源码模式
    set RUN_MODE=source
) else (
    echo ✓ 找到已编译程序
    set RUN_MODE=compiled
)

REM 启动程序
echo [3/3] 启动程序...
echo.
echo ========================================
echo.

if "%RUN_MODE%"=="compiled" (
    dy-live-monitor.exe
) else (
    go run main.go
)

echo.
echo ========================================
echo 程序已退出
echo ========================================
pause
