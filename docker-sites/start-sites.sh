#!/bin/bash
# 构建镜像
sudo docker build -t cmas-light-site:v1 .

# 启动Site1（S1：AR/VR服务）
sudo docker run -d --name cmas-site-1 -p 5001:5000 \
  -e SERVICE_ID="S1" \
  -e SERVICE_NAME="AR/VR轻量服务" \
  -e GAS=3 \
  -e COST=4 \
  -e DELAY=8 \
  -e CSCI_ID="$(sudo docker inspect -f '{{range .NetworkSettings.Networks}}{{.IPAddress}}{{end}}' cmas-site-1):5000" \
  cmas-light-site:v1

# 启动Site2（S2：智能交通服务）
sudo docker run -d --name cmas-site-2 -p 5002:5000 \
  -e SERVICE_ID="S2" \
  -e SERVICE_NAME="智能交通轻量服务" \
  -e GAS=2 \
  -e COST=5 \
  -e DELAY=12 \
  -e CSCI_ID="$(sudo docker inspect -f '{{range .NetworkSettings.Networks}}{{.IPAddress}}{{end}}' cmas-site-2):5000" \
  cmas-light-site:v1

# 启动Site3（S3：大模型服务）
sudo docker run -d --name cmas-site-3 -p 5003:5000 \
  -e SERVICE_ID="S3" \
  -e SERVICE_NAME="大模型轻量服务" \
  -e GAS=1 \
  -e COST=2 \
  -e DELAY=15 \
  -e CSCI_ID="$(sudo docker inspect -f '{{range .NetworkSettings.Networks}}{{.IPAddress}}{{end}}' cmas-site-3):5000" \
  cmas-light-site:v1

# 显示启动状态
echo "3个Docker Site启动完成："
sudo docker ps | grep cmas-site-
