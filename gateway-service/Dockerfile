# 自定义 Kong 镜像，内置声明式配置
FROM kong:3.6
USER root

# kong 官方镜像基于 Debian/Ubuntu，这里用 apt 安装 tzdata 并设置时区
RUN apt-get update && \
    DEBIAN_FRONTEND=noninteractive apt-get install -y --no-install-recommends tzdata && \
    ln -snf /usr/share/zoneinfo/Asia/Shanghai /etc/localtime && \
    echo 'Asia/Shanghai' > /etc/timezone && \
    rm -rf /var/lib/apt/lists/*

# 拷贝自定义插件到 Kong 插件目录
COPY gateway-service/plugins/jwt-user-uuid-header /usr/local/share/lua/5.1/kong/plugins/jwt-user-uuid-header

# 以仓库根作为构建上下文，明确路径
COPY gateway-service/kong.yml /usr/local/kong/declarative/kong.yml
ENV KONG_DATABASE=off
ENV KONG_DECLARATIVE_CONFIG=/usr/local/kong/declarative/kong.yml
ENV KONG_PROXY_LISTEN=0.0.0.0:8000
ENV KONG_ADMIN_LISTEN=0.0.0.0:8001

# 启用自定义插件（在 builtin 插件集合基础上追加）
ENV KONG_PLUGINS=bundled,jwt-user-uuid-header

USER kong
