@echo off
REM ========================================
REM 编译无 WebView2 版本（系统托盘模式）
REM ========================================

echo.
echo ========================================
echo 编译系统托盘版本（无 WebView2）
echo ========================================
echo.
echo 功能说明:
echo   ✅ 数据采集     100%%
echo   ✅ WebSocket    100%%
echo   ✅ 数据存储     100%%
echo   ✅ 许可证系统   100%%
echo   ✅ 主播管理     100%%
echo   ✅ 系统托盘     100%%
echo   ❌ 图形界面     不支持
echo.
echo 优点:
echo   ✅ 编译速度快（30秒）
echo   ✅ 无需 Windows SDK
echo   ✅ 文件体积小
echo   ✅ 核心功能完整
echo.

pause
echo.

REM 1. 打包浏览器插件
echo [1/3] 打包浏览器插件...
cd browser-monitor
if exist pack.bat (
    call pack.bat
) else (
    echo [跳过] pack.bat 不存在
)
cd ..
echo.

REM 2. 编译 server-go（无 WebView2）
echo [2/3] 编译 server-go...
cd server-go

echo [信息] 移除 WebView2 依赖...
go mod edit -droprequire=github.com/webview/webview_go
go mod tidy

echo [信息] 清理缓存...
go clean -cache

echo [信息] 开始编译...
set CGO_ENABLED=1
go build -v -ldflags="-H windowsgui -s -w" -o dy-live-monitor.exe .

if %ERRORLEVEL% NEQ 0 (
    echo.
    echo [错误] server-go 编译失败
    cd ..
    pause
    exit /b 1
)

echo [成功] server-go 编译完成
cd ..
echo.

REM 3. 编译 server-active
echo [3/3] 编译 server-active...
cd server-active

go mod tidy
go build -v -ldflags="-s -w" -o dy-live-license.exe .

if %ERRORLEVEL% NEQ 0 (
    echo.
    echo [错误] server-active 编译失败
    cd ..
    pause
    exit /b 1
)

echo [成功] server-active 编译完成
cd ..
echo.

REM 检查结果
echo ========================================
echo 🎉 编译完成！
echo ========================================
echo.

echo 生成的文件:
if exist "server-go\dy-live-monitor.exe" (
    echo   [✓] server-go\dy-live-monitor.exe
    for %%F in ("server-go\dy-live-monitor.exe") do echo       大小: %%~zF 字节 ^(%%~zF / 1024 / 1024 MB^)
) else (
    echo   [✗] server-go\dy-live-monitor.exe
)

if exist "server-active\dy-live-license.exe" (
    echo   [✓] server-active\dy-live-license.exe
    for %%F in ("server-active\dy-live-license.exe") do echo       大小: %%~zF 字节
) else (
    echo   [✗] server-active\dy-live-license.exe
)

if exist "server-go\assets\browser-monitor.zip" (
    echo   [✓] server-go\assets\browser-monitor.zip
) else (
    echo   [⚠] server-go\assets\browser-monitor.zip 未找到
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
