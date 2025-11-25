# Go Video Platform

这是一个基于 Go 语言生态构建的现代化视频平台示例项目，采用微服务架构 + DDD（领域驱动设计）思想。项目实现了从用户注册登录、视频分片上传、断点续传、到后台异步转码处理及前端播放的完整业务链路。

## ✨ 功能特性

- **微服务架构**：基于 DDD 设计的三个核心后端服务（用户、上传、转码），职责清晰。
- **高性能上传**：支持大文件分片上传、断点续传、秒传，结合 MinIO / RustFS 对象存储。
- **视频处理**：内置 FFmpeg 转码工作流，支持多分辨率转码（可降级为模拟模式），**默认生成 HLS (m3u8) 切片**，实现流畅的流媒体播放体验。
- **服务治理**：使用 etcd 进行服务注册与发现，gRPC 实现高性能服务间通信。
- **安全机制**：基于 JWT 的用户认证，API 网关统一鉴权。
- **现代化前端**：使用 React 18 + Vite + Ant Design 构建的响应式 Web 界面，**仿 Bilibili 风格设计**。

## 🗺️ 产品规划 (Roadmap)

### ✅ 已完成
- [x] 用户注册与登录 (JWT)
- [x] 视频分片上传与断点续传
- [x] 视频转码 (MP4 -> HLS/m3u8)
- [x] 首页视频列表展示 (Bilibili 风格)
- [x] 视频播放器 (支持 HLS)

### 🚧 待开发 (Coming Soon)
- [ ] **增强鉴权**：完善的 RBAC 权限控制，支持 OAuth2 第三方登录。
- [ ] **社交互动**：
    - [ ] 用户关注/粉丝系统
    - [ ] 视频评论与回复
    - [ ] 视频点赞、投币、收藏
- [ ] **内容创作**：
    - [ ] 视频发布草稿箱功能
    - [ ] 创作者中心（数据看板）
- [ ] **运营功能**：
    - [ ] 首页轮播图管理
    - [ ] 视频推荐算法

## 🏗 系统架构

```mermaid
graph TD
    Client[Web Frontend] -->|HTTP| Gateway[API Gateway]
    
    subgraph "Backend Services"
        Gateway -->|HTTP| UserSvc[User Service]
        Gateway -->|HTTP| UploadSvc[Upload Service]
        Gateway -->|HTTP| TranscodeSvc[Transcode Service]
        
        UploadSvc -.->|gRPC| UserSvc
        UploadSvc -.->|gRPC| TranscodeSvc
        TranscodeSvc -.->|gRPC| UploadSvc
    end
    
    subgraph "Infrastructure"
        MySQL[(MySQL)]
        Redis[(Redis)]
        ObjectStorage[(MinIO / RustFS)]
        Etcd[etcd Registry]
    end
    
    UserSvc --> MySQL
    UserSvc --> Redis
    
    UploadSvc --> MySQL
    UploadSvc --> ObjectStorage
    
    TranscodeSvc --> MySQL
    TranscodeSvc --> ObjectStorage
    TranscodeSvc --> Etcd
```

### 服务职责
- **Gateway Service (8080)**: 统一流量入口，负责路由转发、CORS 处理、JWT 鉴权。
- **User Service (8081 / gRPC 9091)**: 用户注册、登录、个人信息管理。
- **Upload Service (8082)**: 视频分片上传、合并、元数据发布（视频发布逻辑在此服务中）。
- **Transcode Service (8083 / gRPC 9092)**: 异步视频转码任务调度与执行，生成 HLS 切片并回传播放地址。

## 🛠 技术栈

- **后端**: Go 1.24+ (Gin, gRPC, GORM, Viper, Wire)
- **数据库/缓存**: MySQL 8.0, Redis 7.0
- **对象存储**: MinIO / RustFS (高性能分布式文件系统)
- **服务发现**: etcd
- **视频工具**: FFmpeg
- **前端**: React 18, TypeScript, Vite, Ant Design

## 🚀 快速开始

### 1. 环境准备
确保本地已安装以下工具：
- Go 1.24+
- Node.js 18+ & pnpm/npm
- Docker & Docker Compose (推荐用于启动基础设施)
- FFmpeg (可选，若未安装转码服务将运行在模拟模式)

### 2. 启动基础设施
使用 Docker 快速启动 MySQL, Redis, MinIO/RustFS, etcd：
```bash
# 假设你已有相关的 docker-compose.yml，或者手动启动这些容器
# 示例 etcd 启动命令:
docker run -d --name etcd -p 2379:2379 \
  -e ALLOW_NONE_AUTHENTICATION=yes \
  -e ETCD_ADVERTISE_CLIENT_URLS=http://0.0.0.0:2379 \
  -e ETCD_LISTEN_CLIENT_URLS=http://0.0.0.0:2379 \
  quay.io/coreos/etcd
```

