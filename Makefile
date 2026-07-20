.PHONY: all build build-master build-agent build-web clean dev run test proto

# 默认目标
all: build

# 构建全部
build: build-web build-master build-agent

# 构建主控
build-master:
	@echo "Building master..."
	CGO_ENABLED=0 go build -ldflags="-s -w" -o bin/xray-master ./cmd/master

# 构建 Agent
build-agent:
	@echo "Building agent..."
	CGO_ENABLED=0 go build -ldflags="-s -w" -o bin/xray-agent ./cmd/agent

# 构建前端
build-web:
	@echo "Building frontend..."
	cd web && npm install && npm run build

# 开发模式 - 启动基础设施
dev-infra:
	docker compose up -d postgres redis

# 开发模式 - 启动后端（需要先 dev-infra）
dev-master:
	@echo "Starting master in dev mode..."
	DB_HOST=localhost go run ./cmd/master -c config.dev.json

# 开发模式 - 启动前端
dev-web:
	@echo "Starting frontend dev server..."
	cd web && npm run dev

# 运行测试
test:
	go test ./...

# 生成 protobuf 代码
proto:
	protoc --go_out=. --go-grpc_out=. pkg/proto/agent.proto

# Docker 构建
docker-build:
	docker compose build

# Docker 启动
docker-up:
	docker compose up -d

# Docker 停止
docker-down:
	docker compose down

# 清理
clean:
	rm -rf bin/
	rm -rf web/dist/

# 交叉编译（Linux amd64/arm64）
release:
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="-s -w" -o bin/xray-master-linux-amd64 ./cmd/master
	GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -ldflags="-s -w" -o bin/xray-master-linux-arm64 ./cmd/master
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="-s -w" -o bin/xray-agent-linux-amd64 ./cmd/agent
	GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -ldflags="-s -w" -o bin/xray-agent-linux-arm64 ./cmd/agent
	@echo "Release binaries built in bin/"
