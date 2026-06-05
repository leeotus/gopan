# GoPan AI 集成方案

GoPan 平台集成了两个 Python AI 微服务，分别承担「语义搜索向量化」和「视频内容听译摘要」两个独立职责。本文档梳理两条链路的架构、数据流、关键代码与部署运维。

---

## 1. 服务定位

| 服务 | 端口 | 模型 | 触发点 | 用途 |
|------|------|------|--------|------|
| `semantic-ai` | 9900 | OFA-Sys/chinese-clip-vit-base-patch16 | 视频索引 ES 时 + 用户搜索时 | 把中文文本/图片转成 512 维向量，供 ES KNN 检索 |
| `summary-ai` | 9920 | OpenAI Whisper tiny + MiniMax/DeepSeek LLM | 视频转码完成后 | 抽音轨 → 听译 → 大模型生成 150 字摘要 |

两个服务完全独立，**互不依赖**。`semantic-ai` 走同步 HTTP 调用（请求耗时 ms 级），`summary-ai` 走 **Kafka 异步任务队列**（请求耗时分钟级，不能阻塞主链路）。

---

## 2. semantic-ai：语义搜索链路

### 2.1 整体数据流

```mermaid
flowchart LR
    subgraph 索引阶段
      A1[用户上传视频] --> A2[转码完成 TranscodeCallback]
      A2 -->|"description (用户原文)"| A3[search-svc.IndexVideo]
      A3 --> A4[es.IndexVideo]
      A4 -->|HTTP POST /embed/text| AI1[semantic-ai:9900]
      AI1 -->|512维向量| A4
      A4 --> A5[(Elasticsearch<br/>video_vector)]
    end

    subgraph 搜索阶段
      B1[用户输入搜索词] --> B2[gateway /search]
      B2 --> B3[search-svc.SearchVideos]
      B3 --> B4[es.SearchVideos]
      B4 -->|HTTP POST /embed/text| AI1
      AI1 -->|512维向量| B4
      B4 -->|KNN 余弦相似度| A5
      A5 -->|Top N 命中| B4
      B4 --> B5[返回视频列表]
    end
```

### 2.2 关键代码

**触发点：`common/es/es.go`**
- `IndexVideo()`：写入索引时调用 `getEmbeddingVector(title + description)` 生成向量并存入 `video_vector` 字段
- `SearchVideos()`：先把 keyword 向量化，再用 ES `knn` 查询

```go
// getEmbeddingVector 发送 HTTP 转换向量
func getEmbeddingVector(ctx context.Context, keyword string) ([]float32, error) {
    aiURL := os.Getenv("AI_SERVICE_URL")
    if aiURL == "" {
        aiURL = "http://127.0.0.1:9900"
    }
    // POST {aiURL}/embed/text → {"vector": [...]}
}
```

**Python 端：`semantic-ai/main.py`**
- 启动时一次性 `from_pretrained` 装载 ChineseCLIP 模型到 GPU/CPU
- `POST /embed/text` → 文本塔前向传播 → L2 归一化 → 返回 512 维向量
- `POST /embed/image` → 视觉塔前向传播 → 同上（当前主链路未用到）

### 2.3 配置

| 位置 | 配置项 | 默认值 |
|------|--------|--------|
| `docker-compose.yml` (search-svc) | `AI_SERVICE_URL` | `http://semantic-ai:9900` |
| `semantic-ai/Dockerfile` | 端口 | 9900 |

**重要**：`description` 在用户上传时填入数据库，**永不被 AI 改写**。`semantic-ai` 只读不写，搜索用的是用户原始简介的语义。

---

## 3. summary-ai：视频内容摘要链路

### 3.1 异步整体数据流

