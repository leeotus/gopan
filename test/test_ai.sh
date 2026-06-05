#!/bin/bash
# GoPan AI 服务端到端测试脚本
# 用法：bash test/test_ai.sh
# 涵盖：
#   1. semantic-ai (9900) 健康检查 + 文本向量化测试
#   2. summary-ai  (9920) 健康检查
#   3. ES 中向量已写入校验
#   4. 端到端语义搜索（gateway /api/search/videos）

set -uo pipefail

AI_SEM="http://localhost:9900"
AI_SUM="http://localhost:9920"
ES="http://localhost:9200"
GATEWAY="http://localhost:8888"
INDEX="gopan_videos"

GREEN=$(printf '\033[32m')
RED=$(printf '\033[31m')
YELLOW=$(printf '\033[33m')
RESET=$(printf '\033[0m')

pass() { echo "${GREEN}✅ $*${RESET}"; }
fail() { echo "${RED}❌ $*${RESET}"; }
warn() { echo "${YELLOW}⚠️  $*${RESET}"; }

echo "═══════════════════════════════════════════════════════════"
echo "  GoPan AI 测试套件"
echo "═══════════════════════════════════════════════════════════"

# ─── 1. semantic-ai 健康检查 ───────────────────────────────────
echo ""
echo "[1/5] semantic-ai 健康检查"
health=$(curl -s "$AI_SEM/health" 2>&1)
echo "  $health"
status=$(echo "$health" | jq -r '.status' 2>/dev/null)
if [ "$status" = "healthy" ]; then pass "semantic-ai 已就绪"; else fail "semantic-ai 不可用"; fi

# ─── 2. semantic-ai 文本向量化 ────────────────────────────────
echo ""
echo "[2/5] semantic-ai 文本向量化"
resp=$(curl -s -X POST "$AI_SEM/embed/text" \
  -H 'Content-Type: application/json' \
  -d '{"text":"一只蜜蜂在花丛中采蜜"}')
dim=$(echo "$resp" | jq -r '.dimension' 2>/dev/null)
head=$(echo "$resp" | jq -c '.vector[0:3]' 2>/dev/null)
if [ "$dim" = "512" ]; then
  pass "返回 512 维向量，前三维 = $head"
else
  fail "向量化失败，响应：$resp"
fi

# ─── 3. summary-ai 健康检查 ───────────────────────────────────
echo ""
echo "[3/5] summary-ai 健康检查"
health=$(curl -s "$AI_SUM/health" 2>&1)
echo "  $health"
ready=$(echo "$health" | jq -r '.whisper_ready' 2>/dev/null)
if [ "$ready" = "true" ]; then pass "summary-ai Whisper 已加载"; else warn "summary-ai 模型未就绪（首次启动需下模型）"; fi

minimax=$(echo "$health" | jq -r '.minimax_configured' 2>/dev/null)
deepseek=$(echo "$health" | jq -r '.deepseek_configured' 2>/dev/null)
if [ "$minimax" = "true" ] || [ "$deepseek" = "true" ]; then
  pass "LLM Key 已配置（MiniMax=$minimax, DeepSeek=$deepseek）"
else
  warn "未配 LLM Key，摘要走本地兜底文案"
fi

# ─── 4. ES 索引向量字段校验 ──────────────────────────────────
echo ""
echo "[4/5] ES 已索引视频 + 向量回填情况"
total=$(curl -s "$ES/$INDEX/_count" | jq -r '.count')
with_vec=$(curl -s "$ES/$INDEX/_count" \
  -H 'Content-Type: application/json' \
  -d '{"query":{"exists":{"field":"video_vector"}}}' | jq -r '.count')
echo "  ES 总视频数: $total"
echo "  已含 video_vector: $with_vec"
if [ "$total" = "$with_vec" ] && [ "$total" -gt 0 ]; then
  pass "全部视频已完成向量化"
elif [ "$total" -gt 0 ]; then
  warn "仍有 $((total - with_vec)) 条缺向量。执行 bash test/reindex_vectors.sh 回填"
else
  warn "ES 索引为空，先上传视频"
fi

# ─── 5. 端到端语义搜索 ──────────────────────────────────────
echo ""
echo "[5/5] 端到端语义搜索（gateway → search-svc → ES KNN）"
for kw in "蜜蜂" "昆虫" "电影" "动物"; do
  resp=$(curl -s --data-urlencode "keyword=$kw" -G "$GATEWAY/api/search/videos" \
    --data "page=1" --data "size=5")
  cnt=$(echo "$resp" | jq -r '.total // 0' 2>/dev/null)
  titles=$(echo "$resp" | jq -r '.videos[]?.title // empty' 2>/dev/null | tr '\n' ',' | sed 's/,$//')
  if [ -z "$titles" ]; then titles="(无)"; fi
  echo "  关键词「$kw」→ $cnt 条 → $titles"
done

# 查 search-svc 日志看走的是 KNN 还是 BM25
echo ""
echo "  search-svc 最近的搜索路径："
docker logs --since 30s gopan-search-svc 2>&1 | grep -iE "AI Search|knn" | tail -5 | sed 's/^/    /'

echo ""
echo "═══════════════════════════════════════════════════════════"
echo "  测试完成"
echo "═══════════════════════════════════════════════════════════"
