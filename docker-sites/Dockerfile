FROM python:3.9-slim
WORKDIR /app
COPY app.py .
RUN pip install --no-cache-dir flask
EXPOSE 5000

# 获取宿主机用户ID（构建时传入，或运行时指定）
ARG USER_ID=1000
ARG GROUP_ID=1000

# 创建与宿主机相同ID的用户
RUN groupadd -g $GROUP_ID appuser && \
    useradd -m -u $USER_ID -g $GROUP_ID appuser

# 切换到appuser用户
USER appuser

CMD ["python", "app.py"]
