#!/bin/bash
# 给所有 ES 中无 video_vector 的视频补向量。
# 用法：bash reindex_vectors.sh
set -euo pipefail

ES="http://localhost:9200"
AI="http://localhost:9900"
INDEX="gopan_videos"

echo "[1/3] 查询所有缺 video_vector 的文档..."
docs=$(curl -s "$ES/$INDEX/_search" -H 'Content-Type: application/json' -d '{
  "_source": ["video_id","title","description"],
  "query": { "bool": { "must_not": { "exists": { "field": "video_vector" } } } },
  "size": 1000
}' | jq -c '.hits.hits[] | {id: ._id, title: ._source.title, description: ._source.description}')

if [ -z "$docs" ]; then
  echo "✅ 全部文档都已有向量，无需回填"
  exit 0
fi

echo "$docs" | while IFS= read -r doc; do
  id=$(echo "$doc"   | jq -r '.id')
  title=$(echo "$doc" | jq -r '.title')
  desc=$(echo "$doc"  | jq -r '.description // ""')
  text="${title} ${desc}"

  echo "[2] 视频 $id 文本=「$text」"

  vec=$(curl -s -X POST "$AI/embed/text" \
    -H 'Content-Type: application/json' \
    -d "$(jq -nc --arg t "$text" '{text:$t}')" | jq -c '.vector')

  if [ "$vec" = "null" ] || [ -z "$vec" ]; then
    echo "  ❌ AI 返回空向量，跳过"
    continue
  fi

  curl -s -X POST "$ES/$INDEX/_update/$id" \
    -H 'Content-Type: application/json' \
    -d "{\"doc\":{\"video_vector\":$vec}}" | jq -r '.result' | xargs -I{} echo "  ✅ ES update: {}"
done

echo "[3/3] 完成。校验："
curl -s "$ES/$INDEX/_search?pretty" -H 'Content-Type: application/json' -d '{
  "_source": ["video_id","title"],
  "query": { "exists": { "field": "video_vector" } }
}' | jq '.hits.total.value'
