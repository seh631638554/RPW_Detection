.PHONY: help build run test clean docker-build docker-run docker-stop install-deps lint format

# 默认目标
help:
	@echo "病虫害检测服务器 - 可用命令:"
	@echo "  build         - 构建Go应用"
	@echo "  run           - 运行Go应用"
	@echo "  test          - 运行测试"
	@echo "  clean         - 清理构建文件"
	@echo "  install-deps  - 安装Go依赖"
	@echo "  lint          - 代码检查"
	@echo "  format        - 代码格式化"
	@echo "  docker-build  - 构建Docker镜像"
	@echo "  docker-run    - 启动Docker服务"
	@echo "  docker-stop   - 停止Docker服务"
	@echo "  docker-logs   - 查看Docker日志"

# 构建Go应用
build:
	@echo "构建Go应用..."
	cd Http && go build -o ../bin/pest-detection-server .

# 运行Go应用
run:
	@echo "运行Go应用..."
	cd Http && go run .

# 运行测试
test:
	@echo "运行测试..."
	cd Http && go test -v ./...

# 清理构建文件
clean:
	@echo "清理构建文件..."
	rm -rf bin/
	rm -rf logs/
	rm -rf uploads/

# 安装Go依赖
install-deps:
	@echo "安装Go依赖..."
	go mod tidy
	go mod download

# 代码检查
lint:
	@echo "代码检查..."
	cd Http && golangci-lint run

# 代码格式化
format:
	@echo "代码格式化..."
	cd Http && go fmt ./...
	cd Http && go vet ./...

# 构建Docker镜像
docker-build:
	@echo "构建Docker镜像..."
	docker build -t pest-detection-server .

# 启动Docker服务
docker-run:
	@echo "启动Docker服务..."
	docker-compose up -d

# 停止Docker服务
docker-stop:
	@echo "停止Docker服务..."
	docker-compose down

# 查看Docker日志
docker-logs:
	@echo "查看Docker日志..."
	docker-compose logs -f

# 重启Docker服务
docker-restart:
	@echo "重启Docker服务..."
	docker-compose restart

# 查看服务状态
docker-status:
	@echo "查看服务状态..."
	docker-compose ps

# 进入应用容器
docker-shell:
	@echo "进入应用容器..."
	docker exec -it pest-detection-server /bin/sh

# 查看应用日志
docker-app-logs:
	@echo "查看应用日志..."
	docker logs -f pest-detection-server

# 数据库备份
db-backup:
	@echo "备份数据库..."
	docker exec pest-detection-mysql mysqldump -u root -ppassword123 pest_detection > backup_$(shell date +%Y%m%d_%H%M%S).sql

# 数据库恢复
db-restore:
	@echo "恢复数据库..."
	@read -p "请输入备份文件名: " backup_file; \
	docker exec -i pest-detection-mysql mysql -u root -ppassword123 pest_detection < $$backup_file

# 性能测试
benchmark:
	@echo "运行性能测试..."
	cd Http && go test -bench=. -benchmem

# 代码覆盖率测试
coverage:
	@echo "运行代码覆盖率测试..."
	cd Http && go test -coverprofile=coverage.out ./...
	cd Http && go tool cover -html=coverage.out

# 生成API文档
docs:
	@echo "生成API文档..."
	@echo "请使用swag或其他工具生成API文档"

# 部署到生产环境
deploy-prod:
	@echo "部署到生产环境..."
	@echo "请根据实际生产环境配置进行部署"

# 开发环境设置
dev-setup:
	@echo "设置开发环境..."
	@echo "1. 安装Go 1.21+"
	@echo "2. 安装Docker和Docker Compose"
	@echo "3. 安装golangci-lint (可选)"
	@echo "4. 配置环境变量"
	@echo "5. 运行 'make install-deps'"
	@echo "6. 运行 'make docker-run'"

# 生产环境设置
prod-setup:
	@echo "设置生产环境..."
	@echo "1. 配置生产环境变量"
	@echo "2. 配置SSL证书"
	@echo "3. 配置反向代理"
	@echo "4. 配置监控和日志"
	@echo "5. 运行 'make docker-run'"