```mermaid
flowchart TB
    A[用户上传 + 转码完成] --> B[video-svc.TranscodeCallback]
    B -->|status=2| B1[更新 ai_summary_status=1]
    B1 --> C[投递 Kafka SummaryTask]
    C --> K[(Kafka<br/>gopan.video.summary.tasks)]
    K --> D[video-svc SummaryConsumer]
    D -->|HTTP POST /analyze<br/>video_url| AI[summary-ai:9920]
    AI -->|1.下载视频| MinIO[(MinIO)]
    AI -->|2.FFmpeg 抽音| AI
    AI -->|3.Whisper 听译| AI
    AI -->|4.LLM 摘要| LLM[MiniMax/DeepSeek]
    LLM --> AI
    AI -->|返回 summary| D
    D -->|UpdateAiSummary| DB[(MySQL videos.ai_summary<br/>ai_summary_status=2)]

    F[前端 VideoDetail 轮询] -->|POST /api/video/ai-analyze| G[gateway]
    G --> H[video-svc.GetVideo]
    H --> DB
    DB -->|status+summary| F
```

### 3.2 状态机

```mermaid
stateDiagram-v2
    [*] --> 未生成 : 视频刚上传
    未生成 --> 生成中 : TranscodeCallback 投递 Kafka
    生成中 --> 已完成 : summary-ai 返回 + UpdateAiSummary
    生成中 --> 失败 : 网络/LLM/FFmpeg 错误
    已完成 --> [*]
    失败 --> [*]

    note right of 未生成 : ai_summary_status=0
    note right of 生成中 : ai_summary_status=1
    note right of 已完成 : ai_summary_status=2
    note right of 失败 : ai_summary_status=3
```

### 3.3 关键代码

**生产者：`rpc/video/internal/logic/transcodecallbacklogic.go`**

```go
// 转码 status=2 时
if vidStatus == 2 {
    v, _ := l.svcCtx.VideoStore.FindById(...)
    if v != nil {
        go l.indexToES(v)          // 异步入 ES
        l.dispatchSummaryTask(v.Id) // 投递 Kafka SummaryTask
    }
}

func (l *TranscodeCallbackLogic) dispatchSummaryTask(videoId int64) {
    // 1. 置 ai_summary_status = 1（生成中）
    l.svcCtx.VideoStore.UpdateAiSummaryStatus(l.ctx, videoId, 1)
    // 2. 投递 Kafka
    body, _ := json.Marshal(commonkafka.SummaryTask{
        VideoId:  videoId,
        VideoUrl: fmt.Sprintf("%s/videos/%d/source.mp4", prefix, videoId),
    })
    writer.WriteMessages(l.ctx, kafkago.Message{Key: ..., Value: body})
}
```

**消费者：`rpc/video/internal/consume/summaryconsumer.go`**

```go
func StartSummaryConsumer(ctx, svcCtx) {
    reader := kafka.NewConsumer(brokers, topic)
    for {
        msg, _ := reader.FetchMessage(ctx)
        var task kafka.SummaryTask
        json.Unmarshal(msg.Value, &task)

        if err := process(ctx, svcCtx, &task); err != nil {
            // 失败：标 status=3
            svcCtx.VideoStore.UpdateAiSummaryStatus(ctx, task.VideoId, 3)
        }
        reader.CommitMessages(ctx, msg) // 成功失败都 commit，避免毒消息死循环
    }
}

func process(ctx, svcCtx, task) error {
    // 1. 标 status=1
    svcCtx.VideoStore.UpdateAiSummaryStatus(ctx, task.VideoId, 1)
    // 2. POST {url}/analyze 携带 video_url，超时 300s
    resp := http.Post(url+"/analyze", {"video_url": task.VideoUrl})
    // 3. UpdateAiSummary(写文本 + 自动置 status=2)
    return svcCtx.VideoStore.UpdateAiSummary(ctx, task.VideoId, aiResp.Summary)
}
```

**Python 端：`summary-ai/main.py`**

