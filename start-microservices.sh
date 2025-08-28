#!/bin/bash

# 微服务启动脚本

echo "🚀 启动微服务架构..."

# 检查Docker是否运行
if ! docker info > /dev/null 2>&1; then
    echo "❌ Docker未运行，请先启动Docker"
    exit 1
fi

# 停止现有服务
echo "🛑 停止现有服务..."
docker-compose -f docker-compose.microservices.yml down

# 清理旧的镜像（可选）
read -p "是否清理旧的Docker镜像？(y/N): " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
    echo "🧹 清理旧镜像..."
    docker-compose -f docker-compose.microservices.yml down --rmi all
fi

# 构建并启动服务
echo "🔨 构建并启动微服务..."
docker-compose -f docker-compose.microservices.yml up --build -d

# 等待服务启动
echo "⏳ 等待服务启动..."
sleep 10

# 检查服务状态
echo "📊 检查服务状态..."
docker-compose -f docker-compose.microservices.yml ps

echo ""
echo "✅ 微服务启动完成！"
echo ""
echo "🌐 服务访问地址："
echo "  API网关:     http://localhost:8080"
echo "  用户服务:     http://localhost:8081"
echo "  上传服务:     http://localhost:8082"
echo "  视频服务:     http://localhost:8083"
echo "  MinIO控制台:  http://localhost:9001 (minioadmin/minioadmin)"
echo ""
echo "🔍 健康检查："
echo "  curl http://localhost:8080/health"
echo "  curl http://localhost:8081/health"
echo "  curl http://localhost:8082/health"
echo "  curl http://localhost:8083/health"
echo ""
echo "📝 查看日志："
echo "  docker-compose -f docker-compose.microservices.yml logs -f [service-name]"
echo ""
echo "🛑 停止服务："
echo "  docker-compose -f docker-compose.microservices.yml down"