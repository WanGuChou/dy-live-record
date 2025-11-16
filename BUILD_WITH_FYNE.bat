@echo off
chcp 65001 >nul 2>&1
setlocal enabledelayedexpansion

echo ========================================
echo 抖音直播监控系统 - Fyne GUI 版本
echo ========================================
echo.
echo 功能特点:
echo   [YES] 现代化图形界面（Fyne）
echo   [YES] 跨平台支持
echo   [YES] 无需 Windows SDK
echo   [YES] 纯 Go 实现
echo   [YES] 完整功能
echo.
pause

REM Step 1: 打包浏览器插件
echo.
echo [1/4] 打包浏览器插件...
if exist browser-monitor\pack.bat (
    cd browser-monitor
    call pack.bat
    cd ..
    if exist server-go\assets\browser-monitor.zip (
        echo [OK] 插件打包成功
    ) else (
        echo [WARNING] 插件打包可能失败，但继续编译
    )
) else (
    echo [SKIP] pack.bat 不存在
)

REM Step 2: 下载 Fyne 依赖
echo.
echo [2/4] 下载 Fyne 依赖...
cd server-go
go mod download
if %ERRORLEVEL% EQU 0 (
    echo [OK] 依赖下载完成
) else (
    echo [WARNING] 依赖下载可能失败，尝试继续...
)

REM Step 3: 检查 GCC
echo.
echo [3/4] 检查 CGO 环境...
where gcc >nul 2>&1
if %ERRORLEVEL% NEQ 0 (
    echo [ERROR] GCC 未找到！Fyne 需要 CGO 支持。
    echo.
    echo 请安装 MinGW-w64:
    echo   方法 1: choco install mingw
    echo   方法 2: https://www.mingw-w64.org/
    echo.
    pause
    exit /b 1
) else (
    echo [OK] GCC 已安装
)

REM Step 4: 编译 server-go
echo.
echo [4/4] 编译 server-go...
echo [INFO] 开始编译（预计 2-3 分钟）...
echo.

set CGO_ENABLED=1
go build -v -o dy-live-monitor.exe .

if %ERRORLEVEL% EQU 0 (
    echo.
    echo ========================================
    echo ✓ 编译成功！
    echo ========================================
    echo.
    echo 输出文件: server-go\dy-live-monitor.exe
    echo.
    echo 启动方法:
    echo   1. 调试模式（跳过 License）:
    echo      cd server-go
    echo      copy config.debug.json config.json
    echo      .\dy-live-monitor.exe
    echo.
    echo   2. 正常模式:
    echo      cd server-go
    echo      .\dy-live-monitor.exe
    echo.
) else (
    echo.
    echo ========================================
    echo ✗ 编译失败
    echo ========================================
    echo.
    echo 可能的原因:
    echo   1. GCC 未安装（Fyne 需要 CGO）
    echo   2. 网络问题（无法下载依赖）
    echo   3. 磁盘空间不足
    echo.
    echo 解决方法:
    echo   1. 安装 MinGW-w64: https://www.mingw-w64.org/
    echo   2. 设置 Go 代理: set GOPROXY=https://goproxy.cn,direct
    echo   3. 查看详细错误: go build -v -x -o dy-live-monitor.exe .
    echo.
    echo 或使用系统托盘版本（无需 Fyne）:
    echo   .\BUILD_NO_WEBVIEW2_FIXED.bat
    echo.
)

cd ..
pause
