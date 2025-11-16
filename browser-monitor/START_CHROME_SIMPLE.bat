@echo off
REM 最简单的启动脚本 - 如果其他脚本都失败，使用这个

echo 正在启动Chrome...
echo.

REM 关闭Chrome
taskkill /F /IM chrome.exe >nul 2>&1
timeout /t 2 >nul

REM 启动Chrome - 使用最常见的路径
if exist "C:\Program Files\Google\Chrome\Application\chrome.exe" (
    start "" "C:\Program Files\Google\Chrome\Application\chrome.exe" --silent-debugger-extension-api
    echo Chrome已启动
) else if exist "C:\Program Files (x86)\Google\Chrome\Application\chrome.exe" (
    start "" "C:\Program Files (x86)\Google\Chrome\Application\chrome.exe" --silent-debugger-extension-api
    echo Chrome已启动
) else (
    echo 错误：找不到Chrome
    echo 请使用 START_CHROME_MANUAL.bat
)

echo.
timeout /t 3
