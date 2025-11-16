@echo off
chcp 65001 >nul 2>&1
setlocal enabledelayedexpansion

echo ========================================
echo WebSocket 服务器连接测试
echo ========================================
echo.

REM 读取配置文件获取端口
set PORT=8080
if exist "config.json" (
    echo [1/3] 读取配置文件...
    for /f "tokens=2 delims=:, " %%a in ('findstr /C:"\"port\"" config.json') do (
        set PORT=%%a
    )
)
echo 配置端口: !PORT!
echo.

echo [2/3] 测试健康检查接口...
curl -s http://localhost:!PORT!/health
if %ERRORLEVEL% EQU 0 (
    echo.
    echo ✅ 健康检查成功！
) else (
    echo.
    echo ❌ 健康检查失败！请确认：
    echo    1. dy-live-monitor.exe 是否正在运行？
    echo    2. 端口 !PORT! 是否被占用？
    echo    3. 防火墙是否阻止了连接？
)
echo.

echo [3/3] WebSocket 连接信息
echo ========================================
echo WebSocket 地址: ws://localhost:!PORT!/ws
echo 健康检查地址:   http://localhost:!PORT!/health
echo ========================================
echo.

echo 测试完成！
echo.
echo 下一步：
echo 1. 打开浏览器安装 browser-monitor 插件
echo 2. 访问抖音直播间 (live.douyin.com/房间号)
echo 3. 插件会自动连接到 WebSocket 服务器
echo.

pause
