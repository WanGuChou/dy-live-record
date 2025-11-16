@echo off
chcp 65001 >nul 2>&1
echo ===================================
echo Batch File Encoding Fix Tool
echo ===================================
echo.
echo This tool will help fix batch file encoding issues.
echo.
echo Detecting system encoding...
chcp
echo.
echo Recommended: Use UTF-8 (Code Page 65001)
echo.
pause
