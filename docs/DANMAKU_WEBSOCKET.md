# 弹幕 WebSocket 实时推送方案

## 整体架构

```
用户 A 发弹幕
  │  POST /api/video/danmaku
  ▼
gateway → interact-svc（写 MySQL + 通过 Redis Pub/Sub 广播）
  │
  │  Redis PUBLISH danmaku:{video_id} {json(danmaku)}
  │
  ▼
gateway WebSocket Server（订阅 Redis Channel）
  │  收到弹幕 → 推送给所有连接到此视频的 WebSocket 客户端
  │
  ▼
用户 B/C/D 浏览器（WebSocket 客户端）
  │  收到弹幕 → 渲染到播放器上方
```

## 数据流

### 发送弹幕（不变）

```
POST /api/video/danmaku
  { "video_id": 123, "content": "哈哈", "time": 45.2, "color": "#fff", "mode": 1 }
  → interact-svc
    → INSERT INTO danmakus
    → Redis PUBLISH danmaku:123 {json}
```

### 接收弹幕（新增）

```
浏览器连接 WebSocket
  ws://localhost:8888/ws/danmaku?video_id=123&token=xxx
  → gateway WebSocket Server
    → 验证 JWT
    → Redis SUBSCRIBE danmaku:123
    → 收到消息 → 推送给浏览器
```

### 历史弹幕（不变）

```
浏览器首次打开视频
  GET /api/video/danmakus?video_id=123&time=0
  → interact-svc → SELECT FROM danmakus WHERE video_id = 123 AND time BETWEEN 0 AND 10
  → 返回当前时间点附近的弹幕列表
```

## 改动点

### 后端

| 文件 | 改动 |
|------|------|
| `rpc/interact/internal/logic/senddanmakulogic.go` | 加 `Redis PUBLISH danmaku:{video_id}` |
| `rpc/interact/internal/svc/servicecontext.go` | 加 Redis 客户端 |
| `rpc/interact/internal/config/config.go` | 加 CacheRedis 字段 |
| `gateway/internal/ws/danmaku.go` | **新文件** WebSocket handler |
| `gateway/gateway.go` | 注册 WebSocket 路由 |

### Redis 设计

| key | 用途 | 类型 |
|-----|------|------|
| `danmaku:{video_id}` | 弹幕推送频道 | Pub/Sub Channel |

### 前端

| 文件 | 改动 |
|------|------|
| `frontend/src/pages/VideoDetail.vue` | 打开视频时连接 WebSocket，收到弹幕后渲染 |

## WebSocket 连接流程

```
1. 前端: new WebSocket("ws://localhost:8888/ws/danmaku?video_id=123&token=xxx")
2. gateway: 验证 token → 提取 user_id
3. gateway: go-redis Subscribe("danmaku:123")
4. gateway: 循环读取 channel 消息 → WriteMessage 到 ws 连接
5. 前端: ws.onmessage → 渲染弹幕到 DOM
6. 前端页面关闭: ws.close()
   gateway: 检测到连接断开 → 清理 goroutine
```

## 需要的依赖

| 包 | 用途 |
|----|------|
| `github.com/gorilla/websocket` | WebSocket 服务端 |
| go-redis 的 PubSub | 已在 interact-svc 使用 |

## 改动量预估

| 文件 | 行数 |
|------|------|
| senddanmaku logic 加 Redis Publish | 5 行 |
| interact svc 加 Redis | 10 行 |
| gateway ws/danmaku.go | 80 行 |
| gateway 注册路由 | 3 行 |
| 前端 VideoDetail | 30 行 |

**总工作量约 30 分钟。**

---

## 广播方案选型

| 方案 | 消息可靠性 | 延迟 | 复杂度 | 适用场景 |
|------|----------|------|--------|---------|
| **Redis Pub/Sub** | 无持久化，订阅者离线消息丢失 | < 1ms | ⭐ | 弹幕（允许丢） |
| Redis Stream | 持久化 + 消费者组 | < 1ms | ⭐⭐ | 弹幕 + 离线消息 |
| Kafka | 持久化 + 重放 + 多消费者 | 5-20ms | ⭐⭐⭐ | 转码任务 |
| NATS | 持久化 + 高性能 | < 1ms | ⭐⭐ | 弹幕 + 高并发 |
| WebSocket 直连 | 无中间件，直连广播 | 0ms | ⭐⭐ | 小规模 |

**选择 Redis Pub/Sub** 的原因：弹幕不需要持久化通道——持久化走 MySQL。Pub/Sub 延迟最低、零运维、go-redis 已引入。

---

## 弹幕存储策略

弹幕两份存储，各司其职：

| 存储 | 用途 | 何时读取 |
|------|------|---------|
| MySQL `danmakus` 表 | 历史弹幕持久化 | 用户打开视频时 `GetDanmakus` 加载 |
| Redis Pub/Sub | 实时推送通道 | 用户观看时 WebSocket 实时接收 |

离线用户不丢弹幕——下次打开视频从 MySQL 加载全部历史弹幕。
