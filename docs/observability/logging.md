# 日志收集与查询方案（Loki + Promtail + Grafana）

本文档说明 go-video 项目中，各个服务的日志是如何产出、采集并集中到 Loki 中的，以及在本地和 K8s 环境下如何查看这些日志。

---

## 1. 整体架构概览

整体链路可以概括为：

> 应用日志（stdout / 文件）→ 容器运行时 / 节点日志文件 → Promtail DaemonSet → Loki（集中存储） → Grafana（查询与可视化）

- 所有 Go 微服务（`user-service`、`upload-service`、`video-service`、`transcode-service`、`notification-service`）使用各自的 `pkg/logger` 输出结构化日志。
- 在 Kubernetes 中，各服务通常将日志写到 `stdout`/`stderr`，由容器运行时落盘到节点的 `/var/log/pods/...` 文件。
- `observability` 命名空间中部署：
  - Loki：集中存储日志。
  - Promtail：在每个节点上采集日志并推送到 Loki。
  - Grafana：通过 Loki 数据源查询日志。

---

## 2. 应用侧日志产出方式

### 2.1 日志组件（pkg/logger）

每个服务都有独立的日志组件实现，例如：

- `user-service/pkg/logger/logger.go`
- `upload-service/pkg/logger/logger.go`
- `video-service/pkg/logger/logger.go`
- `transcode-service/pkg/logger/logger.go`
- `notification-service/pkg/logger/logger.go`

这些实现遵循同一套接口和结构：

- 支持的日志级别：`DEBUG`、`INFO`、`WARN`、`ERROR`、`FATAL`。
- 支持两种格式：
  - `json`：输出 JSON 结构化日志。
  - `text`：输出文本日志（带文件名、行号等）。
- 支持的输出方式：
  - `stdout`：标准输出（推荐 K8s / 容器环境）。
  - `stderr`：标准错误。
  - `file`：写入指定日志文件（本地开发可用）。

典型的日志结构（JSON 模式）类似：

```json
{
  "timestamp": "2024-05-10T12:34:56+08:00",
  "level": "INFO",
  "message": "http request",
  "file": "request_log.go",
  "line": 27,
  "function": "middleware.RequestLogMiddleware.func1",
  "fields": {
    "request_id": "xxx",
    "user_uuid": "yyy",
    "method": "GET",
    "path": "/api/v1/users",
    "status": 200,
    "latency_ms": 54
  }
}
```

日志配置通过各服务的 `configs/config*.yaml` 控制，例如 `user-service/configs/config.dev.yaml` 中：

```yaml
log:
  level: "debug"      # debug, info, warn, error
  format: "json"      # json, text
  output: "stdout"    # stdout, file（K8s 环境必须为 stdout）
  filename: "logs/user-service.log"
  max_size: 100       # MB
  max_backups: 7
  max_age: 30         # days
  compress: true
```

在 K8s 中部署时，推荐统一使用：

```yaml
log:
  level: info
  format: json
  output: stdout
```

### 2.2 HTTP 请求日志

各服务都会在 HTTP 层打印统一格式的访问日志。例如 `notification-service/pkg/middleware/request_log.go`：

- 使用 Gin 中间件记录：
  - `request_id`（来自请求上下文）。
  - `method`、`path`、`status`。
  - `latency_ms`（请求耗时）。
- 日志通过 `logger.WithFields(fields).Info("http request")` 输出，形成统一的 `"message": "http request"` 访问日志。

其他服务（`user-service`、`upload-service` 等）在各自的 `pkg/middleware` 目录下也有类似的 HTTP 访问日志中间件。

### 2.3 gRPC 调用日志与 request_id 传播

gRPC 层使用 `pkg/grpcutil` 中的拦截器（例如 `notification-service/pkg/grpcutil/request_id.go`）来：

- 统一生成/传递 `request_id`：
  - 若 context 中不存在 `request_id`，则生成一个新的 UUID。
  - 将 `request_id` 写入 gRPC metadata（`x-request-id`）。
  - 通过 server 和 client 拦截器自动传递和回写 header。
- 打印 gRPC 调用日志：
  - client 拦截器记录 `method`、`kind=grpc_client`、`duration_ms`、`error`。
  - server 拦截器记录 `method`、`kind=grpc_server`、`duration_ms`、`error`。
  - 日志同样包含 `request_id` 字段，方便跨服务串联调用链路。

对于业务代码，推荐使用 `logger.WithContext(ctx)` 来生成带 `request_id` / `user_uuid` 等字段的日志器：

```go
func (s *UserService) DoSomething(ctx context.Context, req *pb.SomeRequest) (*pb.SomeResponse, error) {
    log := logger.WithContext(ctx)
    log.Infof("start do something, req=%+v", req)
    // ...
    log.Info("done")
    return resp, nil
}
```

### 2.4 本地开发环境

本地运行服务时：

- 直接 `go run main.go` 或 `go run ./app`，默认使用 `configs/config.dev.yaml`。
- 若 `log.output=stdout`，日志直接打印在终端。
- 若配置为 `output=file`，则会写入配置中的 `filename`（例如 `logs/user-service.log`），可通过：

```bash
tail -f logs/user-service.log
```

来查看。

本地一般不依赖 Loki/Promtail，仅用于快速调试。

---

## 3. Kubernetes 中的日志采集链路

### 3.1 节点上的日志文件

在 K8s 中，各服务容器的 stdout/stderr 日志会由 kubelet 写入节点上的文件，路径模式类似：

