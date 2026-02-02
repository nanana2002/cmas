#!/bin/bash

# CMAS System Test Setup Script

echo "Setting up CMAS system for testing..."

# Create required directories if they don't exist
mkdir -p temp/code
mkdir -p web/provider
mkdir -p web/user
mkdir -p services/s1-service/uploads
mkdir -p services/s2-service/uploads
mkdir -p services/s3-service/uploads

# Create a sample Python service file for testing
cat > temp/code/sample_service.py << 'EOF'
from flask import Flask, request, jsonify, send_from_directory
import time
import os
import uuid
from werkzeug.utils import secure_filename

app = Flask(__name__)

# 设置上传文件夹
UPLOAD_FOLDER = '/app/uploads'
os.makedirs(UPLOAD_FOLDER, exist_ok=True)
app.config['UPLOAD_FOLDER'] = UPLOAD_FOLDER

# 模拟不同服务的处理逻辑
SERVICE_HANDLERS = {
    "S1": lambda data: f"AR/VR服务处理结果：{data}（模拟渲染完成）",
    "S2": lambda data: f"智能交通服务分析结果：{data}（模拟流量分析完成）",
    "S3": lambda data: f"大模型服务回答：{data}（模拟大模型生成完成）"
}

@app.route('/run', methods=['POST'])
def run_service():
    try:
        # 检查是否是文件上传请求
        if 'file' in request.files:
            # 处理文件上传
            file = request.files['file']
            service_id = request.form.get('service_id')
            input_text = request.form.get('input', '')
            
            if file.filename == '':
                return jsonify({
                    "success": False,
                    "result": "",
                    "msg": "没有选择文件"
                }), 400
            
            if file:
                filename = secure_filename(file.filename)
                unique_filename = f"{uuid.uuid4()}_{filename}"
                file_path = os.path.join(app.config['UPLOAD_FOLDER'], unique_filename)
                file.save(file_path)
                
                # 返回图片URL和相关信息
                result = {
                    "input_text": input_text,
                    "original_filename": filename,
                    "unique_filename": unique_filename,
                    "image_url": f"http://0.0.0.0:5000/uploads/{unique_filename}",
                    "service_id": service_id
                }
                
                return jsonify({
                    "success": True,
                    "result": result,
                    "msg": "文件上传成功"
                })
        else:
            # 处理普通文本请求
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

@app.route('/uploads/<filename>')
def uploaded_file(filename):
    return send_from_directory(app.config['UPLOAD_FOLDER'], filename)

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
EOF

echo "Setup complete!"
echo "Directories and sample service created."
echo "To run the system:"
echo "1. Build the Docker image: cd docker-sites && docker build -t cmas-service:v1 ."
echo "2. Start the services: docker run commands for S1, S2, S3"
echo "3. Run the Go modules: go run cmd/platform/main.go, go run cmd/c-sma/main.go, go run cmd/c-ps/main.go"