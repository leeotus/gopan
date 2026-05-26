把整个项目的难点和解决方案整理出来。

---

## GoPan 项目阶段性总结

### 一、架构层面

| 难点 | 问题 | 解决方案 |
|------|------|---------|
| **服务拆分的 ServiceContext 理解** | go-zero 依赖注入模式和 Spring 差异大，每个服务有自己的 ServiceContext（gateway 持有 RPC Client，user-svc 持有 DB Store） | 显式在 `NewServiceContext` 中创建所有依赖，通过 handler→logic 闭包传递 |
| **RPC Client 到处 new 导致连接池爆炸** | 不熟悉 go-zero 的 service context 注入模式，容易在每个 logic 里自己 `zrpc.MustNewClient` | ServiceContext 作为单例持有 client，logic 通过 `l.svcCtx.VideoClient` 复用 |
| **重试/异步任务需要 Kafka** | 最初用 Redis 做"假 MQ"，但 Redis 没有削峰/重试/持久化 | 引入 Apache Kafka（KRaft 模式），`video-svc` 做 Producer，`transcode-svc` 做 Consumer |
| **多码率 vs 单码率** | 最初设计 360p/480p/720p/1080p 四档，增加转码复杂度 | 第一版只做 1080p 单码率，后续扩展 |
| **中间件复用到 common/** | Kafka/ES/MinIO 客户端分散在各服务 | 抽到 `common/kafka/`、`common/es/`、`common/storage/` 三个公共库 |

---

### 二、转码模块

| 难点 | 问题 | 解决方案 |
|------|------|---------|
| **FFmpeg 转码实现** | Go 直接调 FFmpeg 需要 libav C 绑定，交叉编译复杂 | 用 `os/exec` 调 ffmpeg 命令行，Docker 镜像内置 ffmpeg |
| **Kafka 异步转码** | transcode-svc Consumer 需要调用 video-svc 回写状态 | 在 transcode-svc 的 ServiceContext 注入 `VideoClient`（gRPC），转码成功后调用 `TranscodeCallback` |
| **HLS 切片存储** | 转码后需要把 .ts 和 .m3u8 上传到 MinIO | transcode-svc 也注入 `MinioClient`，`processTranscode()` 遍历 HLS 输出目录逐个 `PutObject` |

---

### 三、断点上传（最大难点）

| 难点 | 问题 | 解决方案 |
|------|------|---------|
| **Proto 接口设计** | 需要 InitUpload / UploadChunk / UploadStatus / MergeChunks 四个新 RPC | 更新 `video.proto`，`goctl rpc protoc` 重新生成，同时更新 `gateway.api` |
| **MySQL 字段缺失** | 新增 `total_chunks` 和 `upload_id` 字段后，线上 MySQL 容器没列 | `ALTER TABLE` 手动加列，同时更新 `init.sql` |
| **Redis 进度追踪** | 需要记录哪些 chunk 已到达 | Redis Set `upload:{upload_id}:received`，`SADD` 标记，`SMEMBERS` 查询 |
| **MinIO ComposeObject** | 合并分片不能用 Go 服务端拼接（OOM 和带宽问题） | 用 MinIO `ComposeObject` API 服务端合并，零带宽 |
| **前端切片并发和重试** | 5MB chunk，浏览器并发上传，网络中断需重试 | `File.slice()` 切片，`Promise.all` 分批并发（3 个一批），失败重试 3 次 |
| **gRPC 消息大小限制** | **核心 bug**：gRPC 默认 `MaxRecvMsgSize = 4MB`，5MB chunk 被拒绝 `ResourceExhausted` | **改 chunk 为 3MB**，低于 4MB 限制 |
| **Gateway multipart 代理截断** | Vite 代理 `localhost:3000→8888` 对 multipart/form-data 请求可能截断 body | 前端直接用 axios 默认行为（不手动设置 Content-Type），让 axios 自动生成正确的 boundary |
| **Gate 日志二进制污染** | go-zero 中间件把请求 body（5MB 二进制）打到 json 日志，导致日志不可读 | 给 uploadchunk logic 加自定义日志，只打印 chunk_index 和 size，不打 body |
| **docker compose MySQL 8.4 兼容** | `default-authentication-plugin` 废弃 | 去掉 yaml 里的 `command` |
| **Docker Hub 断网** | `alpine:3.21` / `golang:1.26-alpine` 基础镜像拉不下来 | 改用 DaoCloud 镜像代理（`docker.m.daocloud.io/`） |

---

### 四、数据库 & 运维

| 难点 | 问题 | 解决方案 |
|------|------|---------|
| **MySQL 8.4 InnoDB 降级冲突** | 旧 volume 残留高版本 InnoDB 文件，新 8.0 镜像无法降级 | `docker volume rm gopan_mysql-data` 重建 |
| **goctl 代码生成覆盖逻辑** | 每次 `goctl api go` 后 types.go 里的 `M3U8Url` 大小写被重置 | sed 统一改为 `M3U8Url` |
| **`sql.NullTime` scan 失败** | SELECT 列数少于 model 字段数导致 `not matching destination to scan` | 补齐 SELECT 里的 `deleted_at` 列 |

---

### 五、前后端联调

| 难点 | 问题 | 解决方案 |
|------|------|---------|
| **前端登录 500** | `request.js` 拦截器检查 `res.data.code !== 0`，但 gateway 的 login 接口返回裸数据（没有 code 包装） | 改拦截器为：有 `code` 字段时才检查，否则直接通过 |
| **upload-status 返回全零** | gateway logic 是桩，没有调 video-svc gRPC | 注入 `SetRequest(r)`，从 query param 取 upload_id 然后调 `VideoClient.UploadStatus` |
| **merge-chunks 返回桩** | 同上 | 同上报修 |