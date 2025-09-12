# 转码服务 (Transcode Service)

基于DDD架构的分布式视频转码服务，支持多Worker并行处理，具备任务调度、状态管理、失败重试等完整功能。

## 🏗️ 架构设计

### 技术栈

- **语言**: Go 1.21+
- **Web框架**: Gin v1.10+
- **ORM**: GORM v1.25+
- **数据库**: MySQL 8.0
- **缓存**: Redis 7
- **对象存储**: MinIO
- **消息队列**: RabbitMQ (可选)
- **容器化**: Docker & Docker Compose
- **监控**: Prometheus + Grafana

### 架构模式

- **DDD (Domain-Driven Design)**: 领域驱动设计
- **Clean Architecture**: 清洁架构
- **微服务架构**: 调度器 + 多Worker模式
- **事件驱动**: 异步任务处理

## 📁 项目结构

```
transcode-service/
├── cmd/                          # 应用程序入口
│   ├── scheduler/               # 调度器服务
│   └── worker/                  # Worker服务
├── ddd/                         # DDD核心模块
│   ├── adapter/                 # 适配器层
│   │   └── http/               # HTTP控制器
│   ├── application/             # 应用层
│   │   ├── app/                # 应用服务
│   │   └── dto/                # 数据传输对象
│   ├── domain/                  # 领域层
│   │   ├── entity/             # 实体
│   │   ├── vo/                 # 值对象
│   │   ├── repo/               # 仓储接口
│   │   ├── service/            # 领域服务
│   │   └── gateway/            # 网关接口
│   └── infrastructure/          # 基础设施层
│       ├── database/           # 数据库实现
│       ├── ffmpeg/             # FFmpeg实现
│       └── queue/              # 消息队列实现
├── configs/                     # 配置文件
├── deployments/                 # 部署配置
├── scripts/                     # 脚本文件
└── docs/                        # 文档
```

## 🚀 快速开始

### 前置要求

- Docker 20.0+
- Docker Compose 2.0+
- Go 1.21+ (开发环境)

### 一键启动

```bash
# 克隆项目
git clone <repository-url>
cd transcode-service

# 启动所有服务
./start.sh start

# 启动服务并包含监控
./start.sh start --with-monitoring
```

### 手动启动

```bash
# 1. 启动基础设施
docker-compose up -d mysql redis minio rabbitmq

# 2. 等待服务就绪
sleep 30

# 3. 启动调度器
docker-compose up -d scheduler

# 4. 启动Worker
docker-compose up -d worker-1 worker-2
```

## 📊 服务访问

启动成功后，可以通过以下地址访问各项服务：

| 服务 | 地址 | 用户名/密码 |
|------|------|-------------|
| 调度器API | http://localhost:8082 | - |
| 健康检查 | http://localhost:8082/health | - |
| API文档 | http://localhost:8082/swagger/index.html | - |
| MySQL | localhost:3307 | transcode_user/transcode_password |
| Redis | localhost:6380 | - |
| MinIO | http://localhost:9003 | minioadmin/minioadmin123 |
| RabbitMQ | http://localhost:15673 | admin/admin123 |
| Prometheus | http://localhost:9091 | - |
| Grafana | http://localhost:3001 | admin/admin123 |

## 🔧 API 使用示例

### 创建转码任务

```bash
curl -X POST http://localhost:8082/api/v1/tasks \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "user-123",
    "source_video_path": "/path/to/input.mp4",
    "output_path": "/path/to/output.mp4",
    "config": {
      "resolution": "1280x720",
      "bitrate": "2000k",
      "codec": "libx264",
      "preset": "medium",
      "format": "mp4"
    },
    "priority": 5
  }'
```

### 查询任务状态

```bash
curl http://localhost:8082/api/v1/tasks/{task_id}
```

### 获取任务列表

```bash
curl "http://localhost:8082/api/v1/tasks?user_id=user-123&status=processing&limit=10"
```

### 注册Worker

```bash
curl -X POST http://localhost:8082/api/v1/workers \
  -H "Content-Type: application/json" \
  -d '{
    "worker_id": "worker-003",
    "name": "Worker-003",
    "max_tasks": 4
  }'
```

### 获取统计信息

