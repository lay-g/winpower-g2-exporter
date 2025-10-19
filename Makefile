# WinPower G2 Prometheus Exporter Makefile

# 变量定义
BINARY_NAME=winpower-g2-exporter
MIGRATION_TOOL=winpower-config-migrate
VERSION=v0.1.0
BUILD_TIME=$(shell date +%Y-%m-%dT%H:%M:%S%z)
GIT_COMMIT=$(shell git rev-parse --short HEAD)
GO_VERSION=$(shell go version | awk '{print $$3}')

# 构建标志
LDFLAGS=-ldflags "-X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME) -X main.GitCommit=$(GIT_COMMIT) -X main.GoVersion=$(GO_VERSION)"

# 目录
BUILD_DIR=build
DIST_DIR=dist

# Go 相关
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod

.PHONY: help build build-linux build-all build-tools clean test test-coverage test-integration test-all fmt lint deps update-deps dev docker-build docker-run docker-push

# 默认目标
.DEFAULT_GOAL := help

# 帮助信息
help: ## 显示此帮助信息
	@echo 'WinPower G2 Prometheus Exporter Makefile'
	@echo ''
	@echo '可用命令:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-20s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

# 构建命令
build: ## 构建当前平台的二进制文件
	@echo "构建 $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/exporter
	@echo "构建完成: $(BUILD_DIR)/$(BINARY_NAME)"

build-tools: ## 构建工具（包括配置迁移工具）
	@echo "构建工具..."
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(MIGRATION_TOOL) ./cmd/config-migrate
	@echo "迁移工具构建完成: $(BUILD_DIR)/$(MIGRATION_TOOL)"

build-all-tools: ## 构建主程序和所有工具
	@echo "构建主程序和所有工具..."
	$(MAKE) build
	$(MAKE) build-tools
	@echo "所有构建完成"

build-linux: ## 构建 Linux AMD64 二进制文件
	@echo "构建 $(BINARY_NAME) for Linux AMD64..."
	@mkdir -p $(BUILD_DIR)/linux-amd64
	GOOS=linux GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/linux-amd64/$(BINARY_NAME) ./cmd/exporter
	GOOS=linux GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/linux-amd64/$(MIGRATION_TOOL) ./cmd/config-migrate
	@echo "构建完成: $(BUILD_DIR)/linux-amd64/"

build-all: ## 构建所有支持平台的二进制文件
	@echo "构建 $(BINARY_NAME) for multiple platforms..."
	@mkdir -p $(DIST_DIR)

	# Linux AMD64
	GOOS=linux GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(DIST_DIR)/$(BINARY_NAME)-linux-amd64 ./cmd/exporter
	GOOS=linux GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(DIST_DIR)/$(MIGRATION_TOOL)-linux-amd64 ./cmd/config-migrate

	# Linux ARM64
	GOOS=linux GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o $(DIST_DIR)/$(BINARY_NAME)-linux-arm64 ./cmd/exporter
	GOOS=linux GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o $(DIST_DIR)/$(MIGRATION_TOOL)-linux-arm64 ./cmd/config-migrate

	# Darwin AMD64
	GOOS=darwin GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(DIST_DIR)/$(BINARY_NAME)-darwin-amd64 ./cmd/exporter
	GOOS=darwin GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(DIST_DIR)/$(MIGRATION_TOOL)-darwin-amd64 ./cmd/config-migrate

	# Darwin ARM64
	GOOS=darwin GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o $(DIST_DIR)/$(BINARY_NAME)-darwin-arm64 ./cmd/exporter
	GOOS=darwin GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o $(DIST_DIR)/$(MIGRATION_TOOL)-darwin-arm64 ./cmd/config-migrate

	# Windows AMD64
	GOOS=windows GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(DIST_DIR)/$(BINARY_NAME)-windows-amd64.exe ./cmd/exporter
	GOOS=windows GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(DIST_DIR)/$(MIGRATION_TOOL)-windows-amd64.exe ./cmd/config-migrate

	@echo "构建完成，文件位于 $(DIST_DIR)/"
	@ls -la $(DIST_DIR)/

