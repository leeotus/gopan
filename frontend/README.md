# GoPan Frontend

Vue 3 + Vite + Pinia + Vant 4 构建的视频点播平台前端。

## 目录结构

```
frontend/
├── index.html
├── package.json
├── vite.config.js
└── src/
    ├── main.js
    ├── App.vue                       # 根组件 + 底部导航栏
    ├── api/
    │   ├── request.js                # axios 封装 + JWT 注入 + 拦截器
    │   └── index.js                  # user/video/interact/search API
    ├── stores/
    │   ├── auth.js                   # 登录态 + localStorage 持久化
    │   └── video.js                  # 视频列表/详情/点赞/收藏 + mock 数据
    ├── router/
    │   └── index.js                  # 7 个路由 + 鉴权守卫
    ├── composables/
    │   └── utils.js                  # 格式化工具（时长/播放数/时间戳）
    ├── styles/
    │   └── global.css
    ├── components/                   # (预留公共组件)
    └── pages/
        ├── Home.vue                  # 首页：分类Tab + 视频网格 + 加载更多
        ├── Login.vue                 # 登录页
        ├── Register.vue              # 注册页
        ├── VideoDetail.vue           # 视频详情：播放器占位 + 信息 + 互动栏 + 评论
        ├── Search.vue                # 搜索页：关键词搜索 + 结果列表
        ├── Profile.vue               # 个人中心：用户卡片 + 菜单 + 退出
        └── Upload.vue               # 上传页：标题/简介 + 文件选择 + 上传流程
```

## 功能覆盖

| 功能 | 说明 |
|------|------|
| 注册/登录 | 表单校验 + token 持久化 + pinia 状态 |
| 视频列表 | 分类过滤、游标分页加载更多、mock 兜底 |
| 视频详情 | 封面占位、多码率标签、点赞/收藏、评论 |
| 搜索 | 后端优先 + 本地降级 |
| 上传 | 文件选择 → initUpload → uploadChunk → mergeChunks |
| 个人中心 | 登录/未登录双态、退出登录 |
| 路由守卫 | `/upload` 等需要登录的路由自动跳转登录页 |
| Mock | 6 条示例视频（含分类、多码率、点赞收藏状态） |

## 启动方式

```bash
cd frontend
npm install
npm run dev        # → http://localhost:3000
```

后端网关在 `:8888` 时，Vite 自动代理 `/api` 到后端。
