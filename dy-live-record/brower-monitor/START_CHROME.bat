@echo off
REM CDP Monitor - Chrome启动脚本（隐藏调试提示）
REM 使用方法：双击运行此脚本

echo ========================================
echo CDP Monitor - Chrome 启动脚本
echo ========================================
echo.

REM Chrome安装路径（根据实际情况修改）
set CHROME_PATH="C:\Program Files\Google\Chrome\Application\chrome.exe"

REM 检查Chrome是否存在
if not exist %CHROME_PATH% (
    echo [错误] 未找到Chrome，请检查路径：
    echo %CHROME_PATH%
    echo.
    echo 常见路径：
    echo   - C:\Program Files\Google\Chrome\Application\chrome.exe
    echo   - C:\Program Files (x86)\Google\Chrome\Application\chrome.exe
    echo.
    pause
    exit /b 1
)

echo [1/3] 关闭现有Chrome进程...
taskkill /F /IM chrome.exe >nul 2>&1
timeout /t 2 >nul

echo [2/3] 启动Chrome（隐藏调试提示）...
start "" %CHROME_PATH% --silent-debugger-extension-api

echo [3/3] 完成！
echo.
echo ✅ Chrome已启动，不会显示"正在调试此浏览器"提示
echo.
echo 📌 提示：
echo   - 请在Chrome中加载CDP Monitor插件
echo   - 配置并启用监控功能
echo   - 查看详细说明：HIDE_DEBUGGER_BANNER.md
echo.

timeout /t 3 >nul
exit
