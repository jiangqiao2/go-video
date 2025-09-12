#!/bin/bash

# 转码服务启动脚本

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 日志函数
log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

log_debug() {
    echo -e "${BLUE}[DEBUG]${NC} $1"
}

# 检查Docker和Docker Compose
check_dependencies() {
    log_info "检查依赖..."
    
    if ! command -v docker &> /dev/null; then
        log_error "Docker 未安装，请先安装 Docker"
        exit 1
    fi
    
    if ! command -v docker-compose &> /dev/null; then
        log_error "Docker Compose 未安装，请先安装 Docker Compose"
        exit 1
    fi
    
    log_info "依赖检查完成"
}

# 创建必要的目录
create_directories() {
    log_info "创建必要的目录..."
    
    mkdir -p logs
    mkdir -p /tmp/transcode
    mkdir -p deployments/nginx/conf.d
    mkdir -p deployments/prometheus
    mkdir -p deployments/grafana/provisioning
    
    log_info "目录创建完成"
}

# 创建Nginx配置
create_nginx_config() {
    log_info "创建Nginx配置..."
    
    cat > deployments/nginx/nginx.conf << 'EOF'
events {
    worker_connections 1024;
}

http {
    upstream scheduler {
        server scheduler:8082;
    }
    
    server {
        listen 80;
        server_name localhost;
        
        location / {
            proxy_pass http://scheduler;
            proxy_set_header Host $host;
            proxy_set_header X-Real-IP $remote_addr;
            proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
            proxy_set_header X-Forwarded-Proto $scheme;
        }
        
        location /health {
            proxy_pass http://scheduler/health;
        }
    }
}
EOF
    
    log_info "Nginx配置创建完成"
}

# 创建Prometheus配置
create_prometheus_config() {
    log_info "创建Prometheus配置..."
    
    cat > deployments/prometheus/prometheus.yml << 'EOF'
global:
  scrape_interval: 15s
  evaluation_interval: 15s

scrape_configs:
  - job_name: 'transcode-scheduler'
    static_configs:
      - targets: ['scheduler:8082']
    metrics_path: '/metrics'
    scrape_interval: 10s
    
  - job_name: 'transcode-workers'
    static_configs:
      - targets: ['worker-1:8083', 'worker-2:8083']
    metrics_path: '/metrics'
    scrape_interval: 10s
EOF
    
    log_info "Prometheus配置创建完成"
}

# 构建镜像
build_images() {
    log_info "构建Docker镜像..."
    
    log_debug "构建调度器镜像..."
    docker build -f Dockerfile.scheduler -t transcode-scheduler:latest .
    
    log_debug "构建Worker镜像..."
    docker build -f Dockerfile.worker -t transcode-worker:latest .
    
    log_info "镜像构建完成"
}

# 启动服务
start_services() {
    log_info "启动服务..."
    
    # 启动基础设施服务
    log_debug "启动基础设施服务..."
    docker-compose up -d mysql redis minio rabbitmq
    
    # 等待基础设施服务就绪
    log_debug "等待基础设施服务就绪..."
    sleep 30
    
    # 启动应用服务
    log_debug "启动应用服务..."
    docker-compose up -d scheduler
    
    # 等待调度器就绪
    log_debug "等待调度器就绪..."
    sleep 10
    
    # 启动Worker服务
    log_debug "启动Worker服务..."
    docker-compose up -d worker-1 worker-2
    
    # 启动监控服务（可选）
    if [ "$1" = "--with-monitoring" ]; then
        log_debug "启动监控服务..."
        docker-compose up -d nginx prometheus grafana
    fi
    
    log_info "服务启动完成"
}

# 检查服务状态
check_services() {
    log_info "检查服务状态..."
    
    # 等待服务完全启动
    sleep 5
    
    # 检查调度器健康状态
    if curl -f http://localhost:8082/health &> /dev/null; then
        log_info "✓ 调度器服务正常"
    else
        log_warn "✗ 调度器服务异常"
    fi
    
    # 显示服务状态
    docker-compose ps
}

# 显示访问信息
show_access_info() {
    log_info "服务访问信息:"
    echo -e "${BLUE}调度器 API:${NC} http://localhost:8082"
    echo -e "${BLUE}健康检查:${NC} http://localhost:8082/health"
    echo -e "${BLUE}API 文档:${NC} http://localhost:8082/swagger/index.html"
    echo -e "${BLUE}MySQL:${NC} localhost:3307 (用户: transcode_user, 密码: transcode_password)"
    echo -e "${BLUE}Redis:${NC} localhost:6380"
    echo -e "${BLUE}MinIO:${NC} http://localhost:9003 (用户: minioadmin, 密码: minioadmin123)"
    echo -e "${BLUE}RabbitMQ:${NC} http://localhost:15673 (用户: admin, 密码: admin123)"
    
    if docker-compose ps | grep -q nginx; then
        echo -e "${BLUE}Nginx:${NC} http://localhost:80"
    fi
    
    if docker-compose ps | grep -q prometheus; then
        echo -e "${BLUE}Prometheus:${NC} http://localhost:9091"
    fi
    
    if docker-compose ps | grep -q grafana; then
        echo -e "${BLUE}Grafana:${NC} http://localhost:3001 (用户: admin, 密码: admin123)"
    fi
}

# 停止服务
stop_services() {
    log_info "停止服务..."
    docker-compose down
    log_info "服务已停止"
}

# 清理服务
clean_services() {
    log_info "清理服务和数据..."
    docker-compose down -v --remove-orphans
    docker system prune -f
    log_info "清理完成"
}

# 显示日志
show_logs() {
    if [ -n "$1" ]; then
        docker-compose logs -f "$1"
    else
        docker-compose logs -f
    fi
}

# 显示帮助信息
show_help() {
    echo "转码服务管理脚本"
    echo ""
    echo "用法: $0 [命令] [选项]"
    echo ""
    echo "命令:"
    echo "  start [--with-monitoring]  启动服务（可选启动监控）"
    echo "  stop                       停止服务"
    echo "  restart                    重启服务"
    echo "  status                     查看服务状态"
    echo "  logs [service]             查看日志（可指定服务名）"
    echo "  clean                      清理服务和数据"
    echo "  build                      构建镜像"
    echo "  help                       显示帮助信息"
    echo ""
    echo "示例:"
    echo "  $0 start                   # 启动基础服务"
    echo "  $0 start --with-monitoring # 启动服务并包含监控"
    echo "  $0 logs scheduler          # 查看调度器日志"
    echo "  $0 logs worker-1           # 查看Worker-1日志"
}

# 主函数
main() {
    case "$1" in
        start)
            check_dependencies
            create_directories
            create_nginx_config
            create_prometheus_config
            build_images
            start_services "$2"
            check_services
            show_access_info
            ;;
        stop)
            stop_services
            ;;
        restart)
            stop_services
            sleep 2
            start_services "$2"
            check_services
            show_access_info
            ;;
        status)
            docker-compose ps
            ;;
        logs)
            show_logs "$2"
            ;;
        clean)
            clean_services
            ;;
        build)
            build_images
            ;;
        help|--help|-h)
            show_help
            ;;
        "")
            show_help
            ;;
        *)
            log_error "未知命令: $1"
            show_help
            exit 1
            ;;
    esac
}

# 执行主函数
main "$@"