```bash
# 任务统计
curl http://localhost:8082/api/v1/tasks/statistics

# Worker统计
curl http://localhost:8082/api/v1/workers/statistics
```

## 🏛️ 分阶段实现计划

### 阶段1：单机原型 ✅
- [x] 基础DDD架构
- [x] 调度器 + 1个Worker
- [x] 手动插入任务到DB
- [x] Worker拉任务执行转码
- [x] 转码结果写回DB

### 阶段2：多Worker基础主从
- [ ] 支持多台Worker并行转码
- [ ] 调度器负责分配任务
- [ ] Worker上报状态给调度器
- [ ] Worker心跳检测
- [ ] DB统一由调度器更新

### 阶段3：任务调度优化
- [ ] 引入任务队列（Redis Stream）
- [ ] 支持任务失败重试
- [ ] 根据Worker负载分配任务
- [ ] 任务优先级调度

### 阶段4：调度器高可用
- [ ] 使用etcd选举Leader
- [ ] 多调度器实例
- [ ] Leader负责分配任务
- [ ] Leader挂掉自动切换

### 阶段5：存储优化 & 扩展能力
- [ ] 支持多存储副本（MinIO集群）
- [ ] 异步/批量写入DB
- [ ] 秒转逻辑（符合标准直接跳过）
- [ ] 支持任务优先级调度

### 阶段6：监控 & 运维
- [ ] Worker CPU/GPU监控
- [ ] 任务成功率、耗时统计
- [ ] 告警系统（Prometheus + Grafana）
- [ ] 支持Worker弹性扩缩容

## 🛠️ 开发指南

### 本地开发

```bash
# 安装依赖
go mod tidy

# 运行调度器
go run cmd/scheduler/main.go

# 运行Worker
go run cmd/worker/main.go
```

### 添加新功能

1. **领域层**: 在 `domain/` 中定义实体、值对象、仓储接口
2. **基础设施层**: 在 `infrastructure/` 中实现具体的技术细节
3. **应用层**: 在 `application/` 中实现用例和DTO
4. **适配器层**: 在 `adapter/` 中实现HTTP控制器

### 测试

```bash
# 运行所有测试
go test ./...

# 运行特定包的测试
go test ./ddd/domain/...

# 生成测试覆盖率报告
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

## 📋 管理命令

```bash
# 查看服务状态
./start.sh status

# 查看日志
./start.sh logs                # 所有服务日志
./start.sh logs scheduler      # 调度器日志
./start.sh logs worker-1       # Worker-1日志

# 重启服务
./start.sh restart

# 停止服务
./start.sh stop

# 清理所有数据
./start.sh clean

# 重新构建镜像
./start.sh build
```

## 🔍 故障排查

### 常见问题

1. **服务启动失败**
   ```bash
   # 检查Docker状态
   docker ps -a
   
   # 查看服务日志
   docker-compose logs scheduler
   ```

2. **数据库连接失败**
   ```bash
   # 检查MySQL是否就绪
   docker-compose exec mysql mysql -u transcode_user -p -e "SELECT 1"
   ```

3. **Worker无法连接调度器**
   ```bash
   # 检查网络连通性
   docker-compose exec worker-1 curl http://scheduler:8082/health
   ```

### 日志位置

- 应用日志: `./logs/`
- Docker日志: `docker-compose logs [service]`
- 系统日志: `/var/log/transcode-service/`

## 📈 监控指标

### 任务指标
- 任务总数、待处理、处理中、已完成、失败数量
- 任务平均处理时间
- 任务成功率

### Worker指标
- Worker总数、在线、离线、忙碌、空闲数量
- Worker CPU、内存使用率
- Worker负载因子

### 系统指标
- API响应时间
- 数据库连接数
- 队列长度

## 🤝 贡献指南

1. Fork 项目
2. 创建特性分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 打开 Pull Request

## 📄 许可证

本项目采用 MIT 许可证 - 查看 [LICENSE](LICENSE) 文件了解详情。

## 🙏 致谢

- [Gin Web Framework](https://github.com/gin-gonic/gin)
- [GORM](https://github.com/go-gorm/gorm)
- [FFmpeg](https://ffmpeg.org/)
- [Docker](https://www.docker.com/)
- [Domain-Driven Design](https://domainlanguage.com/ddd/)