@echo off
echo ========================================
echo Building server-active (License Service)
echo ========================================

REM Set module name
set MODULE_NAME=dy-live-license

REM Set output binary name
set OUTPUT_NAME=dy-live-license-server.exe

REM Build
echo Building...
go build -o %OUTPUT_NAME% .

if %ERRORLEVEL% EQU 0 (
    echo.
    echo ========================================
    echo Build Success!
    echo Output: %OUTPUT_NAME%
    echo ========================================
) else (
    echo.
    echo ========================================
    echo Build Failed!
    echo ========================================
    exit /b 1
)
