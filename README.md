# Go Video Platform

一个基于 Go 生态的示例级视频平台，采用微服务 + DDD，覆盖用户、上传、转码、通知、视频播放全链路。当前部署形态：Kubernetes（k3s），入口使用 Kong Ingress Controller（KIC）+ NodePort。

## 架构总览
- **网关**：Kong Ingress Controller，NodePort 30080（proxy）/30081（admin）；自定义插件 `jwt-user-uuid-header` 注入 `X-User-UUID`，JWT 使用 RS256 公钥校验。
- **服务**：
  - `user-service` (HTTP 8081 / gRPC 9091)：用户注册登录、资料、关注/粉丝、JWT 发行。
  - `upload-service` (HTTP 8082)：分片上传、断点续传、合并、视频发布。
  - `transcode-service` (HTTP 8083 / gRPC 9092)：转码任务（FFmpeg/NVENC 可选）、生成 HLS。
  - `notification-service` (HTTP 8086 / gRPC 9095 / SSE)：站内通知、SSE 推送。
  - `video-service` (HTTP 8085 / gRPC 9093)：视频列表/详情/互动（点赞、评论）。
- **存储**：MySQL、Redis、对象存储（RustFS）。
- **前端**：React 18 + Vite，Nginx 反代到 `kong-kong-proxy.go-video.svc.cluster.local`（集群内）或 NodePort 30080。

## 代码目录
- `user-service` / `upload-service` / `transcode-service` / `notification-service` / `video-service`: 微服务（Go 1.24），DDD 分层。
- `gateway-service`: 自定义 Kong 插件源码与本地 docker-compose 网关配置（KIC 已接管入口）。
- `frontend`: React + Vite 前端，Nginx 配置在 `k8s/apps/frontend/frontend-nginx-config.yaml`。
- `k8s/apps`: 各服务 Deployment/Service/Ingress/ConfigMap；`k8s/apps/gateway/kong-ingress.yaml` 定义路由与插件；`k8s/apps/gateway/kong-consumer.yaml` 提供 JWT 公钥。
- `build_push.sh`: 构建/推送业务镜像和 `kong-custom` 镜像的脚本（已移除旧 gateway 镜像）。

## 网关与流量
- 入口：NodePort 30080 → Kong → 各服务 ClusterIP。
- Kong 镜像：使用 `gateway-service/Dockerfile` 构建的 `kong-custom`（启用插件 `jwt-user-uuid-header`）。
- Ingress/插件：`k8s/apps/gateway/kong-ingress.yaml`
  - Open/Inner 路由分离；JWT 仅作用于 inner 路由。
  - 自定义插件链：`jwt` 校验 → `request-transformer` 去除客户端 `X-Access-Token` → `jwt-user-uuid-header` 注入 `X-User-UUID`。
- JWT 公钥：`k8s/apps/gateway/kong-consumer.yaml`（consumer `app-client`，key=`go-video`，RS256 公钥与服务一致）。

## 部署（K8s）
1) 确保集群拉取凭证 `acr-cred` 可访问镜像仓库。
2) 部署/升级 Kong（使用 `kong-custom` 镜像，自定义插件启用）：  
   ```bash
   helm upgrade -i kong ./kong-3.0.1.tgz \
     --namespace go-video --create-namespace \
     --set image.repository=<REG>/<NS> \
     --set image.tag=<kong-custom-tag> \
     --set image.effectiveSemver=3.6 \
     --set-string env.plugins=bundled,jwt-user-uuid-header \
     --set env.proxy_read_timeout=600 \
     --set env.nginx_proxy_proxy_buffering=off \
     --set ingressController.enabled=true \
     --set ingressController.ingressClass=kong \
     --set proxy.type=NodePort --set proxy.http.nodePort=30080 \
     --set proxy.tls.enabled=false \
     --set admin.enabled=true --set admin.type=NodePort --set admin.http.nodePort=30081 \
     --set env.database=off
   ```
3) 应用网关路由/插件/consumer：  
   ```bash
   kubectl apply -f k8s/apps/gateway/kong-consumer.yaml
   kubectl apply -f k8s/apps/gateway/kong-ingress.yaml
   ```
4) 部署各服务与前端（按需替换镜像）：  
   ```bash
   kubectl apply -f k8s/apps/user-service
   kubectl apply -f k8s/apps/upload-service
   kubectl apply -f k8s/apps/transcode-service
   kubectl apply -f k8s/apps/video-service
   kubectl apply -f k8s/apps/notification-service
   kubectl apply -f k8s/apps/frontend
   ```
5) 前端 ConfigMap 指向 Kong（已改为 `kong-kong-proxy`）；如更新 ConfigMap：  
   ```bash
   kubectl apply -f k8s/apps/frontend/frontend-nginx-config.yaml
   kubectl -n go-video rollout restart deployment/frontend
   ```
6) 入口访问：`http://<NodeIP>:30080`；Admin API：`http://<NodeIP>:30081`（建议仅集群内访问）。

## 镜像构建
使用 `build_push.sh`（已去掉旧 gateway 镜像）：
```bash
TAG=$(date +%Y%m%d-%H%M%S) REG=<registry> NS=<namespace> ./build_push.sh
```
产出镜像（示例）：`<REG>/<NS>:kong-custom-$TAG`、各服务和前端镜像。发布时更新 Deployment 使用新标签。

## 本地开发
- 服务：进入对应目录 `go mod tidy && go run main.go`，可用 `CONFIG_PATH` 指定配置；HTTP 默认 808x，gRPC 909x。
- 测试：`go test ./...`（按服务）。
- 前端：`cd frontend && npm install && npm run dev`。
- JWT/认证：使用 `user-service` 发的 JWT，iss=`go-video`，RS256 公钥见 `k8s/apps/gateway/kong-consumer.yaml`。

## 观测与日志
- 日志：各服务 stdout JSON，Kong/服务日志可通过 `X-Request-ID` 关联。
- 可选组件：`k8s/observability` 下提供 Loki/Promtail、Grafana、OTEL 示例清单。

## 常见问题
- **401/鉴权失败**：检查 token 的 `iss`、签名密钥是否与 `app-client-jwt` 公钥匹配；确认请求走的是 30080/Kong 入口且命中 inner 路由需要 JWT。
- **SSE 中断/超时**：确认前端 Nginx/SSE 路由指向 Kong，且 Kong 环境变量已关闭 proxy buffering、拉长 read timeout（见 Helm 命令）。
- **旧 gateway 兼容**：`k8s/apps/gateway/gateway-compat-svc.yaml` 可提供 `gateway` 名称指向 Kong（如无需兼容可不应用/删除）。
