# 本地调试指南

## 架构

```
                    ┌──────────────┐
                    │   gateway    │  :8888 (REST API)
                    └──────┬───────┘
        ┌──────────────────┼──────────────────┐
   ┌────┴────┐ ┌──────┴──────┐ ┌──────┴──────┐
   │user-svc │ │ video-svc   │ │ search-svc  │
   │ :8081   │ │ :8082       │ │ :8086       │
   └─────────┘ └──────┬──────┘ └──────┬──────┘
                      │               │
              ┌───────┴───────┐       │
              │ transcode-svc │       │
              │ :8083         │       │
              └───────────────┘       │
              ┌───────┴───────┐       │
              │ stream-svc    │       │
              │ :8084         │       │
              └───────────────┘       │
              ┌───────┴───────┐       │
              │ interact-svc  │       │
              │ :8085         │       │
              └───────────────┘       │
              ┌───────┴───────┐       │
              │ admin-svc     │       │
              │ :8087         │       │
              └───────────────┘       │
                                      │
    ┌─────────┬─────────┬────────────┼──────────┬──────────┐
    │  MySQL  │  Redis  │   MinIO    │   etcd   │  Kafka   │
    │  :3306  │  :6379  │ :9000/9001 │  :2379   │  :9092   │
    └─────────┴─────────┴────────────┴──────────┴──────────┘
    ┌───────────────┐ ┌──────────┐ ┌─────────────┐
    │ Elasticsearch │ │  Jaeger  │ │  Prometheus │
    │ :9200         │ │  :16686  │ │  :9090      │
    └───────────────┘ └──────────┘ └─────────────┘
    ┌──────────┐
    │ Grafana  │
    │ :3001    │
    └──────────┘
```

中间件 6 个通过 Docker 启动，微服务 8 个在 VSCode 中通过 Delve 调试。

---

## 1. 启动基础中间件

```bash
docker compose up -d etcd mysql redis minio kafka elasticsearch
```

可选启动可观测性面板：
```bash
docker compose up -d jaeger prometheus grafana
```

### 各中间件端口

| 中间件 | 端口 | 用途 |
|--------|------|------|
| etcd | 2379 | 服务发现与注册 |
| MySQL | 3306 | 关系型数据（用户、视频、评论） |
| Redis | 6379 | 缓存、会话、限流 |
| MinIO | 9000/9001 | 对象存储（视频文件） |
| Kafka | 9092 | 异步消息（转码、摘要生成） |
| Elasticsearch | 9200 | 全文/语义搜索 |

---

## 2. VSCode 调试

### 方式 A：逐个启动（推荐）

1. 打开 `Run and Debug` 面板（`Ctrl+Shift+D`）
2. 下拉选择要调试的服务
3. 按 `F5` 启动

**推荐启动顺序**：

| 顺序 | 配置名 | 端口 | 用途 |
|------|--------|------|------|
| 1 | Debug user-svc | 8081 | 用户鉴权 |
| 2 | Debug video-svc | 8082 | 视频管理 |
| 3 | Debug search-svc | 8086 | 搜索索引 |
| 4 | Debug stream-svc | 8084 | CDN 签名 |
| 5 | Debug interact-svc | 8085 | 弹幕/评论/点赞 |
| 6 | Debug transcode-svc | 8083 | 视频转码 |
| 7 | Debug admin-svc | 8087 | 管理后台 |
| 8 | Debug gateway | 8888 | API 网关（最后启动） |

> **不需要全启动**。只启动你调试的微服务 + `gateway` 即可。其他依赖的服务会因 go-zero 的 graceful 降级（返回 Internal Error）而不是 panic。

### 方式 B：Makefile 批量启动

```bash
make dev          # 编译所有服务，以 nohup 方式启动
make stop         # 停止所有服务
make status       # 查看运行状态
make logs         # 实时查看所有日志
```

Makefile 方式不支持断点调试，适合快速验证。

---

## 3. local.yaml 配置说明

每个微服务都有两套配置：

| 文件 | 用途 | 地址格式 |
|------|------|----------|
| `*.yaml` | Docker 部署 | 容器名（`mysql`/`etcd`/...） |
| `*.local.yaml` | 本地调试 | `127.0.0.1` |

VSCode 调试会自动选择 `*.local.yaml`。无需手动切换。

---

## 4. 前端调试

```bash
cd frontend
npm install
npm run dev      # 默认 :5173，API 指向 :8888
```

前端开发服务器通过 `vite.config.js` 的 proxy 转发 API 请求到 `localhost:8888`。

---

## 5. 数据库连入

```bash
# 通过 docker exec
docker exec -it gopan-mysql mysql -uroot -pgopan123 gopan

# 或本地 MySQL 客户端
mysql -h 127.0.0.1 -uroot -pgopan123 gopan
```

---

## 6. 常见问题

### Q: 启动微服务报 "context deadline exceeded"

中间件未就绪。确保 `docker compose ps` 显示所有中间件为 `healthy`。

### Q: VSCode 断点不停

编译优化导致。在 `launch.json` 中设置 `"mode": "debug"` 即可。go-zero 使用 `go build -gcflags='all=-N -l'` 禁用内联和优化。

### Q: 端口冲突

检查是否有其他服务占用 `8081-8087`、`8888`：
```bash
lsof -i :8081
```

### Q: 需要修改配置

编辑对应服务的 `etc/*.local.yaml`，无需重启 VSCode（`F5` 重新启动即可）。

---

## 7. 日志查看

```bash
# Docker 中间件日志
docker compose logs -f --tail=50 mysql

# 本地微服务（Makefile 方式启动的）
make logs

# 或单独查看
tail -f logs/gateway.log
```

---

## 8. 清理

```bash
make stop                 # 停止本地微服务
docker compose down       # 停止所有容器
docker compose down -v    # 删除数据卷（清空所有数据）
```
