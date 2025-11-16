@echo off
echo ========================================
echo Packing Browser Monitor Plugin
echo ========================================

set PLUGIN_NAME=browser-monitor
set OUTPUT_DIR=..\server-go\assets

REM Create assets directory
if not exist "%OUTPUT_DIR%" mkdir "%OUTPUT_DIR%"

REM Pack plugin to zip
echo Packing plugin files...
powershell -Command "Compress-Archive -Path manifest.json,background.js,popup.html,popup.js,icons -DestinationPath '%OUTPUT_DIR%\%PLUGIN_NAME%.zip' -Force"

if %ERRORLEVEL% EQU 0 (
    echo.
    echo ========================================
    echo Pack Success!
    echo Output: %OUTPUT_DIR%\%PLUGIN_NAME%.zip
    echo ========================================
) else (
    echo.
    echo ========================================
    echo Pack Failed!
    echo ========================================
    exit /b 1
)
