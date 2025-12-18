# 基础镜像：Python 3.9
FROM python:3.9-slim

# 设置工作目录
WORKDIR /app

# 复制依赖文件并安装
COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

# 复制服务代码
COPY s3_service.py .

# 创建上传目录
RUN mkdir -p /app/uploads

# 暴露端口（容器内端口5000）
EXPOSE 5000

# 环境变量：设置服务IP（部署时可覆盖）
ENV SERVICE_IP=172.17.0.10

# 启动服务
CMD ["python", "s3_service.py"]