# GoPan 源码阅读指南

> 以「用户注册 → 登录」为例，按 go-zero 框架的标准分层，逐层阅读 7 个微服务的代码。

---

## 一、阅读路径总览

```mermaid
flowchart LR
    subgraph step1[" ① 定义层 "]
        PROTO["rpc/**/*.proto"]
        API["api/gateway.api"]
    end

    subgraph step2[" ② 生成代码层 "]
        PB["*.pb.go / *_grpc.pb.go"]
        ZRPC["zrpc 桩: user.go / userclient/"]
        ZAPI["rest 桩: handler/ logic/ types/"]
    end

    subgraph step3[" ③ 配置 & 启动 "]
        YAML["etc/*.yaml"]
        CONFIG["internal/config/"]
        SVC["internal/svc/"]
    end

    subgraph step4[" ④ 基础设施 "]
        MODEL["model/"]
        STORE["store/"]
        MIDDLEWARE["middleware/"]
    end

    subgraph step5[" ⑤ 业务逻辑 "]
        RPC_LOGIC["rpc/*/internal/logic/"]
        GW_LOGIC["gateway/internal/logic/"]
    end

    step1 --> step2 --> step3 --> step4 --> step5
```

---

## 二、一条完整的注册请求是怎么流转的？

以 `POST /api/user/register` 为例，从 HTTP 请求到 MySQL INSERT 全链路。

```mermaid
sequenceDiagram
    autonumber

    participant Client as 浏览器 / curl
    participant GW as gateway<br/>(HTTP :8888)
    participant Handler as RegisterHandler<br/>gateway/internal/handler/user/
    participant GWLogic as RegisterLogic<br/>gateway/internal/logic/user/
    participant UserRPC as user-svc<br/>(gRPC :8081)
    participant RPCLogic as RegisterLogic<br/>rpc/user/internal/logic/
    participant Store as UserStore<br/>rpc/user/store/
    participant DB as MySQL

    Client->>GW: POST /api/user/register<br/>{"username":"alice","password":"123","email":"alice@test.com"}

    rect rgb(240, 248, 255)
        Note over GW: ① 路由注册：gateway/internal/handler/routes.go
        GW->>Handler: server.AddRoutes(...)<br/>匹配到 RegisterHandler
    end

    rect rgb(255, 248, 240)
        Note over Handler: ② Handler 层：解析请求
        Handler->>Handler: httpx.Parse(r, &req)<br/>JSON → types.RegisterReq
        Handler->>GWLogic: NewRegisterLogic(ctx, svcCtx)
        Handler->>GWLogic: l.Register(&req)
    end

    rect rgb(240, 255, 240)
        Note over GWLogic: ③ Gateway Logic：编排层
        GWLogic->>GWLogic: 从 svcCtx 取 UserClient
        GWLogic->>UserRPC: userClient.Register(req)<br/>gRPC 调用 → user-svc
    end

    rect rgb(255, 240, 255)
        Note over UserRPC: ④ RPC Server 接收
        UserRPC->>RPCLogic: user.RegisterUserServer<br/>→ RegisterLogic.Register(in)
    end

    rect rgb(255, 255, 240)
        Note over RPCLogic: ⑤ RPC Logic：业务逻辑
        RPCLogic->>Store: FindByUsername("alice")
        Store->>DB: SELECT * FROM users WHERE username=?
        DB-->>Store: nil (用户不存在，正常)
        RPCLogic->>RPCLogic: bcrypt.GenerateFromPassword(password)
        RPCLogic->>Store: Insert(user)
        Store->>DB: INSERT INTO users (...) VALUES (...)
        DB-->>Store: LastInsertId = 1
        RPCLogic-->>UserRPC: &RegisterResp{UserId: 1}
    end

    UserRPC-->>GWLogic: &RegisterResp{UserId: 1}
    GWLogic-->>Handler: &BaseResp{Message: "注册成功"}
    Handler-->>GW: httpx.OkJson(w, resp)
    GW-->>Client: 200 {"code":0,"message":"注册成功","data":null}
```

---

## 三、按层次分步阅读

### 第 1 步：看定义（不用懂代码，看接口长什么样）