```text
/var/log/pods/<namespace>_<pod>_<pod-uid>/<container>/0.log
```

对于 go-video 项目，核心服务的日志路径大致为：

- `go-video` 命名空间：
  - `user-service`：`/var/log/pods/go-video_user-service-*/user-service/*.log`
  - `upload-service`：`/var/log/pods/go-video_upload-service-*/upload-service/*.log`
  - `video-service`：`/var/log/pods/go-video_video-service-*/video-service/*.log`
  - `transcode-service`：`/var/log/pods/go-video_transcode-service-*/transcode-service/*.log`
  - `gateway (kong)`：`/var/log/pods/go-video_gateway-*/kong/*.log`

> 注意：路径中的 `*` 表示 Pod 实例名和日志文件序号（`0.log` / `1.log` 等）。

### 3.2 Promtail：从节点采集日志到 Loki

Promtail 的配置在：

- `k8s/observability/promtail/promtail.yaml`

该文件包含：

- ServiceAccount / RBAC：授予 Promtail 读取 Pod/Namespace 元数据的权限。
- ConfigMap `promtail-config`：核心采集配置。
- DaemonSet：每个节点运行一个 Promtail 实例。

关键配置片段：

- `kubernetes_sd_configs` + `relabel_configs`：
  - 通过 Kubernetes API 动态发现 `go-video` 命名空间下的 Pod。
  - 只保留 `namespace=go-video` 的日志。
  - 将 Pod/Namespace/Container/Node 等元数据映射为日志标签。
  - 使用 Pod UID 构造 `__path__`，指向 `/var/log/pods/*<pod-uid>*/*/*.log`。
- `static_configs`：
  - 显式为各服务配置静态采集 Job，例如：

    ```yaml
    - job_name: manual-go-video-upload
      static_configs:
        - targets: [localhost]
          labels:
            job: manual-go-video-upload
            namespace: go-video
            app: upload-service
            __path__: /var/log/pods/go-video_upload-service-*/upload-service/*.log
      pipeline_stages:
        - cri: {}
    ```

  - 其他服务（user/video/transcode/gateway）也有类似配置，保证即使自动发现异常，静态日志采集仍然可用。
- `pipeline_stages` 中的 `cri: {}`：
  - 用于解析 CRI 样式的容器日志格式（时间戳 + 流类型 + 实际日志内容），提取出正确的时间和 log line。

Promtail 将解析后的日志发送至 Loki：

```yaml
clients:
  - url: http://loki.observability.svc:3100/loki/api/v1/push
```

### 3.3 Loki：集中存储日志

Loki 的部署与配置在：

- `k8s/observability/loki/loki.yaml`

主要内容：

- Service：`loki.observability.svc`，端口 `3100`。
- ConfigMap `loki-config`：
  - 使用 `boltdb-shipper + filesystem` 作为存储。
  - `retention_period: 168h`（7 天日志保留，可根据需要调整）。
- StatefulSet：
  - 单副本 Loki 实例。
  - 使用 PVC 持久化索引和块数据，挂载到 `/var/loki`。

Promtail 推送过来的日志最终按标签（namespace/app/job 等）存储在 Loki 中，供后续查询。

### 3.4 Grafana：查询与可视化

Grafana 与 Loki 的集成配置在：

- `k8s/observability/grafana/grafana-datasource-loki.yaml`

该 ConfigMap 通过 kube-prometheus-stack 的 sidecar 自动挂载为 Grafana 数据源：

- 数据源名称：`Loki`
- 类型：`loki`
- URL：`http://loki.observability.svc:3100`

部署完成后，你可以通过 NodePort 或 `kubectl port-forward` 访问 Grafana，在 **Explore** 页面选择 `Loki` 数据源并执行 LogQL 查询。

---

## 4. 常用日志查询示例（LogQL）

以下是一些常见的查询场景，方便快速排查问题：

1. 查看 go-video 命名空间下所有服务日志：

   ```logql
   {namespace="go-video"}
   ```

2. 只看某个服务（例如 `upload-service`）的日志：

   ```logql
   {namespace="go-video", app="upload-service"}
   ```

3. 使用静态采集 Job（例如 user-service）筛选：

   ```logql
   {job="manual-go-video-user"}
   ```

4. 过滤 HTTP 访问日志（`message=http request`），并解析 JSON：

   ```logql
   {namespace="go-video"} |= "http request" | json
   ```

5. 根据 `request_id` 追踪一次请求在所有服务中的链路：

   ```logql
   {namespace="go-video"} |= "<your-request-id>" | json
   ```

   将 `<your-request-id>` 替换为实际的 `X-Request-ID` 值，即可看到这次请求在网关与下游各服务的日志。

6. 过滤错误日志：

   ```logql
   {namespace="go-video"} |= "ERROR"
   ```

---

## 5. 各环境下的使用建议

- 本地开发：
  - 直接看终端输出或本地日志文件（`logs/*.log`）。
  - 确保配置中的 `log.level` 为 `debug` 或 `info`，方便调试。
- 测试 / 预发 / 生产（K8s）：
  - 统一使用 `log.format=json`、`log.output=stdout`。
  - 通过 Promtail + Loki + Grafana 集中查询，并结合 `request_id`/`user_uuid` 进行链路分析。

通过上述链路，go-video 可以对各个服务的日志进行统一采集、结构化存储和集中查询，满足日常排查问题与运行观测的需求。

