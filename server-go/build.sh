#!/bin/bash

echo "========================================"
echo "  抖音直播监控系统 - 编译脚本 (Linux/macOS)"
echo "========================================"
echo ""

echo "[1/2] 清理旧的编译文件..."
rm -f dy-live-monitor dy-live-monitor-debug

echo "[2/2] 编译..."
go build -ldflags "-s -w" -o dy-live-monitor main.go
if [ $? -ne 0 ]; then
    echo "❌ 编译失败！"
    exit 1
fi
echo "✅ 编译成功: dy-live-monitor"

chmod +x dy-live-monitor

echo ""
echo "========================================"
echo "  编译完成！"
echo "========================================"
echo "可执行文件: ./dy-live-monitor"
echo ""
echo "运行方式: ./dy-live-monitor"
echo ""
