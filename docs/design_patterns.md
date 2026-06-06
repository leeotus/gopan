# GoPan 项目设计模式分析

> 本文档详细分析 GoPan 视频平台项目中所使用的各类设计模式，按后端 (Go) 和前端 (Vue 3) 分类说明。

---

## 目录

- [GoPan 项目设计模式分析](#gopan-项目设计模式分析)
  - [目录](#目录)
  - [一、后端设计模式 (Go)](#一后端设计模式-go)
    - [1. 中间件模式 / 责任链模式](#1-中间件模式--责任链模式)
    - [2. 策略模式](#2-策略模式)
    - [3. 熔断/降级模式](#3-熔断降级模式)
    - [4. 工厂方法模式](#4-工厂方法模式)
    - [5. 适配器模式](#5-适配器模式)
    - [6. 外观模式](#6-外观模式)
    - [7. 模板方法模式](#7-模板方法模式)
    - [8. 依赖注入 / IoC](#8-依赖注入--ioc)
    - [9. 代理模式](#9-代理模式)
    - [10. 发布-订阅模式](#10-发布-订阅模式)
    - [11. DTO 模式](#11-dto-模式)
    - [12. 仓储模式](#12-仓储模式)
    - [13. 令牌桶模式](#13-令牌桶模式)
    - [14. 前端控制器模式](#14-前端控制器模式)
  - [二、前端设计模式 (Vue 3)](#二前端设计模式-vue-3)
    - [15. 状态管理模式](#15-状态管理模式)
    - [16. 拦截器模式](#16-拦截器模式)
    - [17. 路由守卫模式](#17-路由守卫模式)
  - [三、架构级设计模式](#三架构级设计模式)
    - [微服务架构 (Microservices)](#微服务架构-microservices)
    - [事件驱动架构 (Event-Driven)](#事件驱动架构-event-driven)
    - [CQRS 读写分离](#cqrs-读写分离)
  - [总结表格](#总结表格)

---

## 一、后端设计模式 (Go)

### 1. 中间件模式 / 责任链模式

**文件:** `gateway/gateway.go`、`gateway/internal/middleware/authmiddleware.go`、`gateway/internal/middleware/ratelimitmiddleware.go`

**说明:** 请求经过 CORS → RateLimiter（全局限流）→ Auth（JWT 鉴权）→ Handler 层层传递，每一层都可以中断请求或将其传递给下一层。

```go
// gateway/gateway.go — CORS 中间件（最先注册）
server.Use(func(next http.HandlerFunc) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Access-Control-Allow-Origin", "*")
        w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
        w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
        if r.Method == "OPTIONS" {
            w.WriteHeader(http.StatusNoContent)
            return
        }
        next(w, r) // 传递给下一个中间件
    }
})

// 全局限流中间件（第二个注册）
server.Use(ctx.RateLimiter)
```

```go
// gateway/internal/middleware/authmiddleware.go — JWT 鉴权中间件
func (m *AuthMiddleware) Handle(next http.HandlerFunc) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        // 1. 提取 Authorization header
        authHeader := r.Header.Get("Authorization")
        if authHeader == "" {
            errorResp(w, 1002, "未登录或登录已过期")
            return // 校验失败，中断责任链
        }
        // 2. 校验 JWT 签名
        token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
            return m.secret, nil
        })
        if err != nil || !token.Valid {
            errorResp(w, 1002, "未登录或登录已过期")
            return
        }
        // 3. 注入 user_id/username 到 context
        ctx = context.WithValue(ctx, CtxKeyUserId, int64(uidFloat))
        next(w, r.WithContext(ctx)) // 校验通过，传递给下一个处理者
    }
}
```

**请求链路:**
```
CORS → RateLimiter(全局限流) → [AuthMiddleware(JWT鉴权)] → Handler → Logic → gRPC
```

**本质:** 每个中间件遵循 `func(next http.HandlerFunc) http.HandlerFunc` 接口，形成一条责任链。任一环节可中断请求，否则调用 `next(w, r)` 将请求传递下去。

---

### 2. 策略模式

**文件:** `common/es/es.go`

**说明:** `SearchVideos()` 方法根据 AI 向量服务是否可用，在运行时动态选择"KNN 语义搜索"或"BM25 词频搜索"两种可互换的搜索算法。

```go
// common/es/es.go — SearchVideos 动态选择搜索策略
func (c *Client) SearchVideos(ctx context.Context, keyword string, category string, page, size int) (*SearchResult, error) {
    // 尝试调用 AI 向量服务获取语义向量
    vec, aiErr := getEmbeddingVector(ctx, keyword)
    if aiErr == nil && len(vec) == 512 {
        fmt.Printf("[AI Search OK] Performing Vector k-NN Search for: '%s'\n", keyword)
        // ✅ 策略 A: 向量 K-NN 语义搜索
        return c.searchVideosByKNN(ctx, vec, category, page, size)
    }
    // ❌ 策略 B: 降级到传统 BM25 词频检索
    fmt.Printf("[AI Search Fallback] Performing classical BM25 Search for: '%s'\n", keyword)
    return c.searchVideosByLexical(ctx, keyword, category, page, size)
}
```

**两个具体策略:**

- `searchVideosByKNN()` — ES 8.x 原生 k-NN 语义向量近邻检索
- `searchVideosByLexical()` — 传统 Lucene BM25 倒排词频检索

---

### 3. 熔断/降级模式

**文件:** `common/es/es.go`

**说明:** 当 AI 嵌入服务不可用时，系统自动进行多层降级，保障核心搜索功能绝不中断。

```go
// common/es/es.go — getEmbeddingVector 实现双层降级
func getEmbeddingVector(ctx context.Context, keyword string) ([]float32, error) {
    aiURL := os.Getenv("AI_SERVICE_URL")
    if aiURL == "" {
        aiURL = "http://127.0.0.1:9900" // 默认 GPU 节点
    }

    req, _ := http.NewRequestWithContext(ctx, "POST", aiURL+"/embed/text", ...)
    resp, err := client.Do(req)
    
    // 🔄 第一层降级: GPU:9900 不可用 → 自动切换 CPU:9901
    if err != nil && aiURL == "http://127.0.0.1:9900" {
        req2, _ := http.NewRequestWithContext(ctx, "POST", "http://127.0.0.1:9901/embed/text", ...)
        resp, err = client.Do(req2) // 重试 CPU 节点
    }
    
    // 如果仍失败 → 返回 error → 触发 SearchVideos 的第二层降级
    if err != nil {
        return nil, fmt.Errorf("ai-service offline: %w", err)
    }
    // ...
}
```

**三层降级链路:**
```
GPU:9900 (语义向量) → CPU:9901 (语义向量) → BM25 词频检索 (纯文本全文搜索)
```

---

### 4. 工厂方法模式

**文件:** 遍布全项目，几乎所有结构体通过 `NewXxx()` 构造函数创建

**说明:** 将对象创建逻辑封装在工厂函数中，隔离复杂的初始化流程（如连接检查、Bucket 自动创建等）。

```go
// common/storage/minio.go — MinIO 客户端工厂
func NewMinioClient(endpoint, accessKey, secretKey, bucket string, useSSL bool) (*MinioClient, error) {
    client, err := minio.New(endpoint, &minio.Options{
        Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
        Secure: useSSL,
    })
    if err != nil {
        return nil, fmt.Errorf("minio connect failed: %w", err)
    }
    // 自动检查并创建 Bucket
    exists, err := client.BucketExists(context.Background(), bucket)
    if !exists {
        client.MakeBucket(context.Background(), bucket, minio.MakeBucketOptions{})
    }
    return &MinioClient{client: client, bucket: bucket}, nil
}

// common/kafka/kafka.go — Kafka Producer 工厂
func NewProducer(brokers []string, topic string) *kafkago.Writer {
    return &kafkago.Writer{
        Addr:     kafkago.TCP(brokers...),
        Topic:    topic,
        Balancer: &kafkago.LeastBytes{},
    }
}

// common/storage/upload_progress.go — 上传进度跟踪器工厂
func NewUploadProgress(rdb *redis.Client) *UploadProgress {
    return &UploadProgress{rdb: rdb}
}
```

---

### 5. 适配器模式

**文件:** `common/storage/minio.go`、`common/storage/upload_progress.go`、`common/es/es.go`、`common/kafka/kafka.go`

**说明:** 将第三方 SDK 的复杂接口适配/封装为项目内部统一的简洁接口，降低业务代码与外部库的耦合。

```go
// common/storage/minio.go — 将 minio-go SDK 适配为项目统一接口
type MinioClient struct {
    client *minio.Client  // 被适配的第三方 SDK 对象
    bucket string
}

// 对外暴露简洁的上传接口，隐藏 MinIO SDK 复杂性
func (s *MinioClient) PutObject(ctx context.Context, key string, reader io.Reader, size int64, contentType string) error {
    _, err := s.client.PutObject(ctx, s.bucket, key, reader, size, minio.PutObjectOptions{
        ContentType: contentType,
    })
    return err
}

// 服务端合并接口（用于分片上传合并）
func (s *MinioClient) ComposeObject(ctx context.Context, destKey string, sourceKeys []string) error {
    sources := make([]minio.CopySrcOptions, len(sourceKeys))
    for i, key := range sourceKeys {
        sources[i] = minio.CopySrcOptions{Bucket: s.bucket, Object: key}
    }
    _, err := s.client.ComposeObject(ctx, minio.CopyDestOptions{
        Bucket: s.bucket, Object: destKey,
    }, sources...)
    return err
}
```

```go
// common/storage/upload_progress.go — 将 Redis 适配为上传进度追踪接口
type UploadProgress struct {
    rdb *redis.Client
}
func (p *UploadProgress) MarkReceived(ctx context.Context, uploadId string, index int32) error { ... }
func (p *UploadProgress) GetReceived(ctx context.Context, uploadId string) ([]int32, error) { ... }
func (p *UploadProgress) Clear(ctx context.Context, uploadId string) error { ... }
```

---

### 6. 外观模式

**文件:** `gateway/` 整个层

**说明:** Gateway 作为系统的统一对外入口（Facade），将 7 个后端 gRPC 微服务封装在一致的 REST API 之后，前端只需与 Gateway 交互。

```
前端 (Vue 3)
    │
    ▼
┌──────────────────────────────────────────────┐
│            Gateway (REST /api/*)             │  ← 外观层
│  ┌──────────────────────────────────────┐    │
│  │        ServiceContext (IoC)          │    │
│  ├──────────┬──────────┬────────────────┤    │
│  │ UserClient│VideoClient│SearchClient  │    │  ← gRPC 代理
│  ├──────────┼──────────┼────────────────┤    │
│  │ Interact │ Stream   │ Admin/Transcode│    │
│  └──────────┴──────────┴────────────────┘    │
└──────────────────────────────────────────────┘
    │        │        │        │       │
    ▼        ▼        ▼        ▼       ▼
user-svc  video-svc search-svc ...  (gRPC 微服务)
```

**核心代码:**
```go
// gateway/internal/svc/servicecontext.go — 聚合所有 RPC 客户端
type ServiceContext struct {
    Config          config.Config
    Auth            rest.Middleware
    RateLimiter     rest.Middleware
    UserClient      userclient.User       // user-svc 代理
    VideoClient     videoclient.Video     // video-svc 代理
    TranscodeClient transcodeclient.Transcode
    StreamClient    streamclient.Stream
    InteractClient  interactclient.Interact
    SearchClient    searchclient.Search   // search-svc 代理
    AdminClient     adminclient.Admin
}
```

---

### 7. 模板方法模式

**文件:** `gateway/internal/handler/` + `gateway/internal/logic/` (go-zero 框架生成)

**说明:** go-zero 框架将每个 API 的处理流程固化为统一模板：Handler 接收 HTTP 请求 → 创建 Logic 实例 → 调用 Logic 方法 → 返回响应。所有 30+ 个 API 端点都遵循此模板。

```go
// gateway/internal/handler/video/uploadchunkhandler.go (go-zero 生成)
func UploadChunkHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        l := video.NewUploadChunkLogic(r.Context(), svcCtx) // ① 创建 Logic
        l.SetRequest(r)                                      // ② 注入请求
        resp, err := l.UploadChunk()                         // ③ 执行业务逻辑
        if err != nil {
            httpx.ErrorCtx(r.Context(), w, err)
        } else {
            httpx.OkJsonCtx(r.Context(), w, resp)            // ④ 返回响应
        }
    }
}
```

```go
// gateway/internal/logic/video/uploadchunklogic.go — 所有 Logic 统一的模板结构
type UploadChunkLogic struct {
    logx.Logger                     // ① 继承日志
    ctx    context.Context           // ② 请求上下文
    svcCtx *svc.ServiceContext      // ③ 依赖注入容器
    r      *http.Request            // ④ 原始 HTTP 请求（可选）
}

func (l *UploadChunkLogic) UploadChunk() (resp *types.BaseResp, err error) {
    // ① 解析请求参数
    // ② 调用 gRPC 下游服务
    // ③ 返回统一响应
}
```

---

### 8. 依赖注入 / IoC

**文件:** `gateway/internal/svc/servicecontext.go`、`rpc/video/internal/svc/servicecontext.go`

**说明:** `ServiceContext` 作为 IoC 容器，集中管理所有依赖（Config、RPC 客户端、Redis、中间件等），通过构造函数注入到 Logic 中，避免全局变量和硬编码依赖。

```go
// gateway/internal/svc/servicecontext.go — 集中管理所有依赖
type ServiceContext struct {
    Config      config.Config
    Auth        rest.Middleware
    RateLimiter rest.Middleware
    // RPC 客户端 — 类似 Spring 的 @Autowired
    UserClient      userclient.User
    VideoClient     videoclient.Video
    // ...
}

func NewServiceContext(c config.Config) *ServiceContext {
    rds := redis.MustNewRedis(...)
    return &ServiceContext{
        Auth:        middleware.NewAuthMiddleware(c.Auth.AccessSecret).Handle,
        RateLimiter: middleware.NewRateLimitMiddleware(rds, 3000, 5000, "...").Handle,
    }
}

// 所有 Logic 通过 svcCtx 获取依赖
l.svcCtx.VideoClient.UploadChunk(ctx, &videoclient.UploadChunkReq{...})
l.svcCtx.UserClient.Login(ctx, &userclient.LoginReq{...})
```

---

### 9. 代理模式

**文件:** `gateway/internal/svc/servicecontext.go`、`rpc/video/internal/server/videoserver.go`

**说明:** 两种代理形式共同使用：

**① gRPC 远程代理:**
Gateway 注入的 RPC 客户端是远程微服务的本地代理，对上层透明地发起网络调用。

```go
// gateway 中注入 gRPC 客户端代理
ctx.UserClient = userclient.NewUser(tryNewClient(c.UserRpc))
// UserClient 是 user-svc gRPC 服务的本地代理，隐藏网络通信细节
```

**② Server 静态代理:**
VideoServer 将 gRPC 请求委托给对应的 Logic 处理，起到分发和隔离作用。

```go
// rpc/video/internal/server/videoserver.go — 所有 RPC 方法委托给 Logic
type VideoServer struct {
    svcCtx *svc.ServiceContext
    video.UnimplementedVideoServer
}

func (s *VideoServer) UploadChunk(ctx context.Context, in *video.UploadChunkReq) (*video.UploadChunkResp, error) {
    l := logic.NewUploadChunkLogic(ctx, s.svcCtx) // 创建 Logic
    return l.UploadChunk(in)                       // 委托给 Logic 处理
}
```

---

### 10. 发布-订阅模式

**文件:** `gateway/internal/ws/danmaku.go` (Redis Pub/Sub)、`common/kafka/kafka.go` (Kafka)、`rpc/video/internal/consume/summaryconsumer.go`

**说明:** 使用两种消息中间件实现发布-订阅：

**① Redis Pub/Sub — 弹幕实时推送:**
```go
// gateway/internal/ws/danmaku.go
channel := "danmaku:" + videoId
pubsub := rdb.Subscribe(r.Context(), channel)  // 订阅特定视频频道的弹幕
defer pubsub.Close()

for msg := range pubsub.Channel() {
    // 实时推送弹幕到 WebSocket 客户端
    conn.WriteMessage(websocket.TextMessage, []byte(msg.Payload))
}
```

**② Kafka — 异步任务解耦:**
```go
// common/kafka/kafka.go — 三种 Topic 对应三种异步场景
const (
    TopicTranscodeTasks = "gopan.transcode.tasks"   // Publisher: video-svc → Subscriber: transcode-svc
    TopicMergeTasks     = "gopan.video.merge.tasks"  // Publisher: video-svc → Subscriber: async merge
    TopicSummaryTasks   = "gopan.video.summary.tasks"// Publisher: video-svc → Subscriber: summary-consumer
)

// Producer 工厂
func NewProducer(brokers []string, topic string) *kafkago.Writer { ... }
// Consumer 工厂
func NewConsumer(brokers []string, topic, groupID string) *kafkago.Reader { ... }
```

```go
// rpc/video/internal/consume/summaryconsumer.go — AI 摘要消费者
func StartSummaryConsumer(ctx context.Context, svcCtx *svc.ServiceContext) {
    reader := kafka.NewConsumer(svcCtx.Config.Kafka.Brokers, topic, "gopan-summary-worker")
    for {
        msg, _ := reader.FetchMessage(ctx)
        var task kafka.SummaryTask
        json.Unmarshal(msg.Value, &task)
        process(ctx, svcCtx, &task)  // 调用 summary-ai HTTP 服务
        reader.CommitMessages(ctx, msg)
    }
}
```

---

### 11. DTO 模式

**文件:** `common/response/response.go`、`rpc/video/video.proto`、`gateway/internal/types/types.go`

**说明:** 定义统一的数据传输对象，在不同层之间传递数据。

```go
// common/response/response.go — 统一 HTTP 响应体
type Body struct {
    Code    int    `json:"code"`
    Message string `json:"message"`
    Data    any    `json:"data,omitempty"`
}

// 所有 API 统一返回此格式
func Success(w http.ResponseWriter, r *http.Request, data any) {
    httpx.WriteJson(w, http.StatusOK, &Body{
        Code:    0,
        Message: "success",
        Data:    data,
    })
}
```

```protobuf
// rpc/video/video.proto — gRPC 消息 DTO（protobuf 自动生成 Go 类型）
message UploadChunkReq {
  string upload_id = 1;
  int64 video_id = 2;
  int32 chunk_index = 3;
  int32 file_size = 4;
  bytes data = 5;
}

message UploadChunkResp {
  int32 received_index = 1;
}
```

---

### 12. 仓储模式

**文件:** `rpc/video/store/video.go`

**说明:** `VideoStore` 封装所有数据库 CRUD 操作，隔离业务逻辑层与数据访问层。Logic 层只调用仓库方法，不直接编写 SQL。

```go
// rpc/video/internal/svc/servicecontext.go — 注入仓储
type ServiceContext struct {
    VideoStore *store.VideoStore  // 数据访问层
    // ...
}

// Logic 层只通过仓储操作数据
svcCtx.VideoStore.UpdateAiSummary(ctx, videoId, summary)
svcCtx.VideoStore.UpdateAiSummaryStatus(ctx, videoId, status)
```

---

### 13. 令牌桶模式

**文件:** `gateway/internal/middleware/ratelimitmiddleware.go`

**说明:** 基于 Redis + Lua 脚本实现分布式令牌桶算法，对 API 接口进行分级限流保护。

```go
// gateway/internal/middleware/ratelimitmiddleware.go
type RateLimitMiddleware struct {
    limiter *limit.TokenLimiter // go-zero 内置的分布式令牌桶
}

func NewRateLimitMiddleware(rds *redis.Redis, rate, burst int, keyPrefix string) *RateLimitMiddleware {
    return &RateLimitMiddleware{
        limiter: limit.NewTokenLimiter(rate, burst, rds, keyPrefix),
    }
}

// 分级限流 — 四个不同速率的令牌桶：
//   - RateLimiter:       3000/s, burst 5000  (全局兜底)
//   - RateLimiterList:   2000/s, burst 3000  (视频列表，流量大)
//   - RateLimiterDetail: 1000/s, burst 2000  (视频详情，中等)
//   - RateLimiterLogin:    50/s, burst  100  (登录，防撞库)
```

---

### 14. 前端控制器模式

**文件:** `gateway/internal/handler/routes.go`

**说明:** 所有 HTTP 请求通过统一的 `RegisterHandlers` 入口进行路由分发，便于集中配置鉴权中间件和限流策略。

```go
// gateway/internal/handler/routes.go — 统一路由注册
func RegisterHandlers(server *rest.Server, serverCtx *svc.ServiceContext) {
    // /api/admin — 无鉴权
    server.AddRoutes([]rest.Route{
        {Method: http.MethodPost, Path: "/login", Handler: admin.AdminLoginHandler(serverCtx)},
    }, rest.WithPrefix("/api/admin"))

    // /api/user — 无鉴权
    server.AddRoutes([]rest.Route{
        {Method: http.MethodPost, Path: "/login", Handler: user.LoginHandler(serverCtx)},
        {Method: http.MethodPost, Path: "/register", Handler: user.RegisterHandler(serverCtx)},
    }, rest.WithPrefix("/api/user"))

    // /api/video — 需要 JWT 鉴权
    server.AddRoutes(rest.WithMiddlewares(
        []rest.Middleware{serverCtx.Auth},  // 统配 Auth 中间件
        []rest.Route{
            {Method: http.MethodPost, Path: "/upload-chunk", Handler: video.UploadChunkHandler(serverCtx)},
            {Method: http.MethodGet, Path: "/list", Handler: video.ListVideosHandler(serverCtx)},
            // ... 20+ 路由
        },
    ), rest.WithPrefix("/api/video"))
}
```

---

## 二、前端设计模式 (Vue 3)

### 15. 状态管理模式

**文件:** `frontend/src/stores/auth.js`、`frontend/src/stores/video.js`

**说明:** 使用 Pinia (Composition API 风格) 实现全局状态管理，管理用户认证状态、视频列表等共享数据。

```javascript
// frontend/src/stores/auth.js — Pinia Composition API Store
export const useAuthStore = defineStore("auth", () => {
    const token = ref(localStorage.getItem("token") || "")
    const user = ref(JSON.parse(localStorage.getItem("user") || "null"))
    const isLoggedIn = computed(() => !!token.value) // 计算属性

    async function login(username, password) {
        const res = await userApi.login({ username, password })
        token.value = res.token
        user.value = { userId: res.user_id, username: res.username }
        localStorage.setItem("token", res.token)
        return res
    }

    function logout() {
        token.value = ""
        user.value = null
        localStorage.removeItem("token")
    }

    return { token, user, isLoggedIn, login, register, logout, fetchProfile }
})
```

**关键特性:**
- `ref()` 创建响应式状态
- `computed()` 派生状态
- 自动与 `localStorage` 同步持久化

---

### 16. 拦截器模式

**文件:** `frontend/src/api/request.js`

**说明:** Axios 拦截器在请求和响应两个阶段进行统一处理。

```javascript
// frontend/src/api/request.js
const request = axios.create({ baseURL: "/api", timeout: 10000 })

// ① 请求拦截器 — 自动注入 JWT Token
request.interceptors.request.use((config) => {
    const token = localStorage.getItem("token")
    if (token) config.headers.Authorization = `Bearer ${token}`
    return config
})

// ② 响应拦截器 — 统一解析 code/message
request.interceptors.response.use(
    (res) => {
        const body = res.data
        if (body && typeof body.code === "number" && body.code !== 0) {
            return Promise.reject(new Error(body.message || "请求失败"))
        }
        return body
    },
    (err) => Promise.reject(err)
)
```

**两个阶段:**
- **请求前:** 自动从 localStorage 读取 token 并注入到 Authorization 头
- **响应后:** 统一解析 `{ code, message, data }` 格式，code≠0 时自动 reject

---

### 17. 路由守卫模式

**文件:** `frontend/src/router/index.js`

**说明:** Vue Router 的 `beforeEach` 导航守卫在每次路由跳转前检查登录状态，实现前端层面的权限控制。

```javascript
// frontend/src/router/index.js
const routes = [
    // 需要登录的页面标记 meta.auth = true
    { path: "/upload", name: "Upload", component: () => import("../pages/Upload.vue"), meta: { auth: true } },
    { path: "/profile", name: "Profile", component: () => import("../pages/Profile.vue") },
]

// 全局前置守卫
router.beforeEach((to, from, next) => {
    const token = localStorage.getItem("token")
    if (to.meta.auth && !token) {
        next("/login")  // 未登录 → 重定向到登录页
    } else {
        next()          // 放行
    }
})
```

与后端 `AuthMiddleware` 形成**前后端双重鉴权**。

---

## 三、架构级设计模式

除了上述具体代码层面的设计模式，GoPan 在整体架构上也采用了以下模式：

### 微服务架构 (Microservices)

独立部署的 8 个服务：

| 服务 | 端口 | 功能 |
|------|------|------|
| `gateway` | 8888 | API 网关，统一入口 |
| `rpc/video` | 8082 | 视频 CRUD + 文件上传 |
| `rpc/user` | 8081 | 用户认证与管理 |
| `rpc/search` | 8083 | 全文/语义搜索 |
| `rpc/interact` | 8084 | 点赞/收藏/评论 |
| `rpc/stream` | 8085 | 播放流服务 |
| `rpc/transcode` | 8086 | 视频转码 |
| `rpc/admin` | 8087 | 管理后台 |

### 事件驱动架构 (Event-Driven)

通过 Kafka 实现服务间的异步解耦：

```
video-svc ──(TranscodeTask)──► transcode-svc
video-svc ──(MergeTask)─────► async merge worker
video-svc ──(SummaryTask)───► summary-consumer ──► summary-ai (HTTP)
```

### CQRS 读写分离

- **Command (写):** MySQL (`rpc/video/store/`) 负责视频元数据写入
- **Query (读):** Elasticsearch (`common/es/`) 负责视频搜索查询

---

## 总结表格

| 序号 | 设计模式 | 层级 | 关键文件 |
|------|---------|------|---------|
| 1 | 中间件/责任链 | 后端 | `gateway.go` + `authmiddleware.go` + `ratelimitmiddleware.go` |
| 2 | 策略模式 | 后端 | `common/es/es.go` SearchVideos |
| 3 | 熔断/降级 | 后端 | `common/es/es.go` getEmbeddingVector + SearchVideos |
| 4 | 工厂方法 | 后端 | 全项目 `NewXxx()` 函数 |
| 5 | 适配器 | 后端 | `common/storage/`、`common/es/`、`common/kafka/` |
| 6 | 外观模式 | 后端 | `gateway/` 整个层 |
| 7 | 模板方法 | 后端 | `handler/` + `logic/` go-zero 代码生成 |
| 8 | 依赖注入/IoC | 后端 | `servicecontext.go` |
| 9 | 代理模式 | 后端 | gRPC 客户端 + Server 静态代理 |
| 10 | 发布-订阅 | 后端 | `danmaku.go` (Redis) + `kafka.go` (Kafka) |
| 11 | DTO | 后端 | `response.go` + `video.proto` |
| 12 | 仓储模式 | 后端 | `rpc/video/store/` |
| 13 | 令牌桶 | 后端 | `ratelimitmiddleware.go` |
| 14 | 前端控制器 | 后端 | `handler/routes.go` |
| 15 | 状态管理 | 前端 | `stores/auth.js`、`stores/video.js` |
| 16 | 拦截器 | 前端 | `api/request.js` |
| 17 | 路由守卫 | 前端 | `router/index.js` |

> **总计:** 后端 14 种 + 前端 3 种 = **17 种设计模式**，外加微服务架构、事件驱动、CQRS 等架构级模式。