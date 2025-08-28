# ==========================
# 构建阶段
# ==========================
FROM golang:1.23-alpine AS builder

WORKDIR /app

# 复制 go.mod 和 go.sum
COPY go.mod go.sum ./
RUN go mod download

# 复制源代码
COPY . .

# 构建应用
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main ./cmd/main.go

# ==========================
# 运行阶段
# ==========================
FROM alpine:latest

# 使用国内 APK 镜像源加速，避免下载慢或 SSL 报错
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.tencent.com/g' /etc/apk/repositories \
    && apk --no-cache add ca-certificates

WORKDIR /root/

# 从构建阶段复制二进制文件
COPY --from=builder /app/main .

# 暴露端口
EXPOSE 8081

# 运行应用
CMD ["./main"]
