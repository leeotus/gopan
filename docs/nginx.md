再试试 ts 文件。完美。看实际缓存效果——第二次请求比第一次快多少。**速度提升 2.4x**（11ms → 4.7ms）。6.9MB 的 ts 切片从 nginx 磁盘缓存直接返回，不再走 Docker 网络到 MinIO。

---

## 变更总结

### 文件：`nginx/nginx.conf`

| 改动 | 说明 |
|---|---|
| `proxy_cache_path` 指令 | 定义磁盘缓存路径 ` /tmp/nginx_cache`，最大 1GB，key zone 50MB |
| `proxy_cache hls_cache` | `/videos/` location 启用缓存 |
| `proxy_cache_key $uri` | 按请求 URI 缓存 |
| `proxy_cache_valid 200 1h` | 成功的响应缓存 1 小时 |
| `proxy_cache_lock on` | 并发请求同一文件时，只发一次回源请求 |
| `X-Cache-Status` header | 响应头显示 MISS/HIT，方便调试 |

### 效果

| 请求类型 | 缓存前 | 缓存后 |
|---|---|---|
| `.m3u8` | ~5ms（MinIO 往返） | <1ms（磁盘） |
| `.ts`（6.9MB） | ~11ms | ~4.7ms |
| 1000 人同时播放 | 1000 次 MinIO 请求 | 1 次 MinIO + 999 次磁盘 |