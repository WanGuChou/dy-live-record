@echo off
chcp 65001 >nul
REM ========================================
REM 编译无 WebView2 版本（推荐）
REM ========================================

echo.
echo ========================================
echo 编译系统托盘版本（无 WebView2）
echo ========================================
echo.
echo 功能说明:
echo   [YES] 数据采集     100%%
echo   [YES] WebSocket    100%%
echo   [YES] 数据存储     100%%
echo   [YES] 许可证系统   100%%
echo   [YES] 主播管理     100%%
echo   [YES] 系统托盘     100%%
echo   [NO]  图形界面     不支持
echo.
echo 优点:
echo   [YES] 编译速度快（30秒）
echo   [YES] 无需 Windows SDK
echo   [YES] 文件体积小
echo   [YES] 核心功能完整
echo.

echo 按任意键开始编译...
pause >nul
echo.

REM 1. 打包浏览器插件
echo [1/3] 打包浏览器插件...
cd browser-monitor
if exist pack.bat (
    call pack.bat >nul 2>&1
    if %ERRORLEVEL% EQU 0 (
        echo [OK] 插件打包成功
    ) else (
        echo [SKIP] 插件打包失败，继续编译
    )
) else (
    echo [SKIP] pack.bat 不存在
)
cd ..
echo.

REM 2. 编译 server-go（无 WebView2）
echo [2/3] 编译 server-go...
cd server-go

echo [INFO] 移除 WebView2 依赖...
go mod edit -droprequire=github.com/webview/webview_go >nul 2>&1

echo [INFO] 更新依赖...
go mod tidy >nul 2>&1

echo [INFO] 清理缓存...
go clean -cache >nul 2>&1

echo [INFO] 开始编译（预计 30 秒）...
set CGO_ENABLED=1
go build -v -ldflags="-H windowsgui -s -w" -o dy-live-monitor.exe . 2>&1 | findstr /V "internal"

if %ERRORLEVEL% NEQ 0 (
    echo.
    echo [ERROR] server-go 编译失败
    cd ..
    pause
    exit /b 1
)

echo [OK] server-go 编译完成
cd ..
echo.

REM 3. 编译 server-active
echo [3/3] 编译 server-active...
cd server-active

echo [INFO] 更新依赖...
go mod tidy >nul 2>&1

echo [INFO] 开始编译...
go build -v -ldflags="-s -w" -o dy-live-license.exe . 2>&1 | findstr /V "internal"

if %ERRORLEVEL% NEQ 0 (
    echo.
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
echo   1. 启动主程序: cd server-go ^&^& .\dy-live-monitor.exe
echo   2. 启动授权服务: cd server-active ^&^& .\dy-live-license.exe
echo.
echo 注意:
echo   - 程序启动后在系统托盘（右下角）
echo   - 右键托盘图标可以查看菜单
echo   - 数据采集功能完全正常
echo.
pause
