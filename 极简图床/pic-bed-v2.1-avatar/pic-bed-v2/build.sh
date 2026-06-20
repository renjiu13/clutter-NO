#!/bin/bash
set -e

echo "=== 极轻量图床 - 多架构编译 ==="

mkdir -p bin

echo "编译 amd64..."
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o bin/pic-bed-linux-amd64 ./cmd/pic-bed
echo "✓ amd64 完成"

echo "编译 arm64..."
CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -ldflags="-s -w" -o bin/pic-bed-linux-arm64 ./cmd/pic-bed
echo "✓ arm64 完成"

echo "编译 armv7 (玩客云)..."
CGO_ENABLED=0 GOOS=linux GOARCH=arm GOARM=7 go build -ldflags="-s -w" -o bin/pic-bed-linux-armv7 ./cmd/pic-bed
echo "✓ armv7 完成"

echo ""
echo "=== 编译完成 ==="
ls -lh bin/
