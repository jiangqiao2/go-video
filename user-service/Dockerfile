FROM golang:1.24-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git ca-certificates tzdata

# 复制模块定义（使用仓库根上下文时按路径复制）
COPY user-service/go.mod user-service/go.sum ./
COPY proto/ ../proto/

RUN go mod download

# 复制源码
COPY user-service/ .

# 构建可执行文件
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o user-service .

FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata curl su-exec

RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

WORKDIR /app

COPY --from=builder /app/user-service .

ENV CONFIG_PATH=/app/configs/config.dev.yaml

EXPOSE 8081 9091

USER appuser

ENTRYPOINT ["./user-service"]