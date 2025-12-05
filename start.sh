#!/bin/bash

# 病虫害检测服务器启动脚本
# 使用方法: ./start.sh [dev|prod|docker]

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 打印带颜色的消息
print_message() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

print_header() {
    echo -e "${BLUE}================================${NC}"
    echo -e "${BLUE}  病虫害检测服务器启动脚本${NC}"
    echo -e "${BLUE}================================${NC}"
}

# 检查Go是否安装
check_go() {
    if ! command -v go &> /dev/null; then
        print_error "Go未安装，请先安装Go 1.21+"
        exit 1
    fi
    
    GO_VERSION=$(go version | awk '{print $3}' | sed 's/go//')
    print_message "检测到Go版本: $GO_VERSION"
}

# 检查Docker是否安装
check_docker() {
    if ! command -v docker &> /dev/null; then
        print_error "Docker未安装，请先安装Docker"
        exit 1
    fi
    
    if ! command -v docker-compose &> /dev/null; then
        print_error "Docker Compose未安装，请先安装Docker Compose"
        exit 1
    fi
    
    print_message "Docker和Docker Compose已安装"
}

# 开发环境启动
start_dev() {
    print_message "启动开发环境..."
    
    # 检查环境变量文件
    if [ ! -f .env ]; then
        print_warning "未找到.env文件，将使用默认配置"
        print_message "建议复制env.example为.env并修改配置"
    fi
    
    # 安装依赖
    print_message "安装Go依赖..."
    go mod tidy
    go mod download
    
    # 启动服务器
    print_message "启动HTTP服务器..."
    cd Http
    go run .
}

# 生产环境启动
start_prod() {
    print_message "启动生产环境..."
    
    # 检查环境变量文件
    if [ ! -f .env ]; then
        print_error "生产环境需要.env配置文件"
        exit 1
    fi
    
    # 构建应用
    print_message "构建Go应用..."
    make build
    
    # 启动服务器
    print_message "启动HTTP服务器..."
    ./bin/pest-detection-server
}

# Docker环境启动
start_docker() {
    print_message "启动Docker环境..."
    
    # 检查Docker服务状态
    if ! docker info &> /dev/null; then
        print_error "Docker服务未运行，请先启动Docker"
        exit 1
    fi
    
    # 构建并启动服务
    print_message "构建Docker镜像..."
    make docker-build
    
    print_message "启动Docker服务..."
    make docker-run
    
    print_message "等待服务启动..."
    sleep 10
    
    # 检查服务状态
    print_message "检查服务状态..."
    make docker-status
    
    print_message "Docker服务启动完成！"
    print_message "应用地址: http://localhost:8080"
    print_message "数据库管理: http://localhost:8081"
    print_message "Redis管理: http://localhost:8082"
}

# 停止Docker服务
stop_docker() {
    print_message "停止Docker服务..."
    make docker-stop
    print_message "Docker服务已停止"
}

# 查看日志
show_logs() {
    print_message "显示Docker服务日志..."
    make docker-logs
}

# 主函数
main() {
    print_header
    
    case "${1:-dev}" in
        "dev")
            check_go
            start_dev
            ;;
        "prod")
            check_go
            start_prod
            ;;
        "docker")
            check_docker
            start_docker
            ;;
        "stop")
            check_docker
            stop_docker
            ;;
        "logs")
            check_docker
            show_logs
            ;;
        "help"|"-h"|"--help")
            echo "使用方法: $0 [dev|prod|docker|stop|logs|help]"
            echo ""
            echo "参数说明:"
            echo "  dev     - 启动开发环境 (默认)"
            echo "  prod    - 启动生产环境"
            echo "  docker  - 启动Docker环境"
            echo "  stop    - 停止Docker服务"
            echo "  logs    - 查看Docker日志"
            echo "  help    - 显示帮助信息"
            ;;
        *)
            print_error "未知参数: $1"
            echo "使用 '$0 help' 查看帮助信息"
            exit 1
            ;;
    esac
}

# 捕获中断信号
trap 'print_message "正在退出..."; exit 0' INT TERM

# 执行主函数
main "$@"




