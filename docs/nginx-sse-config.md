## Nginx 中的 SSE 配置说明（通知流）

本项目里，通知服务通过 SSE（Server‑Sent Events）向前端推送事件：

- 路径：`/api/notification/v1/inner/notifications/stream`
- 用途：前端监听 `notification.created` / `notification.updated`，有事件时自动刷新通知列表和未读数。

在多层代理（前端 Nginx → Kong → notification-service）的场景下，如果某一层对响应做了缓冲或压缩，浏览器就可能**长期收不到任何字节**，SSE polyfill 会报：

> `Error: No activity within 45000 milliseconds. No response received. Reconnecting.`

本文件总结了一份在本项目实践过的 Nginx SSE 配置，方便以后参考或复制到其他环境。

---

### 1. SSE 对代理的核心要求

对任何中间层（Nginx、Kong、Ingress）的要求基本一致：

1. **保持长连接，不能随便超时/断开**
   - 使用 HTTP/1.1
   - 拉长 `read_timeout` / `send_timeout`
2. **关闭响应缓冲和缓存**
   - 对 SSE 路径关闭 `proxy_buffering`、`proxy_cache`
   - 给上游返回 `X-Accel-Buffering: no`（有些下游/子代理会尊重该头）
3. **不要压缩 SSE 响应**
   - `Accept-Encoding` 尽量清空/禁止 gzip，避免压缩再触发一层缓冲
4. **透传事件数据**
   - 不要对 `text/event-stream` 做特殊改写

只要满足以上几条，SSE 心跳（`: ping`）和事件（`event: ... data: ...`）就能原样到达浏览器。

---

### 2. 前端 Nginx（Docker 本地）配置示例

文件：`frontend/nginx.conf`

关键片段（只展示与 SSE 有关的部分）：

```nginx
server {
    listen 80;
    root /usr/share/nginx/html;
    index index.html;

    location / {
        try_files $uri $uri/ /index.html;
    }

    # 通知 SSE 专用配置：关闭缓冲/压缩，保持长连接
    location = /api/notification/v1/inner/notifications/stream {
        proxy_pass http://117.50.33.177:30080;  # 本地直连宿主上的 Kong
        proxy_http_version 1.1;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header Connection '';

        # 对 SSE 必须关闭缓冲和缓存，否则心跳和事件会被囤在 nginx 里
        proxy_buffering off;
        proxy_cache off;
        add_header X-Accel-Buffering no;

        # 禁用压缩，避免中间再做一层缓冲
        proxy_set_header Accept-Encoding "";

        # 适当拉长超时，避免长连接被过早切断
        proxy_read_timeout 600s;
        proxy_send_timeout 600s;
        send_timeout 600s;
    }

    # 其他 API 正常反向代理
    location /api/ {
        proxy_pass http://117.50.33.177:30080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    }
}
```

要点：

- 为 SSE 路径单独写一个 `location = /.../notifications/stream`，放在通用 `/api/` 之前；
- 只在这个路径上关闭缓冲/压缩和拉长超时，不影响其他接口。

---

### 3. 前端 Nginx（k8s ConfigMap）配置示例

线上环境使用的是 ConfigMap 注入 Nginx 配置：

- ConfigMap：`k8s/apps/frontend/frontend-nginx-config.yaml`
- Deployment：`k8s/apps/frontend/frontend.yaml`

ConfigMap 中的 `default.conf` 关键片段：

```nginx
default.conf: |
  server {
      listen 80;
      client_max_body_size 128m;
      proxy_request_buffering off;

      root /usr/share/nginx/html;
      index index.html;

      location / {
          try_files $uri $uri/ /index.html;
      }

      # 通知 SSE 专用配置：关闭缓冲/压缩，保持长连接
      location = /api/notification/v1/inner/notifications/stream {
          proxy_http_version 1.1;
          proxy_request_buffering off;
          proxy_read_timeout 600s;
          proxy_send_timeout 600s;
          proxy_pass http://gateway:8000;  # 集群内访问 Kong Service
          proxy_set_header Host $http_host;
          proxy_set_header X-Real-IP $remote_addr;
          proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;

          # 对 SSE 必须关闭缓冲和缓存，否则心跳和事件会被囤在 nginx 里
          proxy_buffering off;
          proxy_cache off;
          add_header X-Accel-Buffering no;

          # 禁用压缩，避免中间再做一层缓冲
          proxy_set_header Accept-Encoding "";
      }

      location /api/ {
          proxy_http_version 1.1;
          proxy_request_buffering off;
          proxy_read_timeout 300s;
          proxy_send_timeout 300s;
          proxy_pass http://gateway:8000;
          proxy_set_header Host $http_host;
          proxy_set_header X-Real-IP $remote_addr;
          proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
      }

      ...
  }
```

部署生效需要两步：

```bash
k3s kubectl -n go-video apply -f k8s/apps/frontend/frontend-nginx-config.yaml
k3s kubectl -n go-video rollout restart deployment/frontend
```

---

### 4. 常见现象与排查思路

#### 4.1 浏览器 Console 报：

> `Error: No activity within 45000 milliseconds. No response received. Reconnecting.`

但：

- 直连通知服务或 Kong（用 `curl -N`）能看到 `: ok` 和 `: ping`；
- Network 里 SSE 请求 `Status = 200`，却总是很快结束。

这通常表示：

- 有某一层代理对 SSE 响应做了缓冲/压缩，导致 45 秒内浏览器没收到任何字节；
- 80% 的情况就是某个 Nginx/Ingress/Kong 没关 `proxy_buffering`。

#### 4.2 本项目推荐的排查顺序

1. 在服务端用 `curl` 直连通知服务：
   ```bash
   curl -N "http://notification-service:8086/api/notification/v1/inner/notifications/stream" \
     -H "X-User-UUID: <user_uuid>"
   ```
   确认可以看到 `: ok` 和周期性的 `: ping`。

2. 直连 Kong：
   ```bash
   curl -N "http://gateway:8000/api/notification/v1/inner/notifications/stream" \
     -H "Authorization: Bearer <access_token>"
   ```

3. 直连前端 Nginx（NodePort）：
   ```bash
   curl -N "http://<node-ip>:30000/api/notification/v1/inner/notifications/stream" \
     -H "Authorization: Bearer <access_token>"
   ```

4. 最后在浏览器里看 Network / Console：
   - Network：`Status=200` 且 SSE 请求长期 pending；
   - Console：收到事件时应有类似：
     ```text
     [SSE] notification event received notification.created { unread_count: ... }
     ```

如果第 1/2 步 OK，第 3 步不 OK，多半是前端 Nginx 配置问题；  
如果第 3 步 OK，但浏览器仍报 45s 无活动，再检查本机 HTTP 代理（如 127.0.0.1:7890）是否在缓冲 SSE。

---

### 5. 小结

要让 SSE 在多层代理下稳定工作，关键是：

- **按路径精确关闭缓冲/缓存/压缩**；
- **拉长超时，避免长连接被中途切断**；
- **必要时从后向前逐层 `curl -N` 验证**。

本项目的前端 Nginx 配置和 Kong 环境变量已经按上述原则调整，后续如果新增 SSE 接口，建议参考同样的 pattern 配置对应的路由。***
