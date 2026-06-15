#!/bin/bash
set -e

echo "开始编译极轻量图床..."
echo "========================"

# Linux amd64（常规x86服务器）
echo "编译 Linux amd64..."
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o pic-bed-linux-amd64 .

# Linux arm64（ARM服务器/树莓派/飞牛NAS等）
echo "编译 Linux arm64..."
CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -ldflags="-s -w" -o pic-bed-linux-arm64 .

echo "========================"
echo "编译完成！生成文件："
ls -lh pic-bed-linux-*
echo ""
echo "提示：纯静态二进制，无任何依赖，上传对应架构文件直接运行即可"
