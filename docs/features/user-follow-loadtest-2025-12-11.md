用户关注压测记录（2025-12-11）
=============================

背景
----

- 服务：`user-service` 的「用户关注（Follow）」写入路径。
- 网关入口：Kong NodePort `http://117.50.223.165:30080`。
- 接口：`POST /api/user/v1/inner/relation/follow`，通过网关转发到 `user-service`。
- 目标：模拟大量用户在短时间内关注同一“明星用户”，找到当前配置下系统的承载上限，并记录失败原因。

压测脚本与数据准备
--------------------

- 压测脚本：`loadtest/follow_stress.js`（k6）
  - 使用 `constant-arrival-rate` 模式控制 QPS。
  - 网关地址：默认 `http://117.50.223.165:30080`，可由 `GATEWAY` 环境变量覆盖。
  - 目标用户：通过 open API 注册登录 `star_user_1`，获取其 `user_uuid`，作为 `TARGET_UUID`。
  - 请求体：

    ```json
    { "target_uuid": "<Star user UUID>" }
    ```

  - Header：`Authorization: Bearer <fan access_token>`。
  - 粉丝账号：使用 `loadtest/run_follow_loadtest.sh` / `prepare_fans.ps1` 批量注册 `fan_0001`、`fan_0002`... 等账号，并登录生成 `tokens.txt`（每行一个 access_token）。

- 数据库配置（`user-service/configs/config.dev.yaml` 中）：

  ```yaml
  database:
    max_idle_conns: 10
    max_open_conns: 100
    conn_max_lifetime: 3600s
  ```

  说明：这是 `user-service` 进程内部的连接池设置；MySQL 服务器本身还有一个全局 `max_connections` 限制（默认通常在 151 左右，取决于实际配置）。


测试 1：500 QPS，5 分钟
-----------------------

- 配置（`follow_stress.js`）：

  ```js
  rate: 500,
  duration: '5m',
  ```

- 结果摘要：

  - 总请求数：`http_reqs ≈ 149,922`（约 500 QPS）
  - 失败率：`http_req_failed ≈ 0.00%`（仅 2 次 timeout）
  - 延迟：
    - `avg ≈ 51 ms`
    - `p90 ≈ 62 ms`
    - `p95 ≈ 77 ms`
    - `max ≈ 5 s`（对应个别超时请求）
  - 日志：`user-service` 基本正常，仅偶发 `context canceled`，可视为客户端超时/取消导致。

- 结论：

  - 在当前配置下，**500 QPS 对同一用户的 Follow 写请求可稳定运行 5 分钟**：
    - 成功率接近 100%
    - p95 延迟 < 80 ms
  - 500 QPS 可以作为目前的「安全基线」。


测试 2：1500 QPS，5 分钟
------------------------

- 配置（`follow_stress.js`）：

  ```js
  rate: 1500,
  duration: '5m',
  ```

- 结果摘要：

  - 总请求数：`http_reqs ≈ 420,662`（约 1,400 QPS 实际达成）
  - 失败率：`http_req_failed ≈ 12.53%`（失败请求约 52,713 个）
  - 延迟：
    - `avg ≈ 253 ms`
    - `med ≈ 128 ms`
    - `p90 ≈ 497 ms`
    - `p95 ≈ 936 ms`
    - `max ≈ 5 s`
  - dropped_iterations：`≈ 29,340`，说明在目标 1500 QPS 下，部分请求因系统压力被 k6 丢弃。
  - k6 日志：大量 `request timeout`，即客户端等待 5 秒仍未收到响应。

- `user-service` 日志中的关键错误：

  1）业务层检查目标用户是否存在时的 `context canceled`：

  ```json
  {
    "level": "ERROR",
    "message": "Follow exists is err context canceled",
    "file": "social_controller.go",
    "function": "http.(*socialControllerImpl).Follow"
  }
  ```

  2）关注写入时的数据库错误：

  ```json
  {
    "level": "ERROR",
    "message": "follow upsert user_uuid: ..., target_uuid: ... error: Error 1040: Too many connections",
    "file": "follow_repository.go",
    "function": "persistence.(*followRepositoryImpl).Follow"
  }
  ```

  其中 `Error 1040: Too many connections` 是 MySQL 返回的典型错误，表示服务器当前总连接数已经达到 `max_connections` 上限，新连接被拒绝。


