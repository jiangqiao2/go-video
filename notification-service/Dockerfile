FROM golang:1.24-alpine AS builder
WORKDIR /app

# 基础依赖
RUN apk add --no-cache ca-certificates tzdata
# 使用国内 Go 模块代理，避免访问 proxy.golang.org 失败
ENV GOPROXY=https://goproxy.cn,direct

# 先复制 go.mod/go.sum 并拉依赖，利用缓存
COPY notification-service/go.mod notification-service/go.sum ./
# 复制本服务 proto（匹配 go.mod 中的 replace notification-service/proto => ./proto）
COPY notification-service/proto/ ./proto/
RUN go mod download

# 复制业务代码
COPY notification-service/ .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /bin/notification-service ./main.go

FROM alpine:3.19
WORKDIR /app

RUN apk add --no-cache tzdata && \
    ln -sf /usr/share/zoneinfo/Asia/Shanghai /etc/localtime && \
    echo 'Asia/Shanghai' > /etc/timezone

COPY --from=builder /bin/notification-service /usr/local/bin/notification-service
# 拷贝配置
COPY --from=builder /app/configs ./configs
# 拷贝 JWT 证书，便于使用 RSA/HS JWT 校验（与其他服务保持一致）
COPY private.pem public.pem /app/
COPY private.pem public.pem /app/certs/

ARG CONFIG_PATH=/app/configs/config.dev.yaml
ENV CONFIG_PATH=${CONFIG_PATH}
EXPOSE 8086
ENTRYPOINT ["notification-service"]