clean: ## 清理构建产物
	@echo "清理构建产物..."
	$(GOCLEAN)
	@rm -rf $(BUILD_DIR)
	@rm -rf $(DIST_DIR)
	@echo "清理完成"

# 测试命令
test: ## 运行单元测试
	@echo "运行单元测试..."
	$(GOTEST) -v ./...

test-coverage: ## 运行测试并生成覆盖率报告
	@echo "运行测试并生成覆盖率报告..."
	$(GOTEST) -v -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "覆盖率报告已生成: coverage.html"

test-integration: ## 运行集成测试
	@echo "运行集成测试..."
	$(GOTEST) -v -tags=integration ./test/integration/...

test-all: ## 运行所有测试
	@echo "运行所有测试..."
	$(MAKE) test
	$(MAKE) test-integration

# 开发命令
fmt: ## 格式化 Go 代码
	@echo "格式化代码..."
	$(GOCMD) fmt ./...

lint: ## 运行代码静态分析
	@echo "运行代码检查..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "golangci-lint 未安装，跳过代码检查"; \
	fi

deps: ## 安装项目依赖
	@echo "安装依赖..."
	$(GOMOD) download
	$(GOMOD) tidy

update-deps: ## 更新依赖
	@echo "更新依赖..."
	$(GOMOD) get -u ./...
	$(GOMOD) tidy

dev: ## 启动开发环境
	@echo "启动开发环境..."
	$(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/exporter
	./$(BUILD_DIR)/$(BINARY_NAME) --log-level=debug

# Docker 命令
docker-build: ## 构建 Docker 镜像
	@echo "构建 Docker 镜像..."
	docker build -t $(BINARY_NAME):$(VERSION) .
	docker tag $(BINARY_NAME):$(VERSION) $(BINARY_NAME):latest
	@echo "Docker 镜像构建完成"

docker-run: ## 运行 Docker 容器
	@echo "运行 Docker 容器..."
	docker run --rm -p 9090:9090 $(BINARY_NAME):latest

docker-push: ## 推送 Docker 镜像到仓库
	@echo "推送 Docker 镜像..."
	@if [ -n "$(REGISTRY)" ]; then \
		docker tag $(BINARY_NAME):$(VERSION) $(REGISTRY)/$(BINARY_NAME):$(VERSION); \
		docker tag $(BINARY_NAME):latest $(REGISTRY)/$(BINARY_NAME):latest; \
		docker push $(REGISTRY)/$(BINARY_NAME):$(VERSION); \
		docker push $(REGISTRY)/$(BINARY_NAME):latest; \
	else \
		echo "请设置 REGISTRY 环境变量"; \
		exit 1; \
	fi

# 发布命令
release: ## 创建发布版本
	@echo "创建发布版本 $(VERSION)..."
	$(MAKE) clean
	$(MAKE) build-all
	@echo "创建发布包..."
	@mkdir -p $(DIST_DIR)/release
	@cd $(DIST_DIR) && for file in $(BINARY_NAME)-*; do \
		tar -czf release/$${file%.exe}.tar.gz $$file; \
	done
	@echo "发布包创建完成: $(DIST_DIR)/release/"

tag: ## 创建 Git 标签
	@echo "创建 Git 标签 $(VERSION)..."
	git tag -a $(VERSION) -m "Release $(VERSION)"
	@echo "标签创建完成，使用 'git push origin $(VERSION)' 推送到远程仓库"

# 安装工具
install-tools: ## 安装开发工具
	@echo "安装开发工具..."
	$(GOGET) -u github.com/golangci/golangci-lint/cmd/golangci-lint
	$(GOGET) -u golang.org/x/tools/cmd/goimports

# 验证
verify: ## 验证代码格式和依赖
	@echo "验证代码格式和依赖..."
	$(GOCMD) mod verify
	@if command -v goimports >/dev/null 2>&1; then \
		goimports -l .; \
	fi
	$(MAKE) fmt
	$(MAKE) lint