| 文件 | 说明 | 重点看什么 |
|------|------|-----------|
| `api/gateway.api` | 所有 HTTP 接口定义 | `@handler` 和 `post/get/put/delete` 路由 |
| `rpc/user/user.proto` | 用户服务 RPC 定义 | `service User` 下的 `rpc` 方法签名 |
| `rpc/video/video.proto` | 视频服务 RPC 定义 | 流的声明 `stream` / 普通 `rpc` |
| `rpc/stream/stream.proto` | 播放服务 RPC 定义 | `GetPlayUrl` 防盗链 |
| `rpc/interact/interact.proto` | 互动服务 RPC 定义 | 点赞/评论/弹幕 |
| `rpc/transcode/transcode.proto` | 转码服务 RPC 定义 | 任务提交与查询 |
| `rpc/search/search.proto` | 搜索服务 RPC 定义 | ES 索引/搜索 |

### 第 2 步：看自动生成代码（理解 go-zero 的分层模式）

**以 user-svc 为例**，代码生成后的文件分工：

```mermaid
flowchart TB
    subgraph generated["goctl 自动生成 (修改 .proto 后 make proto 重新生成)"]
        direction LR
        A["user.proto"] -->|goctl rpc protoc| B["user.pb.go<br/>序列化/反序列化"]
        A -->|goctl rpc protoc| C["user_grpc.pb.go<br/>gRPC Client/Server 接口"]
        A -->|goctl rpc protoc| D["user.go<br/>main() 入口"]
        A -->|goctl rpc protoc| E["userclient/user.go<br/>Client 代理"]
        A -->|goctl rpc protoc| F["internal/server/<br/>Server 注册"]
        A -->|goctl rpc protoc| G["internal/logic/<br/>业务逻辑骨架"]
        A -->|goctl rpc protoc| H["internal/config/<br/>配置结构体"]
        A -->|goctl rpc protoc| I["internal/svc/<br/>ServiceContext"]
    end

    style A fill:#f9f,stroke:#333
```

访问顺序（从外到内）：

```mermaid
flowchart LR
    A["userclient/user.go<br/>(网关调用入口)"] --> B["user_grpc.pb.go<br/>(gRPC传输层)"]
    B --> C["internal/server/<br/>userserver.go<br/>(服务端接收)"]
    C --> D["internal/logic/<br/>xxxlogic.go<br/>(业务逻辑)"]
    D --> E["store/ model/<br/>(数据库访问)"]
```

### 第 3 步：看配置 & 启动过程

```mermaid
flowchart TB
    subgraph startup["服务启动流程"]
        direction LR
        CFG["etc/user.yaml<br/>端口/etcd/DB/密钥"] -->|conf.MustLoad| CONFIG["internal/config/config.go<br/>Config 结构体"]
        CONFIG -->|NewServiceContext| SVC["internal/svc/servicecontext.go<br/>初始化 DB/Redis/RPC Client"]
        SVC -->|传入| MAIN["user.go main()<br/>zrpc.MustNewServer"]
        MAIN -->|Register| SERVER["internal/server/userserver.go<br/>注册到 etcd"]
    end

    style CFG fill:#ffe,stroke:#333
    style CONFIG fill:#eef,stroke:#333
    style SVC fill:#efe,stroke:#333
```

关键文件（每个服务都一样模式）：

| 文件 | 作用 |
|------|------|
| `rpc/user/user.go` | 入口 `main()`，加载配置，启动 gRPC server |
| `rpc/user/internal/config/config.go` | 定义配置结构体（yaml → struct） |
| `rpc/user/internal/svc/servicecontext.go` | 服务上下文，持有 DB 连接、RPC Client 等 |
| `rpc/user/internal/server/userserver.go` | gRPC Server，将请求路由到 logic |

### 第 4 步：看业务逻辑（最核心）

**每条链路都是三段式**：

```mermaid
flowchart LR
    subgraph gateway["API 网关 (gateway/)"]
        H["handler/*handler.go<br/>解析HTTP请求"] --> L["logic/*logic.go<br/>编排层，调用RPC"]
    end

    subgraph rpc["RPC 服务 (rpc/*/)"]
        S["internal/server/*server.go<br/>接收gRPC请求"] --> LL["internal/logic/*logic.go<br/>业务逻辑+DB操作"]
    end

    gateway -->|gRPC| rpc
```

重点阅读文件（以用户注册为例）：

```
gateway/internal/handler/user/registerhandler.go    ← 解析 HTTP 请求
gateway/internal/logic/user/registerlogic.go        ← 调用 user-svc RPC
rpc/user/internal/logic/registerlogic.go            ← 校验用户名 + bcrypt + INSERT
rpc/user/store/user.go                              ← 数据库操作
rpc/user/model/user.go                              ← 数据模型
```

