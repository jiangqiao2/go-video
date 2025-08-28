# ==========================
# 构建阶段
# ==========================
FROM golang:1.23-alpine AS builder

# 设置工作目录
WORKDIR /app

# 复制 go.mod 和 go.sum，下载依赖
COPY go.mod go.sum ./
RUN go mod download

# 复制源代码
COPY . .

# 构建可执行文件
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main ./cmd/main.go

# ==========================
# 运行阶段
# ==========================
FROM alpine:latest

# 替换国内 APK 镜像源，避免 SSL / 网络问题
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.tencent.com/g' /etc/apk/repositories \
    && apk --no-cache add ca-certificates

# 设置工作目录
WORKDIR /root/

# 从构建阶段拷贝二进制文件
COPY --from=builder /app/main .

# 暴露端口
EXPOSE 8082

# 启动应用
CMD ["./main"]
