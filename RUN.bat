@echo off
chcp 65001 >nul 2>&1
cd server-go

REM 检查配置
if not exist config.json copy config.debug.json config.json >nul

REM 运行程序
if exist dy-live-monitor.exe (
    dy-live-monitor.exe
) else (
    go run main.go
)
