@echo off
REM Safe version - No Chinese characters in critical parts
chcp 65001 >nul 2>&1
setlocal enabledelayedexpansion

cls
echo ========================================
echo Douyin Live Monitor - Fyne GUI Build
echo ========================================
echo.
echo Features:
echo   [YES] Modern GUI (Fyne)
echo   [YES] Cross-platform
echo   [YES] No Windows SDK required
echo   [YES] Pure Go implementation
echo.
pause

REM Step 1: Pack browser extension
echo.
echo [1/4] Packing browser extension...
if exist browser-monitor\pack.bat (
    cd browser-monitor
    call pack.bat
    cd ..
    if exist server-go\assets\browser-monitor.zip (
        echo [OK] Plugin packed successfully
    ) else (
        echo [WARNING] Plugin pack may have failed, continuing...
    )
) else (
    echo [SKIP] pack.bat not found
)

REM Step 2: Download Fyne dependencies
echo.
echo [2/4] Downloading Fyne dependencies...
cd server-go
go mod download
if %ERRORLEVEL% EQU 0 (
    echo [OK] Dependencies downloaded
) else (
    echo [WARNING] Dependencies download may have failed, trying to continue...
)

REM Step 3: Check GCC
echo.
echo [3/4] Checking CGO environment...
where gcc >nul 2>&1
if %ERRORLEVEL% NEQ 0 (
    echo [ERROR] GCC not found! Fyne requires CGO support.
    echo.
    echo Please install MinGW-w64:
    echo   Method 1: choco install mingw
    echo   Method 2: https://www.mingw-w64.org/
    echo.
    pause
    exit /b 1
) else (
    gcc --version
    echo [OK] GCC is installed
)

REM Step 4: Build server-go
echo.
echo [4/4] Building server-go...
echo [INFO] Starting build (estimated 2-3 minutes)...
echo.

set CGO_ENABLED=1
go build -v -o dy-live-monitor.exe .

if %ERRORLEVEL% EQU 0 (
    echo.
    echo ========================================
    echo SUCCESS! Build completed!
    echo ========================================
    echo.
    echo Output: server-go\dy-live-monitor.exe
    echo.
    echo How to run:
    echo   1. Debug mode (skip License):
    echo      cd server-go
    echo      copy config.debug.json config.json
    echo      .\dy-live-monitor.exe
    echo.
    echo   2. Normal mode:
    echo      cd server-go
    echo      .\dy-live-monitor.exe
    echo.
) else (
    echo.
    echo ========================================
    echo BUILD FAILED
    echo ========================================
    echo.
    echo Possible reasons:
    echo   1. GCC not installed (Fyne needs CGO)
    echo   2. Network issues (cannot download dependencies)
    echo   3. Insufficient disk space
    echo.
    echo Solutions:
    echo   1. Install MinGW-w64: https://www.mingw-w64.org/
    echo   2. Set Go proxy: set GOPROXY=https://goproxy.cn,direct
    echo   3. View detailed error: go build -v -x -o dy-live-monitor.exe .
    echo.
    echo Or use system tray version (no Fyne required):
    echo   .\BUILD_NO_WEBVIEW2_FIXED.bat
    echo.
)

cd ..
pause
