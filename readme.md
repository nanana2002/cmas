# CMAS系统完整部署手册（S1/S2/S3全服务+前端/后端）
## 一、项目概述
本项目是一套容器化微服务调度系统（CMAS），支持多节点服务（S1/S2/S3）的指标监控、最优节点调度及用户端交互，核心能力如下：
- **核心模块**：
  - Platform（8080端口）：托管前端页面（服务提供者/用户端），提供服务注册与页面访问能力；
  - CSMA（8083端口）：自动拉取所有容器服务（S1/S2/S3）的运行指标（gas/cost/delay）；
  - CPS（8084端口）：基于服务指标决策最优服务节点，返回给用户端调用；
- **业务服务**：
  - S1/S2/S3（Docker容器）：均为轻量级微服务，支持文本回显、图片上传回显（输入即输出），仅IP/端口不同；
- **核心特性**：
  - 容器化部署：S1/S2/S3均通过Docker隔离部署，支持固定IP/自动IP；
  - 指标自动采集：CSMA适配Docker自定义网络，自动获取所有容器IP及指标；
  - 跨网络访问：前端自动将容器内网IP转为宿主机IP，解决浏览器访问容器超时问题；
  - 数据持久化：图片上传后持久化到宿主机目录，容器重启不丢失。

## 二、环境准备
### 前置依赖（必须安装）
| 软件       | 版本要求 | 验证命令          | 安装参考文档                                  |
|------------|----------|-------------------|-----------------------------------------------|
| Docker     | 20.10+   | `docker --version` | https://docs.docker.com/engine/install/ubuntu/ |
| Go         | 1.19+    | `go version`      | https://go.dev/doc/install                    |
| Python     | 3.9+     | `python3 --version`| https://www.python.org/downloads/             |
| 网络       | 服务器公网/内网IP | `ifconfig`/`ip addr` | -                                             |

### 环境初始化命令
```bash
# 更新系统依赖（Ubuntu）
sudo apt update && sudo apt upgrade -y

# 安装Docker依赖（若未安装）
sudo apt install -y apt-transport-https ca-certificates curl software-properties-common
curl -fsSL https://download.docker.com/linux/ubuntu/gpg | sudo apt-key add -
sudo add-apt-repository "deb [arch=amd64] https://download.docker.com/linux/ubuntu $(lsb_release -cs) stable"
sudo apt update && sudo apt install -y docker-ce docker-ce-cli containerd.io

# 配置Docker免sudo（可选，避免后续命令重复加sudo）
sudo usermod -aG docker $USER
newgrp docker

# 安装Go（以1.21为例）
wget https://go.dev/dl/go1.21.0.linux-amd64.tar.gz
sudo rm -rf /usr/local/go && sudo tar -C /usr/local -xzf go1.21.0.linux-amd64.tar.gz
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
source ~/.bashrc
```

## 三、项目目录结构（完整）
```
cmas-cats-go/
├── cmd/                # Go后端核心模块
│   ├── platform/       # Platform模块（8080）- 前端托管
│   │   └── main.go     # 启动入口
│   ├── c-sma/          # CSMA模块（8083）- 指标采集
│   │   └── main.go     # 启动入口
│   └── c-ps/           # CPS模块（8084）- 节点调度
│       └── main.go     # 启动入口
├── config/             # 全局配置目录
│   └── config.go       # 容器IP获取、网络配置等核心配置
├── models/             # 数据模型目录
│   └── service.go      # 服务指标、节点信息模型
├── utils/              # 工具类目录
│   └── http.go         # HTTP请求工具
├── web/                # 前端页面目录
│   ├── provider/       # 服务提供者页面（指标监控）
│   │   └── index.html  # 页面入口
│   └── user/           # 用户端页面（服务调用）
│       └── index.html  # 页面入口
└── services/           # 业务服务Docker部署目录（S1/S2/S3）
    ├── s1-service/     # S1服务
    │   ├── s_service.py   # 服务代码（与S3通用）
    │   ├── requirements.txt # Python依赖
    │   ├── Dockerfile      # Docker构建文件
    │   └── uploads/        # 图片持久化目录
    ├── s2-service/     # S2服务（同S1目录结构）
    └── s3-service/     # S3服务（同S1目录结构）
```

## 四、完整部署步骤（S1/S2/S3+全模块）
### 步骤1：创建完整项目目录
```bash
# 创建根目录
mkdir -p ~/cmas-cats-go
cd ~/cmas-cats-go

# 创建Go模块目录
mkdir -p cmd/platform cmd/c-sma cmd/c-ps config models utils

# 创建前端目录
mkdir -p web/provider web/user

# 创建S1/S2/S3服务目录（统一结构）
mkdir -p services/s1-service services/s2-service services/s3-service
mkdir -p services/s1-service/uploads services/s2-service/uploads services/s3-service/uploads
```

