set -e
cd "$(dirname "$0")/.."

VIDEO_FILE="${1:-test/test.mp4}"
GATEWAY="http://localhost:8888"
RESULTS="test/results"
mkdir -p "$RESULTS"

# ── 可调参数 ──
BENCH_CONCURRENCY=60    # HTTP 压测并发数
BENCH_REQUESTS=80000      # HTTP 压测总请求数
BENCH_URL="$GATEWAY/api/video/list?cursor=0&limit=10"  # 压测目标接口

# 颜色
RED='\033[0;31m'; GREEN='\033[0;32m'; YELLOW='\033[1;33m'; CYAN='\033[0;36m'; NC='\033[0m'

if [ ! -f "$VIDEO_FILE" ]; then
  echo -e "${RED}✗ 测试视频不存在: $VIDEO_FILE${NC}"
  echo "  请将测试视频放到 test/test.mp4"
  exit 1
fi

echo -e "${CYAN}╔══════════════════════════════════════════╗${NC}"
echo -e "${CYAN}║        GoPan 自动化测试                  ║${NC}"
echo -e "${CYAN}╚══════════════════════════════════════════╝${NC}"
echo ""
echo "测试视频: $VIDEO_FILE ($(du -h "$VIDEO_FILE" | cut -f1))"
echo ""

# ── 1. 健康检查 ──
echo -e "${YELLOW}[1/6] 健康检查...${NC}"
HTTP_CODE=$(curl -s -o /dev/null -w "%{http_code}" "$GATEWAY/api/video/list?cursor=0&limit=1" 2>/dev/null || echo "000")
if [ "$HTTP_CODE" = "200" ]; then
  echo -e "  ${GREEN}✓ gateway 可达 (HTTP $HTTP_CODE)${NC}"
else
  echo -e "  ${RED}✗ gateway 不可达 (HTTP $HTTP_CODE)${NC}"
  echo "  请先启动: make docker-up"
  exit 1
fi

# ── 2. 登录 ──
echo -e "${YELLOW}[2/6] 登录...${NC}"
LOGIN=$(curl -s -X POST "$GATEWAY/api/user/login" \
  -H "Content-Type: application/json" \
  -d '{"username":"flareon","password":"123456"}' 2>/dev/null)

TOKEN=$(echo "$LOGIN" | python3 -c "import sys,json; print(json.load(sys.stdin).get('token',''))" 2>/dev/null)
if [ -z "$TOKEN" ]; then
  echo -e "  ${RED}✗ 登录失败${NC}"
  exit 1
fi
echo -e "  ${GREEN}✓ token 获取成功${NC}"

# ── 3. 视频上传 ──
echo -e "${YELLOW}[3/6] 视频分片上传...${NC}"
FILE_SIZE=$(stat -c%s "$VIDEO_FILE")
CHUNK_SIZE=$((3 * 1024 * 1024))
TOTAL_CHUNKS=$(( (FILE_SIZE + CHUNK_SIZE - 1) / CHUNK_SIZE ))

INIT=$(curl -s -X POST "$GATEWAY/api/video/init-upload" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d "{\"filename\":\"$(basename $VIDEO_FILE)\",\"title\":\"性能测试视频\",\"file_size\":$FILE_SIZE,\"total_chunks\":$TOTAL_CHUNKS}" 2>/dev/null)

VIDEO_ID=$(echo "$INIT" | python3 -c "import sys,json; print(json.load(sys.stdin).get('video_id','0'))" 2>/dev/null)
UPLOAD_ID=$(echo "$INIT" | python3 -c "import sys,json; print(json.load(sys.stdin).get('upload_id',''))" 2>/dev/null)

if [ "$VIDEO_ID" = "0" ] || [ -z "$UPLOAD_ID" ]; then
  echo -e "  ${RED}✗ 初始化上传失败${NC}"
  exit 1
fi
echo -e "  ${GREEN}✓ init-upload: video_id=$VIDEO_ID, chunks=$TOTAL_CHUNKS${NC}"

