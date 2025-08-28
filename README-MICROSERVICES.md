# 🎬 Go Video Platform - 微服务架构

## 📋 项目概述

这是一个基于DDD（领域驱动设计）的微服务视频平台，将原有的单体应用拆分为多个独立的微服务。

## 🏗️ 架构设计

### 服务拆分

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   API Gateway   │    │  User Service   │    │ Upload Service  │
│   (网关服务)    │    │   (用户服务)    │    │  (上传服务)     │
│   Port: 8080    │    │   Port: 8081    │    │   Port: 8082    │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                       │                       │
         └───────────────────────┼───────────────────────┘
                                 │
                    ┌─────────────────┐
                    │  Video Service  │
                    │   (视频服务)    │
                    │   Port: 8083    │
                    └─────────────────┘
```

### 服务职责

#### 🔐 API Gateway (端口: 8080)
- **功能**: 统一入口，路由转发
- **职责**: 
  - 请求路由和负载均衡
  - 认证和鉴权
  - 限流和熔断
  - 请求日志和监控

#### 👤 User Service (端口: 8081)
- **功能**: 用户管理和认证
- **职责**:
  - 用户注册、登录
  - 用户资料管理
  - JWT Token生成和验证
  - 权限控制

#### 📤 Upload Service (端口: 8082)
- **功能**: 文件上传和存储
- **职责**:
  - 文件上传处理
  - 上传进度追踪
  - 分片上传支持
  - MinIO存储集成

#### 🎥 Video Service (端口: 8083)
- **功能**: 视频管理和播放
- **职责**:
  - 视频元数据管理
  - 视频播放和检索
  - 推荐算法
  - 播放统计

## 🚀 快速开始

### 前置要求

- Docker 20.0+
- Docker Compose 2.0+
- Git

### 启动服务

1. **克隆项目**
   ```bash
   git clone <repository-url>
   cd go-vedio-1
   ```

2. **启动微服务**
   ```bash
   ./start-microservices.sh
   ```

3. **验证服务状态**
   ```bash
   # 检查所有服务健康状态
   curl http://localhost:8080/health
   curl http://localhost:8081/health
   curl http://localhost:8082/health
   curl http://localhost:8083/health
   ```

### 手动启动（可选）

```bash
# 启动所有服务
docker-compose -f docker-compose.microservices.yml up -d

# 查看服务状态
docker-compose -f docker-compose.microservices.yml ps

# 查看日志
docker-compose -f docker-compose.microservices.yml logs -f

# 停止服务
docker-compose -f docker-compose.microservices.yml down
```

## 📊 数据库设计

### 用户服务数据库 (user_service)
- **users**: 用户基本信息

### 上传服务数据库 (upload_service)
- **upload_tasks**: 上传任务记录
- **upload_chunks**: 分片上传记录

### 视频服务数据库 (video_service)
- **videos**: 视频元数据
- **video_play_records**: 播放记录
- **video_likes**: 点赞记录

## 🔌 API接口

### 通过API网关访问 (推荐)

所有请求都通过API网关 `http://localhost:8080` 进行：

#### 用户相关
```bash
# 用户注册
POST http://localhost:8080/api/v1/auth/register

# 用户登录
POST http://localhost:8080/api/v1/auth/login

# 获取用户信息
GET http://localhost:8080/api/v1/users/profile
```

#### 上传相关
```bash
# 文件上传
POST http://localhost:8080/api/v1/upload/test

# 获取上传进度
GET http://localhost:8080/api/v1/upload/progress/:id
```

#### 视频相关
```bash
# 获取视频列表
GET http://localhost:8080/api/v1/videos/

# 获取视频详情
GET http://localhost:8080/api/v1/videos/:id

# 获取播放地址
GET http://localhost:8080/api/v1/videos/:id/play

# 点赞视频
POST http://localhost:8080/api/v1/videos/:id/like

# 搜索视频
GET http://localhost:8080/api/v1/videos/search?q=keyword

# 推荐视频
GET http://localhost:8080/api/v1/videos/recommend
```

### 直接访问服务 (开发调试)

也可以直接访问各个服务进行调试：

- 用户服务: `http://localhost:8081/api/v1/users/test`
- 上传服务: `http://localhost:8082/api/v1/upload/test`
- 视频服务: `http://localhost:8083/api/v1/videos/`

## 🔧 开发指南

### 服务开发

每个服务都是独立的Go模块，包含：

