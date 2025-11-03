# Go Video Platform

一个围绕 Go 生态构建的视频平台示例项目，涵盖用户管理、视频上传、转码处理和前端展示等完整链路。项目采用微服务 + DDD 分层设计，配合 MinIO、MySQL、Redis、etcd 等基础设施，演示一个可扩展的视频处理流水线。

## 功能特性
- 基于 DDD 的三大后端微服务：用户服务、上传服务、转码服务
- 分片上传 + MinIO 对象存储，支持断点续传与合并
- FFmpeg 驱动的多规格视频转码（可降级为模拟转码，方便本地开发）
- etcd 服务注册与发现，gRPC 打通服务间通信
- React + Ant Design 前端，提供基础的上传发布流程
- 完整的数据库初始化脚本与配置模板，方便快速落地

## 系统架构
```
┌──────────────────────────────┐
│            前端              │
│  Vite + React + Ant Design   │
└─────────────┬────────────────┘
              │ HTTP
┌─────────────▼────────────────┐
│        API Gateway           │
│ Gin Proxy | JWT | CORS       │
└───────┬─────────────┬────────┘
        │             │
┌───────▼─────────────┐          ┌──────────────────────┐
│     Upload Service   │◀─gRPC───▶│   Transcode Service  │
│ Gin REST | MinIO     │          │ gRPC + Worker + FFmpeg│
│ gRPC client (User)   │          │ MinIO 输出 | 任务管理 │
└───────┬─────────────┘          └────────────┬─────────┘
        │                                     │
   MySQL/Redis                            MySQL/MinIO
        │
┌───────▼─────────────┐
│     User Service    │
│ Auth | Profile      │
│ JWT | Metrics       │
└───────────────┬─────┘
                │
           MySQL/Redis
```

## 目录结构
```text
.
├── configs/                  # 全局配置模板（示例）
├── frontend/                 # Vite + React Web 前端
├── proto/                    # gRPC 协议定义与生成代码
├── scripts/mysql/init_all.sql# 一键初始化所有数据库
├── transcode-service/        # 转码微服务（含 Dockerfile）
├── upload-service/           # 上传微服务
├── user-service/             # 用户微服务
├── gateway-service/          # Gin API 网关（路由转发、鉴权）
├── temp/, default.etcd/      # 运行期示例数据目录
└── README.md                 # 本文件
```
各服务内部遵循统一的 `ddd/{adapter,application,domain,infrastructure}` 目录结构，`pkg/` 内放置通用设施（配置、日志、资源管理等）。

## 技术栈
- **后端**：Go 1.21+（go.mod 要求 1.24.x，建议使用最新 Go 版本）
- **框架**：Gin、gRPC、GORM、Logrus、Viper
- **存储**：MySQL、Redis、MinIO
- **注册发现**：etcd（可扩展到 Consul）
- **视频处理**：FFmpeg（容器镜像中已安装，亦可本地配置）
- **消息**：RabbitMQ 配置已预留，当前转码流程以内置调度为主
- **前端**：React 18、Vite、TypeScript、Ant Design

## 开发环境准备
1. **基础依赖**
   - Go ≥ 1.22（最好与 go.mod 指定版本匹配）
   - Node.js ≥ 18 + npm / pnpm（前端）
   - FFmpeg（可选，未安装时会回退为模拟转码）
   - protoc ≥ 3.21 与 `protoc-gen-go`、`protoc-gen-go-grpc`
   - MySQL 8.x、Redis 6/7、MinIO、etcd；可通过 Docker 快速启动

2. **克隆代码并安装依赖**
   ```bash
   git clone <your-repo-url> go-video
   cd go-video
   go env -w GONOSUMDB=*
   ```

3. **初始化数据库**
   ```bash
   mysql -uroot -p < scripts/mysql/init_all.sql
   ```
   脚本会创建并初始化 `user_service`、`upload_service`、`transcode_service` 。

4. **同步对象存储**
   - 在 MinIO 中创建 `uploads` / `videos` 等桶（参考 `configs/config.dev.yaml` 中的名称）
   - 为开发环境准备访问密钥并填入配置

5. **准备 etcd**
   ```bash
   docker run -d --name etcd -p 2379:2379 \
     -e ALLOW_NONE_AUTHENTICATION=yes \
     -e ETCD_ADVERTISE_CLIENT_URLS=http://0.0.0.0:2379 \
     -e ETCD_LISTEN_CLIENT_URLS=http://0.0.0.0:2379 \
     quay.io/coreos/etcd
   ```