UPLOAD_START=$(date +%s%3N)
for i in $(seq 0 $((TOTAL_CHUNKS - 1))); do
  dd if="$VIDEO_FILE" bs=$CHUNK_SIZE skip=$i count=1 of="/tmp/chunk_$i" 2>/dev/null
  CHUNK_SIZE_ACTUAL=$(stat -c%s "/tmp/chunk_$i")

  curl -s -o /dev/null -X POST "$GATEWAY/api/video/upload-chunk" \
    -H "Authorization: Bearer $TOKEN" \
    -F "video_id=$VIDEO_ID" \
    -F "upload_id=$UPLOAD_ID" \
    -F "chunk_index=$i" \
    -F "file_size=$CHUNK_SIZE_ACTUAL" \
    -F "file=@/tmp/chunk_$i" 2>/dev/null

  rm -f "/tmp/chunk_$i"
  printf "\r  进度: %d/%d" $((i+1)) $TOTAL_CHUNKS
done
UPLOAD_END=$(date +%s%3N)
UPLOAD_TIME=$((UPLOAD_END - UPLOAD_START))
echo ""
echo -e "  ${GREEN}✓ 全部 $TOTAL_CHUNKS 个分片上传完成 (${UPLOAD_TIME}ms)${NC}"

# ── 4. 合并视频 ──
echo -e "${YELLOW}[4/6] 合并分片...${NC}"
MERGE=$(curl -s -X POST "$GATEWAY/api/video/merge-chunks" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d "{\"video_id\":$VIDEO_ID,\"upload_id\":\"$UPLOAD_ID\"}" 2>/dev/null)

if echo "$MERGE" | grep -q "merge completed\|complete"; then
  echo -e "  ${GREEN}✓ 合并成功${NC}"
else
  echo -e "  ${YELLOW}⚠ 合并返回: $MERGE${NC}"
fi

# ── 5. HTTP 压测 ──
echo -e "${YELLOW}[5/6] HTTP 压测...${NC}"
echo "  (并发: $BENCH_CONCURRENCY / 请求: $BENCH_REQUESTS)"

go run test/bench/http/http_bench.go -c $BENCH_CONCURRENCY -n $BENCH_REQUESTS -u "$BENCH_URL" 2>&1 | tee "$RESULTS/http_bench.txt"
echo -e "  ${GREEN}✓ 结果已保存: $RESULTS/http_bench.txt${NC}"

# ── 6. 删除测试视频 ──
echo -e "${YELLOW}[6/6] 清理测试数据...${NC}"
curl -s -o /dev/null -X DELETE "$GATEWAY/api/video/delete?video_id=$VIDEO_ID" \
  -H "Authorization: Bearer $TOKEN" 2>/dev/null
echo -e "  ${GREEN}✓ 测试视频已删除 (video_id=$VIDEO_ID)${NC}"

# ── 汇总 ──
echo ""
echo -e "${CYAN}╔══════════════════════════════════════════╗${NC}"
echo -e "${CYAN}║        测试完成                           ║${NC}"
echo -e "${CYAN}╚══════════════════════════════════════════╝${NC}"
echo ""
echo "结果文件: $RESULTS/http_bench.txt"
echo ""

# Jaeger 链路报告
echo -e "${YELLOW}生成 Jaeger 链路延迟报告...${NC}"
bash test/jaeger_trace.sh video-svc 10 2>/dev/null
echo -e "  ${GREEN}✓ Jaeger 报告: $RESULTS/jaeger_trace.md${NC}"

echo ""
echo "上传耗时: ${UPLOAD_TIME}ms (视频 $FILE_SIZE 字节, $TOTAL_CHUNKS 分片)"
echo ""
echo -e "${YELLOW}▶ 建议查看:${NC}"
echo "  Jaeger:      http://localhost:16686"
echo "  Prometheus:  http://localhost:9090/targets"
echo "  Grafana:     http://localhost:3001 (admin/admin)"