```
service-name/
├── cmd/main.go          # 服务入口
├── go.mod              # Go模块定义
├── Dockerfile          # Docker构建文件
└── [DDD架构目录]        # 业务逻辑
```

### 添加新服务

1. 创建服务目录和基础结构
2. 实现业务逻辑
3. 添加到 `docker-compose.microservices.yml`
4. 在API网关中添加路由配置

### 本地开发

```bash
# 进入服务目录
cd user-service

# 安装依赖
go mod tidy

# 运行服务
go run cmd/main.go
```

## 🐳 Docker配置

### 服务端口映射

| 服务 | 内部端口 | 外部端口 | 说明 |
|------|----------|----------|------|
| API Gateway | 8080 | 8080 | 主入口 |
| User Service | 8081 | 8081 | 用户服务 |
| Upload Service | 8082 | 8082 | 上传服务 |
| Video Service | 8083 | 8083 | 视频服务 |
| User DB | 3306 | 3307 | 用户数据库 |
| Upload DB | 3306 | 3308 | 上传数据库 |
| Video DB | 3306 | 3309 | 视频数据库 |
| MinIO | 9000/9001 | 9000/9001 | 对象存储 |
| Redis | 6379 | 6379 | 缓存 |

### 环境变量

每个服务支持以下环境变量：

```bash
# 通用配置
PORT=8081                    # 服务端口
DB_DSN=mysql://...           # 数据库连接

# API网关特有
USER_SERVICE_URL=http://user-service:8081
UPLOAD_SERVICE_URL=http://upload-service:8082
VIDEO_SERVICE_URL=http://video-service:8083

# 上传服务特有
MINIO_ENDPOINT=minio:9000
MINIO_ACCESS_KEY=minioadmin
MINIO_SECRET_KEY=minioadmin
```

## 📈 监控和日志

### 健康检查

每个服务都提供健康检查接口：

```bash
curl http://localhost:8080/health  # API网关
curl http://localhost:8081/health  # 用户服务
curl http://localhost:8082/health  # 上传服务
curl http://localhost:8083/health  # 视频服务
```

### 日志查看

```bash
# 查看所有服务日志
docker-compose -f docker-compose.microservices.yml logs -f

# 查看特定服务日志
docker-compose -f docker-compose.microservices.yml logs -f user-service
docker-compose -f docker-compose.microservices.yml logs -f upload-service
docker-compose -f docker-compose.microservices.yml logs -f video-service
docker-compose -f docker-compose.microservices.yml logs -f api-gateway
```

## 🔒 安全配置

### 认证流程

1. 用户通过API网关登录
2. 用户服务验证凭据并返回JWT Token
3. 后续请求携带Token通过API网关
4. API网关验证Token并转发请求

### 服务间通信

- 所有外部请求通过API网关
- 服务间通过内部网络通信
- 使用Docker网络隔离

## 🚨 故障排除

### 常见问题

1. **服务启动失败**
   ```bash
   # 检查端口占用
   lsof -i :8080
   
   # 检查Docker状态
   docker ps
   docker-compose -f docker-compose.microservices.yml ps
   ```

2. **数据库连接失败**
   ```bash
   # 检查数据库容器状态
   docker-compose -f docker-compose.microservices.yml logs user-db
   
   # 手动连接测试
   docker exec -it go-vedio-1_user-db_1 mysql -u root -p
   ```

3. **服务间通信失败**
   ```bash
   # 检查网络连接
   docker network ls
   docker network inspect go-vedio-1_microservices
   ```

### 重置环境

```bash
# 完全清理环境
docker-compose -f docker-compose.microservices.yml down -v --rmi all
docker system prune -f

# 重新启动
./start-microservices.sh
```

## 🎯 性能优化

### 建议配置

- **生产环境**: 使用外部数据库和Redis集群
- **负载均衡**: 在API网关前添加Nginx
- **缓存策略**: 在各服务中添加Redis缓存
- **监控**: 集成Prometheus + Grafana

### 扩展性

- 每个服务可独立水平扩展
- 数据库可按服务分库分表
- 使用消息队列处理异步任务

## 📚 相关文档

- [原始API文档](./API_DOCUMENTATION.md)
- [DDD架构说明](./README.md)
- [Docker部署指南](./docker-compose.microservices.yml)

## 🤝 贡献指南

1. Fork项目
2. 创建功能分支
3. 提交更改
4. 创建Pull Request

## 📄 许可证

MIT License