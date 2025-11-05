FROM registry.cn-hangzhou.aliyuncs.com/library/golang:1.20-alpine AS builder
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /bin/user-service ./main.go

FROM registry.cn-hangzhou.aliyuncs.com/library/alpine:3.19
WORKDIR /app

COPY --from=builder /bin/user-service /usr/local/bin/user-service
COPY configs ./configs

ENV CONFIG_PATH=/app/configs/config.dev.yaml
EXPOSE 8081
ENTRYPOINT ["user-service"]