# WinPower G2 Prometheus Exporter Makefile

# 变量定义
BINARY_NAME=winpower-g2-exporter
MIGRATION_TOOL=winpower-config-migrate
VERSION=$(shell cat VERSION 2>/dev/null || echo "dev")
DISPLAY_VERSION=v$(shell cat VERSION 2>/dev/null || echo "dev")
BUILD_TIME=$(shell date -u '+%Y-%m-%dT%H:%M:%SZ')
GIT_COMMIT=$(shell git rev-parse HEAD 2>/dev/null || echo "unknown")
GO_VERSION=$(shell go version | awk '{print $$3}')

# 构建标志
LDFLAGS=-ldflags "-X main.version=$(VERSION) -X main.buildTime=$(BUILD_TIME) -X main.commitID=$(GIT_COMMIT)"

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

.PHONY: help build build-linux build-all build-tools clean test test-coverage test-integration test-all fmt lint deps update-deps dev docker-build docker-clean

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
	$(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/winpower-g2-exporter
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
	GOOS=linux GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/linux-amd64/$(BINARY_NAME) ./cmd/winpower-g2-exporter
	@echo "构建完成: $(BUILD_DIR)/linux-amd64/"

build-all: ## 构建所有支持平台的二进制文件
	@echo "构建 $(BINARY_NAME) for multiple platforms..."
	@mkdir -p $(DIST_DIR)

	# Linux AMD64
	GOOS=linux GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(DIST_DIR)/$(BINARY_NAME)-linux-amd64 ./cmd/winpower-g2-exporter

	# Linux ARM64
	GOOS=linux GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o $(DIST_DIR)/$(BINARY_NAME)-linux-arm64 ./cmd/winpower-g2-exporter

	# Darwin AMD64
	GOOS=darwin GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(DIST_DIR)/$(BINARY_NAME)-darwin-amd64 ./cmd/winpower-g2-exporter

	# Darwin ARM64
	GOOS=darwin GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o $(DIST_DIR)/$(BINARY_NAME)-darwin-arm64 ./cmd/winpower-g2-exporter

	# Windows AMD64
	GOOS=windows GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(DIST_DIR)/$(BINARY_NAME)-windows-amd64.exe ./cmd/winpower-g2-exporter

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
	GO_TEST=1 $(GOTEST) -v ./...

test-coverage: ## 运行测试并生成覆盖率报告
	@echo "运行测试并生成覆盖率报告..."
	GO_TEST=1 $(GOTEST) -v -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "覆盖率报告已生成: coverage.html"

test-integration: ## 运行集成测试
	@echo "运行集成测试..."
	GO_TEST=1 $(GOTEST) -v -tags=integration ./test/integration/...

test-all: ## 运行所有测试
	@echo "运行所有测试..."
	$(MAKE) test
	$(MAKE) test-integration

test-quiet: ## 运行静默测试（仅显示错误）
	@echo "运行静默测试..."
	GO_TEST=1 LOG_LEVEL=error $(GOTEST) ./...

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
# GitHub Container Registry 配置
GITHUB_REGISTRY=ghcr.io
GITHUB_USER=lay-g
IMAGE_NAME=$(GITHUB_REGISTRY)/$(GITHUB_USER)/$(BINARY_NAME)

docker-build: ## 构建 Docker 镜像
	@echo "构建 Docker 镜像 $(DISPLAY_VERSION)..."
	docker build -t $(IMAGE_NAME):$(VERSION) \
		--build-arg VERSION=$(VERSION) \
		--build-arg BUILD_TIME=$(BUILD_TIME) \
		--build-arg GIT_COMMIT=$(GIT_COMMIT) \
		.
	docker tag $(IMAGE_NAME):$(VERSION) $(IMAGE_NAME):latest
	@echo "Docker 镜像构建完成:"
	@echo "  - $(IMAGE_NAME):$(VERSION)"
	@echo "  - $(IMAGE_NAME):latest"

docker-clean: ## 清理本地 Docker 镜像
	@echo "清理本地 Docker 镜像..."
	-docker rmi $(IMAGE_NAME):$(VERSION) 2>/dev/null || true
	-docker rmi $(IMAGE_NAME):latest 2>/dev/null || true
	@echo "清理完成"

# 发布命令
release: ## 创建发布版本
	@echo "创建发布版本 $(VERSION)..."
	$(MAKE) clean
	$(MAKE) build-all
	@echo "创建发布包..."
	@mkdir -p $(DIST_DIR)/release
	@mkdir -p $(DIST_DIR)/tmp
	@echo "打包发布文件..."
	@for file in $(DIST_DIR)/$(BINARY_NAME)-*; do \
		platform=$$(basename $$file | sed 's/$(BINARY_NAME)-//'); \
		echo "  打包 $$platform ..."; \
		tmpdir=$(DIST_DIR)/tmp/$(BINARY_NAME)-$$platform; \
		mkdir -p $$tmpdir; \
		cp $$file $$tmpdir/; \
		cp LICENSE $$tmpdir/ 2>/dev/null || echo "警告: LICENSE 文件不存在"; \
		cp README.md $$tmpdir/ 2>/dev/null || echo "警告: README.md 文件不存在"; \
		cp config/config.example.yaml $$tmpdir/ 2>/dev/null || echo "警告: config.example.yaml 文件不存在"; \
		cd $(DIST_DIR)/tmp && tar -czf ../release/$(BINARY_NAME)-$$platform.tar.gz $(BINARY_NAME)-$$platform/; \
		cd ../..; \
	done
	@rm -rf $(DIST_DIR)/tmp
	@echo "发布包创建完成: $(DIST_DIR)/release/"
	@ls -lh $(DIST_DIR)/release/

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