```mermaid
flowchart LR
    A[POST /analyze] --> B[urllib 下载视频]
    B --> C[ffmpeg 抽 16k mono wav]
    C --> D[whisper.transcribe]
    D --> E{LLM Key?}
    E -->|MINIMAX_API_KEY| F[MiniMax abab6.5g]
    E -->|DEEPSEEK_API_KEY| G[DeepSeek chat]
    E -->|都没有| H[本地兜底文案]
    F --> I[返回 summary+vtt+srt]
    G --> I
    H --> I
```

### 3.4 网关查询接口

`POST /api/video/ai-analyze` 由原来的"同步触发 Whisper"改为**纯查询**：

```go
// gateway/internal/logic/video/aianalyzelogic.go
func (l *AIAnalyzeLogic) AIAnalyze(req *types.AIAnalyzeReq) (*types.AIAnalyzeResp, error) {
    r, _ := l.svcCtx.VideoClient.GetVideo(...)
    resp := &types.AIAnalyzeResp{Status: int(r.Video.AiSummaryStatus)}
    if r.Video.AiSummaryStatus == 2 {
        resp.Summary = r.Video.AiSummary
    }
    return resp, nil
}
```

返回字段：

| 字段 | 含义 |
|------|------|
| `status` | 0 未生成 / 1 生成中 / 2 已完成 / 3 失败 |
| `summary` | status=2 时才有值 |

---

## 4. 数据模型

`videos` 表新增两列：

```sql
ALTER TABLE videos
  ADD COLUMN ai_summary TEXT NOT NULL AFTER status,
  ADD COLUMN ai_summary_status TINYINT NOT NULL DEFAULT 0
    COMMENT '0:未生成 1:生成中 2:已完成 3:失败' AFTER ai_summary;
```

`description`（用户原文，搜索语义源）和 `ai_summary`（AI 听译结果，展示用）**彻底分离**，互不影响。

---

## 5. Kafka Topic

| Topic | 生产者 | 消费者 | 消息体 |
|-------|--------|--------|--------|
| `gopan.transcode.tasks` | video-svc | transcode-svc | `TranscodeTask{VideoId, ObjectKey}` |
| `gopan.video.merge.tasks` | video-svc | transcode-svc | `MergeTask{VideoId, ChunkKeys, ...}` |
| `gopan.video.summary.tasks` | **video-svc (新)** | **video-svc (新)** | `SummaryTask{VideoId, VideoUrl}` |

注意 Summary topic 生产者和消费者都在 video-svc 内（解耦同进程：转码回调入口 → Kafka → 消费 goroutine）。

```mermaid
flowchart LR
    TC[TranscodeCallback Logic] -->|投递| K[(Kafka summary.tasks)]
    K -->|消费| SC[SummaryConsumer<br/>goroutine in video-svc]
    SC --> AI[summary-ai HTTP]
    SC --> DB[(MySQL)]
```

为什么生产消费都在 video-svc？
- 解耦同步链路（转码回调 100ms 返回，AI 计算几分钟）
- 消费失败可重试、可监控、不影响转码回调
- 同进程内复用 VideoStore，免去额外 RPC

---

## 6. 前端集成

### 6.1 视频详情页 UI 状态分支

```mermaid
flowchart TB
    A[onMounted: fetchDetail] --> B{ai_summary_status}
    B -->|2| C[打字机渲染 ai_summary]
    B -->|0 或 1| D[骨架屏 + 启动 5s 轮询]
    B -->|3| E[失败提示]

    D --> F[POST /api/video/ai-analyze]
    F --> G{响应 status}
    G -->|2| H[停轮询 + 打字机]
    G -->|3| I[停轮询 + 失败提示]
    G -->|其它| D
```

### 6.2 轮询参数

| 参数 | 值 | 说明 |
|------|----|------|
| 间隔 | 5000ms | 平衡实时性与服务压力 |
| 最大次数 | 120 | 总时长 10 分钟兜底，超时停轮询 |
| 清理 | `onUnmounted` 必清 timer | 防内存泄漏 |

