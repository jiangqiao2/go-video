# 多阶段构建
# 第一阶段：构建
FROM golang:1.24-alpine AS builder
# 使用 Go Alpine 基础镜像作为构建阶段，体积小、包含所需工具
WORKDIR /app
# 构建阶段工作目录设为 /app，后续 COPY/BUILD 均基于此
RUN apk add --no-cache git ca-certificates tzdata
# 安装构建期依赖：git（拉取依赖）、证书（HTTPS）、时区（日志时间正确）
ENV GOPROXY=https://goproxy.cn,direct
# 配置 Go 代理以提升 go mod 下载速度；如有内网代理可按需替换
COPY transcode-service/go.mod transcode-service/go.sum ./
# 仅复制 go.mod/go.sum 以充分利用 Docker 层缓存，加速依赖下载
COPY transcode-service/proto/ ./proto/
COPY upload-service/proto/ ../upload-service/proto/
COPY video-service/proto/ ../video-service/proto/
# 复制各服务的 proto 目录（与 go.mod 中 replace 路径一致），确保本地模块依赖可解析
RUN go mod download
# 预拉取依赖，便于后续源码变更仍可复用缓存
COPY transcode-service/ .
# 复制业务源码到构建容器（包含 main.go、DDD 目录、配置等）
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o transcode-service .
# 构建静态二进制（关闭 CGO）以提升可移植性；目标平台为 Linux

# 第二阶段：提供 NVENC 能力的 FFmpeg
FROM jrottenberg/ffmpeg:6.1-nvidia AS ffmpeg

# 第三阶段：运行
FROM nvidia/cuda:12.4.0-runtime-ubuntu22.04
# 运行阶段使用带 NVIDIA 运行时的 Ubuntu 基础镜像，便于使用 NVENC
# 使用国内镜像源+重试，避免偶发 502
RUN sed -i 's|http://archive.ubuntu.com/ubuntu|http://mirrors.aliyun.com/ubuntu|g' /etc/apt/sources.list
RUN apt-get update -o Acquire::Retries=3 && apt-get install -y \
    ca-certificates \
    tzdata \
    curl \
    libgomp1 \
    libexpat1 && \
    rm -rf /var/lib/apt/lists/*
# 安装运行期依赖：证书/时区/curl/FFmpeg依赖(gomp/expat)
RUN ln -sf /usr/share/zoneinfo/Asia/Shanghai /etc/localtime && \
    echo 'Asia/Shanghai' > /etc/timezone
# 设置容器时区为 Asia/Shanghai，日志时间与本地一致
RUN addgroup --gid 1001 appgroup && \
    adduser --uid 1001 --gid 1001 --disabled-password --gecos "" appuser
# 创建非 root 用户（appuser）提升安全性
WORKDIR /app
# 运行阶段工作目录设为 /app
COPY --from=builder /app/transcode-service .
# 从构建阶段复制可执行文件
COPY --from=builder /app/configs ./configs
# 从构建阶段复制配置目录（包含开发/生产配置示例）
COPY --from=builder /app/entrypoint.sh .
# 复制并授权入口脚本（负责启动服务）
COPY --from=ffmpeg /usr/local/bin/ffmpeg /usr/local/bin/ffprobe /usr/local/bin/
COPY --from=ffmpeg /usr/local/lib/ /usr/local/lib/
# 拷贝支持 NVENC 的 FFmpeg 及依赖库
ARG CONFIG_PATH=/app/configs/config.dev.yaml
ENV CONFIG_PATH=${CONFIG_PATH}
ENV LD_LIBRARY_PATH=/usr/local/lib
# 默认读取开发配置；可通过 -e CONFIG_PATH=... 覆盖为其他环境配置
RUN chmod +x entrypoint.sh && \
    mkdir -p /var/log/transcode-service && \
    mkdir -p /tmp/transcode && \
    mkdir -p /tmp/transcode/transcoded && \
    chown -R appuser:appgroup /var/log/transcode-service && \
    chown -R appuser:appgroup /tmp/transcode && \
    chown -R appuser:appgroup /app
# 预创建日志与临时目录，并将权限下放到非 root 用户
RUN ffmpeg -version
# 构建时验证 FFmpeg 是否可用，便于及早发现镜像问题
USER appuser
EXPOSE 8083 9092
# 暴露端口：HTTP(8083) 健康检查与接口；gRPC(9092)
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD curl -f http://localhost:8083/health || exit 1
# 配置健康检查，确保容器未就绪时不会被认为健康
CMD ["./entrypoint.sh"]
# 容器入口：执行入口脚本启动转码服务
