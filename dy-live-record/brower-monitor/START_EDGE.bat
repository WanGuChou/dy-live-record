@echo off
chcp 65001 >nul
REM CDP Monitor - Edge启动脚本（隐藏调试提示）
REM 使用方法：双击运行此脚本

echo ========================================
echo CDP Monitor - Edge 启动脚本
echo ========================================
echo.

REM Edge安装路径（根据实际情况修改）
set EDGE_PATH=C:\Program Files (x86)\Microsoft\Edge\Application\msedge.exe

REM 检查Edge是否存在
if not exist "%EDGE_PATH%" (
    set EDGE_PATH=C:\Program Files\Microsoft\Edge\Application\msedge.exe
    if not exist "%EDGE_PATH%" (
        echo [错误] 未找到Edge，请检查路径
        echo.
        echo 已尝试的路径：
        echo   - C:\Program Files (x86)\Microsoft\Edge\Application\msedge.exe
        echo   - C:\Program Files\Microsoft\Edge\Application\msedge.exe
        echo.
        echo 请手动指定Edge路径：
        set /p EDGE_PATH="请输入Edge完整路径: "
        if not exist "%EDGE_PATH%" (
            echo [错误] 路径仍然无效
            pause
            exit /b 1
        )
    )
)

echo [1/3] 找到Edge: %EDGE_PATH%
echo.

echo [2/3] 关闭现有Edge进程...
taskkill /F /IM msedge.exe >nul 2>&1
if %errorlevel% equ 0 (
    echo     已关闭Edge进程
) else (
    echo     没有运行中的Edge进程
)
timeout /t 2 >nul
echo.

echo [3/3] 启动Edge（隐藏调试提示）...
start "" "%EDGE_PATH%" --silent-debugger-extension-api
if %errorlevel% equ 0 (
    echo     Edge启动成功！
) else (
    echo     [错误] Edge启动失败
    pause
    exit /b 1
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
echo   4. 选择此文件夹的上级目录
echo   5. 配置并启用CDP Monitor插件
echo.
echo 📖 查看详细说明：
echo   - 完整文档: ..\..\HIDE_DEBUGGER_BANNER.md
echo   - 使用指南: ..\..\CDP_USAGE.md
echo.

pause
