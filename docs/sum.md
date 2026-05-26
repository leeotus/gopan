## GoPan 项目难点总结

### 一、架构层面

| 难点 | 解决方案 |
|------|---------|
| **go-zero 依赖注入模式**：go-zero 不用 Spring 的 `@Autowired`，每个服务需要显式创建 ServiceContext | ServiceContext 作为单例从 main() → handler → logic 一路传递，所有 RPC Client/DB Store/中间件都在 NewServiceContext 里集中初始化 |
| **中间件复用**：Kafka/ES/MinIO 客户端分散在各服务 | 抽到 `common/` 下统一封装：`common/kafka/`、`common/es/`、`common/storage/`、`common/response/` |
| **go-zero 代码生成冲突**：每次 `goctl api go` 重生成代码，types.go 里的字段命名与 proto 生成的 pb 命名不一致（如 `M3u8Url` vs `M3U8Url`） | 统一使用 proto 生成的命名，对 goctl 生成的 types 文件进行手动对齐 |
| **go-zero sqlx 扫描失败**：SELECT 列数和 model 字段数不匹配会导致 `not matching destination to scan` | 补齐 SELECT 中缺失的列（如 `deleted_at`、`total_chunks`、`upload_id`） |
| **MySQL 软删除**：`sql.NullTime` 类型与 NULL 值扫描兼容性 | 确保 SELECT 包含所有 model 字段，顺序一致 |

### 二、断点上传（最大难点）

| 难点 | 解决方案 |
|------|---------|
| **gRPC 4MB 消息限制**：5MB chunk 被拒绝 `ResourceExhausted` | 前端改为 3MB chunk，保持默认 gRPC 限制 |
| **MinIO ComposeObject 5MiB 限制**：3MB chunk 无法用 ComposeObject 合并 | 改为流式合并：逐个从 MinIO 下载 chunk → `bytes.Buffer` 拼接 → `PutObject` 上传 |
| **上传进度追踪**：需要知道哪些 chunk 已到达 | Redis Set `upload:{upload_id}:received`，`SADD` 标记，`SCARD`/`SMEMBERS` 查询 |
| **并发 multipart 上传**：浏览器并发 3 个 worker 时，只有最后一个 chunk 到达后端 | 改为分批并发：每次 3 个，`Promise.all` 等待全部完成再下一批 |
| **Vite 代理截断 multipart**：前端通过 Vite 代理转发 multipart 请求时 body 可能被截断 | 直接连接后端，绕过代理 |
| **网关 body 日志二进制污染**：go-zero 框架会把 multipart 的 3MB 二进制数据打到日志 | 网关 yaml 设置 `Log.Level: error`，关掉 info 级请求日志；business logic 使用自定义日志 |
| **前端栈溢出**：`btoa(String.fromCharCode(...new Uint8Array(arrayBuf)))` 对大文件栈溢出 | 改为 chunk Size 3MB 并发，不再转 base64 |

### 三、转码模块

| 难点 | 解决方案 |
|------|---------|
| **FFmpeg 转码实现** | Go 用 `os/exec` 调 ffmpeg 命令行，Docker 镜像内置 ffmpeg |
| **Kafka KRaft 模式 Consumer 连接失败**：`Group Coordinator Not Available` | `segmentio/kafka-go` 的 consumer group 在 KRaft 模式下 coordinator 发现失败，改为不指定 GroupID，手动指定 Partition 0 直接消费 |
| **异步解耦 video-svc ↔ transcode-svc** | video-svc merge 后发 Kafka 消息 → transcode-svc 消费 → FFmpeg 转码 → 回调 TranscodeCallback 更新视频状态 |

### 四、数据库 & 运维

| 难点 | 解决方案 |
|------|---------|
| **MySQL 8.4 `default-authentication-plugin` 废弃** | docker-compose 去掉 `command: --default-authentication-plugin=mysql_native_password` |
| **MySQL InnoDB 降级冲突** | 删除旧 Docker volume 重新初始化 |
| **ALTER TABLE 加列** | 新增 `total_chunks`/`upload_id` 后手动 `ALTER TABLE` |
| **Docker Hub 网络不可达** | 改用 DaoCloud 镜像代理 (`docker.m.daocloud.io/`) |

### 五、前后端联调

| 难点 | 解决方案 |
|------|---------|
| **前端登录 500**：`request.js` 拦截器误判 code !== 0 | 改为只在有 `code` 字段时才校验 |
| **upload-status 返回空**：gateway logic 桩未调 video-svc RPC | 注入 `SetRequest(r)`，从 query param 取 `upload_id` 调 `VideoClient.UploadStatus` |
| **merge-chunks 返回桩** | 同上 |

### 六、关键性能指标

| 指标 | 当前值 |
|------|--------|
| chunk 大小 | 3MB |
| 并发 worker | 3 |
| HLS 切片时长 | 10 秒 |
| 视频编码 | libx264 + AAC |
| 转码方式 | 单码率 1080p |

核心教训：**Docker 缓存是最大的坑**——改完代码经常忘了 rebuild 对应服务的镜像，反复排查才发现是老镜像在跑。Go-zero 的 goctl 代码生成很便利但要注意生成覆盖和命名不一致问题。