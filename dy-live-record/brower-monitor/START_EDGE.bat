@echo off
chcp 65001 >nul
REM CDP Monitor - Edge启动脚本（隐藏调试提示）

echo ========================================
echo CDP Monitor - Edge 启动脚本
echo ========================================
echo.

REM 尝试路径1: 常见32位路径
set "EDGE_PATH_1=C:\Program Files (x86)\Microsoft\Edge\Application\msedge.exe"

REM 尝试路径2: 64位路径
set "EDGE_PATH_2=C:\Program Files\Microsoft\Edge\Application\msedge.exe"

REM 检测Edge路径
set "EDGE_PATH="
if exist "%EDGE_PATH_1%" (
    set "EDGE_PATH=%EDGE_PATH_1%"
    goto :found
)

if exist "%EDGE_PATH_2%" (
    set "EDGE_PATH=%EDGE_PATH_2%"
    goto :found
)

REM 未找到Edge
echo [错误] 未找到Edge
echo.
echo 已尝试的路径：
echo   1. %EDGE_PATH_1%
echo   2. %EDGE_PATH_2%
echo.
echo 请确认Edge是否已安装
echo.
pause
exit /b 1

:found
echo [1/3] 找到Edge
echo       路径: %EDGE_PATH%
echo.

echo [2/3] 关闭现有Edge进程...
taskkill /F /IM msedge.exe >nul 2>&1
if errorlevel 1 (
    echo       没有运行中的Edge进程
) else (
    echo       已关闭Edge进程
)
timeout /t 2 >nul
echo.

echo [3/3] 启动Edge（隐藏调试提示）...
start "" "%EDGE_PATH%" --silent-debugger-extension-api
if errorlevel 1 (
    echo       [错误] Edge启动失败
    pause
    exit /b 1
) else (
    echo       Edge启动成功！
)
echo.

echo ========================================
echo 完成！
echo ========================================
echo.
echo ✅ Edge已启动，不会显示"正在调试此浏览器"提示
echo.
echo 📌 接下来的步骤：
echo   1. 在Edge中打开 edge://extensions/
echo   2. 启用"开发者模式"
echo   3. 点击"加载已解压的扩展程序"
echo   4. 选择此文件夹（brower-monitor）
echo   5. 配置并启用CDP Monitor插件
echo.
echo 📖 查看详细说明：README_STARTUP_SCRIPTS.md
echo.

timeout /t 5
exit /b 0
