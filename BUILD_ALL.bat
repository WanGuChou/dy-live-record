@echo off
echo ========================================
echo Building All Components
echo ========================================

set ERROR_COUNT=0

REM Build server-go
echo.
echo [1/3] Building server-go...
cd server-go
call build.bat
if %ERRORLEVEL% NEQ 0 (
    echo [FAILED] server-go build failed
    set /a ERROR_COUNT+=1
) else (
    echo [SUCCESS] server-go built successfully
)
cd ..

REM Pack browser-monitor
echo.
echo [2/3] Packing browser-monitor...
cd browser-monitor
call pack.bat
if %ERRORLEVEL% NEQ 0 (
    echo [FAILED] browser-monitor pack failed
    set /a ERROR_COUNT+=1
) else (
    echo [SUCCESS] browser-monitor packed successfully
)
cd ..

REM Build server-active
echo.
echo [3/3] Building server-active...
cd server-active
call build.bat
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
    echo   - server-go/dy-live-monitor.exe
    echo   - server-go/assets/browser-monitor.zip
    echo   - server-active/dy-live-license-server.exe
) else (
    echo Status: %ERROR_COUNT% BUILD(S) FAILED
    exit /b 1
)
echo ========================================
