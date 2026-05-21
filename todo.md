# GoPan AI 功能规划

## 1. 智能视频标签 & 摘要

**场景**：用户上传视频后，自动生成标签 + 摘要，提升搜索和推荐质量。

**实现**：

- transcode-svc 在 FFmpeg 转码时抽取音频 WAV
- 调用 OpenAI Whisper API 或本地 whisper.cpp 转文字
- 文本 → 调用 LLM（DeepSeek 等）生成 3-5 个标签 + 50 字摘要
- 回写 `videos.description` 或新增 `ai_tags` JSON 字段

**成本**：Whisper ≈ ¥0.5/小时；LLM ≈ ¥0.02/条

---

## 2. 个性化视频推荐

**场景**：首页推荐可能感兴趣的视频，提升留存。

**实现**：

- Phase 1（冷启动）：基于 `videos.category` + `ai_tags` 做内容相似度推荐
- Phase 2（数据积累后）：用户行为（点赞/收藏/观看时长）→ `interact-svc` 写 Kafka `user-behavior` topic → `recommend-svc` 消费计算协同过滤
- 推荐结果存 Redis，每天离线更新一次

**成本**：纯计算，无外部 API 调用

---

## 3. AI 评论审核

**场景**：用户发表评论时自动检测违规内容（广告、辱骂、色情）。

**实现**：

- `interact-svc.PostComment` 发评论后，同步调用 `moderate-svc` gRPC
- 用阿里云内容安全 API 或本地敏感词库过滤
- 检测到违规 → 标记 `status=blocked`，不展示
- 不确定的 → 标为 `pending_review`

**成本**：阿里云内容安全 ≈ ¥0.5/千条；本地敏感词库免费

---

## 4. 视频内容理解

**场景**：理解视频画面内容（人物、场景、动作），生成精确时间轴标签。

**实现**：

- 定时抽取视频帧（每 5 秒一帧）→ CLIP/OpenAI Vision 识别
- 生成时间轴标签：`{ "0:00-0:05": "会议室", "0:05-0:10": "PPT讲解" }`
- 前端播放时可按标签跳转

**成本**：OpenAI Vision API ≈ ¥0.1/帧，1 小时视频 ≈ 720 帧 ≈ ¥72

---

## 5. 智能语义搜索

**场景**：用户搜 "怎么优化数据库查询" 能匹配到标题是 "MySQL 性能调优" 的视频。

**实现**：

- Phase 1：视频摘要/标签用 text-embedding 模型（BGE）转向量 → 存 ES `dense_vector` 字段
- Phase 2（RAG）：用户搜索 → 向量相似度 → top-5 送 LLM 重排序 → 返回最佳匹配

**成本**：embedding API ≈ ¥0.001/条；ES dense_vector 免费

---

## 推荐实施顺序

| 阶段 | 功能 | 依赖 | 工作量 |
|------|------|------|--------|
| 1 | 智能标签 & 摘要 | Whisper + LLM API | 2-3 天 |
| 2 | AI 评论审核 | 敏感词库 / 阿里云 API | 1 天 |
| 3 | 个性化推荐 | Kafka + Redis | 3-5 天 |
| 4 | 语义搜索 | embedding + ES | 2-3 天 |
| 5 | 视频内容理解 | Vision API | 5-7 天 |
