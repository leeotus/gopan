该项目所有服务（gateway + 7 个 RPC）都通过 go-zero 自动暴露 `/metrics`，由 Prometheus 每 15 秒采集。以下按关注维度列出值得在 Grafana 建面板的 PromQL：

---

### 一、QPS & 延迟

| 面板 | PromQL |
|------|--------|
| **HTTP QPS** | `rate(http_server_requests_duration_ms_count{job="gateway"}[1m])` |
| **HTTP P99 延迟** | `histogram_quantile(0.99, rate(http_server_requests_duration_ms_bucket{job="gateway"}[1m]))` |
| **HTTP P50 延迟** | `histogram_quantile(0.50, rate(http_server_requests_duration_ms_bucket{job="gateway"}[1m]))` |
| **gRPC QPS（按接口）** | `rate(grpc_server_handled_total{job="video-svc"}[1m])` |
| **gRPC P99（按接口）** | `histogram_quantile(0.99, rate(grpc_server_handled_seconds_bucket{job="video-svc"}[1m]))` |

---

### 二、错误率

| 面板 | PromQL |
|------|--------|
| **HTTP 5xx 错误率** | `rate(http_server_requests_duration_ms_count{job="gateway",status="500"}[1m]) / rate(http_server_requests_duration_ms_count{job="gateway"}[1m]) * 100` |
| **HTTP 429 限流率** | `rate(http_server_requests_duration_ms_count{job="gateway",status="429"}[1m]) / rate(http_server_requests_duration_ms_count{job="gateway"}[1m]) * 100` |
| **gRPC 错误率** | `rate(grpc_server_handled_total{job="video-svc",grpc_code!="OK"}[1m]) / rate(grpc_server_handled_total{job="video-svc"}[1m]) * 100` |

---

### 三、熔断器

go-zero 内置 adaptive breaker，关键指标：

| 面板 | PromQL |
|------|--------|
| **熔断触发次数** | `rate(breaker_requests_total{result="drop"}[1m])` |
| **熔断通过次数** | `rate(breaker_requests_total{result="success"}[1m])` |
| **总请求量** | `rate(breaker_requests_total[1m])` |

熔断触发 > 0 说明下游（如 video-svc → MySQL）响应超时或被拖慢。

---

### 四、运行时

| 面板 | PromQL | 告警阈值建议 |
|------|--------|------------|
| **内存占用** | `go_memstats_alloc_bytes{job="video-svc"} / 1024 / 1024` | > 500MB |
| **堆内存** | `go_memstats_heap_inuse_bytes{job="video-svc"} / 1024 / 1024` | 持续增长 → 疑似泄漏 |
| **Goroutine 数** | `go_goroutines{job="video-svc"}` | > 10000 → 泄漏 |
| **GC Pause P50** | `rate(go_gc_duration_seconds_sum{job="video-svc"}[1m]) / rate(go_gc_duration_seconds_count{job="video-svc"}[1m])` | > 10ms |
| **CPU 使用率** | `rate(process_cpu_seconds_total{job="transcode-svc"}[1m]) * 100` | > 80% |

---

### 五、业务指标

需要应用层自行暴露，目前项目已有：

| 面板 | 数据源 | PromQL |
|------|--------|--------|
| **Redis 连接数** | Redis Exporter | `redis_connected_clients` |
| **Kafka Consumer Lag** | Kafka Exporter | `kafka_consumer_group_lag{group="transcode-group"}` |

> 注：**Consumer Lag 是最值得加的一个**。转码消费者积压说明 FFmpeg 处理速度跟不上上传速度，需要 scale up `transcode-svc`。

---

### 六、建议 Grafana Dashboard 布局

按优先级排列四行：

```
Row 1: HTTP QPS 曲线 | P99 延迟曲线 | 5xx/429 错误率
Row 2: gRPC QPS × 5（video/stream/interact 等，每人一小格）
Row 3: 熔断触发次数 | Kafka Lag | Goroutine 数
Row 4: 内存占用 × 8 服务（heatmap）
```

核心盯 3 个数：**P99 延迟**、**5xx 错误率**、**熔断触发**。任一个异常就顺着 Jaeger trace 看具体瓶颈。