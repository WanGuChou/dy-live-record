@echo off
chcp 65001 >nul
REM CDP Monitor - Chrome手动路径启动脚本
REM 如果自动检测失败，使用此脚本手动指定Chrome路径

echo ========================================
echo CDP Monitor - Chrome 手动启动脚本
echo ========================================
echo.
echo 此脚本允许你手动指定Chrome的安装路径
echo.

REM 让用户输入Chrome路径
echo 请输入Chrome的完整路径
echo 常见路径示例：
echo   C:\Program Files\Google\Chrome\Application\chrome.exe
echo   C:\Program Files (x86)\Google\Chrome\Application\chrome.exe
echo.

set /p CHROME_PATH="Chrome路径: "

REM 移除可能的引号
set CHROME_PATH=%CHROME_PATH:"=%

REM 检查路径是否存在
if not exist "%CHROME_PATH%" (
    echo.
    echo [错误] 找不到指定的Chrome路径：
    echo %CHROME_PATH%
    echo.
    echo 请检查：
    echo   1. 路径是否正确
    echo   2. Chrome是否已安装
    echo   3. 路径中是否有拼写错误
    echo.
    pause
    exit /b 1
)

echo.
echo [1/3] 找到Chrome: %CHROME_PATH%
echo.

echo [2/3] 关闭现有Chrome进程...
taskkill /F /IM chrome.exe >nul 2>&1
timeout /t 2 >nul
echo.

echo [3/3] 启动Chrome（隐藏调试提示）...
start "" "%CHROME_PATH%" --silent-debugger-extension-api
echo.

echo ========================================
echo 完成！
echo ========================================
echo.
echo ✅ Chrome已启动
echo.
echo 💡 提示：如果此路径经常使用，可以修改 START_CHROME.bat
echo    将第10行的路径改为：
echo    set CHROME_PATH=%CHROME_PATH%
echo.

pause
