@echo off
REM WebView2 测试脚本
REM 用于快速测试 WebView2 功能

echo ========================================
echo WebView2 测试工具
echo ========================================
echo.

REM 检查是否在 Windows 上
if not "%OS%"=="Windows_NT" (
    echo 错误: WebView2 仅支持 Windows 平台
    pause
    exit /b 1
)

echo 选择测试模式:
echo 1. 运行单元测试
echo 2. 启动演示程序（浏览器模式）
echo 3. 运行所有测试
echo 4. 运行性能基准测试
echo 5. 退出
echo.

set /p choice="请输入选项 (1-5): "

if "%choice%"=="1" goto :unit_tests
if "%choice%"=="2" goto :demo
if "%choice%"=="3" goto :all_tests
if "%choice%"=="4" goto :benchmark
if "%choice%"=="5" goto :end
goto :invalid_choice

:unit_tests
echo.
echo ========================================
echo 运行 WebView2 单元测试
echo ========================================
echo.
go test -v ./internal/ui -run TestWebView2
goto :end

:demo
echo.
echo ========================================
echo 启动 WebView2 演示程序
echo ========================================
echo.
echo 提示: 程序将启动本地服务器
echo 浏览器访问: http://localhost:18889
echo 按 Ctrl+C 停止服务器
echo.
cd cmd\webview_demo
go run main.go
cd ..\..
goto :end

:all_tests
echo.
echo ========================================
echo 运行所有 WebView2 测试
echo ========================================
echo.

echo [1/6] 基础窗口测试...
go test -v ./internal/ui -run TestWebView2BasicWindow

echo.
echo [2/6] HTML 加载测试...
go test -v ./internal/ui -run TestWebView2WithHTML

echo.
echo [3/6] 通信测试...
go test -v ./internal/ui -run TestWebView2Communication

echo.
echo [4/6] 本地服务器测试...
go test -v ./internal/ui -run TestWebView2WithLocalServer

echo.
echo [5/6] 多窗口测试...
go test -v ./internal/ui -run TestWebView2MultipleWindows

echo.
echo [6/6] 性能测试...
go test -v ./internal/ui -run TestWebView2Performance

echo.
echo ========================================
echo 所有测试完成！
echo ========================================
goto :end

:benchmark
echo.
echo ========================================
echo 运行性能基准测试
echo ========================================
echo.
go test -v ./internal/ui -bench BenchmarkWebView2Creation -benchmem
goto :end

:invalid_choice
echo.
echo 错误: 无效的选项
echo.
goto :end

:end
echo.
pause
