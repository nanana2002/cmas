from flask import Flask, request, jsonify, send_file
from flask_cors import CORS  # 解决跨域
import os
import uuid
from datetime import datetime

# 初始化Flask应用
app = Flask(__name__)
CORS(app)  # 允许所有跨域请求（适配前端调用）

# 配置：图片上传目录（容器内路径）
UPLOAD_FOLDER = "/app/uploads"
os.makedirs(UPLOAD_FOLDER, exist_ok=True)
app.config['UPLOAD_FOLDER'] = UPLOAD_FOLDER

# ========== 1. 指标接口（供CSMA拉取） ==========
@app.route('/metrics', methods=['GET'])
def get_metrics():
    """返回S3服务的指标（固定值，模拟大模型轻量服务）"""
    return jsonify({
        "service_id": "S3",
        "gas": 1,          # 可用实例数
        "cost": 2,         # 服务成本
        "csci_id": f"{os.environ.get('SERVICE_IP', '172.17.0.10')}:5000",  # 容器IP:端口
        "delay": 15        # 延迟（ms）
    })

# ========== 2. 核心服务接口（供用户调用） ==========
@app.route('/run', methods=['POST'])
def run_service():
    """处理用户请求：文本回显 / 图片回显（兼容JSON/FormData）"""
    try:
        # ------------- 处理FormData请求（图片上传）-------------
        if 'file' in request.files:
            file = request.files['file']
            service_id = request.form.get('service_id', 'S3')
            input_text = request.form.get('input', '')
            
            if file.filename == '':
                return jsonify({"success": False, "msg": "未选择图片文件"}), 400
            
            # 保存图片（生成唯一文件名）
            file_ext = file.filename.rsplit('.', 1)[1].lower()
            filename = f"{uuid.uuid4()}.{file_ext}"
            file_path = os.path.join(app.config['UPLOAD_FOLDER'], filename)
            file.save(file_path)
            
            # 图片回显逻辑：返回图片访问URL + 原文件名
            return jsonify({
                "success": True,
                "result": {
                    "image_url": f"http://{os.environ.get('SERVICE_IP', '172.18.0.2')}:5000/uploads/{filename}",
                    "original_filename": file.filename,
                    "input_text": input_text,
                    "message": "图片已接收，返回原图片"
                },
                "msg": "图片请求处理成功",
                "service_id": service_id,
                "timestamp": datetime.now().strftime("%Y-%m-%d %H:%M:%S")
            })
        
        # ------------- 处理JSON请求（文本）-------------
        elif request.is_json:
            data = request.get_json()
            input_text = data.get("input", "")
            service_id = data.get("service_id", "S3")
            
            # 文本回显逻辑
            return jsonify({
                "success": True,
                "result": input_text,  # 输入什么返回什么
                "msg": "文本请求处理成功",
                "service_id": service_id,
                "timestamp": datetime.now().strftime("%Y-%m-%d %H:%M:%S")
            })
        
        # ------------- 无效请求 -------------
        else:
            return jsonify({"success": False, "msg": "不支持的请求格式（仅支持JSON文本/FormData图片上传）"}), 400
    
    except Exception as e:
        return jsonify({"success": False, "msg": f"服务处理失败：{str(e)}"}), 500

# ========== 3. 图片访问接口（供前端查看回显图片） ==========
@app.route('/uploads/<filename>', methods=['GET'])
def get_uploaded_image(filename):
    """返回上传的图片文件"""
    file_path = os.path.join(app.config['UPLOAD_FOLDER'], filename)
    if os.path.exists(file_path):
        return send_file(file_path)
    else:
        return jsonify({"success": False, "msg": "图片不存在"}), 404

# 启动服务（监听所有网卡，端口5000）
if __name__ == '__main__':
    app.run(host='0.0.0.0', port=5000, debug=True)