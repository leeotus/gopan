<div align="center">

<img src="imgs/logo.svg" alt="GoPan" width="420"/>

# GoPan VOD Platform

**Go-zero 微服务 + Vue 3 + 双 AI 引擎 的全栈视频点播平台**

[![Go](https://img.shields.io/badge/Go-1.21+-00ADD8?logo=go)](https://go.dev)
[![Vue](https://img.shields.io/badge/Vue-3-42b883?logo=vuedotjs)](https://vuejs.org)
[![Docker](https://img.shields.io/badge/Docker-Compose-2496ED?logo=docker)](https://docs.docker.com/compose/)
[![Kafka](https://img.shields.io/badge/Kafka-KRaft-231f20?logo=apachekafka)](https://kafka.apache.org)
[![Elasticsearch](https://img.shields.io/badge/Elasticsearch-8.x-005571?logo=elasticsearch)](https://elastic.co)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)

</div>

---

## 📌 简介

GoPan 是一套**生产级**视频点播平台，覆盖从「分片上传 → 异步转码 → HLS 流播 → 弹幕互动 → AI 摘要 → 语义搜索」的完整链路。

| 维度 | 选型 |
|------|------|
| 后端 | go-zero v1.10 + gRPC + Protobuf |
| 前端 | Vue 3 + Vite + Pinia + Vant 4，暗色赛博朋克主题 |
| 中间件 | MySQL 8 / Redis 7 / MinIO / Kafka (KRaft) / Elasticsearch 8 / etcd |
| AI | ChineseCLIP（语义向量化）+ Whisper + MiniMax/DeepSeek（视频摘要）|
| 可观测 | Prometheus + Grafana + Jaeger（OTLP） |
| 部署 | Docker Compose（18 容器）/ Nginx 反向代理 |

---

## 🧬 整体架构

```mermaid
flowchart TB
    User[浏览器 / App]
    User -->|HTTPS| Nginx
    Nginx -->|/api| Gateway[Gateway :8888<br/>JWT 鉴权 + 路由]
    Nginx -->|/| FE[Vue 3 SPA]

    Gateway -->|gRPC| US[user-svc :8081]
    Gateway -->|gRPC| VS[video-svc :8082]
    Gateway -->|gRPC| TS[transcode-svc :8083]
    Gateway -->|gRPC| SS[stream-svc :8084]
    Gateway -->|gRPC| IS[interact-svc :8085]
    Gateway -->|gRPC| SRC[search-svc :8086]
    Gateway -->|gRPC| AS[admin-svc]

    VS -->|Kafka transcode.tasks| TS
    VS -->|Kafka summary.tasks| SUMC[SummaryConsumer<br/>同进程 goroutine]
    SUMC -->|HTTP| SUM[summary-ai :9920<br/>Whisper + LLM]
    SRC -->|HTTP /embed/text| SEM[semantic-ai :9900<br/>ChineseCLIP]
    SRC <-->|KNN| ES[(Elasticsearch)]

    US <--> DB[(MySQL)]
    VS <--> DB
    IS <--> DB
    IS <--> RDS[(Redis Pub/Sub)]
    VS <--> MinIO[(MinIO)]
    TS <--> MinIO
    SS --> MinIO

    style SUM fill:#1a0a2e,stroke:#b347ea,color:#fff
    style SEM fill:#1a0a2e,stroke:#00f0ff,color:#fff
    style Gateway fill:#0a1a2e,stroke:#00f0ff,color:#fff
```

---

## ✨ 核心特性

### 视频链路
- **分片上传** 浏览器切片 → Redis 进度追踪 → 失败续传
- **异步合并** Kafka `merge.tasks` 解耦请求线程
- **HLS 转码** Kafka `transcode.tasks` 触发 FFmpeg 1080p / H.264 + AAC，10s 切片
- **防盗链播放** MD5 签名 URL，Nginx 二级代理 MinIO

### AI 双引擎（详见 [`docs/ai.md`](docs/ai.md)）
- **语义搜索** ChineseCLIP 512 维向量 + ES KNN，长句搜画面（如「一只小狗在沙滩冲浪」）
- **视频摘要** Whisper tiny 听译音轨 → LLM 提炼 150 字大纲，**异步 Kafka 触发**，前端轮询渲染
- **状态机** `ai_summary_status` 0未生成 / 1生成中 / 2完成 / 3失败，前端三分支 UI

### 互动
- **点赞 / 收藏 / 评论 / 弹幕** 四表 CRUD
- **实时弹幕推送** Redis Pub/Sub + WebSocket，gateway 多订阅，详见 [`docs/DANMAKU_WEBSOCKET.md`](docs/DANMAKU_WEBSOCKET.md)

### 可观测
- **Metrics** go-zero 内置 Prometheus exporter（gateway 9102 / 各 svc 91xx）
- **Tracing** OTLP → Jaeger
- **Dashboards** Grafana
- 详见 [`docs/OBSERVABILITY.md`](docs/OBSERVABILITY.md)

---

## 📁 项目结构

```
gopan/
├── api/gateway.api               # 网关 API 定义（goctl）
├── gateway/                      # HTTP API 网关 :8888
├── rpc/                          # 7 个 gRPC 微服务
│   ├── user/   video/  transcode/  stream/  interact/  search/  admin/
├── common/
│   ├── es/                       # Elasticsearch 封装 + 向量化客户端
│   ├── storage/                  # MinIO 封装
│   ├── kafka/                    # Topic 常量 + 任务消息体
│   └── response/                 # 统一响应
├── semantic-ai/                  # Python · ChineseCLIP 向量化服务 :9900
├── summary-ai/                   # Python · Whisper + LLM 摘要服务 :9920
├── frontend/                     # Vue 3 SPA
├── etc/init.sql                  # 数据库建表（含 ai_summary 字段）
├── nginx/                        # Nginx 反向代理 + MinIO 回源
├── docker-compose.yml            # 全栈编排（含 AI 服务）
├── Makefile                      # build/run/stop/logs/proto
├── docs/                         # 各类文档（架构 / AI / Kafka / 弹幕 / 部署）
└── imgs/                         # logo + 文档配图
```

---

## 🚀 快速开始

### 方式一：一键 Docker 全栈（推荐）

```bash
# 1. 启动全部 18 个容器（首次 build ~10 分钟，含 AI 镜像）
docker compose up -d --build

# 2. MinIO 桶设公开读（summary-ai 直拉视频必须）
docker run --rm --network gopan_gopan-net \
  -e MC_HOST_local=http://minioadmin:minioadmin@minio:9000 \
  docker.m.daocloud.io/minio/mc anonymous set download local/gopan-videos

# 3. 检查状态
docker compose ps
curl http://localhost:9900/health   # semantic-ai
curl http://localhost:9920/health   # summary-ai
```

可选：配置 LLM Key 让摘要走真实大模型（否则用本地兜底文案）：

```bash
echo 'DEEPSEEK_API_KEY=sk-xxx' >> .env
# 或 MINIMAX_API_KEY=xxx
docker compose up -d summary-ai
```

### 方式二：中间件 docker + Go 本地（开发调试）

```bash
# 1. 仅起中间件
docker compose up -d etcd mysql redis minio elasticsearch kafka jaeger prometheus grafana
docker compose up -d semantic-ai summary-ai

# 2. Go 服务本地编译运行
make run         # 走 *.local.yaml，连 127.0.0.1
make status
make logs-video  # 查 video-svc 日志（含 SummaryConsumer 输出）
make stop
```

### 前端

```bash
cd frontend && npm install && npm run dev
# → http://localhost:5173
```

---

## 🔌 端口速查

| 类别 | 服务 | 端口 |
|------|------|------|
| **入口** | nginx | 80 |
| | gateway | 8888 |
| **业务 gRPC** | user-svc | 8081 |
| | video-svc | 8082 |
| | transcode-svc | 8083 |
| | stream-svc | 8084 |
| | interact-svc | 8085 |
| | search-svc | 8086 |
| **AI** | semantic-ai | 9900 |
| | summary-ai | 9920 |
| **中间件** | mysql / redis / etcd | 3306 / 6379 / 2379 |
| | minio API / Console | 9000 / 9001 |
| | kafka | 9092 |
| | elasticsearch | 9200 |
| **观测** | prometheus / grafana / jaeger | 9090 / 3001 / 16686 |

---

## 📡 API 速览

完整定义见 [`api/gateway.api`](api/gateway.api)。

### 用户 `/api/user`
`POST /register`、`POST /login`、`GET /profile`、`PUT /profile`

### 视频 `/api/video`（JWT）
`GET /list`、`GET /detail`、`PUT /update`、`DELETE /delete`<br/>
`POST /init-upload`、`POST /upload-chunk`、`POST /merge-chunks`、`GET /upload-status`<br/>
`GET /play-url`、`POST /play-progress`<br/>
`POST /like` / `DELETE /like`、`POST /favorite` / `DELETE /favorite`<br/>
`POST /comment`、`GET /comments`、`DELETE /comment`<br/>
`POST /danmaku`、`GET /danmakus`<br/>
**`POST /ai-analyze`** ← 查询/轮询 AI 摘要状态

### 搜索 `/api/search`
`GET /videos` ← ChineseCLIP 语义向量 + ES KNN

### 弹幕 WebSocket
`ws://host/ws/danmaku?video_id=xxx&token=xxx`

### 管理员 `/api/admin`
`POST /login`、`GET /videos`、`POST /videos/:id/approve|reject|delete`

---

## 📚 文档

| 文档 | 内容 |
|------|------|
| [`docs/ai.md`](docs/ai.md) | **AI 双链路完整文档（含 mermaid 图、状态机、排查指南）** |
| [`docs/CODE_READING_GUIDE.md`](docs/CODE_READING_GUIDE.md) | 源码阅读指南 |
| [`docs/kafka.md`](docs/kafka.md) | Kafka 集成 & topic 设计 |
| [`docs/DANMAKU_WEBSOCKET.md`](docs/DANMAKU_WEBSOCKET.md) | 弹幕实时推送方案 |
| [`docs/OBSERVABILITY.md`](docs/OBSERVABILITY.md) | Prometheus + Jaeger 可观测 |
| [`docs/deploy_multi_machine.md`](docs/deploy_multi_machine.md) | 多机部署架构 |
| [`docs/nginx.md`](docs/nginx.md) | Nginx 反代 + MinIO 回源 |
| [`docs/qa.md`](docs/qa.md) | 面试 / 设计问答 |

---

## 🧪 测试与基准

```bash
# go vet & 单元测试
make lint
make test

# HTTP 压测（wrk + curls）
bash test/curls.sh
bash test/run_test.sh
cat test/results/http_bench.txt
```

---

## 🛠️ 开发调试

```bash
# VSCode：F5 选服务即可断点（.vscode/launch.json 已配 7 服务）
# 重新生成 proto / api 桩代码
make proto       # 所有 rpc/*.proto → *.pb.go
make api         # gateway.api → gateway types/handler
make gen         # 二合一
```

gRPC 反射调试（DevMode）：

```bash
grpcurl -plaintext localhost:8081 list
grpcurl -plaintext -d '{"username":"test","password":"123"}' localhost:8081 user.User/Login
```

---

## 🗺️ Roadmap

| 优先级 | 任务 | 状态 |
|--------|------|------|
| ✅ | 分片上传 + 异步合并 + Kafka 转码 | Done |
| ✅ | HLS 流播 + 防盗链 | Done |
| ✅ | Elasticsearch 全文 + 语义 KNN | Done |
| ✅ | 弹幕 WebSocket | Done |
| ✅ | AI 摘要异步 Kafka 链路 | Done |
| ✅ | Prometheus + Jaeger 可观测 | Done |
| 🔧 | 直播推流（RTMP/SRS） | Planned |
| 🔧 | CDN 接入（阿里云 / Cloudflare） | Planned |
| 🔧 | 视频帧级 AI 检索（image embedding） | Planned |

---

## 📜 License

MIT © 2026 GoPan

<div align="center">
<sub>Built with ☕ + Mermaid + AI 双引擎</sub>
</div>