## 配置说明
- `configs/config.dev.yaml`：全局示例模板，可复制到各服务目录下作为本地配置。注意替换其中公开 IP、数据库密码等敏感信息。
- `<service>/configs/config.dev.yaml`：服务级默认配置。若通过环境变量 `CONFIG_PATH` 指向其他文件，加载逻辑会优先使用自定义路径。
- 建议为不同环境创建独立配置文件（如 `config.local.yaml`），并通过环境变量或启动参数切换。

## 启动后端服务
> 以下命令默认在仓库根目录执行，确保所需基础设施已就绪。

1. **网关服务（HTTP:8080）**
   ```bash
   cd gateway-service
   go mod tidy
   go run ./main.go
   ```
   - 统一 `/api` 前缀路由，转发到用户、上传、转码服务
   - 对 `routes.*.auth_required=true` 的路径进行JWT鉴权，自动透传 `X-User-UUID`

2. **用户服务（HTTP:8081 / gRPC:9091）**
   ```bash
   cd user-service
   go mod tidy
   go run ./main.go
   ```

3. **上传服务（HTTP:8082）**
   ```bash
   cd upload-service
   go mod tidy
   go run ./main.go
   ```
   - 通过请求头 `X-User-UUID` 传入用户身份；服务会调用 user-service 校验。
   - `/api/v1/inner/upload` 下提供初始化、分片上传、合并等接口。

4. **转码服务（HTTP:8083 / gRPC:9092）**
   ```bash
   cd transcode-service
   go mod tidy
   go run ./main.go
   ```
   - Worker 会定期拉取待处理任务，并在有 FFmpeg 时执行真实转码；若未安装 FFmpeg 则使用模拟模式生成占位文件。
   - 暴露 `/health` 及 `/api/v1` 路由，同时向 etcd 注册自身服务信息。

4. **容器化运行（仅转码服务示例）**
   ```bash
   cd transcode-service
   docker build -t go-video/transcode-service:dev .
   docker compose up -d
   ```
   确认挂载的配置文件、日志目录与 `/tmp/transcode` 在宿主机可写。

## 启动前端
```bash
cd frontend
npm install    # 或 pnpm install
npm run dev    # 默认 5173 端口
```
前端通过统一的 `/api` 代理访问 `gateway-service`，如需直连单个服务可在 `vite.config.ts` 中调整目标地址。

## gRPC 与 Proto 文件
- 所有协议定义位于 `proto/`，`go.mod` 中通过 `replace go-vedio-1/proto => ../proto` 共享同一模块。
- 修改 proto 后需重新生成：
  ```bash
  cd proto
  protoc --go_out=. --go-grpc_out=. transcode/transcode_service.proto
  protoc --go_out=. --go-grpc_out=. user_service.proto
  ```
- 引用方只需 `go mod tidy` 即可同步最新生成代码。

## 常用端点与端口
| 服务 | HTTP 端口 | gRPC 端口 | Health Check | 说明 |
|------|-----------|-----------|--------------|------|
| gateway-service | 8080 | - | `/health` | 统一入口，负责路由转发、鉴权、CORS |
| user-service | 8081 | 9091 | `/health` | 用户登录、信息查询、gRPC 用户查询接口 |
| upload-service | 8082 | - | `/health` | 视频分片上传、合并、发布流程 |
| transcode-service | 8083 | 9092 | `/health` | 转码任务调度、状态查询、worker |

更多内部路由可参考各服务 `ddd/adapter/http` 下的控制器实现。

## 监控与运维
- 所有服务在 `configs/config.dev.yaml` 中提供 `monitoring.metrics` 配置，启用后将暴露 `/metrics`（Prometheus 格式）。
- etcd 服务注册可通过 `etcdctl get /services --prefix` 检查。
- `upload-service/ddd/task/chunk_cleanup_task.go` 会启动定期清理任务，避免残留分片。

## 开发建议
- 每次修改配置后记得重启对应服务；若使用 Docker 请重新挂载或 `docker compose up -d --build`。
- go.mod 中的 Go 版本处于前瞻状态（1.24.x），若尚未正式发布，可在本地使用 `go env -w GOVERSION=1.23` 并更新 go.mod，以便编译通过。
- 敏感配置请放置到未入库的文件中（如 `.env.local`），仓库中的 IP/密钥仅供示例，实际部署务必替换。
- 服务启动日志中包含 `[STARTUP]` 与结构化日志，可协助排查依赖加载顺序。

## TODO / 后续计划
- [ ] 补充自动化测试与 CI/CD 流程
- [ ] 完整的 API 文档与 OpenAPI/Swagger 支持
- [ ] 引入消息队列驱动的异步转码派发
- [ ] 扩展用户认证流程（OAuth、短信等）
- [ ] 添加前端播放、转码状态实时推送

欢迎提交 Issue 或 PR 共同完善项目。祝开发顺利!
