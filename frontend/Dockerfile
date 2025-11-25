# 构建阶段
FROM node:18-alpine AS builder
WORKDIR /app
COPY frontend/package*.json ./
RUN npm ci
COPY frontend/. .
RUN npm run build

# 运行阶段
FROM nginx:alpine
COPY --from=builder /app/dist /usr/share/nginx/html
# 如需反代 API，可在这里 COPY 自定义 nginx.conf
EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]
