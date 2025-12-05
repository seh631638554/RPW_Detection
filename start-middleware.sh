#!/bin/bash

# 启动中间件服务脚本
# 使用方法: ./start-middleware.sh [start|stop|restart|status|logs]

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
    echo -e "${BLUE}  中间件服务管理脚本${NC}"
    echo -e "${BLUE}================================${NC}"
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

# 启动服务
start_services() {
    print_message "启动中间件服务..."
    
    # 启动基础服务
    docker-compose up -d mysql zookeeper
    
    print_message "等待基础服务启动..."
    sleep 10
    
    # 启动Redis主从
    print_message "启动Redis主从服务..."
    docker-compose up -d redis-master redis-slave-1 redis-slave-2
    
    print_message "等待Redis服务启动..."
    sleep 10
    
    # 启动Kafka集群
    print_message "启动Kafka集群..."
    docker-compose up -d kafka-1 kafka-2 kafka-3
    
    print_message "等待Kafka启动..."
    sleep 20
    
    print_message "所有服务启动完成！"
    show_status
}

# 停止服务
stop_services() {
    print_message "停止所有服务..."
    docker-compose down
    print_message "所有服务已停止"
}

# 重启服务
restart_services() {
    print_message "重启所有服务..."
    stop_services
    sleep 5
    start_services
}

# 显示服务状态
show_status() {
    print_message "服务状态："
    docker-compose ps
    
    echo ""
    print_message "访问地址："
    echo "  MySQL: localhost:3306"
    echo "  Redis主节点: localhost:6379"
    echo "  Redis从节点1: localhost:6380"
    echo "  Redis从节点2: localhost:6381"
    echo "  Kafka集群: localhost:9092-9094"
    echo "  Zookeeper: localhost:2181"
    echo ""
    print_message "连接信息："
    echo "  MySQL: root/password123@localhost:3306/pest_detection"
    echo "  Redis: localhost:6379 (主节点)"
    echo "  Kafka: localhost:9092,localhost:9093,localhost:9094"
}

# 显示日志
show_logs() {
    print_message "显示服务日志..."
    docker-compose logs -f
}

# 清理服务
clean_services() {
    print_warning "这将删除所有数据，确定继续吗？(y/N)"
    read -r response
    if [[ "$response" =~ ^([yY][eE][sS]|[yY])$ ]]; then
        print_message "清理所有服务..."
        docker-compose down -v
        docker system prune -f
        print_message "清理完成"
    else
        print_message "取消清理操作"
    fi
}

# 检查服务健康状态
check_health() {
    print_message "检查服务健康状态..."
    
    # 检查MySQL
    if docker-compose exec mysql mysqladmin ping -h localhost -u root -ppassword123 &> /dev/null; then
        print_message "✅ MySQL: 运行正常"
    else
        print_error "❌ MySQL: 连接失败"
    fi
    
    # 检查Redis主节点
    if docker-compose exec redis-master redis-cli ping &> /dev/null; then
        print_message "✅ Redis主节点: 运行正常"
    else
        print_error "❌ Redis主节点: 连接失败"
    fi
    
    # 检查Redis从节点
    if docker-compose exec redis-slave-1 redis-cli ping &> /dev/null; then
        print_message "✅ Redis从节点1: 运行正常"
    else
        print_error "❌ Redis从节点1: 连接失败"
    fi
    
    if docker-compose exec redis-slave-2 redis-cli ping &> /dev/null; then
        print_message "✅ Redis从节点2: 运行正常"
    else
        print_error "❌ Redis从节点2: 连接失败"
    fi
    
    # 检查Kafka
    if docker-compose exec kafka-1 kafka-topics --bootstrap-server localhost:9092 --list &> /dev/null; then
        print_message "✅ Kafka集群: 运行正常"
    else
        print_error "❌ Kafka集群: 连接失败"
    fi
}

# 主函数
main() {
    print_header
    check_docker
    
    case "${1:-start}" in
        "start")
            start_services
            ;;
        "stop")
            stop_services
            ;;
        "restart")
            restart_services
            ;;
        "status")
            show_status
            ;;
        "logs")
            show_logs
            ;;
        "health")
            check_health
            ;;
        "clean")
            clean_services
            ;;
        "help"|"-h"|"--help")
            echo "使用方法: $0 [start|stop|restart|status|logs|health|clean|help]"
            echo ""
            echo "参数说明:"
            echo "  start   - 启动所有中间件服务 (默认)"
            echo "  stop    - 停止所有服务"
            echo "  restart - 重启所有服务"
            echo "  status  - 显示服务状态"
            echo "  logs    - 显示服务日志"
            echo "  health  - 检查服务健康状态"
            echo "  clean   - 清理所有服务和数据"
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
