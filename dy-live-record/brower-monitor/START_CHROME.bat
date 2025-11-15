@echo off
chcp 65001 >nul
REM CDP Monitor - Chrome启动脚本（隐藏调试提示）
REM 使用方法：双击运行此脚本

echo ========================================
echo CDP Monitor - Chrome 启动脚本
echo ========================================
echo.

REM Chrome安装路径（根据实际情况修改）
set CHROME_PATH=C:\Program Files\Google\Chrome\Application\chrome.exe

REM 检查Chrome是否存在
if not exist "%CHROME_PATH%" (
    set CHROME_PATH=C:\Program Files (x86)\Google\Chrome\Application\chrome.exe
    if not exist "%CHROME_PATH%" (
        echo [错误] 未找到Chrome，请检查路径
        echo.
        echo 已尝试的路径：
        echo   - C:\Program Files\Google\Chrome\Application\chrome.exe
        echo   - C:\Program Files (x86)\Google\Chrome\Application\chrome.exe
        echo.
        echo 请手动指定Chrome路径：
        set /p CHROME_PATH="请输入Chrome完整路径: "
        if not exist "%CHROME_PATH%" (
            echo [错误] 路径仍然无效
            pause
            exit /b 1
        )
    )
)

echo [1/3] 找到Chrome: %CHROME_PATH%
echo.

echo [2/3] 关闭现有Chrome进程...
taskkill /F /IM chrome.exe >nul 2>&1
if %errorlevel% equ 0 (
    echo     已关闭Chrome进程
) else (
    echo     没有运行中的Chrome进程
)
timeout /t 2 >nul
echo.

echo [3/3] 启动Chrome（隐藏调试提示）...
start "" "%CHROME_PATH%" --silent-debugger-extension-api
if %errorlevel% equ 0 (
    echo     Chrome启动成功！
) else (
    echo     [错误] Chrome启动失败
    pause
    exit /b 1
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
echo   4. 选择此文件夹的上级目录
echo   5. 配置并启用CDP Monitor插件
echo.
echo 📖 查看详细说明：
echo   - 完整文档: ..\..\HIDE_DEBUGGER_BANNER.md
echo   - 使用指南: ..\..\CDP_USAGE.md
echo.

pause
