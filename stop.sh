#!/bin/bash

# 停止CMAS系统相关进程

echo "==================================="
echo "    停止CMAS系统相关进程"
echo "==================================="

# 定义服务名称
services=(
    "go run cmd/platform/main.go"
    "go run cmd/c-sma/main.go"
    "go run cmd/c-ps/main.go"
)

# 遍历服务并终止进程
for service in "${services[@]}"; do
    pids=$(pgrep -f "$service")
    if [ -n "$pids" ]; then
        echo "正在停止服务: $service"
        kill -9 $pids
        echo "服务已停止: $service"
    else
        echo "未找到运行中的服务: $service"
    fi
done

# 检查是否有残留进程
sleep 1
echo "检查是否有残留进程..."
for service in "${services[@]}"; do
    pids=$(pgrep -f "$service")
    if [ -n "$pids" ]; then
        echo "残留进程未停止: $service (PID: $pids)"
    else
        echo "所有相关进程已成功停止: $service"
    fi
done

echo "==================================="
echo "    所有服务已停止"
echo "==================================="