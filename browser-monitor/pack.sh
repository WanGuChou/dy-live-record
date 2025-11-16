#!/bin/bash

echo "========================================"
echo "Packing Browser Monitor Plugin"
echo "========================================"

PLUGIN_NAME="browser-monitor"
OUTPUT_DIR="../server-go/assets"

# Create assets directory
mkdir -p "$OUTPUT_DIR"

# Pack plugin to zip
echo "Packing plugin files..."
zip -r "$OUTPUT_DIR/$PLUGIN_NAME.zip" manifest.json background.js popup.html popup.js icons/

if [ $? -eq 0 ]; then
    echo ""
    echo "========================================"
    echo "Pack Success!"
    echo "Output: $OUTPUT_DIR/$PLUGIN_NAME.zip"
    echo "========================================"
else
    echo ""
    echo "========================================"
    echo "Pack Failed!"
    echo "========================================"
    exit 1
fi
