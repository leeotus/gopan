# GoPan VOD Platform

基于 go-zero 微服务框架搭建的视频点播平台（方案B标准生产版）。

## 项目结构

```
gopan/
├── api/
│   └── gateway.api              # API网关定义
├── gateway/                     # → API网关 (HTTP, 8888端口)
│   ├── etc/gateway.yaml
│   ├── gateway.go               # 入口
│   └── internal/
│       ├── config/              # 配置(含6个RPC client)
│       ├── handler/             # HTTP handler (user/video/search)
│       ├── logic/               # 业务编排层(调RPC)
│       ├── middleware/          # JWT鉴权
│       ├── svc/                 # ServiceContext
│       └── types/               # 请求/响应类型
├── rpc/
│   ├── user/                    # → 用户服务 (gRPC)
│   ├── video/                   # → 视频服务 (gRPC, 含model/store)
│   ├── transcode/               # → 转码服务 (gRPC, FFmpeg桩)
│   ├── stream/                  # → 流媒体服务 (gRPC, 防盗链签名)
│   ├── interact/                # → 互动服务 (gRPC, 点赞/评论/弹幕)
│   └── search/                  # → 搜索服务 (gRPC, ES桩)
├── common/response/             # 统一响应封装
└── etc/init.sql                 # 数据库建表脚本
```

## 已实现

| 模块 | 状态 |
|------|------|
| 7个服务骨架(goctl生成) | ✅ |
| 全部protobuf定义 | ✅ |
| API网关路由(23个接口) | ✅ |
| user-svc 注册/登录/个人信息(JWT+bcrypt) | ✅ 有DB |
| video-svc 列表/详情/更新/删除/转码回调 | ✅ 有DB+Store |
| stream-svc 播放地址+防盗链签名 | ✅ |
| transcode-svc 转码任务提交/查询 | 桩实现 |
| interact-svc 点赞/收藏/评论/弹幕 | 桩实现 |
| search-svc 搜索/索引 | 桩实现 |
| JWT Auth中间件 | ✅ |
| SQL建表脚本 | ✅ 6张表 |

## 服务架构

```
                    ┌─────────────────┐
                    │     gateway     │  (HTTP API, :8888)
                    │  JWT Auth 中间件  │
                    └────────┬────────┘
                             │ (gRPC via etcd)
          ┌──────────────────┼──────────────────┐
    ┌─────▼─────┐      ┌─────▼─────┐      ┌─────▼─────┐
    │  user-svc │      │ video-svc │      │stream-svc │
    │ 注册/登录  │      │ 上传/管理  │      │ 播放/签名  │
    └─────┬─────┘      └─────┬─────┘      └─────┬─────┘
          │                  │                    │
    ┌─────▼─────┐      ┌─────▼──────┐           │
    │interact-  │      │transcode-  │           │
    │   svc     │      │   svc      │           │
    │点赞/评论/  │      │ FFmpeg转码  │           │
    │弹幕/收藏   │      │ HLS多码率   │           │
    └───────────┘      └────────────┘           │
                             │                    │
                      ┌──────▼────────────────────▼──┐
                      │     MinIO / CDN / MySQL      │
                      └──────────────────────────────┘
```

## 技术栈

| 组件 | 选型 |
|------|------|
| 微服务框架 | go-zero v1.10.1 |
| RPC协议 | gRPC + Protobuf |
| API网关 | go-zero rest (HTTP) |
| 服务注册发现 | etcd |
| 数据库 | MySQL |
| 密码加密 | bcrypt |
| 鉴权 | JWT (golang-jwt/jwt/v4) |
| 对象存储 | MinIO |
| 转码 | FFmpeg (HLS多码率) |
| 搜索 | Elasticsearch |
| 消息队列 | Kafka / asynq |
| 缓存 | Redis |

## 数据库

6张核心表：`users`, `videos`, `transcodes`, `likes`, `favorites`, `comments`, `danmakus`

建表脚本：`etc/init.sql`

## 下一步待做

1. **配置 MinIO** → 完成 `video-svc` 的 `upload`/`merge` 实际存储
2. **配置 FFmpeg** → 完成 `transcode-svc` 的 HLS 多码率转码
3. **配置 Redis** → 完成 `interact-svc` 的点赞计数、`stream-svc` 的播放计数
4. **配置 Elasticsearch** → 完成 `search-svc` 的全文搜索
5. **配置 etcd** → 服务注册发现
6. **配置 Kafka/asynq** → video ↔ transcode 异步解耦
7. **完善 Auth** → JWT 签名验证、从 context 提取 user_id
8. **WebSocket** → 弹幕实时推送

## 快速开始

### 方式一：Docker Compose（推荐，一键启动全部中间件+服务）

```bash
# 1. 克隆项目
git clone <repo-url> gopan && cd gopan

# 2. 构建并启动所有容器（MySQL + Redis + etcd + MinIO + ES + 7个微服务）
make docker-up

# 等价于
docker compose up -d

# 3. 查看运行状态
docker compose ps

# 4. 查看日志
docker compose logs -f gateway
docker compose logs -f user-svc

# 5. 停止
make docker-down
# 等价于
docker compose down
```

启动后各服务端口：

