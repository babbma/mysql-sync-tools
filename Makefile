.PHONY: build run clean test fmt help

# 变量定义
BINARY_NAME=db-sync
BINARY_UNIX=$(BINARY_NAME)-unix
BINARY_WINDOWS=$(BINARY_NAME).exe
MAIN_PATH=cmd/db-sync/main.go

# 默认目标
all: build

# 编译
build:
	@echo "开始编译..."
	go build -o $(BINARY_NAME) $(MAIN_PATH)
	@echo "编译完成: $(BINARY_NAME)"

# 编译 Linux 版本
build-linux:
	@echo "编译 Linux 版本..."
	GOOS=linux GOARCH=amd64 go build -o $(BINARY_UNIX) $(MAIN_PATH)
	@echo "编译完成: $(BINARY_UNIX)"

# 编译 Windows 版本
build-windows:
	@echo "编译 Windows 版本..."
	GOOS=windows GOARCH=amd64 go build -o $(BINARY_WINDOWS) $(MAIN_PATH)
	@echo "编译完成: $(BINARY_WINDOWS)"

# 编译所有平台
build-all: build-linux build-windows
	@echo "所有平台编译完成"

# 运行
run: build
	@echo "运行程序..."
	./$(BINARY_NAME) -config config.yaml

# 运行（不编译）
run-dev:
	@echo "开发模式运行..."
	go run $(MAIN_PATH) -config config.yaml

# 清理
clean:
	@echo "清理编译文件..."
	go clean
	rm -f $(BINARY_NAME) $(BINARY_UNIX) $(BINARY_WINDOWS)
	rm -f *.log
	@echo "清理完成"

# 测试
test:
	@echo "运行测试..."
	go test -v ./...

# 测试覆盖率
test-coverage:
	@echo "运行测试并生成覆盖率报告..."
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out

# 代码格式化
fmt:
	@echo "格式化代码..."
	go fmt ./...
	@echo "格式化完成"

# 代码检查
lint:
	@echo "检查代码..."
	golangci-lint run ./...

# 下载依赖
deps:
	@echo "下载依赖..."
	go mod download
	@echo "依赖下载完成"

# 更新依赖
deps-update:
	@echo "更新依赖..."
	go get -u ./...
	go mod tidy
	@echo "依赖更新完成"

# 安装
install: build
	@echo "安装到 GOPATH..."
	go install $(MAIN_PATH)
	@echo "安装完成"

# 版本信息
version:
	@./$(BINARY_NAME) -version

# 帮助
help:
	@echo "可用的 make 命令:"
	@echo "  make build         - 编译程序"
	@echo "  make build-linux   - 编译 Linux 版本"
	@echo "  make build-windows - 编译 Windows 版本"
	@echo "  make build-all     - 编译所有平台版本"
	@echo "  make run           - 编译并运行"
	@echo "  make run-dev       - 开发模式运行（不编译）"
	@echo "  make clean         - 清理编译文件"
	@echo "  make test          - 运行测试"
	@echo "  make test-coverage - 运行测试并生成覆盖率报告"
	@echo "  make fmt           - 格式化代码"
	@echo "  make lint          - 代码检查"
	@echo "  make deps          - 下载依赖"
	@echo "  make deps-update   - 更新依赖"
	@echo "  make install       - 安装到 GOPATH"
	@echo "  make version       - 显示版本信息"
	@echo "  make help          - 显示帮助信息"

