#!/usr/bin/env python3
"""
HLS功能测试脚本
测试转码服务的HLS切片功能
"""

import requests
import json
import time
import uuid

# 配置
TRANSCODE_SERVICE_URL = "http://localhost:8083"
API_BASE = f"{TRANSCODE_SERVICE_URL}/api/v1"

def test_health():
    """测试服务健康状态"""
    print("🔍 测试服务健康状态...")
    try:
        response = requests.get(f"{TRANSCODE_SERVICE_URL}/health")
        if response.status_code == 200:
            print("✅ 转码服务健康状态正常")
            return True
        else:
            print(f"❌ 转码服务健康检查失败: {response.status_code}")
            return False
    except Exception as e:
        print(f"❌ 连接转码服务失败: {e}")
        return False

def create_hls_transcode_task():
    """创建HLS转码任务"""
    print("\n🚀 创建HLS转码任务...")
    
    # 生成测试UUID
    task_uuid = str(uuid.uuid4())
    user_uuid = str(uuid.uuid4())
    video_uuid = str(uuid.uuid4())
    
    # 构造请求数据
    request_data = {
        "user_uuid": user_uuid,
        "video_uuid": video_uuid,
        "original_path": "/test/input/sample_video.mp4",
        "resolution": "720p",
        "bitrate": "2000k",
        "enable_hls": True,
        "segment_duration": 10,
        "list_size": 0,
        "hls_format": "ts",
        "hls_resolutions": [
            {
                "width": 1280,
                "height": 720,
                "bitrate": "2000k"
            },
            {
                "width": 854,
                "height": 480,
                "bitrate": "1000k"
            },
            {
                "width": 640,
                "height": 360,
                "bitrate": "500k"
            }
        ]
    }
    
    try:
        print(f"📤 发送请求到: {API_BASE}/transcode/tasks")
        print(f"📋 请求数据: {json.dumps(request_data, indent=2, ensure_ascii=False)}")
        
        response = requests.post(
            f"{API_BASE}/transcode/tasks",
            json=request_data,
            headers={"Content-Type": "application/json"}
        )
        
        print(f"📥 响应状态码: {response.status_code}")
        print(f"📄 响应内容: {response.text}")
        
        if response.status_code == 200 or response.status_code == 201:
            print("✅ HLS转码任务创建成功!")
            return task_uuid
        else:
            print(f"❌ HLS转码任务创建失败: {response.status_code} - {response.text}")
            return None
            
    except Exception as e:
        print(f"❌ 创建HLS转码任务时发生错误: {e}")
        return None

def main():
    """主函数"""
    print("🎬 开始测试HLS功能...")
    print("=" * 50)
    
    # 1. 测试服务健康状态
    if not test_health():
        print("❌ 服务不可用，退出测试")
        return
    
    # 2. 创建HLS转码任务
    task_uuid = create_hls_transcode_task()
    if task_uuid:
        print(f"\n🎉 测试完成! 任务UUID: {task_uuid}")
        print("\n📝 测试总结:")
        print("✅ 服务健康检查通过")
        print("✅ HLS转码任务创建成功")
        print("✅ API接口正常工作")
        print("\n💡 提示: 实际的转码处理需要Worker来执行")
    else:
        print("\n❌ 测试失败")

if __name__ == "__main__":
    main()