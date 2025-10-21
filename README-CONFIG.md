# 配置文件说明

## 概述

本项目使用配置文件模板（`.example` 文件）来管理敏感配置信息。实际的配置文件已被添加到 `.gitignore` 中，不会被提交到版本控制系统。

## 配置文件结构

```
├── configs/
│   ├── config.dev.yaml.example     # 根目录配置模板
│   └── config.prod.yaml.example    # 生产环境配置模板（如果存在）
├── user-service/configs/
│   ├── config.dev.yaml.example     # 用户服务配置模板
│   └── config.prod.yaml.example    # 生产环境配置模板（如果存在）
├── upload-service/configs/
│   ├── config.dev.yaml.example     # 上传服务配置模板
│   └── config.prod.yaml.example    # 生产环境配置模板（如果存在）
├── api-gateway/configs/
│   ├── config.dev.yaml.example     # API网关配置模板
│   └── config.prod.yaml.example    # 生产环境配置模板（如果存在）
└── transcode-service/configs/
    ├── config.dev.yaml.example     # 转码服务配置模板
    └── config.prod.yaml.example    # 生产环境配置模板（如果存在）
```

## 首次设置

### 1. 复制配置模板

为每个服务创建实际的配置文件：

```bash
# 根目录配置
cp configs/config.dev.yaml.example configs/config.dev.yaml

# 用户服务
cp user-service/configs/config.dev.yaml.example user-service/configs/config.dev.yaml

# 上传服务
cp upload-service/configs/config.dev.yaml.example upload-service/configs/config.dev.yaml

# API网关
cp api-gateway/configs/config.dev.yaml.example api-gateway/configs/config.dev.yaml

# 转码服务
cp transcode-service/configs/config.dev.yaml.example transcode-service/configs/config.dev.yaml
```

### 2. 配置敏感信息

编辑每个配置文件，替换以下占位符：

- `YOUR_DATABASE_PASSWORD` - 数据库密码
- `YOUR_REDIS_PASSWORD` - Redis密码
- `YOUR_JWT_SECRET_KEY` - JWT密钥
- `YOUR_MINIO_ACCESS_KEY` - MinIO访问密钥
- `YOUR_MINIO_SECRET_KEY` - MinIO密钥
- `YOUR_ETCD_ENDPOINT` - etcd服务地址
- `YOUR_ETCD_USERNAME` - etcd用户名
- `YOUR_ETCD_PASSWORD` - etcd密码
- `YOUR_EMAIL_USERNAME` - 邮件服务用户名
- `YOUR_EMAIL_PASSWORD` - 邮件服务密码
- `YOUR_SMS_ACCESS_KEY` - 短信服务访问密钥
- `YOUR_SMS_SECRET_KEY` - 短信服务密钥

## 环境特定配置

### 开发环境 (config.dev.yaml)
- 数据库：使用本地MySQL容器
- Redis：使用本地Redis实例
- 日志级别：debug
- 服务发现：使用etcd

### 生产环境 (config.prod.yaml)
- 数据库：使用生产数据库集群
- Redis：使用Redis集群
- 日志级别：info或warn
- 启用监控和追踪

## 安全注意事项

1. **永远不要提交实际的配置文件**到版本控制系统
2. **定期更换密钥和密码**，特别是在生产环境中
3. **使用强密码**和复杂的JWT密钥
4. **限制数据库和Redis的访问权限**
5. **在生产环境中启用SSL/TLS**

## 配置验证

启动服务前，确保：

1. 所有必需的配置项都已填写
2. 数据库连接信息正确
3. 外部服务（Redis、etcd、MinIO）可访问
4. 端口没有冲突
5. 日志目录有写权限

## 故障排除

### 常见问题

1. **数据库连接失败**
   - 检查数据库服务是否运行
   - 验证连接信息（主机、端口、用户名、密码）
   - 确认数据库已创建

2. **Redis连接失败**
   - 检查Redis服务状态
   - 验证密码和数据库编号

3. **服务发现失败**
   - 检查etcd服务状态
   - 验证etcd连接信息

4. **文件上传失败**
   - 检查MinIO服务状态
   - 验证访问密钥和存储桶配置

## 配置更新

当配置模板更新时：

1. 查看 `.example` 文件的变更
2. 相应地更新你的实际配置文件
3. 重启相关服务以应用新配置

## 联系方式

如有配置相关问题，请联系开发团队。