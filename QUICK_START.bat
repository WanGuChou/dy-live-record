@echo off
chcp 65001 >nul
REM ========================================
REM 一键编译脚本（推荐使用）
REM ========================================

cls
echo.
echo ========================================
echo 抖音直播监控系统 - 一键编译
echo ========================================
echo.
echo 请选择编译方式:
echo.
echo   [1] Fyne GUI 版本（推荐）
echo       - 编译时间: 2-3分钟
echo       - 无需 Windows SDK
echo       - 现代化图形界面
echo       - 跨平台支持
echo.
echo   [2] 系统托盘版本（轻量）
echo       - 编译时间: 30秒
echo       - 无需 Windows SDK
echo       - 纯后台运行
echo       - 核心功能完整
echo.
echo   [3] 退出
echo.

set /p choice=请输入选项 (1/2/3): 

if "%choice%"=="1" goto fast_build
if "%choice%"=="2" goto full_build
if "%choice%"=="3" goto end
echo 无效选项，请重新运行
pause
goto end

:fast_build
echo.
echo ========================================
echo 开始编译 Fyne GUI 版本...
echo ========================================
echo.
call BUILD_WITH_FYNE.bat
goto end

:full_build
echo.
echo ========================================
echo 开始编译系统托盘版本...
echo ========================================
echo.
call BUILD_NO_WEBVIEW2_FIXED.bat
goto end

:end
echo.
echo 完成！
