#!/bin/bash

echo "========================================"
echo "Building All Components"
echo "========================================"

ERROR_COUNT=0

# Build server-go
echo ""
echo "[1/3] Building server-go..."
cd server-go
./build.sh
if [ $? -ne 0 ]; then
    echo "[FAILED] server-go build failed"
    ((ERROR_COUNT++))
else
    echo "[SUCCESS] server-go built successfully"
fi
cd ..

# Pack browser-monitor
echo ""
echo "[2/3] Packing browser-monitor..."
cd browser-monitor
./pack.sh
if [ $? -ne 0 ]; then
    echo "[FAILED] browser-monitor pack failed"
    ((ERROR_COUNT++))
else
    echo "[SUCCESS] browser-monitor packed successfully"
fi
cd ..

# Build server-active
echo ""
echo "[3/3] Building server-active..."
cd server-active
go build -o dy-live-license-server .
if [ $? -ne 0 ]; then
    echo "[FAILED] server-active build failed"
    ((ERROR_COUNT++))
else
    echo "[SUCCESS] server-active built successfully"
fi
cd ..

echo ""
echo "========================================"
echo "Build Summary"
echo "========================================"
if [ $ERROR_COUNT -eq 0 ]; then
    echo "Status: ALL BUILDS SUCCEEDED!"
    echo ""
    echo "Output files:"
    echo "  - server-go/dy-live-monitor"
    echo "  - server-go/assets/browser-monitor.zip"
    echo "  - server-active/dy-live-license-server"
else
    echo "Status: $ERROR_COUNT BUILD(S) FAILED"
    exit 1
fi
echo "========================================"
