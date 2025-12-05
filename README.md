# 病虫害检测服务器 (RPW Detection Server)

## 项目描述
这是一个用于检测树干虫害的智能服务器系统，通过分析采集到的音频数据，使用机器学习模型判断树干是否感染虫害。

## 技术栈
- **后端框架**: Golang + Gin
- **认证**: JWT (JSON Web Token)
- **数据库**: MySQL
- **缓存**: Redis
- **消息队列**: Kafka
- **容器化**: Docker
- **其他**: gRPC, Protobuf, Epoll

## 主要功能模块

### 1. 用户认证模块
- 用户登录/注册
- JWT Token生成与验证
- Token自动刷新策略

### 2. 音频检测模块
- 音频文件上传
- 音频特征提取
- 虫害检测分析
- 检测结果查询

### 3. 设备管理模块
- 设备注册与管理
- 设备状态监控
- 设备信息查询

### 4. 数据查询模块
- 检测结果缓存
- 布隆过滤器防缓存穿透
- 设备ID和日期联合索引优化

## 项目结构
```
RPW_Detection/
├── Http/                 # HTTP服务器模块
│   ├── http.go          # 主服务器文件
│   ├── handlers.go      # 请求处理函数
│   ├── middleware.go    # 中间件
│   └── config.go        # 配置文件
├── go.mod               # Go模块依赖
└── README.md            # 项目说明文档
```

## 快速开始

### 环境要求
- Go 1.21+
- MySQL 8.0+
- Redis 6.0+
- Kafka 2.8+

### 安装依赖
```bash
go mod tidy
```

### 配置环境变量
```bash
# 服务器配置
export SERVER_PORT=8080
export SERVER_MODE=debug

# 数据库配置
export DB_HOST=localhost
export DB_PORT=3306
export DB_USERNAME=root
export DB_PASSWORD=your_password
export DB_DATABASE=pest_detection

# Redis配置
export REDIS_HOST=localhost
export REDIS_PORT=6379
export REDIS_PASSWORD=
export REDIS_DATABASE=0

# JWT配置
export JWT_SECRET_KEY=your-secret-key-here
export JWT_EXPIRE_TIME=24h

# Kafka配置
export KAFKA_BROKERS=localhost:9092
export KAFKA_TOPIC=audio_detection
export KAFKA_GROUP_ID=detection_group
```

### 运行服务器
```bash
cd Http
go run .
```

### 测试API接口

#### 健康检查
```bash
curl http://localhost:8080/api/v1/health
```

#### 用户登录
```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"123456"}'
```

#### 音频上传
```bash
curl -X POST http://localhost:8080/api/v1/detection/upload \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -F "audio_file=@/path/to/audio.wav" \
  -F "device_id=dev_001" \
  -F "audio_type=wav"
```

#### 查询检测结果
```bash
curl http://localhost:8080/api/v1/detection/result/TASK_ID \
  -H "Authorization: Bearer YOUR_TOKEN"
```

## API接口文档

### 认证接口
- `POST /api/v1/auth/login` - 用户登录
- `POST /api/v1/auth/register` - 用户注册
- `GET /api/v1/auth/verify` - Token验证

### 检测接口
- `POST /api/v1/detection/upload` - 音频上传
- `GET /api/v1/detection/result/:id` - 获取检测结果
- `GET /api/v1/detection/status/:id` - 获取检测状态

### 设备接口
- `GET /api/v1/device/list` - 设备列表
- `GET /api/v1/device/:id` - 设备信息
- `POST /api/v1/device/register` - 设备注册

## 中间件特性

### 1. JWT认证中间件
- 自动验证请求头中的Authorization token
- 支持Bearer token格式
- Token过期自动拒绝请求

### 2. 请求日志中间件
- 记录所有HTTP请求的详细信息
- 包含请求时间、方法、路径、状态码、延迟等

### 3. 错误处理中间件
- 自动捕获panic并返回友好的错误信息
- 防止服务器崩溃

### 4. 性能监控中间件
- 记录请求处理时间
- 标记慢请求（超过1秒）

### 5. CORS中间件
- 支持跨域请求
- 可配置允许的域名、方法、头部

## 开发计划

### 第一阶段 (已完成)
- [x] Gin框架基础架构
- [x] 基本路由配置
- [x] JWT认证中间件
- [x] 请求处理函数框架

### 第二阶段 (进行中)
- [ ] 数据库连接与模型
- [ ] Redis缓存集成
- [ ] Kafka消息队列
- [ ] 音频文件处理

### 第三阶段 (计划中)
- [ ] gRPC服务集成
- [ ] 机器学习模型集成
- [ ] 性能优化
- [ ] 监控告警

## 贡献指南
1. Fork 项目
2. 创建功能分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 打开 Pull Request

## 许可证
本项目采用 MIT 许可证 - 查看 [LICENSE](LICENSE) 文件了解详情

## 联系方式
如有问题或建议，请提交 Issue 或联系开发团队。
