.PHONY: all build run stop clean proto api help docker-build docker-up docker-down

# 项目名称
APP_NAME := gopan
BUILD_DIR := build

# 服务列表
SERVICES := gateway user video transcode stream interact search

# 服务端口映射
GATEWAY_PORT    := 8888
USER_PORT       := 8081
VIDEO_PORT      := 8082
TRANSCODE_PORT  := 8083
STREAM_PORT     := 8084
INTERACT_PORT   := 8085
SEARCH_PORT     := 8086

# 默认目标
all: build

# ─── help ────────────────────────────────────────
help:
	@echo "GoPan VOD Platform - Makefile"
	@echo ""
	@echo "  编译:"
	@echo "    make build                编译所有服务到 build/ 目录"
	@echo "    make build-gateway        仅编译 gateway"
	@echo "    make build-user           仅编译 user-svc"
	@echo "    make build-video          仅编译 video-svc"
	@echo "    make build-transcode      仅编译 transcode-svc"
	@echo "    make build-stream         仅编译 stream-svc"
	@echo "    make build-interact       仅编译 interact-svc"
	@echo "    make build-search         仅编译 search-svc"
	@echo ""
	@echo "  运行 (本地开发):"
	@echo "    make run                  后台启动所有服务，日志写入 logs/"
	@echo "    make run-gateway          仅启动 gateway"
	@echo "    make run-user             仅启动 user-svc"
	@echo "    make run-video            仅启动 video-svc"
	@echo "    make run-transcode        仅启动 transcode-svc"
	@echo "    make run-stream           仅启动 stream-svc"
	@echo "    make run-interact         仅启动 interact-svc"
	@echo "    make run-search           仅启动 search-svc"
	@echo ""
	@echo "  停止:"
	@echo "    make stop                 停止所有后台运行的服务"
	@echo ""
	@echo "  查看:"
	@echo "    make status               查看服务运行状态"
	@echo "    make logs                 实时查看全部日志"
	@echo "    make logs-gateway         查看 gateway 日志"
	@echo ""
	@echo "  清理:"
	@echo "    make clean                删除 build/ 和 logs/"
	@echo ""
	@echo "  代码生成:"
	@echo "    make proto                重新生成所有 protobuf 桩代码"
	@echo "    make api                  重新生成 gateway API 桩代码"
	@echo "    make gen                  生成 proto + api"
	@echo ""
	@echo "  Docker:"
	@echo "    make docker-build         构建 Docker 镜像"
	@echo "    make docker-up            启动 docker-compose (所有服务+中间件)"
	@echo "    make docker-down          停止并清理 docker-compose"
	@echo ""
	@echo "  测试:"
	@echo "    make test                 运行所有测试"
	@echo "    make lint                 运行 go vet"
	@echo ""
	@echo "  依赖:"
	@echo "    make deps                 下载并整理依赖"
	@echo ""

# ─── 依赖 ────────────────────────────────────────
deps:
	@echo ">>> 下载依赖..."
	go mod tidy
	go mod download
	@echo ">>> 依赖整理完成"

# ─── 编译 ────────────────────────────────────────
build: deps $(addprefix build-,$(SERVICES))
	@echo ">>> 所有服务编译完成: $(BUILD_DIR)/"

build-gateway:
	@mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/gateway ./gateway/
	@echo "  ✓ gateway"

build-user:
	@mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/user-svc ./rpc/user/
	@echo "  ✓ user-svc"

build-video:
	@mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/video-svc ./rpc/video/
	@echo "  ✓ video-svc"

build-transcode:
	@mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/transcode-svc ./rpc/transcode/
	@echo "  ✓ transcode-svc"

build-stream:
	@mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/stream-svc ./rpc/stream/
	@echo "  ✓ stream-svc"

build-interact:
	@mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/interact-svc ./rpc/interact/
	@echo "  ✓ interact-svc"

build-search:
	@mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/search-svc ./rpc/search/
	@echo "  ✓ search-svc"

# ─── 运行 ────────────────────────────────────────
LOG_DIR := logs
PID_DIR := $(LOG_DIR)/pids

# 本地运行时优先使用 .local.yaml（连接 localhost），否则用 Docker 版 yaml
GW_YAML := $(shell [ -f gateway/etc/gateway.local.yaml ] && echo "gateway.local.yaml" || echo "gateway.yaml")
USR_YAML := $(shell [ -f rpc/user/etc/user.local.yaml ] && echo "user.local.yaml" || echo "user.yaml")
VID_YAML := $(shell [ -f rpc/video/etc/video.local.yaml ] && echo "video.local.yaml" || echo "video.yaml")
TRANS_YAML := $(shell [ -f rpc/transcode/etc/transcode.local.yaml ] && echo "transcode.local.yaml" || echo "transcode.yaml")
STRM_YAML := $(shell [ -f rpc/stream/etc/stream.local.yaml ] && echo "stream.local.yaml" || echo "stream.yaml")
INT_YAML := $(shell [ -f rpc/interact/etc/interact.local.yaml ] && echo "interact.local.yaml" || echo "interact.yaml")
SRCH_YAML := $(shell [ -f rpc/search/etc/search.local.yaml ] && echo "search.local.yaml" || echo "search.yaml")

