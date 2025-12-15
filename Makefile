# Makefile for maven-proxy

# 变量定义
BINARY_NAME=maven-proxy
DOCKER_IMAGE=maven-proxy
DOCKER_TAG=latest
VERSION?=1.0.0

# Go 相关变量
GOOS?=linux
GOARCH?=amd64
CGO_ENABLED?=0

# 构建目录
BUILD_DIR=build
DIST_DIR=dist

.PHONY: help build build-local build-cross docker docker-build clean test run

# 默认目标
help:
	@echo "可用的构建目标:"
	@echo "  build        - 构建二进制文件"
	@echo "  build-local  - 本地构建二进制文件"
	@echo "  build-cross  - 交叉编译二进制文件"
	@echo "  docker       - 构建 Docker 镜像"
	@echo "  docker-build - 构建 Docker 镜像 (别名)"
	@echo "  clean        - 清理构建文件"
	@echo "  test         - 运行测试"
	@echo "  run          - 运行应用程序"

# 构建二进制文件
build:
	@echo "构建二进制文件..."
	@mkdir -p $(BUILD_DIR)
	go mod tidy
	go build -ldflags "-X main.version=$(VERSION)" -o $(BUILD_DIR)/$(BINARY_NAME) cmd/maven-proxy/main.go
	@echo "构建完成: $(BUILD_DIR)/$(BINARY_NAME)"

# 本地构建
build-local:
	@echo "本地构建二进制文件..."
	go mod tidy
	go build -o $(BINARY_NAME) cmd/maven-proxy/main.go
	@echo "构建完成: $(BINARY_NAME)"

# 交叉编译
build-cross:
	@echo "交叉编译二进制文件..."
	@mkdir -p $(DIST_DIR)

	# Linux AMD64
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "-X main.version=$(VERSION)" -o $(DIST_DIR)/$(BINARY_NAME)-linux-amd64 cmd/maven-proxy/main.go

	# Linux ARM64
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -ldflags "-X main.version=$(VERSION)" -o $(DIST_DIR)/$(BINARY_NAME)-linux-arm64 cmd/maven-proxy/main.go

	# Windows AMD64
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -ldflags "-X main.version=$(VERSION)" -o $(DIST_DIR)/$(BINARY_NAME)-windows-amd64.exe cmd/maven-proxy/main.go

	# Darwin AMD64
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -ldflags "-X main.version=$(VERSION)" -o $(DIST_DIR)/$(BINARY_NAME)-darwin-amd64 cmd/maven-proxy/main.go

	@echo "交叉编译完成，文件位于 $(DIST_DIR)/"

# 构建 Docker 镜像
docker: docker-build

docker-build:
	@echo "构建 Docker 镜像..."
	docker build -t $(DOCKER_IMAGE):$(DOCKER_TAG) .
	docker tag $(DOCKER_IMAGE):$(DOCKER_TAG) $(DOCKER_IMAGE):$(VERSION)
	@echo "Docker 镜像构建完成: $(DOCKER_IMAGE):$(DOCKER_TAG)"

# 清理构建文件
clean:
	@echo "清理构建文件..."
	rm -rf $(BUILD_DIR)
	rm -rf $(DIST_DIR)
	rm -f $(BINARY_NAME)
	go clean
	@echo "清理完成"

# 运行测试
test:
	@echo "运行测试..."
	go test -v ./...

# 运行应用程序
run:
	@echo "运行应用程序..."
	go run cmd/maven-proxy/main.go -c config.yaml

# 安装依赖
deps:
	@echo "安装依赖..."
	go mod download
	go mod tidy

# 格式化代码
fmt:
	@echo "格式化代码..."
	go fmt ./...

# 代码检查
lint:
	@echo "运行代码检查..."
	golangci-lint run

# 生成模拟数据
mocks:
	@echo "生成模拟数据..."
	go generate ./...

# 发布准备
release: clean test build-cross docker
	@echo "发布准备完成"
