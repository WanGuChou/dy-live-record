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
echo   [1] 快速编译 - 系统托盘版本（推荐）
echo       - 编译时间: 30秒
echo       - 无需 Windows SDK
echo       - 核心功能完整
echo       - 使用系统托盘
echo.
echo   [2] 完整编译 - 包含图形界面
echo       - 编译时间: 5-10分钟
echo       - 需要 Windows SDK
echo       - 包含图形主界面
echo       - 文件较大
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
echo 开始快速编译...
echo ========================================
echo.
call BUILD_NO_WEBVIEW2_FIXED.bat
goto end

:full_build
echo.
echo ========================================
echo 开始完整编译...
echo ========================================
echo.
echo [检查] 检测 Windows SDK...

REM 检查 SDK 是否存在
set SDK_PATH=C:\Program Files (x86)\Windows Kits\10\Include\10.0.26100.0
if exist "%SDK_PATH%\winrt\EventToken.h" (
    echo [OK] Windows SDK 已安装
    echo.
    call FIX_CGO_PATHS.bat
) else (
    echo [ERROR] 未找到 Windows SDK
    echo.
    echo 请选择:
    echo   [1] 安装 Windows SDK 后重试
    echo   [2] 改用快速编译（无需 SDK）
    echo.
    set /p fallback=请输入选项 (1/2): 
    if "!fallback!"=="2" (
        echo.
        echo 切换到快速编译...
        call BUILD_NO_WEBVIEW2_FIXED.bat
    ) else (
        echo.
        echo 请先安装 Windows SDK:
        echo https://developer.microsoft.com/en-us/windows/downloads/windows-sdk/
        pause
    )
)
goto end

:end
echo.
echo 完成！
