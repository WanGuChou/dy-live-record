#!/bin/bash
# WebView2 测试脚本（Linux/Mac - 仅用于语法检查）

echo "========================================"
echo "WebView2 测试工具"
echo "========================================"
echo ""

# 检查操作系统
if [[ "$OSTYPE" != "msys" && "$OSTYPE" != "win32" ]]; then
    echo "⚠️  警告: WebView2 仅支持 Windows 平台"
    echo "在当前平台上只能运行语法检查，无法运行实际测试"
    echo ""
fi

echo "选择操作:"
echo "1. 运行语法检查"
echo "2. 查看测试代码"
echo "3. 退出"
echo ""

read -p "请输入选项 (1-3): " choice

case $choice in
    1)
        echo ""
        echo "========================================"
        echo "运行语法检查"
        echo "========================================"
        echo ""
        go build ./internal/ui/webview_test.go && echo "✅ 语法检查通过"
        ;;
    2)
        echo ""
        echo "========================================"
        echo "查看测试代码"
        echo "========================================"
        echo ""
        cat internal/ui/webview_test.go | less
        ;;
    3)
        echo "退出"
        exit 0
        ;;
    *)
        echo "无效的选项"
        ;;
esac

echo ""
