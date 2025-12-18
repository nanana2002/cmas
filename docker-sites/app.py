from flask import Flask, request, jsonify
import os

app = Flask(__name__)

# 从环境变量读取配置（启动时传入）
SERVICE_ID = os.getenv("SERVICE_ID", "S1")
GAS = int(os.getenv("GAS", 2))
COST = int(os.getenv("COST", 3))
DELAY = int(os.getenv("DELAY", 10))
CSCI_ID = os.getenv("CSCI_ID", "172.17.0.10:5000")
SERVICE_NAME = os.getenv("SERVICE_NAME", "轻量计算服务")

# 1. 指标接口（供C-SMA拉取）
@app.route("/metrics")
def metrics():
    return jsonify({
        "service_id": SERVICE_ID,
        "gas": GAS,
        "cost": COST,
        "csci_id": CSCI_ID,
        "delay": DELAY,
        "service_name": SERVICE_NAME
    })

# 2. 业务接口（接收请求并返回结果）
@app.route("/run", methods=["POST"])
def run():
    data = request.json
    result = f"服务[{SERVICE_ID}]处理成功：输入={data}，延迟={DELAY}ms"
    return jsonify({
        "success": True,
        "result": result,
        "service_id": SERVICE_ID,
        "csci_id": CSCI_ID
    })

if __name__ == "__main__":
    app.run(host="0.0.0.0", port=5000)