### 步骤2：编写核心代码（文件内容需提前准备）
| 目录路径                          | 文件作用                                                                 |
|-----------------------------------|--------------------------------------------------------------------------|
| `config/config.go`                | 自动获取容器IP（兼容Docker自定义网络）、服务端口配置                     |
| `models/service.go`               | 定义Service结构体（ID/IP/Port/Gas/Cost/Delay等指标）                     |
| `cmd/platform/main.go`            | 启动HTTP服务，托管web目录下的前端页面，支持跨域                          |
| `cmd/c-sma/main.go`               | 定时拉取S1/S2/S3的/metrics接口，更新服务指标                             |
| `cmd/c-ps/main.go`                | 提供/select接口，根据指标返回最优服务节点（S1/S2/S3）                    |
| `web/provider/index.html`         | 服务提供者页面，展示所有服务（S1/S2/S3）的实时指标                       |
| `web/user/index.html`             | 用户端页面，支持选择S1/S2/S3服务，文本输入/图片上传调用                  |
| `services/s*/s_service.py`        | 业务服务核心代码（/metrics返回指标，/run支持文本/图片回显）              |
| `services/s*/requirements.txt`    | Python依赖（flask）                                                      |
| `services/s*/Dockerfile`          | Docker构建文件（Python镜像、代码复制、端口暴露、权限配置）               |

### 步骤3：部署S1/S2/S3容器服务（统一流程）
#### 3.1 构建通用业务服务镜像（S1/S2/S3共用）
```bash
# 进入任意服务目录（以S3为例，构建后S1/S2直接复用镜像）
cd ~/cmas-cats-go/services/s3-service

# 构建Docker镜像（命名为cmas-service:v1，S1/S2/S3共用）
# 【命令原因】：将Python服务打包为容器镜像，实现环境隔离，S1/S2/S3仅IP/端口不同，无需重复构建
docker build -t cmas-service:v1 .

# 【易错点1】：报错“no such file or directory”
# 解决：确保当前目录有Dockerfile、requirements.txt、s_service.py文件
# 【易错点2】：pip安装超时
# 解决：修改requirements.txt，添加国内源（如pip install -i https://pypi.tuna.tsinghua.edu.cn/simple flask）
```

#### 3.2 创建Docker自定义网络（避免IP冲突）
```bash
# 创建带子网的自定义网络（S1/S2/S3均接入此网络）
# 【命令原因】：默认Docker网络无子网，无法指定固定IP；自定义子网（172.18.0.0/16）可给S1/S2/S3分配固定IP，便于CMAS识别
# 【克服问题】：解决“Pool overlaps with other one on this address space”子网冲突问题
docker network create \
  --driver bridge \
  --subnet=172.18.0.0/16 \
  --gateway=172.18.0.1 \
  cmas-network
```

#### 3.3 启动S1/S2/S3容器（固定IP+端口映射）
```bash
# 通用函数：启动服务容器（避免重复命令）
start_service() {
  local SERVICE_NAME=$1    # 容器名（s1/s2/s3）
  local IP=$2              # 固定IP
  local HOST_PORT=$3       # 宿主机映射端口
  local SERVICE_DIR=$4     # 宿主机上传目录

  # 赋予目录权限（解决容器写入权限问题）
  # 【克服问题】：解决“Operation not permitted”权限错误，确保容器能写入宿主机目录
  sudo chown -R $USER:$USER ${SERVICE_DIR}
  chmod 777 ${SERVICE_DIR}

  # 启动容器
  docker run -d \
    --name cmas-${SERVICE_NAME} \
    --network cmas-network \
    --ip ${IP} \
    -p ${HOST_PORT}:5000 \
    -v ${SERVICE_DIR}:/app/uploads \
    -e SERVICE_IP=${IP} \
    cmas-service:v1

  # 验证启动状态
  echo "启动${SERVICE_NAME}服务，容器状态："
  docker ps | grep cmas-${SERVICE_NAME}
}

# 启动S1（IP:172.18.0.8，宿主机端口:5001）
start_service "site-1" "172.18.0.8" "5001" "~/cmas-cats-go/services/s1-service/uploads"

# 启动S2（IP:172.18.0.9，宿主机端口:5002）
start_service "site-2" "172.18.0.9" "5002" "~/cmas-cats-go/services/s2-service/uploads"

# 启动S3（IP:172.18.0.10，宿主机端口:5003）
start_service "site-3" "172.18.0.10" "5003" "~/cmas-cats-go/services/s3-service/uploads"

# 【易错点1】：报错“invalid endpoint settings”
# 解决：确保网络已配置子网（步骤3.2），若仍报错则删除--ip参数（自动分配IP）
# 【易错点2】：挂载路径错误（如s1_service而非s1-service）
# 解决：检查-v参数路径，确保宿主机目录名称与代码一致
```

