项目名称：GoPan 视频点播平台
技术栈：Go-zero / gRPC / MySQL / Redis / Kafka / MinIO / Elasticsearch / Docker Compose / Vue 3 / Prometheus / Jaeger
项目地址：https://github.com/xxx/gopan

【项目简介】
从零搭建的微服务视频点播平台，前后端分离，支持断点上传、异步转码、弹幕 WebSocket 实时推送、管理员后台审核、ES 全文搜索、Prometheus + Jaeger 可观测性。

【核心亮点】

1. 断点上传
   - 前端 File.slice 切片，3MB chunk 并发上传，失败重试 3 次，断网恢复后补传缺失分片
   - 后端 Redis Set 追踪进度，MinIO ComposeObject 合并分片
   - 支持 200MB+ 大文件，gRPC 4MB 消息限制在服务端优雅绕过

2. Kafka 异步削峰
   - video-svc 上传完成后 Produce 转码任务到 Kafka，立即返回，用户感知延迟从 30s 降到 <10ms
   - transcode-svc 异步消费，FFmpeg 1080p HLS 转码，完成后回调更新视频状态
   - 系统吞吐从 0.05 转码/s 提升到 6500 req/s

3. 弹幕 WebSocket 实时推送
   - interact-svc 写 MySQL 持久化 + Redis Pub/Sub 广播
   - gateway WebSocket server 订阅 Redis channel，转发给同房间观众
   - 三层数据流：MySQL 永久存储 / Redis 实时推送 / WebSocket 浏览器通道

4. 全链路可观测性
   - Prometheus 采集 7 个微服务指标 + Grafana 仪表盘
   - Jaeger 全链路追踪，自动记录 HTTP → gRPC → SQL → Kafka 每一步延迟
   - 自研压测脚本，支持 50 并发 100 万请求，记录 P50/P95/P99 + 错误率

5. 服务设计
   - 8 个 gRPC 服务（user/video/transcode/stream/interact/search/admin + gateway）
   - etcd 服务发现，JWT 鉴权，bcrypt 密码加密
   - 管理员后台（视频审核通过/下架），ES 全文搜索，播放进度断点续播

6. 工程化
   - Makefile 一键编译启动，*.local.yaml 本地调试环境
   - Docker Compose 14 容器一体编排（MySQL/Redis/MinIO/Kafka/ES/Jaeger/Prometheus/Grafana）
   - VSCode launch.json 8 服务独立 debug 配置

【技术难点】

| 难点 | 解决方案 |
|------|---------|
| gRPC 默认 4MB 消息限制导致 5MB chunk 被拒绝 | 改为 3MB chunk + MinIO ComposeObject 服务端合并 |
| MinIO ComposeObject 要求 source ≥ 5MiB，3MB chunk 无法合并 | 改为流式下载拼接：逐个 GetObject → bytes.Buffer → PutObject |
| Kafka KRaft 模式 consumer coordinator 发现失败 | 去掉 GroupID，手动指定 Partition 消费 |
| go-zero/goctl 代码生成后 M3U8Url 大小写不一致 | 统一对齐为 M3U8Url |
| Docker Hub 网络不可达 | 改用 DaoCloud 镜像代理 |
| 网关 multipart 并发上传只有最后一个 chunk 到达 | 改为并发分批 + 每批等待完成后再下一批 |

【性能数据】

- HTTP 压测：QPS 6500，P99 46ms
- merge-chunks 异步改造后：1008ms → 1ms（提升 1000 倍）
- Kafka 削峰：用户感知延迟 30s → <10ms（提升 3000 倍）
- 上传 200MB 视频：8 个 chunk，并发 3 个 worker，约 2 秒完成