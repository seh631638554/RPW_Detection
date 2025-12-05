package httpserver

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// ==================== 文件上传处理器测试 ====================

func TestCreateUploadJob(t *testing.T) {
	// 设置测试模式
	gin.SetMode(gin.TestMode)
	
	// 创建测试请求
	reqBody := CreateUploadJobRequest{
		DeviceID:    "dev_001",
		FileName:    "test_audio.wav",
		FileSize:    1024000,
		FileType:    "wav",
		ContentType: "audio/wav",
		Description: "测试音频文件",
	}
	
	jsonData, _ := json.Marshal(reqBody)
	
	// 创建HTTP请求
	req, _ := http.NewRequest("POST", "/api/v1/jobs", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	
	// 创建响应记录器
	w := httptest.NewRecorder()
	
	// 创建Gin上下文
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	
	// 调用处理函数
	CreateUploadJob(c)
	
	// 验证响应状态码
	assert.Equal(t, http.StatusOK, w.Code)
	
	// 解析响应
	var response APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	
	// 验证响应结构
	assert.Equal(t, 200, response.Code)
	assert.Equal(t, "操作成功", response.Message)
	assert.NotNil(t, response.Data)
	
	// 验证数据字段
	data, ok := response.Data.(map[string]interface{})
	assert.True(t, ok)
	
	assert.Contains(t, data, "job_id")
	assert.Contains(t, data, "upload_url")
	assert.Contains(t, data, "bucket")
	assert.Contains(t, data, "key")
	assert.Contains(t, data, "ttl")
	assert.Contains(t, data, "expires_at")
	assert.Contains(t, data, "content_type")
	assert.Contains(t, data, "max_file_size")
	assert.Contains(t, data, "required_fields")
	assert.Contains(t, data, "status")
	assert.Contains(t, data, "created_at")
}

func TestCreateUploadJobInvalidFileType(t *testing.T) {
	// 设置测试模式
	gin.SetMode(gin.TestMode)
	
	// 创建测试请求 - 无效的文件类型
	reqBody := CreateUploadJobRequest{
		DeviceID:    "dev_001",
		FileName:    "test_audio.txt",
		FileSize:    1024000,
		FileType:    "txt",
		ContentType: "text/plain",
		Description: "测试文本文件",
	}
	
	jsonData, _ := json.Marshal(reqBody)
	
	// 创建HTTP请求
	req, _ := http.NewRequest("POST", "/api/v1/jobs", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	
	// 创建响应记录器
	w := httptest.NewRecorder()
	
	// 创建Gin上下文
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	
	// 调用处理函数
	CreateUploadJob(c)
	
	// 验证响应状态码
	assert.Equal(t, http.StatusBadRequest, w.Code)
	
	// 解析响应
	var response APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	
	// 验证错误响应
	assert.Equal(t, 400, response.Code)
	assert.Contains(t, response.Message, "不支持的文件类型")
}

func TestCreateUploadJobFileTooLarge(t *testing.T) {
	// 设置测试模式
	gin.SetMode(gin.TestMode)
	
	// 创建测试请求 - 文件过大
	reqBody := CreateUploadJobRequest{
		DeviceID:    "dev_001",
		FileName:    "large_audio.wav",
		FileSize:    200 * 1024 * 1024, // 200MB
		FileType:    "wav",
		ContentType: "audio/wav",
		Description: "大文件测试",
	}
	
	jsonData, _ := json.Marshal(reqBody)
	
	// 创建HTTP请求
	req, _ := http.NewRequest("POST", "/api/v1/jobs", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	
	// 创建响应记录器
	w := httptest.NewRecorder()
	
	// 创建Gin上下文
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	
	// 调用处理函数
	CreateUploadJob(c)
	
	// 验证响应状态码
	assert.Equal(t, http.StatusBadRequest, w.Code)
	
	// 解析响应
	var response APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	
	// 验证错误响应
	assert.Equal(t, 400, response.Code)
	assert.Contains(t, response.Message, "文件大小超出限制")
}

func TestGetUploadJobStatus(t *testing.T) {
	// 设置测试模式
	gin.SetMode(gin.TestMode)
	
	// 创建HTTP请求
	req, _ := http.NewRequest("GET", "/api/v1/jobs/job_123", nil)
	
	// 创建响应记录器
	w := httptest.NewRecorder()
	
	// 创建Gin上下文
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: "job_123"}}
	
	// 调用处理函数
	GetUploadJobStatus(c)
	
	// 验证响应状态码
	assert.Equal(t, http.StatusOK, w.Code)
	
	// 解析响应
	var response APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	
	// 验证响应结构
	assert.Equal(t, 200, response.Code)
	assert.Equal(t, "操作成功", response.Message)
	assert.NotNil(t, response.Data)
}

func TestListUploadJobs(t *testing.T) {
	// 设置测试模式
	gin.SetMode(gin.TestMode)
	
	// 创建HTTP请求
	req, _ := http.NewRequest("GET", "/api/v1/jobs", nil)
	
	// 创建响应记录器
	w := httptest.NewRecorder()
	
	// 创建Gin上下文
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	
	// 调用处理函数
	ListUploadJobs(c)
	
	// 验证响应状态码
	assert.Equal(t, http.StatusOK, w.Code)
	
	// 解析响应
	var response APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	
	// 验证响应结构
	assert.Equal(t, 200, response.Code)
	assert.Equal(t, "操作成功", response.Message)
	assert.NotNil(t, response.Data)
	
	// 验证分页数据
	data, ok := response.Data.(map[string]interface{})
	assert.True(t, ok)
	
	assert.Contains(t, data, "total")
	assert.Contains(t, data, "page")
	assert.Contains(t, data, "page_size")
	assert.Contains(t, data, "total_pages")
	assert.Contains(t, data, "data")
}

func TestDeleteUploadJob(t *testing.T) {
	// 设置测试模式
	gin.SetMode(gin.TestMode)
	
	// 创建HTTP请求
	req, _ := http.NewRequest("DELETE", "/api/v1/jobs/job_123", nil)
	
	// 创建响应记录器
	w := httptest.NewRecorder()
	
	// 创建Gin上下文
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: "job_123"}}
	
	// 调用处理函数
	DeleteUploadJob(c)
	
	// 验证响应状态码
	assert.Equal(t, http.StatusOK, w.Code)
	
	// 解析响应
	var response APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	
	// 验证响应结构
	assert.Equal(t, 200, response.Code)
	assert.Equal(t, "操作成功", response.Message)
	assert.NotNil(t, response.Data)
	
	// 验证删除消息
	data, ok := response.Data.(map[string]interface{})
	assert.True(t, ok)
	
	assert.Equal(t, "任务删除成功", data["message"])
	assert.Equal(t, "job_123", data["job_id"])
}

