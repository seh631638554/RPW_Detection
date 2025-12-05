package httpserver

import (
	"time"
)

// ==================== 文件上传相关模型 ====================

// 创建上传任务请求
type CreateUploadJobRequest struct {
	DeviceID    string `json:"device_id" binding:"required"`    // 设备ID
	FileName    string `json:"file_name" binding:"required"`    // 文件名
	FileSize    int64  `json:"file_size" binding:"required"`    // 文件大小(字节)
	FileType    string `json:"file_type" binding:"required"`    // 文件类型 (wav, mp3, flac等)
	ContentType string `json:"content_type" binding:"required"` // MIME类型
	Description string `json:"description"`                     // 文件描述
}

// 上传任务响应
type CreateUploadJobResponse struct {
	JobID          string    `json:"job_id"`          // 任务ID
	UploadURL      string    `json:"upload_url"`      // 预签名上传URL
	Bucket         string    `json:"bucket"`          // 存储桶名称
	Key            string    `json:"key"`             // 对象键
	TTL            int64     `json:"ttl"`             // 预签名URL过期时间(秒)
	ExpiresAt      time.Time `json:"expires_at"`      // 过期时间点
	ContentType    string    `json:"content_type"`    // 要求的Content-Type
	MaxFileSize    int64     `json:"max_file_size"`   // 最大文件大小
	RequiredFields []string  `json:"required_fields"` // 必需的表单字段
	Status         string    `json:"status"`          // 任务状态
	CreatedAt      time.Time `json:"created_at"`      // 创建时间
}

// 上传任务状态
type UploadJobStatus string

const (
	JobStatusPending   UploadJobStatus = "pending"   // 等待上传
	JobStatusUploading UploadJobStatus = "uploading" // 上传中
	JobStatusCompleted UploadJobStatus = "completed" // 上传完成
	JobStatusFailed    UploadJobStatus = "failed"    // 上传失败
	JobStatusExpired   UploadJobStatus = "expired"   // 已过期
)

// 上传任务记录
type UploadJob struct {
	ID          string          `json:"id" db:"id"`                     // 任务ID
	DeviceID    string          `json:"device_id" db:"device_id"`       // 设备ID
	FileName    string          `json:"file_name" db:"file_name"`       // 文件名
	FileSize    int64           `json:"file_size" db:"file_size"`       // 文件大小
	FileType    string          `json:"file_type" db:"file_type"`       // 文件类型
	ContentType string          `json:"content_type" db:"content_type"` // MIME类型
	Description string          `json:"description" db:"description"`   // 文件描述
	Bucket      string          `json:"bucket" db:"bucket"`             // 存储桶
	Key         string          `json:"key" db:"key"`                   // 对象键
	Status      UploadJobStatus `json:"status" db:"status"`             // 任务状态
	UploadURL   string          `json:"upload_url" db:"upload_url"`     // 预签名URL
	TTL         int64           `json:"ttl" db:"ttl"`                   // 过期时间(秒)
	ExpiresAt   time.Time       `json:"expires_at" db:"expires_at"`     // 过期时间点
	CreatedAt   time.Time       `json:"created_at" db:"created_at"`     // 创建时间
	UpdatedAt   time.Time       `json:"updated_at" db:"updated_at"`     // 更新时间
}

// 对象存储配置
type ObjectStorageConfig struct {
	Provider    string `json:"provider"`     // 存储提供商 (minio, s3, oss)
	Endpoint    string `json:"endpoint"`     // 存储服务端点
	AccessKey   string `json:"access_key"`   // 访问密钥
	SecretKey   string `json:"secret_key"`   // 秘密密钥
	Bucket      string `json:"bucket"`       // 默认存储桶
	Region      string `json:"region"`       // 存储区域
	UseSSL      bool   `json:"use_ssl"`      // 是否使用SSL
	ExpireHours int    `json:"expire_hours"` // 预签名URL过期时间(小时)
}

// 预签名URL生成参数
type PresignedURLParams struct {
	Bucket      string            `json:"bucket"`       // 存储桶
	Key         string            `json:"key"`          // 对象键
	Method      string            `json:"method"`       // HTTP方法 (PUT, POST)
	Expires     time.Duration     `json:"expires"`      // 过期时间
	ContentType string            `json:"content_type"` // 内容类型
	Metadata    map[string]string `json:"metadata"`     // 元数据
}

// 上传完成通知
type UploadCompletionNotification struct {
	JobID       string    `json:"job_id"`       // 任务ID
	Bucket      string    `json:"bucket"`       // 存储桶
	Key         string    `json:"key"`          // 对象键
	ETag        string    `json:"etag"`         // 文件ETag
	Size        int64     `json:"size"`         // 文件大小
	CompletedAt time.Time `json:"completed_at"` // 完成时间
}

// ==================== 响应结构体 ====================

// 标准API响应
type APIResponse struct {
	Code    int         `json:"code"`           // 响应状态码
	Message string      `json:"message"`        // 响应消息
	Data    interface{} `json:"data,omitempty"` // 响应数据
	Time    time.Time   `json:"time"`           // 响应时间
}

// 分页响应
type PaginatedResponse struct {
	Total      int         `json:"total"`       // 总记录数
	Page       int         `json:"page"`        // 当前页码
	PageSize   int         `json:"page_size"`   // 每页大小
	TotalPages int         `json:"total_pages"` // 总页数
	Data       interface{} `json:"data"`        // 数据列表
}

// ==================== 错误定义 ====================

// 自定义错误
type AppError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

func (e AppError) Error() string {
	return e.Message
}

// 预定义错误
var (
	ErrInvalidRequest      = AppError{Code: 400, Message: "无效的请求参数"}
	ErrUnauthorized        = AppError{Code: 401, Message: "未授权访问"}
	ErrForbidden           = AppError{Code: 403, Message: "禁止访问"}
	ErrNotFound            = AppError{Code: 404, Message: "资源不存在"}
	ErrFileTooLarge        = AppError{Code: 413, Message: "文件过大"}
	ErrUnsupportedFileType = AppError{Code: 415, Message: "不支持的文件类型"}
	ErrInternalServer      = AppError{Code: 500, Message: "服务器内部错误"}
	ErrStorageService      = AppError{Code: 503, Message: "存储服务不可用"}
)
