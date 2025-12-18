# 安装依赖（容器内执行）
pip install flask-cors

# 修改app.py
from flask import Flask, request, jsonify
from flask_cors import CORS  # 新增
import time

app = Flask(__name__)
CORS(app) 

# 模拟不同服务的处理逻辑
SERVICE_HANDLERS = {
    "S1": lambda data: f"AR/VR服务处理结果：{data}（模拟渲染完成）",
    "S2": lambda data: f"智能交通服务分析结果：{data}（模拟流量分析完成）",
    "S3": lambda data: f"大模型服务回答：{data}（模拟大模型生成完成）"
}

@app.route('/run', methods=['POST'])
def run_service():
    try:
        data = request.get_json()
        service_id = data.get('service_id')
        input_data = data.get('input')
        
        # 模拟处理耗时
        time.sleep(0.1)
        
        # 调用对应服务的处理逻辑
        result = SERVICE_HANDLERS.get(service_id, lambda x: f"未知服务处理结果：{x}")(input_data)
        
        return jsonify({
            "success": True,
            "result": result,
            "msg": "处理成功"
        })
    except Exception as e:
        return jsonify({
            "success": False,
            "result": "",
            "msg": str(e)
        }), 500

@app.route('/metrics', methods=['GET'])
def get_metrics():
    # 原有的metrics接口（返回gas/cost/delay等）
    import os
    service_id = os.environ.get('SERVICE_ID', 'S1')
    metrics = {
        "S1": {"service_id": "S1", "gas": 3, "cost": 4, "csci_id": "172.17.0.8:5000", "delay": 8},
        "S2": {"service_id": "S2", "gas": 2, "cost": 5, "csci_id": "172.17.0.9:5000", "delay": 12},
        "S3": {"service_id": "S3", "gas": 1, "cost": 2, "csci_id": "172.17.0.10:5000", "delay": 15}
    }
    return jsonify(metrics[service_id])

if __name__ == '__main__':
    app.run(host='0.0.0.0', port=5000)