#!/bin/bash
JAEGER="${JAEGER_URL:-http://localhost:16686}"
SERVICE="${1:-video-svc}"
LIMIT="${2:-10}"
OUTPUT="test/results/jaeger_trace.md"

echo "查询 Jaeger: $JAEGER (service=$SERVICE, limit=$LIMIT)"

TRACES=$(curl -s "$JAEGER/api/traces?service=$SERVICE&limit=$LIMIT&lookback=1h" 2>/dev/null)
COUNT=$(echo "$TRACES" | python3 -c 'import json,sys; d=json.load(sys.stdin); print(len(d.get("data",[])))' 2>/dev/null || echo "0")

if [ "$COUNT" = "0" ]; then
  echo "⚠ 无 trace 数据"
  exit 0
fi

echo "✓ 获取到 $COUNT 条 trace"

cat > "$OUTPUT" << MDEOF
# GoPan 链路延迟报告

> 服务: $SERVICE | 采样: $COUNT 条 | 时间: $(date '+%Y-%m-%d %H:%M:%S')

MDEOF

echo "$TRACES" | python3 test/jaeger_parse.py >> "$OUTPUT"

echo ""
echo "✓ 报告已生成: $OUTPUT"
