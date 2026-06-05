# GoPan 可观测性手册

## 一、架构概览

```
微服务 (gateway + 7 RPC)
  │
  ├─ traces → Jaeger (16686 面板)
  ├─ metrics → Prometheus (9090 面板)
  └─ dashboards → Grafana (3001 面板, 数据源: Prometheus)
```

| 面板 | 地址 | 用途 |
|------|------|------|
| Jaeger | http://localhost:16686 | 每次请求的完整调用链 |
| Prometheus | http://localhost:9090 | 指标查询（QPS/P99/内存） |
| Grafana | http://localhost:3001 | 仪表盘可视化（admin / admin） |

---

## 二、链路追踪 (Jaeger)

### 2.1 怎么看一个请求的完整链路

1. 打开 `http://localhost:16686`
2. Service 下拉选 `user-svc`（或其他服务名）
3. 点击 **Find Traces**
4. 点击任一 trace，展开瀑布图

### 2.2 瀑布图解读

```
POST /api/video/upload (128ms)
├─ gateway: 1ms
├─ video-svc.InitUpload: 15ms
│   └─ MySQL INSERT: 10ms
├─ video-svc.UploadChunk × 3: 600ms
│   └─ MinIO PutObject × 3: 500ms
├─ video-svc.MergeChunks: 800ms
│   ├─ MinIO GetObject × 3: 300ms
│   └─ MinIO PutObject: 400ms
├─ Kafka Write: 5ms
│   (transcode-svc 异步消费)
└─ transcode-svc.ffmpeg: 30s
```

- 每个色块 = 一个 Span（跨服务调用）
- 色块越宽 = 耗时越长
- 红色 = 错误
- 点击 span 看详细信息（headers / tags / logs）

### 2.3 配置说明

各服务 yaml 中的 Telemetry 配置：

```yaml
Telemetry:
  Name: user-svc           # 服务名，Jaeger 面板按此筛选
  Endpoint: http://jaeger:4318  # OTLP HTTP 端口
  Sampler: 1.0             # 采样率，1.0 = 全量
  Batcher: otlphttp        # 协议，go-zero 支持 zipkin/otlpgrpc/otlphttp/file
```

**关键参数**：

| 参数 | 说明 | 建议值 |
|------|------|--------|
| Sampler | 采样率 | 开发环境 1.0，生产 0.1 |
| Batcher | 上报协议 | otlphttp（Jaeger 兼容） |
| Endpoint | 收集器地址 | Jaeger 容器 jaeger:4318 |

### 2.4 本地开发调试

本地启动 Jaeger（Docker）：

```bash
docker compose up -d jaeger
```

Jaeger 面板自带，无需额外安装。

---

## 三、指标监控 (Prometheus + Grafana)

### 3.1 Prometheus

**访问**：`http://localhost:9090`

**查询示例**：

| 查询 | 含义 |
|------|------|
| `grpc_server_handled_total` | gRPC 请求总数 |
| `http_server_requests_duration_ms_bucket` | HTTP 请求延迟分桶 |
| `go_memstats_alloc_bytes` | 内存占用 |
| `process_cpu_seconds_total` | CPU 占用 |

**查看所有指标**：`http://localhost:9090/targets` → 确认 7 个 RPC 服务 + gateway 都是 UP 状态。

### 3.2 Grafana 仪表盘

1. 打开 `http://localhost:3001`（admin / admin）
2. 左侧菜单 → **Administration** → **Data sources** → 添加 Prometheus，URL 填 `http://prometheus:9090`
3. 发布自己定制仪表盘

### 3.3 配置说明

Prometheus 抓取配置在 `etc/prometheus.yml`：

```yaml
scrape_configs:
  - job_name: gateway
    static_configs:
      - targets: ["gateway:9102"]   # HTTP 服务 /metrics 端口 = 9102
  - job_name: user-svc
    static_configs:
      - targets: ["user-svc:9101"]  # RPC 服务 /metrics 端口 = 9101
```

go-zero 自动暴露 `/metrics` 端点，无需添加代码。

---

## 四、常见问题

### Q1：Jaeger 看不到 trace 数据

1. 确认 Jaeger 容器 `docker compose ps jaeger` 是 Up 状态
2. 确认每个微服务的 yaml 里 Telemetry.Endpoint 是 `http://jaeger:4318`
3. 确认 Batcher 是 `otlphttp`

### Q2：Prometheus target 是 DOWN

检查 Prometheus 配置文件 `etc/prometheus.yml` 里的 targets 是否包含正确的服务名和端口。

### Q3：telemetry 字段冲突

`RpcServerConf` / `RestConf` 中内置了 Telemetry 字段，**不要在 config 结构体中再次声明 `Telemetry` 字段**，只需在 yaml 中配置即可。

### Q4：`batcher "jaeger" is not defined`

go-zero 不直接支持 `jaeger` batcher，应使用 `otlphttp`，Jaeger 会自动接收 OTLP 协议数据。

确认所有微服务 target 都 UP 之后，Prometheus 就是给自己用的——四种方式查数据。

# Prometheus使用方法

## 1. 查所有服务是否健康

`http://localhost:9090/targets` → 确认全部 State 为 UP。

## 2. 语义查询

打开 `http://localhost:9090` → Scope 框，输入以下查询：

| 你想知道 | 输入 |
|---------|------|
| user-svc 总请求数 | `grpc_server_handled_total{job="user-svc"}` |
| user-svc 每个接口平均耗时 | `rate(grpc_server_handled_seconds_sum{job="user-svc"}[1m]) / rate(grpc_server_handled_seconds_count{job="user-svc"}[1m])` |
| gateway HTTP 错误率 | `rate(http_server_requests_duration_ms_count{job="gateway",status="500"}[1m]) / rate(http_server_requests_duration_ms_count{job="gateway"}[1m]) * 100` |
| CPU 占用 | `process_cpu_seconds_total{job="user-svc"}` |
| 内存占用 | `go_memstats_alloc_bytes{job="user-svc"}` |
| 熔断次数 | `breaker_requests_total{result="drop"}` |

---

## 3. 看曲线图

输入查询后点击 **Graph** 标签，自动生成时间曲线。

---

## 4. Grafana 里建仪表盘

`http://localhost:3001`（admin/admin）→ Add data source → Prometheus → URL = `http://prometheus:9090`。然后建 dashboard，每个面板贴一条 PromQL 查询就行。