run: stop build
	@mkdir -p $(LOG_DIR) $(PID_DIR) logs/gateway logs/user logs/video logs/transcode logs/stream logs/interact logs/search
	@echo ">>> 启动所有服务 (使用 *.local.yaml)..."

	@nohup $(BUILD_DIR)/user-svc -f rpc/user/etc/$(USR_YAML) > logs/user/user.log 2>&1 & echo $$! > $(PID_DIR)/user.pid
	@sleep 1
	@nohup $(BUILD_DIR)/video-svc -f rpc/video/etc/$(VID_YAML) > logs/video/video.log 2>&1 & echo $$! > $(PID_DIR)/video.pid
	@sleep 1
	@nohup $(BUILD_DIR)/transcode-svc -f rpc/transcode/etc/$(TRANS_YAML) > logs/transcode/transcode.log 2>&1 & echo $$! > $(PID_DIR)/transcode.pid
	@sleep 1
	@nohup $(BUILD_DIR)/stream-svc -f rpc/stream/etc/$(STRM_YAML) > logs/stream/stream.log 2>&1 & echo $$! > $(PID_DIR)/stream.pid
	@sleep 1
	@nohup $(BUILD_DIR)/interact-svc -f rpc/interact/etc/$(INT_YAML) > logs/interact/interact.log 2>&1 & echo $$! > $(PID_DIR)/interact.pid
	@sleep 1
	@nohup $(BUILD_DIR)/search-svc -f rpc/search/etc/$(SRCH_YAML) > logs/search/search.log 2>&1 & echo $$! > $(PID_DIR)/search.pid
	@sleep 2
	@nohup $(BUILD_DIR)/gateway -f gateway/etc/$(GW_YAML) > logs/gateway/gateway.log 2>&1 & echo $$! > $(PID_DIR)/gateway.pid

	@echo ">>> 全部启动完成"
	@make status

run-gateway:
	@mkdir -p $(LOG_DIR) $(PID_DIR)
	@nohup $(BUILD_DIR)/gateway -f gateway/etc/$(GW_YAML) > logs/gateway.log 2>&1 & echo $$! > $(PID_DIR)/gateway.pid
	@echo "  gateway → :$(GATEWAY_PORT)"

run-user:
	@mkdir -p $(LOG_DIR) $(PID_DIR)
	@nohup $(BUILD_DIR)/user-svc -f rpc/user/etc/$(USR_YAML) > logs/user.log 2>&1 & echo $$! > $(PID_DIR)/user.pid
	@echo "  user-svc → :$(USER_PORT)"

run-video:
	@mkdir -p $(LOG_DIR) $(PID_DIR)
	@nohup $(BUILD_DIR)/video-svc -f rpc/video/etc/$(VID_YAML) > logs/video.log 2>&1 & echo $$! > $(PID_DIR)/video.pid
	@echo "  video-svc → :$(VIDEO_PORT)"

run-transcode:
	@mkdir -p $(LOG_DIR) $(PID_DIR)
	@nohup $(BUILD_DIR)/transcode-svc -f rpc/transcode/etc/$(TRANS_YAML) > logs/transcode.log 2>&1 & echo $$! > $(PID_DIR)/transcode.pid
	@echo "  transcode-svc → :$(TRANSCODE_PORT)"

run-stream:
	@mkdir -p $(LOG_DIR) $(PID_DIR)
	@nohup $(BUILD_DIR)/stream-svc -f rpc/stream/etc/$(STRM_YAML) > logs/stream.log 2>&1 & echo $$! > $(PID_DIR)/stream.pid
	@echo "  stream-svc → :$(STREAM_PORT)"

run-interact:
	@mkdir -p $(LOG_DIR) $(PID_DIR)
	@nohup $(BUILD_DIR)/interact-svc -f rpc/interact/etc/$(INT_YAML) > logs/interact.log 2>&1 & echo $$! > $(PID_DIR)/interact.pid
	@echo "  interact-svc → :$(INTERACT_PORT)"

run-search:
	@mkdir -p $(LOG_DIR) $(PID_DIR)
	@nohup $(BUILD_DIR)/search-svc -f rpc/search/etc/$(SRCH_YAML) > logs/search.log 2>&1 & echo $$! > $(PID_DIR)/search.pid
	@echo "  search-svc → :$(SEARCH_PORT)"