---

## 7. 部署

### 7.1 docker-compose 服务条目

```yaml
semantic-ai:
  build: ./semantic-ai
  container_name: gopan-semantic-ai
  ports:
    - "9900:9900"
  volumes:
    - hf-cache:/root/.cache/huggingface   # 持久化 ChineseCLIP 模型
  networks:
    - gopan-net

summary-ai:
  build: ./summary-ai
  container_name: gopan-summary-ai
  ports:
    - "9920:9920"
  environment:
    - MINIMAX_API_KEY=${MINIMAX_API_KEY:-}
    - DEEPSEEK_API_KEY=${DEEPSEEK_API_KEY:-}
  volumes:
    - whisper-cache:/root/.cache          # 持久化 Whisper 模型
  networks:
    - gopan-net
```

### 7.2 配置入口

| 服务 | 入口 | 配置内容 |
|------|------|---------|
| search-svc | `AI_SERVICE_URL` env | `http://semantic-ai:9900` |
| video-svc | `SummaryAI.URL` yaml | `http://summary-ai:9920` |
| video-svc | `SummaryAI.MinIO` yaml | `http://minio:9000/gopan-videos` |
| video-svc | `Kafka.SummaryTopic` yaml | `gopan.video.summary.tasks` |

### 7.3 MinIO 桶必须公开读

`summary-ai` 用 `urllib.request.urlretrieve` 直接 GET 视频 URL，无鉴权。需要把桶设为匿名 download：

```bash
docker run --rm --network gopan_gopan-net \
  -e MC_HOST_local=http://minioadmin:minioadmin@minio:9000 \
  docker.m.daocloud.io/minio/mc anonymous set download local/gopan-videos
```

### 7.4 启动顺序

```bash
# 1. 全栈
docker compose up -d --build

# 2. MinIO 桶公开
docker run --rm --network gopan_gopan-net ... mc anonymous set download local/gopan-videos

# 3. 验证
curl http://localhost:9900/health    # semantic-ai
curl http://localhost:9920/health    # summary-ai
docker logs gopan-video-svc | grep SummaryConsumer
```

---

## 8. 排查指南

### 8.1 端到端日志关键字

| 阶段 | 容器 | 关键日志 |
|------|------|---------|
| 投递 | video-svc | `[Summary] summary task sent: video_id=` |
| 消费 | video-svc | `[SummaryConsumer] received task` |
| 调 AI | summary-ai | `[Step 1] Loading video file from MinIO` |
| 听译 | summary-ai | `[Step 3] Transcription finished in X seconds` |
| 摘要 | summary-ai | `[LLM Summary] Calling MiniMax/DeepSeek API` |
| 落库 | video-svc | `[SummaryConsumer] done: video_id=` |

### 8.2 故障速查表

| 现象 | 原因 | 排查命令 |
|------|------|---------|
| `ai_summary_status` 一直为 0 | 转码未完成 / Kafka 投递失败 | `docker logs gopan-video-svc \| grep -E "Summary\|Transcode"` |
| 一直为 1 | summary-ai 不可达 / 超时 | `curl http://localhost:9920/health` |
| 一直为 3 | summary-ai 内部错误 | `docker logs gopan-summary-ai \| tail -50` |
| summary-ai 拉视频 403 | MinIO 桶没设 anonymous | 重跑 `mc anonymous set download` |
| summary-ai 摘要为 mock 文案 | 未配 LLM Key | `docker exec gopan-summary-ai env \| grep API_KEY` |
| 搜索结果不准 | semantic-ai 不可达 / 向量未入库 | `curl http://localhost:9900/health`，检查 ES `video_vector` 字段 |

### 8.3 DB 直查

```bash
docker exec gopan-mysql mysql -uroot -pgopan123 gopan -e \
  "SELECT id, title, ai_summary_status, LEFT(ai_summary, 80) AS summary
   FROM videos ORDER BY id DESC LIMIT 10;"
```

