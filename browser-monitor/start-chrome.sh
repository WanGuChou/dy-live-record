#!/bin/bash
# CDP Monitor - Chrome启动脚本（隐藏调试提示）
# 使用方法：chmod +x start-chrome.sh && ./start-chrome.sh

echo "========================================"
echo "CDP Monitor - Chrome 启动脚本"
echo "========================================"
echo ""

# 检测操作系统
if [[ "$OSTYPE" == "darwin"* ]]; then
    # macOS
    CHROME_PATH="/Applications/Google Chrome.app/Contents/MacOS/Google Chrome"
elif [[ "$OSTYPE" == "linux-gnu"* ]]; then
    # Linux
    if command -v google-chrome &> /dev/null; then
        CHROME_PATH="google-chrome"
    elif command -v google-chrome-stable &> /dev/null; then
        CHROME_PATH="google-chrome-stable"
    elif command -v chromium-browser &> /dev/null; then
        CHROME_PATH="chromium-browser"
    else
        echo "[错误] 未找到Chrome/Chromium"
        echo ""
        echo "请安装Chrome："
        echo "  Ubuntu/Debian: sudo apt install google-chrome-stable"
        echo "  Fedora: sudo dnf install google-chrome-stable"
        echo "  Arch: yay -S google-chrome"
        echo ""
        exit 1
    fi
else
    echo "[错误] 不支持的操作系统：$OSTYPE"
    exit 1
fi

# macOS特殊处理
if [[ "$OSTYPE" == "darwin"* ]]; then
    if [ ! -f "$CHROME_PATH" ]; then
        echo "[错误] 未找到Chrome："
        echo "  路径：$CHROME_PATH"
        echo ""
        echo "请先安装Chrome："
        echo "  https://www.google.com/chrome/"
        echo ""
        exit 1
    fi
fi

echo "[1/3] 关闭现有Chrome进程..."
if [[ "$OSTYPE" == "darwin"* ]]; then
    killall "Google Chrome" 2>/dev/null
else
    killall chrome chromium-browser 2>/dev/null
fi
sleep 2

echo "[2/3] 启动Chrome（隐藏调试提示）..."
if [[ "$OSTYPE" == "darwin"* ]]; then
    "$CHROME_PATH" --silent-debugger-extension-api &
else
    $CHROME_PATH --silent-debugger-extension-api &
fi

echo "[3/3] 完成！"
echo ""
echo "✅ Chrome已启动，不会显示'正在调试此浏览器'提示"
echo ""
echo "📌 提示："
echo "  - 请在Chrome中加载CDP Monitor插件"
echo "  - 配置并启用监控功能"
echo "  - 查看详细说明：HIDE_DEBUGGER_BANNER.md"
echo ""