# ─── 停止 ────────────────────────────────────────
stop:
	@echo ">>> 停止所有服务..."
	@for svc in $(SERVICES); do \
		if [ -f $(PID_DIR)/$$svc.pid ]; then \
			pid=$$(cat $(PID_DIR)/$$svc.pid); \
			if kill -0 $$pid 2>/dev/null; then \
				kill $$pid 2>/dev/null && echo "  $$svc (pid=$$pid) 已停止" || true; \
			fi; \
			rm -f $(PID_DIR)/$$svc.pid; \
		fi; \
	done
	@rm -rf $(PID_DIR)
	@echo ">>> 全部停止"

# ─── 状态 & 日志 ────────────────────────────────────
status:
	@echo "================================"
	@echo "  服务运行状态"
	@echo "================================"
	@for svc in $(SERVICES); do \
		pid_file=$(PID_DIR)/$$svc.pid; \
		if [ -f $$pid_file ]; then \
			pid=$$(cat $$pid_file); \
			if kill -0 $$pid 2>/dev/null; then \
				echo "  ✓ $$svc  → 运行中 (pid=$$pid)"; \
			else \
				echo "  ✗ $$svc  → 已退出 (pid=$$pid)"; \
			fi; \
		else \
			echo "  - $$svc  → 未启动"; \
		fi; \
	done
	@echo "================================"

logs:
	@tail -f logs/gateway/gateway.log \
		logs/user/user.log \
		logs/video/video.log \
		logs/transcode/transcode.log \
		logs/stream/stream.log \
		logs/interact/interact.log \
		logs/search/search.log 2>/dev/null

logs-gateway:
	@tail -f logs/gateway/gateway.log
logs-user:
	@tail -f logs/user/user.log
logs-video:
	@tail -f logs/video/video.log
logs-transcode:
	@tail -f logs/transcode/transcode.log
logs-stream:
	@tail -f logs/stream/stream.log
logs-interact:
	@tail -f logs/interact/interact.log
logs-search:
	@tail -f logs/search/search.log

# ─── 清理 ────────────────────────────────────────
clean:
	@echo ">>> 清理编译产物和日志..."
	rm -rf $(BUILD_DIR)
	rm -rf $(LOG_DIR)
	rm -rf logs/
	rm -rf go-build-cache/
	@echo ">>> 清理完成"

# ─── 代码生成 ────────────────────────────────────────
GOCTL := $(shell which goctl 2>/dev/null || echo $(HOME)/go/bin/goctl)

proto: guard-goctl
	@echo ">>> 生成 protobuf 桩代码..."
	cd rpc/user        && $(GOCTL) rpc protoc user.proto     --go_out=. --go-grpc_out=. --zrpc_out=. 2>&1 | tail -1
	cd rpc/video       && $(GOCTL) rpc protoc video.proto    --go_out=. --go-grpc_out=. --zrpc_out=. 2>&1 | tail -1
	cd rpc/transcode   && $(GOCTL) rpc protoc transcode.proto --go_out=. --go-grpc_out=. --zrpc_out=. 2>&1 | tail -1
	cd rpc/stream      && $(GOCTL) rpc protoc stream.proto   --go_out=. --go-grpc_out=. --zrpc_out=. 2>&1 | tail -1
	cd rpc/interact    && $(GOCTL) rpc protoc interact.proto --go_out=. --go-grpc_out=. --zrpc_out=. 2>&1 | tail -1
	cd rpc/search      && $(GOCTL) rpc protoc search.proto   --go_out=. --go-grpc_out=. --zrpc_out=. 2>&1 | tail -1
	@echo ">>> protobuf 生成完成"

api: guard-goctl
	@echo ">>> 生成 API 网关桩代码..."
	$(GOCTL) api go -api api/gateway.api -dir gateway 2>&1
	@echo ">>> API 网关生成完成"

gen: proto api
	@echo ">>> 代码生成全部完成"

guard-goctl:
	@if [ ! -x "$(GOCTL)" ]; then \
		echo ">>> goctl not found, installing..."; \
		go install github.com/zeromicro/go-zero/tools/goctl@latest; \
	fi

# ─── Docker ──────────────────────────────────────
docker-build:
	@echo ">>> 构建 Docker 镜像..."
	docker compose build

docker-up:
	@echo ">>> 启动 docker-compose..."
	docker compose up -d
	@echo ">>> 等待服务就绪..."
	@sleep 10
	docker compose ps

docker-down:
	@echo ">>> 停止 docker-compose..."
	docker compose down

# ─── 测试 & 检查 ──────────────────────────────────
test:
	@echo ">>> 运行测试..."
	go test ./... -count=1 -timeout 30s

lint:
	@echo ">>> 静态检查..."
	go vet ./...

fmt:
	@echo ">>> 格式化代码..."
	gofmt -s -w .

# ─── 快捷命令 ────────────────────────────────────
dev: stop build run
	@echo ">>> 开发模式启动完成"
