package httpserver

import (
	"fmt"
	"log"
	"path/filepath"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/google/uuid"
)

// ==================== 对象存储服务 ====================

// StorageService 对象存储服务接口
type StorageService interface {
	// 生成预签名上传URL
	GeneratePresignedUploadURL(params PresignedURLParams) (string, error)

	// 检查文件是否存在
	FileExists(bucket, key string) (bool, error)

	// 获取文件信息
	GetFileInfo(bucket, key string) (*FileInfo, error)

	// 删除文件
	DeleteFile(bucket, key string) error
}

// FileInfo 文件信息
type FileInfo struct {
	Key          string            `json:"key"`
	Size         int64             `json:"size"`
	ETag         string            `json:"etag"`
	ContentType  string            `json:"content_type"`
	LastModified time.Time         `json:"last_modified"`
	Metadata     map[string]string `json:"metadata"`
}

// MinIOStorageService MinIO存储服务实现
type MinIOStorageService struct {
	config    *ObjectStorageConfig
	s3Client  *s3.S3
	s3Session *session.Session
}

// NewMinIOStorageService 创建MinIO存储服务
func NewMinIOStorageService(config *ObjectStorageConfig) (*MinIOStorageService, error) {
	// 创建AWS会话配置
	awsConfig := &aws.Config{
		Credentials:      credentials.NewStaticCredentials(config.AccessKey, config.SecretKey, ""),
		Endpoint:         aws.String(config.Endpoint),
		Region:           aws.String(config.Region),
		S3ForcePathStyle: aws.Bool(true), // MinIO需要这个设置
		DisableSSL:       aws.Bool(!config.UseSSL),
	}

	// 创建会话
	sess, err := session.NewSession(awsConfig)
	if err != nil {
		return nil, fmt.Errorf("创建AWS会话失败: %v", err)
	}

	// 创建S3客户端
	s3Client := s3.New(sess)

	// 测试连接
	_, err = s3Client.ListBuckets(&s3.ListBucketsInput{})
	if err != nil {
		return nil, fmt.Errorf("连接MinIO失败: %v", err)
	}

	log.Printf("MinIO存储服务初始化成功: %s", config.Endpoint)

	return &MinIOStorageService{
		config:    config,
		s3Client:  s3Client,
		s3Session: sess,
	}, nil
}

// GeneratePresignedUploadURL 生成预签名上传URL
func (s *MinIOStorageService) GeneratePresignedUploadURL(params PresignedURLParams) (string, error) {
	// 设置默认值
	if params.Expires == 0 {
		params.Expires = time.Duration(s.config.ExpireHours) * time.Hour
	}

	// 创建预签名URL请求
	req, _ := s.s3Client.PutObjectRequest(&s3.PutObjectInput{
		Bucket:      aws.String(params.Bucket),
		Key:         aws.String(params.Key),
		ContentType: aws.String(params.ContentType),
		Metadata:    aws.StringMap(params.Metadata),
	})

	// 生成预签名URL
	url, err := req.Presign(params.Expires)
	if err != nil {
		return "", fmt.Errorf("生成预签名URL失败: %v", err)
	}

	return url, nil
}

// FileExists 检查文件是否存在
func (s *MinIOStorageService) FileExists(bucket, key string) (bool, error) {
	_, err := s.s3Client.HeadObject(&s3.HeadObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})

	if err != nil {
		// 检查是否是"文件不存在"错误
		if strings.Contains(err.Error(), "NotFound") {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

// GetFileInfo 获取文件信息
func (s *MinIOStorageService) GetFileInfo(bucket, key string) (*FileInfo, error) {
	result, err := s.s3Client.HeadObject(&s3.HeadObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})

	if err != nil {
		return nil, err
	}

	fileInfo := &FileInfo{
		Key:          key,
		Size:         *result.ContentLength,
		ETag:         strings.Trim(*result.ETag, `"`),
		ContentType:  *result.ContentType,
		LastModified: *result.LastModified,
		Metadata:     aws.StringValueMap(result.Metadata),
	}

	return fileInfo, nil
}

// DeleteFile 删除文件
func (s *MinIOStorageService) DeleteFile(bucket, key string) error {
	_, err := s.s3Client.DeleteObject(&s3.DeleteObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})

	return err
}

// ==================== 工具函数 ====================

// GenerateJobID 生成任务ID
func GenerateJobID() string {
	return fmt.Sprintf("job_%s", uuid.New().String()[:8])
}

// GenerateStorageKey 生成存储键
func GenerateStorageKey(deviceID, fileName string) string {
	// 生成时间戳
	timestamp := time.Now().Format("2006/01/02/15")

	// 生成唯一文件名
	ext := filepath.Ext(fileName)
	baseName := strings.TrimSuffix(fileName, ext)
	uniqueName := fmt.Sprintf("%s_%s%s", baseName, uuid.New().String()[:8], ext)

	// 返回存储路径: device_id/年/月/日/时/文件名
	return fmt.Sprintf("%s/%s/%s", deviceID, timestamp, uniqueName)
}

// ValidateFileType 验证文件类型
func ValidateFileType(fileType string) bool {
	allowedTypes := []string{"wav", "mp3", "flac", "m4a", "aac"}

	for _, allowed := range allowedTypes {
		if strings.ToLower(fileType) == allowed {
			return true
		}
	}

	return false
}

// ValidateFileSize 验证文件大小
func ValidateFileSize(fileSize int64, maxSize int64) bool {
	if maxSize <= 0 {
		maxSize = 100 * 1024 * 1024 // 默认100MB
	}

	return fileSize > 0 && fileSize <= maxSize
}

// GetContentType 根据文件扩展名获取Content-Type
func GetContentType(fileName string) string {
	ext := strings.ToLower(filepath.Ext(fileName))

	switch ext {
	case ".wav":
		return "audio/wav"
	case ".mp3":
		return "audio/mpeg"
	case ".flac":
		return "audio/flac"
	case ".m4a":
		return "audio/mp4"
	case ".aac":
		return "audio/aac"
	default:
		return "application/octet-stream"
	}
}

// ==================== 配置相关 ====================

// DefaultObjectStorageConfig 默认对象存储配置
func DefaultObjectStorageConfig() *ObjectStorageConfig {
	return &ObjectStorageConfig{
		Provider:    "minio",
		Endpoint:    "localhost:9000",
		AccessKey:   "minioadmin",
		SecretKey:   "minioadmin",
		Bucket:      "pest-detection",
		Region:      "us-east-1",
		UseSSL:      false,
		ExpireHours: 24,
	}
}

// LoadObjectStorageConfig 从环境变量加载对象存储配置
func LoadObjectStorageConfig() *ObjectStorageConfig {
	config := DefaultObjectStorageConfig()

	// TODO: 从环境变量加载配置
	// config.Endpoint = os.Getenv("STORAGE_ENDPOINT")
	// config.AccessKey = os.Getenv("STORAGE_ACCESS_KEY")
	// config.SecretKey = os.Getenv("STORAGE_SECRET_KEY")
	// config.Bucket = os.Getenv("STORAGE_BUCKET")

	return config
}
