# 文件上传功能说明

## 🎯 功能概述

本系统实现了基于预签名URL的文件上传功能，支持音频文件的安全上传和任务管理。

## 🔄 上传流程

```
1. 前端 → POST /api/v1/jobs (元信息)
2. 后端 → 生成job_id、预签名URL
3. 前端 → 使用预签名URL直接上传到对象存储
4. 后端 → 接收上传完成通知
```

## 📋 API接口说明

### 1. 创建上传任务

**接口**: `POST /api/v1/jobs`

**请求体**:
```json
{
  "device_id": "dev_001",
  "file_name": "audio_sample.wav",
  "file_size": 1024000,
  "file_type": "wav",
  "content_type": "audio/wav",
  "description": "音频文件描述"
}
```

**响应**:
```json
{
  "code": 200,
  "message": "操作成功",
  "data": {
    "job_id": "job_abc123",
    "upload_url": "https://minio.example.com/pest-detection/...",
    "bucket": "pest-detection",
    "key": "dev_001/2024/01/15/14/audio_sample_abc123.wav",
    "ttl": 86400,
    "expires_at": "2024-01-16T14:00:00Z",
    "content_type": "audio/wav",
    "max_file_size": 1024000,
    "required_fields": ["file"],
    "status": "pending",
    "created_at": "2024-01-15T14:00:00Z"
  },
  "time": "2024-01-15T14:00:00Z"
}
```

### 2. 获取任务状态

**接口**: `GET /api/v1/jobs/:id`

**响应**:
```json
{
  "code": 200,
  "message": "操作成功",
  "data": {
    "id": "job_abc123",
    "device_id": "dev_001",
    "file_name": "audio_sample.wav",
    "file_size": 1024000,
    "file_type": "wav",
    "content_type": "audio/wav",
    "status": "completed",
    "created_at": "2024-01-15T14:00:00Z",
    "updated_at": "2024-01-15T14:30:00Z"
  }
}
```

### 3. 列出所有任务

**接口**: `GET /api/v1/jobs?device_id=dev_001&status=pending&page=1&page_size=20`

**响应**:
```json
{
  "code": 200,
  "message": "操作成功",
  "data": {
    "total": 2,
    "page": 1,
    "page_size": 20,
    "total_pages": 1,
    "data": [
      {
        "id": "job_abc123",
        "device_id": "dev_001",
        "file_name": "audio_sample.wav",
        "status": "completed"
      }
    ]
  }
}
```

### 4. 删除任务

**接口**: `DELETE /api/v1/jobs/:id`

**响应**:
```json
{
  "code": 200,
  "message": "操作成功",
  "data": {
    "message": "任务删除成功",
    "job_id": "job_abc123"
  }
}
```

### 5. 上传完成回调

**接口**: `POST /api/v1/jobs/:id/complete`

**请求体**:
```json
{
  "job_id": "job_abc123",
  "bucket": "pest-detection",
  "key": "dev_001/2024/01/15/14/audio_sample_abc123.wav",
  "etag": "abc123def456",
  "size": 1024000,
  "completed_at": "2024-01-15T14:30:00Z"
}
```

## 🔧 配置说明

### 对象存储配置

```go
type ObjectStorageConfig struct {
    Provider    string // 存储提供商 (minio, s3, oss)
    Endpoint    string // 存储服务端点
    AccessKey   string // 访问密钥
    SecretKey   string // 秘密密钥
    Bucket      string // 默认存储桶
    Region      string // 存储区域
    UseSSL      bool   // 是否使用SSL
    ExpireHours int    // 预签名URL过期时间(小时)
}
```

### 默认配置

```go
{
    Provider:    "minio",
    Endpoint:    "localhost:9000",
    AccessKey:   "minioadmin",
    SecretKey:   "minioadmin",
    Bucket:      "pest-detection",
    Region:      "us-east-1",
    UseSSL:      false,
    ExpireHours: 24
}
```

## 📁 文件存储结构

