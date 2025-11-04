FROM golang:1.24.7-alpine AS builder
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /bin/gateway-service ./main.go

FROM alpine:3.19
WORKDIR /app

COPY --from=builder /bin/gateway-service /usr/local/bin/gateway-service
COPY configs ./configs

ENV CONFIG_PATH=/app/configs/config.dev.yaml
EXPOSE 8080
ENTRYPOINT ["gateway-service"]