#### 3.4 验证S1/S2/S3容器服务
```bash
# 验证容器IP（确保S1/S2/S3 IP正确）
echo "S1 IP: $(docker inspect -f '{{range .NetworkSettings.Networks}}{{.IPAddress}}{{end}}' cmas-site-1)"
echo "S2 IP: $(docker inspect -f '{{range .NetworkSettings.Networks}}{{.IPAddress}}{{end}}' cmas-site-2)"
echo "S3 IP: $(docker inspect -f '{{range .NetworkSettings.Networks}}{{.IPAddress}}{{end}}' cmas-site-3)"

# 验证服务指标接口（S1/S2/S3通用）
echo "验证S1指标："
curl http://192.168.235.48:5001/metrics  # 替换为实际服务器IP
echo "验证S2指标："
curl http://192.168.235.48:5002/metrics
echo "验证S3指标："
curl http://192.168.235.48:5003/metrics

# 验证文本回显（以S3为例）
curl -X POST http://192.168.235.48:5003/run \
  -H "Content-Type: application/json" \
  -d '{"service_id":"S3","input":"测试文本"}'

# 验证图片上传（以S3为例，替换为实际图片路径）
curl -X POST http://192.168.235.48:5003/run \
  -F "service_id=S3" \
  -F "file=@/home/$USER/test.jpg" \
  -F "input=测试图片"

# 【预期结果】：
# 1. 指标接口返回JSON（包含gas/cost/delay）；
# 2. 文本请求返回输入的文本；
# 3. 图片请求返回image_url，且宿主机services/s3-service/uploads目录出现图片文件；
# 4. 【克服问题】：解决“图片未保存到宿主机”问题，确保挂载路径+权限正确
```

### 步骤4：启动CMAS核心后端模块
#### 4.1 启动Platform模块（前端托管）
```bash
cd ~/cmas-cats-go
# 启动Platform（8080端口）
# 【命令原因】：托管web目录下的前端页面，支持跨域访问，用户通过8080端口访问前端
# 【后台运行】：若需后台运行，添加nohup和&（nohup go run cmd/platform/main.go &）
go run cmd/platform/main.go
```

#### 4.2 启动CSMA模块（指标采集）
```bash
cd ~/cmas-cats-go
# 启动CSMA（8083端口）
# 【命令原因】：定时拉取S1/S2/S3的/metrics接口，更新服务指标，为CPS提供决策依据
# 【克服问题】：解决“容器IP获取失败”问题（config.go已适配自定义网络IP读取）
go run cmd/c-sma/main.go

# 【验证】：查看日志，确认S1/S2/S3 IP获取成功
# 预期日志：
# [CONFIG] 容器cmas-site-1 IP获取成功：172.18.0.8
# [CONFIG] 容器cmas-site-2 IP获取成功：172.18.0.9
# [CONFIG] 容器cmas-site-3 IP获取成功：172.18.0.10
```

#### 4.3 启动CPS模块（最优节点调度）
```bash
cd ~/cmas-cats-go
# 启动CPS（8084端口）
# 【命令原因】：提供/select接口，接收服务ID后返回最优节点（基于gas/cost/delay）
go run cmd/c-ps/main.go

# 【验证】：测试CPS接口
curl -X POST http://192.168.235.48:8084/select \
  -H "Content-Type: application/json" \
  -d '{"service_id":"S3","max_accept_cost":10,"max_accept_delay":30}'

# 【预期结果】：返回最优节点的CSCIID（如172.18.0.10:5000）
```

### 步骤5：访问前端页面（完整功能验证）
#### 5.1 服务提供者页面（指标监控）
- 访问地址：`http://<服务器IP>:8080/provider`
- 【验证内容】：
  1. 页面显示S1/S2/S3的服务信息（IP、端口、gas、cost、delay）；
  2. 指标实时更新（CSMA定时采集）；
  3. 无404/500错误（Platform模块正常托管页面）。

#### 5.2 用户端页面（服务调用）
- 访问地址：`http://<服务器IP>:8080/user`
- 【操作流程】：
  1. 选择服务：下拉框选择S1/S2/S3；
  2. 输入文本：填写任意文本（如“测试S1服务”）；
  3. （可选）上传图片：选择本地图片文件；
  4. 点击“提交请求”；
