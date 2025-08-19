# Go Video Project Makefile

# 变量定义
APP_NAME=go-video
API_BINARY=api
WORKER_BINARY=worker
MIGRATION_BINARY=migration
DOCKER_IMAGE=$(APP_NAME)
DOCKER_TAG=latest

# 默认目标
.DEFAULT_GOAL := help

# 帮助信息
.PHONY: help
help: ## 显示帮助信息
	@echo "可用的命令:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

# 开发相关命令
.PHONY: deps
deps: ## 安装依赖
	go mod download
	go mod tidy

.PHONY: build
build: ## 构建所有二进制文件
	@echo "构建API服务..."
	go build -o bin/$(API_BINARY) ./cmd/api
	@echo "构建Worker服务..."
	go build -o bin/$(WORKER_BINARY) ./cmd/worker
	@echo "构建Migration工具..."
	go build -o bin/$(MIGRATION_BINARY) ./cmd/migration
	@echo "构建完成!"

.PHONY: build-api
build-api: ## 构建API服务
	go build -o bin/$(API_BINARY) ./cmd/api

.PHONY: build-worker
build-worker: ## 构建Worker服务
	go build -o bin/$(WORKER_BINARY) ./cmd/worker

.PHONY: build-migration
build-migration: ## 构建Migration工具
	go build -o bin/$(MIGRATION_BINARY) ./cmd/migration

.PHONY: run-api
run-api: ## 运行API服务
	go run ./cmd/api

.PHONY: run-worker
run-worker: ## 运行Worker服务
	go run ./cmd/worker

.PHONY: run-migration
run-migration: ## 运行数据库迁移
	go run ./cmd/migration

# 测试相关命令
.PHONY: test
test: ## 运行所有测试
	go test -v ./...

.PHONY: test-cover
test-cover: ## 运行测试并生成覆盖率报告
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "覆盖率报告已生成: coverage.html"

.PHONY: benchmark
benchmark: ## 运行基准测试
	go test -bench=. -benchmem ./...

# 代码质量检查
.PHONY: fmt
fmt: ## 格式化代码
	go fmt ./...

.PHONY: vet
vet: ## 代码静态检查
	go vet ./...

.PHONY: lint
lint: ## 代码风格检查 (需要安装golangci-lint)
	golangci-lint run

.PHONY: check
check: fmt vet lint test ## 运行所有代码检查

# Docker相关命令
.PHONY: docker-build
docker-build: ## 构建Docker镜像
	docker build -t $(DOCKER_IMAGE):$(DOCKER_TAG) .

.PHONY: docker-run
docker-run: ## 运行Docker容器
	docker run -p 8080:8080 $(DOCKER_IMAGE):$(DOCKER_TAG)

.PHONY: docker-push
docker-push: ## 推送Docker镜像
	docker push $(DOCKER_IMAGE):$(DOCKER_TAG)

# Docker Compose相关命令
.PHONY: up
up: ## 启动所有服务
	docker-compose up -d

.PHONY: down
down: ## 停止所有服务
	docker-compose down

.PHONY: logs
logs: ## 查看服务日志
	docker-compose logs -f

.PHONY: ps
ps: ## 查看服务状态
	docker-compose ps

.PHONY: restart
restart: ## 重启所有服务
	docker-compose restart

.PHONY: rebuild
rebuild: ## 重新构建并启动服务
	docker-compose down
	docker-compose build --no-cache
	docker-compose up -d

# 数据库相关命令
.PHONY: db-up
db-up: ## 启动数据库服务
	docker-compose up -d mysql redis

.PHONY: db-migrate
db-migrate: ## 运行数据库迁移
	./bin/$(MIGRATION_BINARY) || go run ./cmd/migration

.PHONY: db-reset
db-reset: ## 重置数据库
	docker-compose down mysql
	docker volume rm go-vedio-1_mysql_data || true
	docker-compose up -d mysql
	@echo "等待MySQL启动..."
	sleep 10
	make db-migrate

# 清理命令
.PHONY: clean
clean: ## 清理构建文件
	rm -rf bin/
	rm -f coverage.out coverage.html
	go clean -cache

.PHONY: clean-docker
clean-docker: ## 清理Docker资源
	docker-compose down -v
	docker system prune -f
	docker volume prune -f

# 生产部署相关
.PHONY: deploy-prod
deploy-prod: ## 生产环境部署
	@echo "开始生产环境部署..."
	make build
	make docker-build
	make docker-push
	@echo "部署完成!"

# 开发环境快速启动
.PHONY: dev
dev: ## 快速启动开发环境
	make db-up
	@echo "等待数据库启动..."
	sleep 10
	make db-migrate
	make run-api

# 安装开发工具
.PHONY: install-tools
install-tools: ## 安装开发工具
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install github.com/swaggo/swag/cmd/swag@latest
	go install github.com/golang-migrate/migrate/v4/cmd/migrate@latest

# 生成API文档
.PHONY: swagger
swagger: ## 生成Swagger文档
	swag init -g ./cmd/api/main.go -o ./docs/swagger

.PHONY: all
all: deps check build ## 执行完整的构建流程