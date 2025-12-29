.PHONY: build clean install run-init run-deploy help

BINARY_NAME=talos-deployer
GO=go

help:
	@echo "Talos Proxmox Deployer - Makefile"
	@echo ""
	@echo "可用命令:"
	@echo "  make build       - 编译项目"
	@echo "  make install     - 安装到 /usr/local/bin"
	@echo "  make clean       - 清理编译文件"
	@echo "  make run-init    - 运行配置向导"
	@echo "  make run-deploy  - 运行部署"
	@echo "  make test        - 运行测试"

build:
	@echo "编译 $(BINARY_NAME)..."
	$(GO) build -o $(BINARY_NAME) .
	@echo "✓ 编译完成: $(BINARY_NAME)"

install: build
	@echo "安装到 /usr/local/bin..."
	sudo cp $(BINARY_NAME) /usr/local/bin/
	@echo "✓ 安装完成"

clean:
	@echo "清理..."
	rm -f $(BINARY_NAME)
	rm -rf *-config/
	rm -f *.qcow2 *.raw *.xz
	@echo "✓ 清理完成"

run-init: build
	./$(BINARY_NAME) init

run-deploy: build
	./$(BINARY_NAME) deploy

test:
	$(GO) test -v ./...

deps:
	$(GO) mod download
	$(GO) mod tidy