### 8.4 Kafka 直查

```bash
# 查看 topic 是否存在
docker exec gopan-kafka /opt/kafka/bin/kafka-topics.sh \
  --bootstrap-server localhost:9092 --list | grep summary

# 查看消息堆积
docker exec gopan-kafka /opt/kafka/bin/kafka-consumer-groups.sh \
  --bootstrap-server localhost:9092 --describe --all-groups | grep summary
```

### 8.5 ES 向量检查

```bash
curl -s 'http://localhost:9200/gopan_videos/_search?pretty' \
  -H 'Content-Type: application/json' -d'{
    "_source": ["video_id","title","video_vector"],
    "size": 3
  }'
# 期望每条都有 video_vector 数组（512 维 float）
```

---

## 9. 设计权衡 FAQ

### Q1: 为什么 summary 不直接覆盖 description？

`description` 是用户上传时的原创简介，是搜索语义的「事实源」。AI 摘要是"对内容的二次解读"，两者应**并行展示**：用户可以看到「作者怎么写」和「AI 怎么总结」对比，也避免 AI 幻觉污染搜索质量。

### Q2: 为什么用 Kafka 而不是 video-svc 直接同步调 summary-ai？

- Whisper 听译 + LLM 摘要总耗时几十秒到几分钟，**不能阻塞**转码回调（gRPC 默认 60s 超时）
- 失败需要重试、需要监控、需要可观测，Kafka 天生提供
- 后续可水平扩展消费者，吞吐线性提升

### Q3: 为什么消费者不是独立服务？

- 当前只有一个消费者职责，独立部署运维成本不划算
- video-svc 内嵌 goroutine 复用了 `VideoStore`，免一次 RPC
- 后续业务变复杂可平移到独立 `summary-worker` 服务

### Q4: 前端为什么不用 WebSocket 推送？

- AI 摘要一个视频只触发一次，**不是高频事件**
- 5s 轮询的 HTTP 开销小于维护 WS 连接
- 现有 WS 已用于弹幕，避免 channel 复用复杂度

### Q5: 模型缓存挂卷的必要性？

| 模型 | 大小 | 不挂卷 | 挂卷后 |
|------|------|--------|--------|
| ChineseCLIP | ~600MB | 每次容器重建都重下 | 首次后 0 重下 |
| Whisper tiny | ~70MB | 同上 | 同上 |

`hf-cache` 和 `whisper-cache` 两个 named volume 必须保留。

---

## 10. 改动文件清单

| 类型 | 文件 |
|------|------|
| DB | `etc/init.sql`（加 `ai_summary` + `ai_summary_status`） |
| Model | `rpc/video/model/video.go` |
| Store | `rpc/video/store/video.go`（`UpdateAiSummary` / `UpdateAiSummaryStatus`） |
| Proto | `rpc/video/video.proto` → 重新生成 pb.go |
| Logic | `rpc/video/internal/logic/{transcodecallbacklogic,getvideologic}.go` |
| 新建消费者 | `rpc/video/internal/consume/summaryconsumer.go` |
| Main | `rpc/video/video.go`（启 SummaryConsumer goroutine） |
| Config | `rpc/video/internal/config/config.go`、`etc/video*.yaml` |
| Common | `common/kafka/kafka.go`（`SummaryTask` + `TopicSummaryTasks`） |
| Gateway | `gateway/internal/logic/video/aianalyzelogic.go`（改查询模式） |
| Gateway types | `gateway/internal/types/types.go`、`api/gateway.api` |
| 前端 | `frontend/src/pages/VideoDetail.vue`（状态分支 + 轮询） |
| Docker | `semantic-ai/Dockerfile`、`summary-ai/Dockerfile`、`docker-compose.yml` |

---

至此 GoPan AI 双链路落成，前端、后端、AI 服务、消息队列、数据库均完成对接，可水平扩展、可观测、可降级。
