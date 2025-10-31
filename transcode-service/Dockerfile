# 多阶段构建
# 第一阶段：构建
FROM golang:1.24-alpine AS builder

# 设置工作目录
WORKDIR /app

# 安装必要的包
RUN apk add --no-cache git ca-certificates tzdata

# 复制go mod文件
COPY transcode-service/go.mod transcode-service/go.sum ./

# 复制proto目录到正确位置（匹配replace路径）
COPY proto/ ../proto/

# 下载依赖
RUN go mod download

# 复制源代码
COPY transcode-service/ .

# 构建应用
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o transcode-service .

# 第二阶段：运行
FROM alpine:latest

# 安装必要的包，包括FFmpeg
RUN apk --no-cache add \
    ca-certificates \
    tzdata \
    curl \
    ffmpeg

# 设置时区
RUN ln -sf /usr/share/zoneinfo/Asia/Shanghai /etc/localtime
RUN echo 'Asia/Shanghai' > /etc/timezone

# 创建非root用户
RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

# 设置工作目录
WORKDIR /app

# 从构建阶段复制二进制文件
COPY --from=builder /app/transcode-service .

# 创建必要的目录
RUN mkdir -p /var/log/transcode-service && \
    mkdir -p /tmp/transcode && \
    chown -R appuser:appgroup /var/log/transcode-service && \
    chown -R appuser:appgroup /tmp/transcode && \
    chown -R appuser:appgroup /app

# 验证FFmpeg安装
RUN ffmpeg -version

# 切换到非root用户
USER appuser

# 暴露端口
EXPOSE 8083 50051

# 健康检查
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD curl -f http://localhost:8083/health || exit 1

# 启动命令
CMD ["./transcode-service"]