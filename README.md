# Go Video Platform - DDD架构实现

基于领域驱动设计(DDD)和Clean Architecture的Go视频平台项目。

## 🏗️ 架构设计

### 技术栈

- **语言**: Go 1.18+
- **Web框架**: Gin v1.7.7
- **ORM**: GORM v1.22.5
- **数据库**: MySQL
- **缓存**: Redis (go-redis/v9)
- **认证**: JWT
- **配置管理**: YAML
- **日志**: 结构化日志

### 架构模式

- **DDD (Domain-Driven Design)**: 领域驱动设计
- **Clean Architecture**: 清洁架构
- **CQRS**: 命令查询职责分离
- **Plugin Pattern**: 插件化架构
- **Singleton Pattern**: 单例模式

## 📁 项目结构

```
go-video/
├── cmd/                    # 应用程序入口
│   └── api/               # API服务入口
│       └── main.go
├── ddd/                   # DDD核心模块
│   ├── internal/          # 内部资源
│   │   └── resource/      # 资源管理(MySQL、JWT等)
│   └── user/              # 用户领域模块
│       ├── adapter/       # 适配器层
│       │   ├── http/      # HTTP控制器
│       │   └── event/     # 事件处理器
│       ├── application/   # 应用层
│       │   ├── app/       # 应用服务
│       │   ├── cqe/       # 命令查询事件
│       │   └── dto/       # 数据传输对象
│       ├── domain/        # 领域层
│       │   ├── entity/    # 实体
│       │   ├── vo/        # 值对象
│       │   ├── repo/      # 仓储接口
│       │   └── service/   # 领域服务
│       ├── infrastructure/ # 基础设施层
│       │   └── database/  # 数据库实现
│       │       ├── dao/   # 数据访问对象
│       │       ├── po/    # 持久化对象
│       │       ├── persistence/ # 仓储实现
│       │       └── convertor/   # 转换器
│       └── init.go        # 模块初始化
├── pkg/                   # 公共包
│   ├── assert/           # 断言工具
│   ├── config/           # 配置管理
│   ├── database/         # 数据库管理
│   ├── errno/            # 错误处理
│   ├── manager/          # 组件管理器
│   ├── restapi/          # REST API响应
│   └── utils/            # 工具包
├── configs/              # 配置文件
├── docs/                 # 文档
└── scripts/              # 脚本
```

## 🚀 快速开始

### 环境要求

- Go 1.18+
- MySQL 8.0+
- Redis 6.0+ (可选)

### 安装依赖

```bash
go mod download
```

### 配置数据库

1. 创建MySQL数据库:
```sql
CREATE DATABASE go_video CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
```

2. 修改配置文件 `configs/config.dev.yaml`:
```yaml
server:
  host: "0.0.0.0"
  port: 8080
  mode: "debug"

database:
  host: "localhost"
  port: 3306
  username: "root"
  password: "password"
  database: "go_video"
  max_idle_conns: 10
  max_open_conns: 100
  conn_max_lifetime: 3600s

jwt:
  secret: "your-secret-key"
  expire_time: 24h
  refresh_expire_time: 168h
```

### 运行应用

```bash
# 开发模式
go run cmd/api/main.go

# 编译运行
go build -o bin/api cmd/api/main.go
./bin/api
```

### 健康检查

```bash
curl http://localhost:8080/health
```

## 📚 API文档

### 用户相关API

#### 开放API (无需认证)

- `POST /api/v1/users/register` - 用户注册
- `POST /api/v1/users/login` - 用户登录

#### 内部API (需要认证)

- `GET /api/v1/users/me` - 获取当前用户信息
- `PUT /api/v1/users/me` - 更新用户资料
- `PUT /api/v1/users/me/password` - 修改密码

#### 运维API (需要管理员权限)

- `GET /ops/v1/users` - 获取用户列表
- `GET /ops/v1/users/:id` - 获取用户详情
- `PUT /ops/v1/users/:id/activate` - 激活用户
- `PUT /ops/v1/users/:id/disable` - 禁用用户
- `DELETE /ops/v1/users/:id` - 删除用户

#### 调试API

- `POST /debug/v1/users/validate-token` - 验证令牌

### 请求示例

#### 用户注册

```bash
curl -X POST http://localhost:8080/api/v1/users/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "john_doe",
    "password": "password123",
    "email": "john@example.com"
  }'
```

#### 用户登录

```bash
curl -X POST http://localhost:8080/api/v1/users/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "john_doe",
    "password": "password123"
  }'
```

## 🏛️ 架构特性

### DDD四层架构

1. **适配器层 (Adapter)**: 处理外部接口，如HTTP控制器、事件处理器
2. **应用层 (Application)**: 协调领域对象，实现用例
3. **领域层 (Domain)**: 核心业务逻辑，包含实体、值对象、领域服务
4. **基础设施层 (Infrastructure)**: 技术实现，如数据库访问、外部服务调用

### 单例模式实现

```go
var (
    componentOnce sync.Once
    singletonComponent ComponentType
)

func DefaultComponent() ComponentType {
    assert.NotCircular()  // 循环依赖检查
    componentOnce.Do(func() {
        singletonComponent = &componentImpl{
            // 依赖注入
        }
    })
    assert.NotNil(singletonComponent)
    return singletonComponent
}
```

### 错误处理机制

- **BizError**: 业务错误类型，包含错误码和调用栈
- **统一响应格式**: 标准化的REST API响应
- **错误码定义**: 预定义的业务错误码

### 插件化架构

- **ServicePlugin**: 服务插件接口
- **ControllerPlugin**: 控制器插件接口
- **ComponentPlugin**: 组件插件接口
- **依赖注入**: 支持构造函数注入和单例模式

## 🔧 开发指南

### 添加新领域模块

1. 在 `ddd/` 目录下创建新的领域目录
2. 按照四层架构创建子目录
3. 实现领域实体、值对象、仓储接口
4. 实现基础设施层的数据访问
5. 实现应用层的用例
6. 实现适配器层的控制器
7. 创建插件初始化文件

### 代码规范

- 使用Go标准代码格式
- 接口定义在使用方包内
- 所有公共方法添加注释
- 遵循DDD设计原则
- 使用依赖注入

### 测试

```bash
# 运行所有测试
go test ./...

# 运行特定包的测试
go test ./ddd/user/...

# 生成测试覆盖率报告
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

## 🚀 部署

### Docker部署

```bash
# 构建镜像
docker build -t go-video .

# 运行容器
docker run -p 8080:8080 go-video
```

### Docker Compose

```bash
docker-compose up -d
```

## 📈 监控和日志

- 健康检查端点: `/health`
- 结构化日志记录
- 错误追踪和调用栈
- 性能监控钩子

## 🤝 贡献指南

1. Fork 项目
2. 创建特性分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 打开 Pull Request

## 📄 许可证

本项目采用 MIT 许可证 - 查看 [LICENSE](LICENSE) 文件了解详情。

## 📞 联系方式

- 项目链接: [https://github.com/your-username/go-video](https://github.com/your-username/go-video)
- 问题反馈: [Issues](https://github.com/your-username/go-video/issues)

## 🙏 致谢

- [Gin Web Framework](https://github.com/gin-gonic/gin)
- [GORM](https://github.com/go-gorm/gorm)
- [JWT-Go](https://github.com/golang-jwt/jwt)
- [Domain-Driven Design](https://domainlanguage.com/ddd/)
- [Clean Architecture](https://blog.cleancoder.com/uncle-bob/2012/08/13/the-clean-architecture.html)