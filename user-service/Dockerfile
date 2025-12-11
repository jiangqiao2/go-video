FROM golang:1.24-alpine AS builder
WORKDIR /app

# 安装证书等基础依赖，避免拉取模块时缺组件
RUN apk add --no-cache ca-certificates tzdata

# 先复制 go.mod/go.sum 并拉依赖，利用缓存
COPY user-service/go.mod user-service/go.sum ./
COPY user-service/proto/ ./proto/
RUN go mod download

# 复制业务代码
COPY user-service/ .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /bin/user-service ./main.go

FROM alpine:3.19
WORKDIR /app

RUN apk add --no-cache tzdata && \
    ln -sf /usr/share/zoneinfo/Asia/Shanghai /etc/localtime && \
    echo 'Asia/Shanghai' > /etc/timezone

COPY --from=builder /bin/user-service /usr/local/bin/user-service
# 从构建阶段拷贝配置，避免第二阶段再依赖宿主路径
COPY --from=builder /app/configs ./configs
# 拷贝 JWT 证书（同时放到根目录与 certs 目录，便于相对路径引用）
COPY private.pem public.pem /app/
COPY private.pem public.pem /app/certs/

ARG CONFIG_PATH=/app/configs/config.dev.yaml
ENV CONFIG_PATH=${CONFIG_PATH}
EXPOSE 8081
ENTRYPOINT ["user-service"]
