#! /bin/bash
# 登录
TOKEN=$(curl -s -X POST http://localhost:8888/api/user/login -H 'Content-Type: application/json' -d '{"username":"flareon","password":"123456"}' | python3 -c 'import sys,json; print(json.load(sys.stdin)["token"])')

# init
INIT=$(curl -s -X POST http://localhost:8888/api/video/init-upload -H "Content-Type: application/json" -H "Authorization: Bearer $TOKEN" -d '{"filename":"t.mp4","title":"t","file_size":15360,"total_chunks":3}')
VID=$(echo $INIT | python3 -c 'import sys,json; print(json.load(sys.stdin)["video_id"])')
UPL=$(echo $INIT | python3 -c 'import sys,json; print(json.load(sys.stdin)["upload_id"])')

# 3 个 chunk 一个一个发
for i in 0 1 2; do
  dd if=/dev/urandom bs=5120 count=1 of=/tmp/c$i 2>/dev/null
  echo "chunk $i:"
  curl -s -w "\nHTTP %{http_code}\n" -X POST http://localhost:8888/api/video/upload-chunk -H "Authorization: Bearer $TOKEN" -F "video_id=$VID" -F "upload_id=$UPL" -F "chunk_index=$i" -F "file_size=5120" -F "file=@/tmp/c$i"
done

# 查 Redis + MinIO
echo "Redis:"
docker compose exec redis redis-cli SMEMBERS "upload:${UPL}:received"
echo "MinIO parts:"
docker compose exec minio ls -R /data/gopan-videos/parts/$VID/ 2>/dev/null || echo "no parts dir"