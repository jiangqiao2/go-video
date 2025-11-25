# 自定义 Kong 镜像，内置声明式配置
FROM kong:3.6
USER root
# 以仓库根作为构建上下文，明确路径
COPY gateway-service/kong.yml /usr/local/kong/declarative/kong.yml
ENV KONG_DATABASE=off
ENV KONG_DECLARATIVE_CONFIG=/usr/local/kong/declarative/kong.yml
ENV KONG_PROXY_LISTEN=0.0.0.0:8000
ENV KONG_ADMIN_LISTEN=0.0.0.0:8001
USER kong
