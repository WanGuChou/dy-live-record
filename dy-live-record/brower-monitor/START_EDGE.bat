@echo off
REM CDP Monitor - Edge启动脚本（隐藏调试提示）
REM 使用方法：双击运行此脚本

echo ========================================
echo CDP Monitor - Edge 启动脚本
echo ========================================
echo.

REM Edge安装路径（根据实际情况修改）
set EDGE_PATH="C:\Program Files (x86)\Microsoft\Edge\Application\msedge.exe"

REM 检查Edge是否存在
if not exist %EDGE_PATH% (
    echo [错误] 未找到Edge，尝试其他路径...
    set EDGE_PATH="C:\Program Files\Microsoft\Edge\Application\msedge.exe"
    if not exist %EDGE_PATH% (
        echo [错误] 仍未找到Edge，请检查路径：
        echo %EDGE_PATH%
        echo.
        echo 常见路径：
        echo   - C:\Program Files (x86)\Microsoft\Edge\Application\msedge.exe
        echo   - C:\Program Files\Microsoft\Edge\Application\msedge.exe
        echo.
        pause
        exit /b 1
    )
)

echo [1/3] 关闭现有Edge进程...
taskkill /F /IM msedge.exe >nul 2>&1
timeout /t 2 >nul

echo [2/3] 启动Edge（隐藏调试提示）...
start "" %EDGE_PATH% --silent-debugger-extension-api

echo [3/3] 完成！
echo.
echo ✅ Edge已启动，不会显示"正在调试此浏览器"提示
echo.
echo 📌 提示：
echo   - 请在Edge中加载CDP Monitor插件
echo   - 配置并启用监控功能
echo   - 查看详细说明：HIDE_DEBUGGER_BANNER.md
echo.

timeout /t 3 >nul
exit
