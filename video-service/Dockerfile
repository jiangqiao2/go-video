# 多阶段构建
# 第一阶段：构建
FROM golang:1.24-alpine AS builder

# 设置工作目录
WORKDIR /app

# 安装必要的包
RUN apk add --no-cache git ca-certificates tzdata
# 使用国内 Go 模块代理，避免访问 proxy.golang.org 失败
ENV GOPROXY=https://goproxy.cn,direct

# 复制 go.mod / go.sum（使用仓库根作为构建上下文）
COPY video-service/go.mod video-service/go.sum ./

# 复制本服务 proto（匹配 replace video-service/proto => ./proto）
COPY video-service/proto/ ./proto/

# 复制通知服务 proto（匹配 replace notification-service/proto => ../notification-service/proto）
COPY notification-service/proto/ /notification-service/proto/

# 下载依赖（此时 go.mod 中的 replace 路径已经在镜像里存在）
RUN go mod download

# 复制源代码
COPY video-service/ .

# 构建二进制
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o video-service .

# 第二阶段：运行
FROM alpine:latest

# 安装必要的包
RUN apk --no-cache add \
    ca-certificates \
    tzdata \
    curl \
    su-exec

# 设置时区
RUN ln -sf /usr/share/zoneinfo/Asia/Shanghai /etc/localtime && \
    echo 'Asia/Shanghai' > /etc/timezone

# 创建非 root 用户
RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

# 工作目录
WORKDIR /app

# 复制二进制与配置
COPY --from=builder /app/video-service .
COPY --from=builder /app/configs ./configs
# 拷贝 JWT 证书，便于 RSA 校验
COPY private.pem public.pem /app/certs/

# 默认配置路径（可通过 -e CONFIG_PATH 覆盖）
ARG CONFIG_PATH=/app/configs/config.dev.yaml
ENV CONFIG_PATH=${CONFIG_PATH}

# 目录权限
RUN chown -R appuser:appgroup /app

# 切换到非 root 用户
USER appuser

# 暴露端口（HTTP / gRPC）
EXPOSE 8083 9094

# 健康检查
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD curl -f http://localhost:8083/health || exit 1

# 启动命令
CMD ["./video-service"]

