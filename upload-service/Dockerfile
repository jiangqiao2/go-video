# ==========================
# 构建阶段
# ==========================
FROM golang:1.23-alpine AS builder

# 设置工作目录
WORKDIR /app


COPY go.mod go.sum ./
COPY . .

# 构建可执行文件
COPY . .
FROM alpine:latest

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main ./cmd/main.go
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.tencent.com/g' /etc/apk/repositories \
    && apk --no-cache add ca-certificates

# 设置工作目录
WORKDIR /root/

# 从构建阶段拷贝二进制文件
COPY --from=builder /app/main .

# 暴露端口
EXPOSE 8082

# 启动应用
# 从构建阶段拷贝二进制文件
