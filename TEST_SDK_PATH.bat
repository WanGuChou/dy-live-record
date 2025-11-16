@echo off
REM 测试 Windows SDK 短路径

echo 检测 Windows SDK 短路径...
echo.

REM 显示 Program Files (x86) 的短路径
echo Program Files (x86) 的短路径:
dir /x "C:\" | findstr "Program Files"
echo.

REM 显示 Windows Kits 的短路径
echo Windows Kits 的短路径:
dir /x "C:\Program Files (x86)" | findstr "Windows Kits"
echo.

REM 显示 10 目录的短路径
echo 10 目录的短路径:
dir /x "C:\Program Files (x86)\Windows Kits" | findstr " 10"
echo.

echo 请根据上面的输出，找到正确的短路径名
echo 示例: PROGRA~2\WI459C~1\10
pause
