# 多阶段构建
# 第一阶段：构建二进制
FROM golang:1.24-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git ca-certificates tzdata

# 复制模块定义（使用仓库根上下文时按路径复制）
COPY upload-service/go.mod upload-service/go.sum ./
# 复制proto目录到replace路径
COPY proto/ ../proto/

RUN go mod download

# 复制源码（upload-service目录）
COPY upload-service/ .

# 构建可执行文件
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o upload-service .

# 第二阶段：运行时
FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata curl su-exec

# 创建非root用户
RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

WORKDIR /app

# 从构建阶段拷贝可执行文件
COPY --from=builder /app/upload-service .

# 默认配置路径（可通过环境变量覆盖）
ENV CONFIG_PATH=/app/configs/config.dev.yaml

# 暴露HTTP与gRPC端口
EXPOSE 8082 9093

# 以非root运行
USER appuser

ENTRYPOINT ["./upload-service"]