失败原因分析
------------

综合 k6 和服务日志，可以得出：

1. 1500 QPS 下，`user-service` 的 Follow 请求显著增加了对数据库的并发访问：
   - 每个请求都会调用 `userRepo.ExistsByUUID` 检查目标用户是否存在；
   - 然后调用 `followRepo.Follow`，最终通过 `FollowDao.Upsert` 写入/更新 `user_follow` 表。

2. MySQL 端出现 `Error 1040: Too many connections`：
   - 说明 MySQL 全局连接数已经达到 `max_connections` 上限；
   - 新的连接/查询被拒绝，GORM 将该错误返回给应用；
   - `user-service` 记录了 `follow upsert ... error: Error 1040: Too many connections` 的错误日志。

3. 同时，部分请求在执行过程中因为客户端超时或连接中断，导致服务端 `context` 被取消：
   - `social_app.go` 中 `ExistsByUUID` 使用的是带 context 的 GORM 查询；
   - 当上游（k6 或网关）在等待超时后主动断开连接，Gin 会取消请求的 context；
   - GORM 随之返回 `context canceled`，日志中出现 `"Follow exists is err context canceled"`。

4. 对应到 k6 的表现：
   - 约 12.5% 的请求失败，主要原因是：
     - 数据库拒绝连接（`Too many connections`）导致 Follow 写入失败；
     - 请求处理时间超过 k6 设置的 5 秒超时，客户端主动放弃，记录为 `request timeout`。
   - 延迟分布明显恶化（p95 接近 1 秒），说明系统已接近或超过当前配置下的容量上限。

综上，**这次压测在 1500 QPS 时的主要失败原因是：MySQL 连接数达到上限（max_connections），导致 Follow 写入和部分查询失败，从而引发请求超时和错误率上升。**


潜在优化方向
------------

1. 数据库层面
   - 根据机器配置适当提升 MySQL 的 `max_connections`，并确保：
     - MySQL 实例有足够的 CPU / 内存支撑更高并发；
     - 监控连接数、QPS 和慢查询，避免盲目放大。

2. 应用层连接池
   - 根据数据库配置，适当调优 `user-service` 中的：

     ```yaml
     max_open_conns: 100
     max_idle_conns: 10
     ```

   - 避免单个服务进程持有过多闲置连接，同时又不能太小导致频繁建连。

3. Follow 写入路径优化
   - 当前路径：每次 Follow 都是同步检查 + 写入数据库：
     - 检查目标用户是否存在（`ExistsByUUID`）
     - Upsert `user_follow` 记录
   - 可以考虑：
     - 用缓存/本地缓存减少对用户表的存在性检查压力；
     - 对写入做削峰（排队 / 限流）或异步化（消息队列）；
     - 配合批量写/合并写，降低高峰期对 MySQL 的直接冲击。

4. 压测策略
   - 保留本次的 500 QPS / 1500 QPS 结果作为基准：
     - 500 QPS：当前配置下稳定；
     - 1500 QPS：触发 MySQL 连接数瓶颈。
   - 以后做新一轮调优（DB 配置或应用改动）后，可以复用 `loadtest` 脚本，以同样的 QPS 组合重新对比：
     - 是否降低了错误率；
     - 是否改善了延迟分布；
     - 容量上限是否提升。


结论
----

这次压测验证了「同一用户被大量关注」场景下系统的表现：

- 当前配置下，**500 QPS 可以稳定支撑，1500 QPS 会触发数据库连接数上限**；
- 失败的直接原因是 MySQL `max_connections` 被打满，导致 Follow 写入出现 `Error 1040: Too many connections`；
- 这为后续的容量规划、数据库配置调优和关注写入路径优化提供了一个清晰的参考点。

这次事故级别的压测结果是非常宝贵的实践样本，后续所有与关注系统容量相关的改动，建议都以本次记录为基线做对比回归。