| 服务 | 端口 | 说明 |
|------|------|------|
| gateway | 8888 | HTTP API 网关 |
| user-svc | 8081 | 用户服务 gRPC |
| video-svc | 8082 | 视频服务 gRPC |
| transcode-svc | 8083 | 转码服务 gRPC |
| stream-svc | 8084 | 流媒体服务 gRPC |
| interact-svc | 8085 | 互动服务 gRPC |
| search-svc | 8086 | 搜索服务 gRPC |
| mysql | 3306 | 数据库 (root/gopan123) |
| redis | 6379 | 缓存 |
| etcd | 2379 | 服务注册发现 |
| minio | 9000 | S3 对象存储 API |
| minio console | 9001 | MinIO Web 管理界面 |
| elasticsearch | 9200 | 搜索引擎 |

---

### 方式二：Make 本地开发

#### 前置依赖

```bash
# 需要本地安装以下服务并保持运行：
# - MySQL (3306, 数据库: gopan)
# - Redis (6379)
# - etcd (2379)
# - MinIO (9000)
# - Elasticsearch (9200, 可选)

# 初始化数据库
mysql -u root -p gopan < etc/init.sql
```

#### Make 命令速查

```bash
make help          # 查看所有可用命令
```

##### 编译

```bash
make build             # 编译所有服务到 build/ 目录
make build-gateway     # 仅编译 gateway
make build-user        # 仅编译 user-svc
make build-video       # 仅编译 video-svc
make build-transcode   # 仅编译 transcode-svc
make build-stream      # 仅编译 stream-svc
make build-interact    # 仅编译 interact-svc
make build-search      # 仅编译 search-svc
```

##### 运行

```bash
make run               # 一键编译并后台启动所有 7 个服务
make dev               # make stop + build + run (开发首选)
make run-gateway       # 仅启动 gateway
make run-user          # 仅启动 user-svc
```

##### 停止 & 状态

```bash
make stop              # 停止所有后台运行的服务
make status            # 查看所有服务运行状态 (✓ 运行中 / ✗ 已退出 / - 未启动)
```

##### 日志

```bash
make logs              # 实时查看所有服务日志
make logs-gateway      # 查看 gateway 日志
make logs-video        # 查看 video-svc 日志
```

##### 清理

```bash
make clean             # 删除 build/ 目录和 logs/ 目录
```

##### 代码生成

```bash
make proto             # 重新生成所有 protobuf 桩代码（修改 .proto 后执行）
make api               # 重新生成 gateway API 桩代码（修改 .api 后执行）
make gen               # proto + api 一起重新生成
```

##### 依赖

```bash
make deps              # go mod tidy + go mod download
```

##### 测试 & 检查

```bash
make test              # 运行所有测试 (go test ./...)
make lint              # 静态检查 (go vet ./...)
make fmt               # 格式化代码 (gofmt -s -w .)
```

---

### 方式三：手动启动（逐步调试）

```bash
# 0. 前置：确保 etcd / MySQL / Redis 已启动

# 1. 安装依赖
go mod tidy

# 2. 初始化数据库
mysql -u root -p < etc/init.sql

# 3. 按顺序启动各服务（每个终端一个，或后台运行）

# 先启动 RPC 服务
go run rpc/user/user.go -f rpc/user/etc/user.yaml &
go run rpc/video/video.go -f rpc/video/etc/video.yaml &
go run rpc/transcode/transcode.go -f rpc/transcode/etc/transcode.yaml &
go run rpc/stream/stream.go -f rpc/stream/etc/stream.yaml &
go run rpc/interact/interact.go -f rpc/interact/etc/interact.yaml &
go run rpc/search/search.go -f rpc/search/etc/search.yaml &

# 最后启动网关
go run gateway/gateway.go -f gateway/etc/gateway.yaml

# 4. 测试
curl http://localhost:8888/api/user/register \
  -X POST -H "Content-Type: application/json" \
  -d '{"username":"test","password":"123456","email":"test@test.com"}'

curl http://localhost:8888/api/user/login \
  -X POST -H "Content-Type: application/json" \
  -d '{"username":"test","password":"123456"}'

# 5. 停止
killall user video transcode stream interact search gateway
```

---

### API 接口一览

#### 用户模块 `prefix: /api/user`

| 方法 | 路径 | 说明 | 认证 |
|------|------|------|------|
| POST | `/api/user/register` | 注册 | ❌ |
| POST | `/api/user/login` | 登录 | ❌ |
| GET | `/api/user/profile` | 获取个人信息 | ❌ |
| PUT | `/api/user/profile` | 更新个人信息 | ❌ |

#### 视频模块 `prefix: /api/video` (需认证)

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/api/video/init-upload` | 初始化上传 |
| POST | `/api/video/upload-chunk` | 上传分片 |
| POST | `/api/video/merge-chunks` | 合并分片 |
| GET | `/api/video/list` | 视频列表 |
| GET | `/api/video/detail` | 视频详情 |
| PUT | `/api/video/update` | 更新视频信息 |
| DELETE | `/api/video/delete` | 删除视频 |
| GET | `/api/video/play-url` | 获取播放地址 |
| POST | `/api/video/like` | 点赞 |
| DELETE | `/api/video/like` | 取消点赞 |
| POST | `/api/video/favorite` | 收藏 |
| DELETE | `/api/video/favorite` | 取消收藏 |
| POST | `/api/video/comment` | 发表评论 |
| GET | `/api/video/comments` | 评论列表 |
| DELETE | `/api/video/comment` | 删除评论 |
| POST | `/api/video/danmaku` | 发送弹幕 |

#### 搜索模块 `prefix: /api/search`

| 方法 | 路径 | 说明 | 认证 |
|------|------|------|------|
| GET | `/api/search/videos` | 搜索视频 | ❌ |
