@echo off
echo ========================================
echo Building All Components (FIXED)
echo ========================================

set ERROR_COUNT=0

REM Step 1: Pack browser-monitor
echo.
echo [1/5] Packing browser-monitor...
cd browser-monitor
call pack.bat
if %ERRORLEVEL% NEQ 0 (
    echo [FAILED] browser-monitor pack failed
    set /a ERROR_COUNT+=1
) else (
    echo [SUCCESS] browser-monitor packed successfully
)
cd ..

REM Step 2: Fix server-go dependencies
echo.
echo [2/5] Fixing server-go dependencies...
cd server-go
if exist go.sum del go.sum
go mod tidy
if %ERRORLEVEL% NEQ 0 (
    echo [WARNING] go mod tidy had issues, continuing...
)
cd ..

REM Step 3: Build server-go
echo.
echo [3/5] Building server-go...
cd server-go
set CGO_ENABLED=1
go build -v -o dy-live-monitor.exe .
if %ERRORLEVEL% NEQ 0 (
    echo [FAILED] server-go build failed
    set /a ERROR_COUNT+=1
    cd ..
    goto check_active
) else (
    echo [SUCCESS] server-go built successfully
)
cd ..

:check_active
REM Step 4: Fix server-active dependencies
echo.
echo [4/5] Fixing server-active dependencies...
cd server-active
if exist go.sum del go.sum
go mod tidy
if %ERRORLEVEL% NEQ 0 (
    echo [WARNING] go mod tidy had issues, continuing...
)
cd ..

REM Step 5: Build server-active
echo.
echo [5/5] Building server-active...
cd server-active
go build -v -o dy-live-license-server.exe .
if %ERRORLEVEL% NEQ 0 (
    echo [FAILED] server-active build failed
    set /a ERROR_COUNT+=1
) else (
    echo [SUCCESS] server-active built successfully
)
cd ..

echo.
echo ========================================
echo Build Summary
echo ========================================
if %ERROR_COUNT% EQU 0 (
    echo Status: ALL BUILDS SUCCEEDED!
    echo.
    echo Output files:
    if exist "server-go\dy-live-monitor.exe" echo   - server-go\dy-live-monitor.exe
    if exist "server-go\assets\browser-monitor.zip" echo   - server-go\assets\browser-monitor.zip
    if exist "server-active\dy-live-license-server.exe" echo   - server-active\dy-live-license-server.exe
) else (
    echo Status: %ERROR_COUNT% BUILD(S) FAILED
    echo.
    echo Please check:
    echo   1. MinGW-w64 is installed: gcc --version
    echo   2. CGO is enabled: go env CGO_ENABLED
    echo   3. Network connection for go mod download
    exit /b 1
)
echo ========================================