- 【验证内容】：
  1. 文本请求：返回输入的文本，页面显示“请求成功”；
  2. 图片请求：显示图片预览，宿主机对应服务的uploads目录出现图片文件；
  3. 节点调度：CPS返回最优节点，前端自动将容器IP转为宿主机IP（解决“ERR_CONNECTION_TIMED_OUT”超时问题）；
  4. 【克服问题】：解决“前端上传图片未保存”“图片无法显示”问题。

## 五、核心问题与解决方案（全量）
| 问题现象                                  | 根本原因                                  | 解决方案                                                                 |
|-------------------------------------------|-------------------------------------------|--------------------------------------------------------------------------|
| CSMA启动提示“容器IP获取失败”              | 默认IP读取逻辑仅支持Docker默认网络        | 修改config/config.go，遍历容器所有网络IP（兼容自定义网络）               |
| 前端访问容器IP超时（ERR_CONNECTION_TIMED_OUT） | 容器内网IP仅宿主机可访问，浏览器无法穿透  | 前端代码将容器IP替换为宿主机IP+映射端口（如172.18.0.10:5000→192.168.235.48:5003） |
| 图片上传后未保存到宿主机                  | 挂载路径错误/目录权限为root所有           | 1. 修正-v挂载路径（s1-service而非s1_service）；2. chown -R $USER:$USER 目录 |
| Docker网络子网冲突（Pool overlaps）       | 自定义子网与默认172.17.0.0/16冲突         | 改用非冲突子网（如172.18.0.0/16）                                       |
| 容器启动报错“invalid endpoint settings”  | 自定义网络未配置子网，无法指定固定IP       | 1. 给网络配置子网；2. 删除--ip参数，让Docker自动分配IP                  |
| 前端图片无法显示                          | 图片URL使用容器内网IP                     | 前端替换image_url中的容器IP为宿主机IP+映射端口                          |
| pip安装依赖超时                           | 国外源访问慢                              | requirements.txt中添加国内PyPI源（如-i https://pypi.tuna.tsinghua.edu.cn/simple） |

## 六、运维与管理（全服务）
### 1. 重启服务
```bash
# 重启单个容器（如S3）
docker restart cmas-site-3

# 重启所有业务容器
docker restart cmas-site-1 cmas-site-2 cmas-site-3

# 重启CMAS后端模块（需重新执行go run命令）
```

### 2. 查看日志
```bash
# 查看容器日志（实时）
docker logs -f cmas-site-3

# 查看Go模块日志（若后台运行）
tail -f nohup.out  # 需启动时添加nohup（如nohup go run cmd/c-sma/main.go &）
```

### 3. 数据备份
```bash
# 备份所有服务的上传图片
mkdir -p ~/cmas-backup
cp -r ~/cmas-cats-go/services/s1-service/uploads ~/cmas-backup/
cp -r ~/cmas-cats-go/services/s2-service/uploads ~/cmas-backup/
cp -r ~/cmas-cats-go/services/s3-service/uploads ~/cmas-backup/
```

### 4. 清理无用资源
```bash
# 停止并删除所有业务容器
docker stop cmas-site-1 cmas-site-2 cmas-site-3
docker rm cmas-site-1 cmas-site-2 cmas-site-3

# 删除自定义网络
docker network rm cmas-network

# 删除无用镜像
docker rmi cmas-service:v1
```

## 七、功能验证清单（全量）
| 验证项                | 验证方法                                  | 预期结果                                  |
|-----------------------|-------------------------------------------|-------------------------------------------|
| S1/S2/S3容器启动      | `docker ps`                               | 三个容器状态均为Up                         |
| 容器IP获取            | `docker inspect`                          | S1:172.18.0.8、S2:172.18.0.9、S3:172.18.0.10 |
| 服务指标接口          | `curl 服务器IP:5001/metrics`              | 返回JSON格式指标数据                      |
| CSMA指标采集          | 查看CSMA日志                              | 成功获取S1/S2/S3 IP及指标                 |
| CPS节点调度           | `curl 服务器IP:8084/select`               | 返回最优节点CSCIID                         |
| Platform页面托管      | 访问8080端口                              | 前端页面正常加载                          |
| 前端文本调用          | 用户端选择服务，输入文本提交              | 返回相同文本，请求成功                    |
| 前端图片调用          | 用户端上传图片提交                        | 图片预览正常，宿主机uploads目录有文件      |
| 图片持久化            | 重启容器后访问图片URL                     | 图片仍可显示，未丢失                      |