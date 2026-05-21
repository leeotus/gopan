# GoPan VOD Platform

基于 go-zero 微服务框架搭建的视频点播平台。  
后端 7 个 gRPC 服务 + 1 个 API 网关，前端 Vue 3 + Vant 暗色主题 SPA。

---

## 项目结构

```
gopan/
├── api/
│   └── gateway.api                # API 网关定义（goctl 源文件）
├── gateway/                       # API 网关 (HTTP :8888)
│   └── internal/
│       ├── config/                # 配置
│       ├── handler/               # HTTP handler（解析请求、调用 logic）
│       ├── logic/                 # 业务编排层（调下游 RPC）
│       ├── middleware/            # JWT 鉴权中间件
│       ├── svc/                   # ServiceContext（依赖注入容器）
│       └── types/                 # 请求/响应类型定义
├── rpc/
│   ├── user/                      # 用户服务 (gRPC :8081)
│   ├── video/                     # 视频服务 (gRPC :8082)
│   ├── transcode/                 # 转码服务 (gRPC :8083) Kafka 消费 + FFmpeg
│   ├── stream/                    # 流媒体服务 (gRPC :8084) 防盗链签名
│   ├── interact/                  # 互动服务 (gRPC :8085) 四表 CRUD
│   └── search/                    # 搜索服务 (gRPC :8086) Elasticsearch
├── common/
│   ├── storage/minio.go           # MinIO 客户端封装（video/transcode 共享）
│   ├── es/es.go                   # Elasticsearch 客户端封装
│   └── response/                  # 统一错误码
├── frontend/                      # Vue 3 前端
│   └── src/
│       ├── api/                   # axios 封装 + 后端接口
│       ├── stores/                # Pinia 状态（auth / video）
│       ├── router/                # 路由
│       ├── pages/                 # 页面（Home/Login/Register/VideoDetail/Search/Profile/Upload）
│       └── styles/                # 暗色紫色主题
├── etc/
│   └── init.sql                   # 数据库建表脚本（7 张表）
├── docs/
│   └── CODE_READING_GUIDE.md      # 源码阅读指南（含 Mermaid 流程图）
├── Dockerfile                     # 多阶段构建
├── docker-compose.yml             # 12 容器编排
├── Makefile                       # 编译/运行/停止/日志/代码生成
├── todo.md                        # AI 功能规划
└── .vscode/launch.json            # VSCode Debug 配置
```

---

## 模块状态

| 模块 | 状态 | 说明 |
|------|------|------|
| user-svc | ✅ | 注册/登录（bcrypt + JWT）、个人信息 CRUD |
| video-svc | ✅ | 视频列表/详情/更新/删除、MinIO 存储、Kafka Producer、ES 索引 |
| transcode-svc | ✅ | Kafka Consumer、FFmpeg 1080p HLS 转码、回调 video-svc |
| stream-svc | ✅ | 播放地址 + MD5 防盗链签名 |
| interact-svc | ✅ | 点赞/收藏/评论/弹幕四表 CRUD |
| search-svc | ✅ | Elasticsearch 全文搜索（multi_match）、索引管理 |
| Gateway | ✅ | JWT Auth、路由转发、23 个 API |
| Docker Compose | ✅ | etcd/MySQL/Redis/MinIO/Kafka/ES + 7 微服务 |
| 前端 Vue 3 | ✅ | 暗色紫色主题、卡片化 UI、注册/登录/列表/详情/搜索/个人中心 |
| 视频上传 | 🔧 | 前端占位提示，后端分片上传待实现 |
| CDN | ⏳ | 当前 MinIO 直连，后续接入 CDN |

---

## 服务架构

```
                        浏览器 / App
                             │
                    ┌────────▼────────┐
                    │    gateway      │  HTTP :8888
                    │  JWT Auth 中间件  │
                    └───────┬─────────┘
                            │ gRPC (etcd 服务发现)
          ┌─────────────────┼──────────────────┐
    ┌─────▼─────┐    ┌──────▼──────┐    ┌──────▼──────┐
    │ user-svc  │    │  video-svc  │    │ stream-svc  │
    │ :8081     │    │  :8082      │    │  :8084      │
    └─────┬─────┘    └──┬──┬──┬───┘    └─────────────┘
          │             │  │  │
    ┌─────▼─────┐       │  │  └──────────────┐
    │interact-  │       │  │                 │
    │   svc     │       │  │  Kafka          │
    │  :8085    │       │  │  gopan.         │
    └───────────┘       │  │  transcode.     │
                        │  │  tasks          │
                        │  │         ┌───────▼──────┐
                        │  │         │ transcode-svc│
                        │  │         │    :8083     │
                        │  │         │ FFmpeg HLS   │
                        │  │         └───────┬──────┘
                        │  │                 │ callback
                        │  └─────────────────┘
          ┌─────────────┼──────────────┐
    ┌─────▼─────┐  ┌────▼─────┐  ┌─────▼─────┐
    │  MySQL    │  │  MinIO   │  │  Redis    │
    │  :3306    │  │  :9000   │  │  :6379    │
    └───────────┘  └──────────┘  └───────────┘
    ┌──────────┐   ┌──────────┐
    │   etcd   │   │   ES     │
    │  :2379   │   │  :9200   │
    └──────────┘   └──────────┘
```

---

## 技术栈

