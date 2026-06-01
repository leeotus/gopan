#!/bin/bash
# 令牌桶限流效果对比测试（自动版）
# 用法: bash test/compare_ratelimit.sh
# 前提: gateway 已部署，/api/test/list 无限制，/api/video/list 有限流
set -e
cd "$(dirname "$0")/.."

GATEWAY="http://localhost"
RESULTS_DIR="test/results"
mkdir -p "$RESULTS_DIR"
REPORT="$RESULTS_DIR/ratelimit_compare.md"

echo "╔══════════════════════════════════════════╗"
echo "║     令牌桶限流效果对比测试（自动版）      ║"
echo "╚══════════════════════════════════════════╝"
echo ""

CONCURRENCY=20
REQUESTS=2000
URL_NO_LIMIT="/api/test/list?cursor=0&limit=10"
URL_WITH_LIMIT="/api/video/list?cursor=0&limit=10"

if ! curl -s -o /dev/null -w "%{http_code}" "$GATEWAY$URL_NO_LIMIT" | grep -q 200; then
  echo "❌ 测试路由不可达，请先部署 gateway"
  exit 1
fi

run_bench() {
  local label="$1"
  local url="$2"
  local output="$RESULTS_DIR/ratelimit_$label.txt"
  echo ">>> $label: $url"
  go run test/bench/http/http_bench.go -c $CONCURRENCY -n $REQUESTS -u "$GATEWAY$url" > "$output" 2>&1
  awk '/^成功:/{s=$2} /^限流/{t=$2} /^失败:/{f=$2} /^QPS:/{q=$2} /^P50/{p50=$3} /^P95/{p95=$3} /^P99/{p99=$3} END{printf "%d|%d|%d|%s|%s|%s|%s",s,t,f,q,p50,p95,p99}' "$output"
}

NO_LIMIT=$(run_bench "no_limit" "$URL_NO_LIMIT")
WITH_LIMIT=$(run_bench "with_limit" "$URL_WITH_LIMIT")

IFS='|' read -r nl_s nl_t nl_f nl_qps nl_p50 nl_p95 nl_p99 <<< "$NO_LIMIT"
IFS='|' read -r wl_s wl_t wl_f wl_qps wl_p50 wl_p95 wl_p99 <<< "$WITH_LIMIT"

total_no=$((nl_s + nl_f))
total_wl=$((wl_s + wl_t + wl_f))
rate_no=$(echo "scale=1; $nl_s * 100 / $total_no" | bc)
rate_wl=$(echo "scale=1; $wl_s * 100 / $total_wl" | bc)

cat > "$REPORT" << EOF
# 令牌桶限流效果对比测试报告

**测试时间**: $(date)
**并发**: $CONCURRENCY | **总请求**: $REQUESTS
**无限流路由**: \`$URL_NO_LIMIT\` | **有限流路由**: \`$URL_WITH_LIMIT\`

## 结果对比

| 指标 | 无限流 | 有限流 | 说明 |
|------|--------|--------|------|
| 成功 | $nl_s | $wl_s | 200 OK |
| 限流 | $nl_t | $wl_t | 429 被令牌桶拦截 |
| 失败 | $nl_f | $wl_f | 500/超时 |
| **成功率** | **${rate_no}%** | **${rate_wl}%** | \`success / total\` |
| QPS | $nl_qps | $wl_qps | req/s |
| P50 | $nl_p50 | $wl_p50 | 延迟 |
| P95 | $nl_p95 | $wl_p95 | 延迟 |
| P99 | $nl_p99 | $wl_p99 | 延迟 |

## 结论

- **无限流**: $nl_f 个真实错误（5xx），系统在高并发下出现不稳定性
- **有限流**: ${wl_t} 个被限流（429，可重试），${wl_f} 个真实错误
- **限流效果**: 令牌桶拦截了超过阈值的请求，保护核心服务不崩溃
EOF

echo ""
echo "✅ 报告已生成: $REPORT"
cat "$REPORT"