### 第 5 步：看中间件 & 公共层

| 文件 | 说明 |
|------|------|
| `gateway/internal/middleware/authmiddleware.go` | JWT 鉴权中间件，拦截 `/api/video/*` 路由 |
| `common/response/response.go` | 统一错误码和响应格式 |

---

## 四、7 个服务依赖关系

```mermaid
flowchart TB
    Client["浏览器/App"]
    GW["gateway :8888<br/>HTTP API 网关"]

    U["user-svc :8081<br/>注册/登录/用户信息"]
    V["video-svc :8082<br/>上传/列表/详情/管理"]
    T["transcode-svc :8083<br/>FFmpeg转码/多码率HLS"]
    S["stream-svc :8084<br/>播放地址/防盗链签名"]
    I["interact-svc :8085<br/>点赞/收藏/评论/弹幕"]
    SE["search-svc :8086<br/>ES全文搜索"]

    MYSQL[("MySQL :3306")]
    REDIS[("Redis :6379")]
    MINIO[("MinIO :9000")]
    ETCD[("etcd :2379")]
    ES[("Elasticsearch :9200")]

    Client -->|HTTP| GW

    GW -->|gRPC| U
    GW -->|gRPC| V
    GW -->|gRPC| T
    GW -->|gRPC| S
    GW -->|gRPC| I
    GW -->|gRPC| SE

    GW -.->|注册发现| ETCD
    U -.->|注册发现| ETCD
    V -.->|注册发现| ETCD
    T -.->|注册发现| ETCD
    S -.->|注册发现| ETCD
    I -.->|注册发现| ETCD
    SE -.->|注册发现| ETCD

    U --> MYSQL
    V --> MYSQL
    V --> MINIO
    T --> MINIO
    I --> MYSQL
    I --> REDIS
    SE --> ES
    S --> REDIS

    style GW fill:#4fc3f7,color:#000
    style U fill:#81c784,color:#000
    style V fill:#81c784,color:#000
    style T fill:#81c784,color:#000
    style S fill:#81c784,color:#000
    style I fill:#81c784,color:#000
    style SE fill:#81c784,color:#000
    style ETCD fill:#ffb74d,color:#000
```

---

## 五、推荐阅读顺序

```mermaid
flowchart TD
    A["1. 读 api/gateway.api<br/>了解有哪些接口"] --> B["2. 读 rpc/user/user.proto<br/>了解 RPC 定义格式"]
    B --> C["3. 读 gateway/internal/handler/routes.go<br/>看路由怎么注册的"]
    C --> D["4. 读 user-svc 完整链路<br/>注册 Handler → Logic → RPC → Logic → Store → DB"]
    D --> E["5. 读 gateway/internal/svc/servicecontext.go<br/>理解依赖注入"]
    E --> F["6. 读 gateway/internal/middleware/authmiddleware.go<br/>看中间件怎么拦截"]
    F --> G["7. 读 video-svc (model + store)<br/>看 DB 层封装"]
    G --> H["8. 读 stream-svc (防盗链签名)<br/>MD5 时间戳签名"]
    H --> I["9. 读其余服务<br/>(transcode/interact/search) 桩实现"]
    I --> J["10. 运行 docker-compose up -d<br/>看效果"]
```

---

## 六、核心概念速查

| go-zero 概念 | 对应位置 | 说明 |
|-------------|---------|------|
| `.api` 文件 | `api/gateway.api` | HTTP 接口定义，goctl 生成 handler/logic/types |
| `.proto` 文件 | `rpc/*/*.proto` | gRPC 接口定义，goctl 生成 pb + zrpc 桩 |
| Handler | `gateway/internal/handler/` | 解析 HTTP 请求/响应，不做业务 |
| Logic | `gateway/internal/logic/` + `rpc/*/internal/logic/` | 业务逻辑，写代码的地方 |
| ServiceContext | `*/internal/svc/servicecontext.go` | 依赖注入容器（DB、Redis、RPC Client 等） |
| Config | `*/internal/config/config.go` | yaml → struct 映射 |
| Routes | `gateway/internal/handler/routes.go` | 路由注册，goctl 自动生成勿手动改 |
| Middleware | `gateway/internal/middleware/` | HTTP 中间件（鉴权） |
| Model | `rpc/*/model/` | 数据库表结构映射 |
| Store | `rpc/*/store/` | 数据库 CRUD 操作封装 |