```
pest-detection/
├── dev_001/
│   ├── 2024/
│   │   ├── 01/
│   │   │   ├── 15/
│   │   │   │   ├── 14/
│   │   │   │   │   ├── audio_sample_abc123.wav
│   │   │   │   │   └── audio_sample_def456.mp3
│   │   │   │   └── 15/
│   │   │   │       └── audio_sample_ghi789.flac
│   │   │   └── 16/
│   │   └── 02/
│   └── 2023/
└── dev_002/
    └── 2024/
        └── 01/
            └── 15/
                └── 14/
                    └── audio_sample_jkl012.m4a
```

## ✅ 支持的文件类型

- **WAV**: `audio/wav`
- **MP3**: `audio/mpeg`
- **FLAC**: `audio/flac`
- **M4A**: `audio/mp4`
- **AAC**: `audio/aac`

## 📏 文件大小限制

- **默认最大大小**: 100MB
- **可配置**: 通过环境变量或配置文件设置

## 🔐 安全特性

1. **预签名URL**: 临时访问权限，24小时后自动过期
2. **文件类型验证**: 只允许音频文件上传
3. **文件大小限制**: 防止恶意大文件上传
4. **元数据记录**: 记录上传时间、设备ID等信息

## 🚀 使用示例

### 前端JavaScript示例

```javascript
// 1. 创建上传任务
const createJob = async (fileInfo) => {
  const response = await fetch('/api/v1/jobs', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify(fileInfo)
  });
  
  const result = await response.json();
  return result.data;
};

// 2. 使用预签名URL上传文件
const uploadFile = async (uploadURL, file) => {
  const response = await fetch(uploadURL, {
    method: 'PUT',
    body: file,
    headers: {
      'Content-Type': file.type,
    }
  });
  
  if (response.ok) {
    // 上传成功，通知后端
    await notifyCompletion(jobId, response.headers.get('ETag'));
  }
};

// 3. 通知上传完成
const notifyCompletion = async (jobId, etag) => {
  await fetch(`/api/v1/jobs/${jobId}/complete`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify({
      job_id: jobId,
      etag: etag,
      completed_at: new Date().toISOString()
    })
  });
};

// 完整流程
const uploadAudioFile = async (file) => {
  try {
    // 创建任务
    const job = await createJob({
      device_id: 'dev_001',
      file_name: file.name,
      file_size: file.size,
      file_type: file.name.split('.').pop(),
      content_type: file.type,
      description: '音频文件上传'
    });
    
    // 上传文件
    await uploadFile(job.upload_url, file);
    
    console.log('文件上传成功:', job.job_id);
  } catch (error) {
    console.error('上传失败:', error);
  }
};
```

### cURL示例

```bash
# 1. 创建上传任务
curl -X POST http://localhost:8080/api/v1/jobs \
  -H "Content-Type: application/json" \
  -d '{
    "device_id": "dev_001",
    "file_name": "audio.wav",
    "file_size": 1024000,
    "file_type": "wav",
    "content_type": "audio/wav",
    "description": "测试音频"
  }'

# 2. 获取任务状态
curl http://localhost:8080/api/v1/jobs/job_abc123

# 3. 列出所有任务
curl "http://localhost:8080/api/v1/jobs?device_id=dev_001&status=pending"

# 4. 删除任务
curl -X DELETE http://localhost:8080/api/v1/jobs/job_abc123
```

## 🧪 测试

运行测试：

```bash
# 运行所有测试
go test ./...

# 运行特定测试
go test -v ./Http -run TestCreateUploadJob

# 运行测试并显示覆盖率
go test -v -cover ./...
```

## 🔍 故障排除

### 常见问题

1. **存储服务连接失败**
   - 检查MinIO/S3服务是否运行
   - 验证访问密钥和端点配置

2. **预签名URL生成失败**
   - 检查存储服务权限
   - 验证存储桶是否存在

3. **文件上传失败**
   - 检查预签名URL是否过期
   - 验证文件类型和大小限制

### 日志查看

```bash
# 查看应用日志
tail -f logs/app.log

# 查看存储服务日志
docker logs pest-detection-minio
```

## 📚 相关文档

- [Gin框架文档](https://gin-gonic.com/docs/)
- [AWS SDK Go文档](https://docs.aws.amazon.com/sdk-for-go/)
- [MinIO文档](https://docs.min.io/)
- [预签名URL说明](https://docs.aws.amazon.com/AmazonS3/latest/userguide/ShareObjectPreSignedURL.html)
