@echo off
chcp 65001 >nul
REM CDP Monitor - Chrome启动脚本（隐藏调试提示）

echo ========================================
echo CDP Monitor - Chrome 启动脚本
echo ========================================
echo.

REM 尝试路径1: 64位系统默认路径
set "CHROME_PATH_1=C:\Program Files\Google\Chrome\Application\chrome.exe"

REM 尝试路径2: 32位或旧版路径
set "CHROME_PATH_2=C:\Program Files (x86)\Google\Chrome\Application\chrome.exe"

REM 检测Chrome路径
set "CHROME_PATH="
if exist "%CHROME_PATH_1%" (
    set "CHROME_PATH=%CHROME_PATH_1%"
    goto :found
)

if exist "%CHROME_PATH_2%" (
    set "CHROME_PATH=%CHROME_PATH_2%"
    goto :found
)

REM 未找到Chrome
echo [错误] 未找到Chrome
echo.
echo 已尝试的路径：
echo   1. %CHROME_PATH_1%
echo   2. %CHROME_PATH_2%
echo.
echo 请确认Chrome是否已安装
echo 或使用 START_CHROME_MANUAL.bat 手动指定路径
echo.
pause
exit /b 1

:found
echo [1/3] 找到Chrome
echo       路径: %CHROME_PATH%
echo.

echo [2/3] 关闭现有Chrome进程...
taskkill /F /IM chrome.exe >nul 2>&1
if errorlevel 1 (
    echo       没有运行中的Chrome进程
) else (
    echo       已关闭Chrome进程
)
timeout /t 2 >nul
echo.

echo [3/3] 启动Chrome（隐藏调试提示）...
start "" "%CHROME_PATH%" --silent-debugger-extension-api
if errorlevel 1 (
    echo       [错误] Chrome启动失败
    pause
    exit /b 1
) else (
    echo       Chrome启动成功！
)
echo.

echo ========================================
echo 完成！
echo ========================================
echo.
echo ✅ Chrome已启动，不会显示"正在调试此浏览器"提示
echo.
echo 📌 接下来的步骤：
echo   1. 在Chrome中打开 chrome://extensions/
echo   2. 启用"开发者模式"
echo   3. 点击"加载已解压的扩展程序"
echo   4. 选择此文件夹（brower-monitor）
echo   5. 配置并启用CDP Monitor插件
echo.
echo 📖 查看详细说明：README_STARTUP_SCRIPTS.md
echo.

timeout /t 5
exit /b 0
