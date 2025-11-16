@echo off
chcp 65001 >nul
REM ========================================
REM 编译 Fyne GUI 版本
REM ========================================

echo.
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
echo.

REM 1. 打包浏览器插件
echo [1/4] 打包浏览器插件...
cd browser-monitor
if exist pack.bat (
    call pack.bat >nul 2>&1
    if %ERRORLEVEL% EQU 0 (
        echo [OK] 插件打包成功
    ) else (
        echo [WARN] 插件打包失败，继续编译
    )
) else (
    echo [SKIP] pack.bat 不存在
)
cd ..
echo.

REM 2. 下载 Fyne 依赖
echo [2/4] 下载 Fyne 依赖...
cd server-go
go mod tidy
if %ERRORLEVEL% NEQ 0 (
    echo [ERROR] 依赖下载失败
    cd ..
    pause
    exit /b 1
)
echo [OK] 依赖下载完成
echo.

REM 3. 编译 server-go（Fyne GUI）
echo [3/4] 编译 server-go（Fyne GUI）...
echo [INFO] 开始编译（预计 2-3 分钟）...

set CGO_ENABLED=1
go build -v -ldflags="-H windowsgui -s -w" -o dy-live-monitor.exe . 2>&1 | findstr /V "internal"

if %ERRORLEVEL% NEQ 0 (
    echo.
    echo [ERROR] server-go 编译失败
    echo.
    echo 可能的原因:
    echo   1. GCC 未安装（Fyne 需要 CGO）
    echo   2. 网络问题（无法下载 Fyne 依赖）
    echo.
    echo 解决方法:
    echo   1. 安装 MinGW-w64: https://www.mingw-w64.org/
    echo   2. 设置 Go 代理: set GOPROXY=https://goproxy.cn,direct
    echo.
    cd ..
    pause
    exit /b 1
)

echo [OK] server-go 编译完成
cd ..
echo.

REM 4. 编译 server-active
echo [4/4] 编译 server-active...
cd server-active

go mod tidy >nul 2>&1
go build -v -ldflags="-s -w" -o dy-live-license.exe . 2>&1 | findstr /V "internal"

if %ERRORLEVEL% NEQ 0 (
    echo [ERROR] server-active 编译失败
    cd ..
    pause
    exit /b 1
)

echo [OK] server-active 编译完成
cd ..
echo.

REM 检查结果
echo ========================================
echo 编译成功！
echo ========================================
echo.

echo 生成的文件:
if exist "server-go\dy-live-monitor.exe" (
    echo   [YES] server-go\dy-live-monitor.exe
    for %%F in ("server-go\dy-live-monitor.exe") do (
        set size=%%~zF
        set /a sizeMB=!size! / 1024 / 1024
        echo         大小: !sizeMB! MB
    )
) else (
    echo   [NO] server-go\dy-live-monitor.exe
)

if exist "server-active\dy-live-license.exe" (
    echo   [YES] server-active\dy-live-license.exe
) else (
    echo   [NO] server-active\dy-live-license.exe
)

if exist "server-go\assets\browser-monitor.zip" (
    echo   [YES] server-go\assets\browser-monitor.zip
) else (
    echo   [WARN] server-go\assets\browser-monitor.zip 未找到
)

echo.
echo 运行程序:
echo   cd server-go
echo   .\dy-live-monitor.exe
echo.
echo 特点:
echo   - 图形界面启动（Fyne）
echo   - 系统托盘同时可用
echo   - 无需 Windows SDK
echo   - 跨平台支持
echo.
pause
