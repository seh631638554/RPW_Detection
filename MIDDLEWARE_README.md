# 中间件服务使用说明

## 🎯 服务概述

这个项目使用Docker容器来提供以下中间件服务：

- **MySQL 8.0** - 主数据库
- **Redis 6.2 主从** - 缓存和分布式锁
- **Kafka 7.3 集群** - 消息队列
- **Zookeeper** - Kafka集群协调服务

## 🚀 快速开始

### 启动所有服务
```bash
./start-middleware.sh start
```

### 查看服务状态
```bash
./start-middleware.sh status
```

### 停止所有服务
```bash
./start-middleware.sh stop
```

### 重启所有服务
```bash
./start-middleware.sh restart
```

## 📋 服务详情

### MySQL 数据库
- **端口**: 3306
- **用户名**: root
- **密码**: password123
- **数据库**: pest_detection
- **连接字符串**: root:password123@localhost:3306/pest_detection

### Redis 主从
- **主节点**: 1个 (端口: 6379)
- **从节点**: 2个 (端口: 6380, 6381)
- **模式**: 主从复制
- **特点**: 读写分离，高可用
- **主节点连接**: localhost:6379

### Kafka 集群
- **Broker 1**: 端口 9092
- **Broker 2**: 端口 9093  
- **Broker 3**: 端口 9094
- **副本因子**: 3
- **连接地址**: localhost:9092,localhost:9093,localhost:9094

### Zookeeper
- **端口**: 2181
- **用途**: Kafka集群协调

## 🔧 管理命令

### 查看服务日志
```bash
./start-middleware.sh logs
```

### 检查服务健康状态
```bash
./start-middleware.sh health
```

### 清理所有数据（谨慎使用）
```bash
./start-middleware.sh clean
```

### 查看帮助
```bash
./start-middleware.sh help
```

## 📊 访问地址

| 服务 | 地址 | 说明 |
|------|------|------|
| MySQL | localhost:3306 | 数据库服务 |
| Redis主节点 | localhost:6379 | 主节点，支持读写 |
| Redis从节点1 | localhost:6380 | 从节点，只读 |
| Redis从节点2 | localhost:6381 | 从节点，只读 |
| Kafka集群 | localhost:9092-9094 | 消息队列 |
| Zookeeper | localhost:2181 | 集群协调 |

## ⚠️ 注意事项

1. **Redis主从**: 主节点支持读写，从节点只读，自动同步数据
2. **端口冲突**: 确保本地端口 3306, 6379-6381, 9092-9094, 2181 未被占用
3. **数据持久化**: 所有数据都保存在Docker volumes中，重启容器数据不会丢失
4. **资源要求**: 建议至少4GB内存和2核CPU
5. **无管理界面**: 所有服务通过命令行或程序直接连接

## 🐛 故障排除

### 服务启动失败
```bash
# 查看详细日志
docker-compose logs [服务名]

# 重启特定服务
docker-compose restart [服务名]

# 完全重新启动
./start-middleware.sh restart
```

### 连接问题
```bash
# 检查服务状态
./start-middleware.sh status

# 检查健康状态
./start-middleware.sh health

# 查看网络配置
docker network ls
docker network inspect pest-detection-network
```

### 清理环境
```bash
# 停止所有服务
./start-middleware.sh stop

# 清理所有数据
./start-middleware.sh clean

# 重新启动
./start-middleware.sh start
```

## 🔄 开发流程

1. **启动中间件服务**: `./start-middleware.sh start`
2. **等待服务就绪**: 检查健康状态 `./start-middleware.sh health`
3. **开发应用代码**: 连接到相应的服务端口
4. **测试完成后**: 停止服务 `./start-middleware.sh stop`

## 📚 更多信息

- [Docker Compose 文档](https://docs.docker.com/compose/)
- [Redis 主从复制文档](https://redis.io/topics/replication)
- [Kafka 集群文档](https://kafka.apache.org/documentation/)
- [MySQL 8.0 文档](https://dev.mysql.com/doc/refman/8.0/en/)
