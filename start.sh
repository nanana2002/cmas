#!/bin/bash

# CMAS系统一键启动脚本

echo "==================================="
echo "    CMAS系统一键启动脚本"
echo "==================================="

# 检查Docker是否运行
if ! command -v docker &> /dev/null; then
    echo "错误: Docker未安装或未在PATH中"
    exit 1
fi

if ! docker info &> /dev/null; then
    echo "错误: Docker守护进程未运行，请启动Docker"
    exit 1
fi

# 检查Go是否安装
if ! command -v go &> /dev/null; then
    echo "错误: Go未安装或未在PATH中"
    exit 1
fi

echo "1. 检查并创建Docker网络..."
if ! docker network ls | grep -q "cmas-network"; then
    echo "创建cmas-network网络..."
    docker network create --driver bridge --subnet=172.18.0.0/16 --gateway=172.18.0.1 cmas-network
    if [ $? -ne 0 ]; then
        echo "错误: 创建Docker网络失败"
        exit 1
    fi
else
    echo "Docker网络cmas-network已存在"
fi

echo "2. 检查并构建Docker镜像..."
if ! docker images | grep -q "cmas-service"; then
    echo "构建cmas-service:v1镜像..."
    cd docker-sites && docker build -t cmas-service:v1 . && cd ..
    if [ $? -ne 0 ]; then
        echo "错误: 构建Docker镜像失败"
        exit 1
    fi
else
    echo "Docker镜像cmas-service:v1已存在"
fi

echo "3. 检查并启动S1/S2/S3容器服务..."

# 检查容器是否已存在
for i in {1..3}; do
    if [ "$(docker ps -q -f name=cmas-site-$i)" ]; then
        echo "容器cmas-site-$i已在运行"
    elif [ "$(docker ps -aq -f name=cmas-site-$i)" ]; then
        echo "启动已存在的容器cmas-site-$i..."
        docker start cmas-site-$i
    else
        echo "启动新的容器cmas-site-$i..."
        # 创建上传目录
        mkdir -p services/s$i-service/uploads
        # 启动容器
        docker run -d \
            --name cmas-site-$i \
            --network cmas-network \
            --ip 172.18.0.$((7+i)) \
            -p $((5000+i)):5000 \
            -v $(pwd)/services/s$i-service/uploads:/app/uploads \
            -e SERVICE_ID="S$i" \
            cmas-service:v1
    fi
done

# 等待容器启动
echo "等待容器启动..."
sleep 5

# 检查容器状态
for i in {1..3}; do
    if [ "$(docker ps -q -f name=cmas-site-$i)" ]; then
        echo "✓ 容器cmas-site-$i 启动成功"
    else
        echo "✗ 容器cmas-site-$i 启动失败"
    fi
done

echo "4. 启动Go后端服务..."

# 启动Platform服务 (8081端口)
echo "启动Platform服务 (端口8081)..."
if pgrep -f "go run cmd/platform/main.go" > /dev/null; then
    echo "Platform服务已在运行"
else
    nohup go run cmd/platform/main.go > platform.log 2>&1 &
    PLATFORM_PID=$!
    disown $PLATFORM_PID
    echo "Platform服务已启动，PID: $PLATFORM_PID"
fi

# 等待Platform服务启动
sleep 3

# 启动CSMA服务 (8083端口)
echo "启动CSMA服务 (端口8083)..."
if pgrep -f "go run cmd/c-sma/main.go" > /dev/null; then
    echo "CSMA服务已在运行"
else
    nohup go run cmd/c-sma/main.go > csma.log 2>&1 &
    CSMA_PID=$!
    disown $CSMA_PID
    echo "CSMA服务已启动，PID: $CSMA_PID"
fi

# 等待CSMA服务启动
sleep 3

# 启动CPS服务 (8084端口)
echo "启动CPS服务 (端口8084)..."
if pgrep -f "go run cmd/c-ps/main.go" > /dev/null; then
    echo "CPS服务已在运行"
else
    nohup go run cmd/c-ps/main.go > cps.log 2>&1 &
    CPS_PID=$!
    disown $CPS_PID
    echo "CPS服务已启动，PID: $CPS_PID"
fi

# 等待服务完全启动
sleep 5

echo "==================================="
echo "    服务启动完成"
echo "==================================="

# 显示服务状态
echo "服务状态:"
echo "- Platform: http://localhost:8081 (前端托管)"
echo "- CSMA: http://localhost:8083 (指标采集)"
echo "- CPS: http://localhost:8084 (节点调度)"
echo ""
echo "前端页面:"
echo "- 服务提供者: http://localhost:8081/provider"
echo "- 用户端: http://localhost:8081/user"
echo "- 服务展示页面: http://localhost:8081/show"
echo ""
echo "Docker容器状态:"
docker ps --filter "name=cmas-site-" --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}"

echo ""
echo "日志文件:"
echo "- platform.log: Platform服务日志"
echo "- csma.log: CSMA服务日志" 
echo "- cps.log: CPS服务日志"
echo ""
echo "注意: 如果需要停止服务，请使用 kill 命令终止对应的进程"