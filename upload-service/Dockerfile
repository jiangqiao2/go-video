# ==========================
# 构建阶段
# ==========================
FROM golang:1.23-alpine AS builder
# 构建阶段
# 设置工作目录
# ==========================


COPY go.mod go.sum ./
FROM golang:1.23-alpine AS builder

# 构建可执行文件
COPY . .
FROM alpine:latest

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main ./cmd/main.go
COPY . .
    && apk --no-cache add ca-certificates

# 设置工作目录
WORKDIR /root/

# 从构建阶段拷贝二进制文件
COPY --from=builder /app/main .

# 暴露端口
EXPOSE 8082

# 启动应用
# 从构建阶段拷贝二进制文件
    && apk --no-cache add ca-certificates

# 设置工作目录
WORKDIR /root/

# 从构建阶段拷贝二进制文件
COPY --from=builder /app/main .

# 暴露端口
EXPOSE 8082

# 启动应用
# 从构建阶段拷贝二进制文件