| 组件 | 选型 |
|------|------|
| 微服务框架 | go-zero v1.10.1 |
| RPC | gRPC + Protobuf |
| API 网关 | go-zero rest (HTTP) |
| 服务发现 | etcd |
| 数据库 | MySQL 8.x |
| 对象存储 | MinIO |
| 消息队列 | Apache Kafka (KRaft) |
| 转码 | FFmpeg 1080p HLS (libx264 + AAC) |
| 搜索引擎 | Elasticsearch 8.x |
| 缓存 | Redis 7 |
| 鉴权 | JWT (golang-jwt/v4) + bcrypt |
| 前端 | Vue 3 + Vite + Pinia + Vant 4 |
| 部署 | Docker Compose (12 容器) |

---

## 数据库

7 张表：`users` `videos` `transcodes` `likes` `favorites` `comments` `danmakus`

建表脚本：`etc/init.sql`（MySQL 容器首次启动自动执行）

---

## 快速开始

### 方式一：Docker Compose（推荐）

```bash
# 构建并启动全部 12 个容器
make docker-up

# 等价于
docker compose up -d --build

# 查看状态
docker compose ps

# 查看日志
docker compose logs -f gateway

# 停止
make docker-down
```

**启动后端口**：见下方端口表。

### 方式二：Make 本地开发

前置条件：本地运行 etcd / MySQL / Redis / MinIO / Kafka / ES。

```bash
make dev        # 停止旧进程 → 编译 → 后台启动所有服务（使用 *.local.yaml）
make status     # 查看运行状态
make logs       # 实时查看全部日志
make stop       # 停止所有服务
make clean      # 清理 build/ 和 logs/
```

---

## 端口速查

| 服务 | 端口 | 说明 |
|------|------|------|
| gateway | 8888 | HTTP API 网关 |
| user-svc | 8081 | 用户服务 gRPC |
| video-svc | 8082 | 视频服务 gRPC |
| transcode-svc | 8083 | 转码服务 gRPC + Kafka Consumer |
| stream-svc | 8084 | 流媒体服务 gRPC |
| interact-svc | 8085 | 互动服务 gRPC |
| search-svc | 8086 | 搜索服务 gRPC |
| frontend | 3000 | Vue 开发服务器 (`npm run dev`) |
| mysql | 3306 | 数据库 (root / gopan123) |
| redis | 6379 | 缓存 |
| etcd | 2379 | 服务注册发现 |
| minio | 9000 | S3 API |
| minio console | 9001 | MinIO Web 管理界面 |
| kafka | 9092 | 消息队列 |
| elasticsearch | 9200 | 搜索引擎 |

---

## API 接口一览

### 用户模块 `/api/user`

| 方法 | 路径 | 说明 | 认证 |
|------|------|------|------|
| POST | `/api/user/register` | 注册 | ❌ |
| POST | `/api/user/login` | 登录 | ❌ |
| GET | `/api/user/profile` | 获取个人信息 | ❌ |
| PUT | `/api/user/profile` | 更新个人信息 | ❌ |

### 视频模块 `/api/video`（需 JWT 认证）

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/video/list` | 视频列表（分类/排序/游标分页） |
| GET | `/api/video/detail` | 视频详情（含转码信息） |
| PUT | `/api/video/update` | 更新视频 |
| DELETE | `/api/video/delete` | 删除视频 |
| GET | `/api/video/play-url` | 获取播放地址（防盗链签名） |
| POST | `/api/video/like` | 点赞 |
| DELETE | `/api/video/like` | 取消点赞 |
| POST | `/api/video/favorite` | 收藏 |
| DELETE | `/api/video/favorite` | 取消收藏 |
| POST | `/api/video/comment` | 发表评论 |
| GET | `/api/video/comments` | 评论列表 |
| DELETE | `/api/video/comment` | 删除评论 |
| POST | `/api/video/danmaku` | 发送弹幕 |
| POST | `/api/video/init-upload` | 初始化上传 🔧 |
| POST | `/api/video/upload-chunk` | 上传分片 🔧 |
| POST | `/api/video/merge-chunks` | 合并分片 🔧 |

### 搜索模块 `/api/search`

| 方法 | 路径 | 说明 | 认证 |
|------|------|------|------|
| GET | `/api/search/videos` | 搜索视频 | ❌ |

---

## 待完成

| 优先级 | 任务 |
|--------|------|
| P0 | 视频分片上传（video-svc uploadChunk + gateway upload handler） |
| P1 | 视频播放器集成 hls.js |
| P1 | 播放计数接入 Redis |
| P2 | 弹幕 WebSocket 实时推送 |
| P2 | CDN 接入（Nginx proxy_cache / 阿里云 CDN / Cloudflare） |
| P3 | AI 功能（见 `todo.md`） |

---

## 前端

```bash
cd frontend
npm install
npm run dev       # → http://localhost:3000
```

Vite 自动代理 `/api` → `http://localhost:8888`。详细说明见 `frontend/README.md`。

---

## Debug

VSCode：F5 选择服务即可断点调试（已配置 `.vscode/launch.json`，7 个服务各一个 configuration）。

gRPC 调试（需 DevMode）：
```bash
grpcurl -plaintext localhost:8081 list
grpcurl -plaintext -d '{"username":"test","password":"123"}' localhost:8081 user.User/Login
```

---

## 相关文档

- [源码阅读指南](docs/CODE_READING_GUIDE.md)
- [AI 功能规划](todo.md)
- [前端说明](frontend/README.md)