### 3. 初始化数据库
执行初始化脚本创建数据库和表结构：
```bash
mysql -h 127.0.0.1 -P 3306 -u root -p < scripts/mysql/init_all.sql
```
> **注意**: 脚本会创建 `user_service`, `upload_service`, `transcode_service` 等数据库。其中 `upload_service` 库包含了视频发布相关的 `video_publish` 表。

### 4. 配置文件
复制并修改配置文件（参考 `configs/config.dev.yaml`）：
- 确保数据库、Redis、MinIO/RustFS 的连接信息正确。
- **关键配置**：确保存储服务 (MinIO 或 RustFS) 中存在 `upload` 和 `transcode` 桶（Bucket）。
    - `upload`: 用于存储原始上传文件和转码后的 HLS 切片。
    - `transcode`: (可选) 用于存储中间产物或其他转码资源。

### 5. 启动后端服务
建议在不同的终端窗口中分别启动：

**Gateway Service**
```bash
cd gateway-service
go mod tidy
go run main.go
```

**User Service**
```bash
cd user-service
go mod tidy
go run main.go
```

**Upload Service**
```bash
cd upload-service
go mod tidy
go run main.go
```

**Transcode Service**
```bash
cd transcode-service
go mod tidy
go run main.go
```

### 6. 启动前端
```bash
cd frontend
npm install
npm run dev
```
访问 `http://localhost:5173` 开始体验。

### 7. 构建 Docker 镜像（本地/K8s）
在仓库根目录执行，确保能访问到 `proto/` 与各服务源码：
```bash
# 用户服务镜像
docker build -f user-service/Dockerfile -t go-video/user-service:dev .

# 上传服务镜像
docker build -f upload-service/Dockerfile -t go-video/upload-service:dev .

# 转码服务镜像（如需）
docker build -f transcode-service/Dockerfile -t go-video/transcode-service:dev .
```

### 8. 运行容器（示例）
假设 MySQL/Redis 在宿主机，可用 `host.docker.internal` 访问；密钥以挂载方式提供：
```bash
# 用户服务
docker rm -f user-service 2>/dev/null
docker run -d --name user-service \
  -p 8081:8081 \
  -v "$(pwd)/user-service/configs/config.dev.yaml:/app/configs/config.dev.yaml" \
  -v "$(pwd)/private.pem:/app/certs/private.pem" \
  -v "$(pwd)/public.pem:/app/certs/public.pem" \
  -e CONFIG_PATH=/app/configs/config.dev.yaml \
  -e GO_VIDEO_JWT_RSA_PRIVATE_KEY_PATH=/app/certs/private.pem \
  -e GO_VIDEO_JWT_RSA_PUBLIC_KEY_PATH=/app/certs/public.pem \
  go-video/user-service:dev

# 上传服务
docker rm -f upload-service 2>/dev/null
docker run -d --name upload-service \
  -p 8082:8082 \
  -p 9094:9094 \
  -v "$(pwd)/upload-service/configs/config.dev.yaml:/app/configs/config.dev.yaml" \
  -e CONFIG_PATH=/app/configs/config.dev.yaml \
  go-video/upload-service:dev
```
若依赖运行在其他容器，请将配置中的 host 改为对应容器名并放到同一网络；启动失败时用 `docker logs <name>` 排查。

### 📦 不使用 etcd 的本地 / 云端直连部署
- 在三个服务的配置中将 `etcd.endpoints` 置空，`service_registry.enabled`（以及 upload-service 的 `grpc_service_registry.enabled`）设为 `false`。
- 为依赖服务填好直连地址：`upload-service` -> `dependencies.user_service.address=host:9091`、`dependencies.transcode_service.address=host:9092`；`transcode-service` -> `dependencies.upload_service.address=host:9093`。在 K8s 中可直接使用 Service DNS，如 `user-service:9091`。
- 仍可在有 etcd 的环境下开启注册发现，保持兼容。

## 📂 目录结构

```text
.
├── configs/                # 全局配置示例
├── frontend/               # React 前端项目
├── gateway-service/        # API 网关
├── upload-service/         # 上传与视频发布服务
├── user-service/           # 用户服务
├── transcode-service/      # 转码服务
├── proto/                  # gRPC Protobuf 定义
├── scripts/                # 数据库初始化脚本等
└── README.md               # 项目文档
```

## 📝 注意事项
- **Go 版本**: 项目使用了 Go 1.24 的新特性，请确保本地 Go 版本符合要求。
- **FFmpeg**: 转码服务会检查本地是否安装 `ffmpeg`。如果未安装，会自动降级为"模拟转码"模式（仅复制文件并重命名），方便非生产环境开发。
- **服务依赖**: 启动顺序建议为：基础设施 -> User/Upload/Transcode -> Gateway -> Frontend。

## 🤝 贡献
欢迎提交 Issue 和 Pull Request！
