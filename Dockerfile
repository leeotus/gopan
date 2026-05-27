# ---- Build Stage ----
FROM docker.m.daocloud.io/library/golang:1.26-alpine AS builder

RUN apk add --no-cache git protobuf-dev gcc musl-dev
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# 编译所有服务
RUN CGO_ENABLED=0 GOOS=linux go build -o /build/gateway ./gateway/
RUN CGO_ENABLED=0 GOOS=linux go build -o /build/user-svc ./rpc/user/
RUN CGO_ENABLED=0 GOOS=linux go build -o /build/video-svc ./rpc/video/
RUN CGO_ENABLED=0 GOOS=linux go build -o /build/transcode-svc ./rpc/transcode/
RUN CGO_ENABLED=0 GOOS=linux go build -o /build/stream-svc ./rpc/stream/
RUN CGO_ENABLED=0 GOOS=linux go build -o /build/interact-svc ./rpc/interact/
RUN CGO_ENABLED=0 GOOS=linux go build -o /build/search-svc ./rpc/search/
RUN CGO_ENABLED=0 GOOS=linux go build -o /build/admin-svc ./rpc/admin/

# ---- Runtime Stage ----
FROM docker.m.daocloud.io/library/alpine:3.21

RUN apk add --no-cache ca-certificates tzdata ffmpeg
ENV TZ=Asia/Shanghai

# 先创建所有目录，再分别 COPY 二进制 + 配置
RUN mkdir -p /app/gateway/etc /app/rpc/user/etc /app/rpc/video/etc \
             /app/rpc/transcode/etc /app/rpc/stream/etc \
             /app/rpc/interact/etc /app/rpc/search/etc /app/rpc/admin/etc

# gateway
COPY --from=builder /build/gateway /app/gateway/gateway
COPY gateway/etc/gateway.yaml /app/gateway/etc/gateway.yaml

# user-svc
COPY --from=builder /build/user-svc /app/rpc/user/user-svc
COPY rpc/user/etc/user.yaml /app/rpc/user/etc/user.yaml

# video-svc
COPY --from=builder /build/video-svc /app/rpc/video/video-svc
COPY rpc/video/etc/video.yaml /app/rpc/video/etc/video.yaml

# transcode-svc
COPY --from=builder /build/transcode-svc /app/rpc/transcode/transcode-svc
COPY rpc/transcode/etc/transcode.yaml /app/rpc/transcode/etc/transcode.yaml

# stream-svc
COPY --from=builder /build/stream-svc /app/rpc/stream/stream-svc
COPY rpc/stream/etc/stream.yaml /app/rpc/stream/etc/stream.yaml

# interact-svc
COPY --from=builder /build/interact-svc /app/rpc/interact/interact-svc
COPY rpc/interact/etc/interact.yaml /app/rpc/interact/etc/interact.yaml

# search-svc
COPY --from=builder /build/search-svc /app/rpc/search/search-svc
COPY rpc/search/etc/search.yaml /app/rpc/search/etc/search.yaml

# admin-svc
COPY --from=builder /build/admin-svc /app/rpc/admin/admin-svc
COPY rpc/admin/etc/admin.yaml /app/rpc/admin/etc/admin.yaml

WORKDIR /app
