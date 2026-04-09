.PHONY: all build build-web build-server run test clean deps sqlc lint docker-build docker-run

APP_NAME := mengri-flow
BUILD_DIR := ./bin
MAIN_PATH := ./cmd/server

# Go 构建参数
GO := go
GOFLAGS := -v
LDFLAGS := -s -w

all: build

## ===== 完整构建：先编译前端，再编译后端（前端产物内嵌到二进制） =====
build: build-web build-server

## 仅构建前端
build-web:
	@echo ">>> Building frontend..."
	cd web && npm install && npm run build
	@echo ">>> Frontend built to web/dist/"

## 仅构建后端（需要 web/dist 已存在）
build-server:
	@echo ">>> Building backend..."
	$(GO) build $(GOFLAGS) -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/$(APP_NAME) $(MAIN_PATH)
	@echo ">>> Binary output: $(BUILD_DIR)/$(APP_NAME)"

## 运行（开发模式，前端用 vite dev server 代理）
run:
	$(GO) run $(MAIN_PATH)/main.go

## 开发模式：同时启动后端和前端 dev server
dev:
	@echo "Starting backend and frontend dev server..."
	@echo "Backend: http://localhost:8080"
	@echo "Frontend: http://localhost:3000 (proxy /api -> :8080)"
	@$(GO) run $(MAIN_PATH)/main.go &
	cd web && npm run dev

## 测试
test:
	$(GO) test ./... -v -cover

## 代码检查
lint:
	golangci-lint run ./...

## 清理构建产物
clean:
	rm -rf $(BUILD_DIR)
	rm -rf web/dist

## 下载依赖
deps:
	$(GO) mod tidy
	$(GO) mod download
	cd web && npm install

## SQLC 生成代码
sqlc:
	sqlc generate

## 数据库迁移（需要安装 golang-migrate）
migrate-up:
	migrate -path migrations -database "mysql://root:123456@tcp(127.0.0.1:3306)/mengri_flow" up

migrate-down:
	migrate -path migrations -database "mysql://root:123456@tcp(127.0.0.1:3306)/mengri_flow" down

migrate-create:
	@read -p "Enter migration name: " name; \
	migrate create -ext sql -dir migrations -seq $$name

## Swagger 文档生成
swagger:
	swag init -g cmd/server/main.go -o docs

## Docker
## 构建Docker镜像（包含Console和Executor双角色）
docker-build:
	docker build -t $(APP_NAME) .

## 运行Docker容器（默认Console角色）
docker-run:
	docker run -p 8080:8080 --env-file .env $(APP_NAME) --role=console

## 运行Docker容器（Executor角色）
docker-run-executor:
	docker run -d --name $(APP_NAME)-executor \
		--env-file .env \
		$(APP_NAME) \
		--role=executor \
		--etcd-endpoints=etcd:2379 \
		--cluster-id=cluster-prod-001

## 帮助
help:
	@echo "可用命令："
	@echo "  make build         - 完整构建（前端 + 后端，产出单一二进制）"
	@echo "  make build-web     - 仅构建前端到 web/dist/"
	@echo "  make build-server  - 仅构建后端（需 web/dist 已存在）"
	@echo "  make run           - 运行后端"
	@echo "  make dev           - 开发模式（后端 + 前端 dev server）"
	@echo "  make test          - 运行测试"
	@echo "  make lint          - 代码检查"
	@echo "  make clean         - 清理构建产物"
	@echo "  make deps          - 下载所有依赖（Go + npm）"
	@echo "  make sqlc          - 生成 SQLC 代码"
	@echo "  make migrate-up    - 执行数据库迁移"
	@echo "  make migrate-down  - 回滚数据库迁移"
	@echo "  make swagger       - 生成 Swagger 文档"
	@echo "  make docker-build  - 构建 Docker 镜像"
	@echo "  make docker-run    - 运行 Docker 容器"