// ==================== 工具函数测试 ====================

func TestGenerateJobID(t *testing.T) {
	jobID1 := GenerateJobID()
	jobID2 := GenerateJobID()
	
	// 验证格式
	assert.Contains(t, jobID1, "job_")
	assert.Contains(t, jobID2, "job_")
	
	// 验证唯一性
	assert.NotEqual(t, jobID1, jobID2)
}

func TestGenerateStorageKey(t *testing.T) {
	deviceID := "dev_001"
	fileName := "audio_sample.wav"
	
	storageKey := GenerateStorageKey(deviceID, fileName)
	
	// 验证格式
	assert.Contains(t, storageKey, deviceID)
	assert.Contains(t, storageKey, fileName)
	assert.Contains(t, storageKey, "/")
	
	// 验证时间戳格式
	assert.Contains(t, storageKey, "2024") // 假设是2024年
}

func TestValidateFileType(t *testing.T) {
	// 测试有效文件类型
	assert.True(t, ValidateFileType("wav"))
	assert.True(t, ValidateFileType("mp3"))
	assert.True(t, ValidateFileType("flac"))
	assert.True(t, ValidateFileType("m4a"))
	assert.True(t, ValidateFileType("aac"))
	
	// 测试无效文件类型
	assert.False(t, ValidateFileType("txt"))
	assert.False(t, ValidateFileType("pdf"))
	assert.False(t, ValidateFileType("doc"))
	
	// 测试大小写
	assert.True(t, ValidateFileType("WAV"))
	assert.True(t, ValidateFileType("MP3"))
}

func TestValidateFileSize(t *testing.T) {
	maxSize := int64(100 * 1024 * 1024) // 100MB
	
	// 测试有效文件大小
	assert.True(t, ValidateFileSize(1024, maxSize))           // 1KB
	assert.True(t, ValidateFileSize(50*1024*1024, maxSize))   // 50MB
	assert.True(t, ValidateFileSize(maxSize, maxSize))         // 100MB
	
	// 测试无效文件大小
	assert.False(t, ValidateFileSize(0, maxSize))             // 0字节
	assert.False(t, ValidateFileSize(-1024, maxSize))         // 负数
	assert.False(t, ValidateFileSize(150*1024*1024, maxSize)) // 150MB
	
	// 测试默认最大大小
	assert.True(t, ValidateFileSize(50*1024*1024, 0))        // 使用默认100MB
}

func TestGetContentType(t *testing.T) {
	// 测试音频文件类型
	assert.Equal(t, "audio/wav", GetContentType("audio.wav"))
	assert.Equal(t, "audio/mpeg", GetContentType("audio.mp3"))
	assert.Equal(t, "audio/flac", GetContentType("audio.flac"))
	assert.Equal(t, "audio/mp4", GetContentType("audio.m4a"))
	assert.Equal(t, "audio/aac", GetContentType("audio.aac"))
	
	// 测试未知文件类型
	assert.Equal(t, "application/octet-stream", GetContentType("audio.xyz"))
	assert.Equal(t, "application/octet-stream", GetContentType("file.txt"))
	
	// 测试大小写
	assert.Equal(t, "audio/wav", GetContentType("AUDIO.WAV"))
	assert.Equal(t, "audio/mpeg", GetContentType("AUDIO.MP3"))
}
