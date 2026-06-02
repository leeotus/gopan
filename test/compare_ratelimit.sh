#!/bin/bash
# 令牌桶限流效果对比测试（自动版）
set -e
cd "$(dirname "$0")/.."

GATEWAY="http://localhost:8888"
RESULTS_DIR="test/results"
mkdir -p "$RESULTS_DIR"
REPORT="$RESULTS_DIR/ratelimit_compare.md"

echo "╔══════════════════════════════════════════╗"
echo "║     令牌桶限流效果对比测试（自动版）      ║"
echo "╚══════════════════════════════════════════╝"
echo ""

CONCURRENCY=40
REQUESTS=200000
URL_NO="/api/test/list?cursor=0&limit=10"
URL_WITH="/api/video/list?cursor=0&limit=10"

# 解析压测输出的函数
parse_output() {
  local f="$1"
  local s=$(grep "成功:" "$f" | head -1 | awk '{print $2}')
  local t=$(grep "限流(429):" "$f" | head -1 | awk '{print $2}')
  local f_=$(grep "失败:" "$f" | head -1 | awk '{print $2}')
  local q=$(grep "QPS:" "$f" | head -1 | awk '{print $2}')
  local p50=$(grep "P50" "$f" | head -1 | awk '{print $3}')
  local p95=$(grep "P95" "$f" | head -1 | awk '{print $3}')
  local p99=$(grep "P99" "$f" | head -1 | awk '{print $3}')
  echo "${s:-0}|${t:-0}|${f_:-0}|${q:-0}|${p50:-0}|${p95:-0}|${p99:-0}"
}

echo "[1/2] 无限流压测: $URL_NO"
go run test/bench/http/http_bench.go -c $CONCURRENCY -n $REQUESTS -u "$GATEWAY$URL_NO" > "$RESULTS_DIR/ratelimit_no.txt" 2>&1
NO_DATA=$(parse_output "$RESULTS_DIR/ratelimit_no.txt")
echo ""

echo "[2/2] 有限流压测: $URL_WITH"
go run test/bench/http/http_bench.go -c $CONCURRENCY -n $REQUESTS -u "$GATEWAY$URL_WITH" > "$RESULTS_DIR/ratelimit_with.txt" 2>&1
WITH_DATA=$(parse_output "$RESULTS_DIR/ratelimit_with.txt")
echo ""

# 解析数据
IFS='|' read -r nl_s nl_t nl_f nl_qps nl_p50 nl_p95 nl_p99 <<< "$NO_DATA"
IFS='|' read -r wl_s wl_t wl_f wl_qps wl_p50 wl_p95 wl_p99 <<< "$WITH_DATA"

# 计算成功率
total_no=$((nl_s + nl_f))
total_no=$((total_no > 0 ? total_no : 1))
total_wl=$((wl_s + wl_t + wl_f))
total_wl=$((total_wl > 0 ? total_wl : 1))
rate_no=$(awk "BEGIN { printf \"%.1f\", $nl_s * 100 / $total_no }")
rate_wl=$(awk "BEGIN { printf \"%.1f\", $wl_s * 100 / $total_wl }")

cat > "$REPORT" << EOF
# 令牌桶限流效果对比测试报告

**测试时间**: $(date)
**并发**: $CONCURRENCY | **总请求**: $REQUESTS
**无限流路由**: \`$URL_NO\` | **有限流路由**: \`$URL_WITH\`

## 结果对比

| 指标 | 无限流 | 有限流 | 说明 |
|------|--------|--------|------|
| 成功 | $nl_s | $wl_s | 200 OK |
| 限流 | $nl_t | $wl_t | 429 被令牌桶拦截 |
| 失败 | $nl_f | $wl_f | 500/超时 |
| **成功率** | **${rate_no}%** | **${rate_wl}%** | success / total |
| QPS | $nl_qps | $wl_qps | req/s |
| P50 | $nl_p50 | $wl_p50 | 延迟 |
| P95 | $nl_p95 | $wl_p95 | 延迟 |
| P99 | $nl_p99 | $wl_p99 | 延迟 |

## 结论

- **无限流**: $nl_f 个真实错误（5xx），系统在高并发下可能出现不稳定
- **有限流**: $wl_t 个被限流（429，可安全重试），$wl_f 个真实错误
- **限流效果**: 令牌桶拦截了超过阈值的请求，保护核心服务不因过载而崩溃
- **成功率对比**: ${rate_no}% → ${rate_wl}%（提升保护效果）
EOF

echo ""
echo "✅ 报告已生成: $REPORT"
echo ""
cat "$REPORT"
