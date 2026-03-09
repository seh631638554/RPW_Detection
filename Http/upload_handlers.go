package httpserver

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// ==================== 文件上传处理器 ====================

// 全局存储服务实例
var storageService StorageService
var uploadHandler *UploadHandler

// UploadHandler handles upload APIs.
type UploadHandler struct {
	service *UploadService
}

func NewUploadHandler(service *UploadService) *UploadHandler {
	return &UploadHandler{service: service}
}

// InitStorageService 初始化存储服务
func InitStorageService() error {
	config := LoadObjectStorageConfig()

	var err error
	storageService, err = NewMinIOStorageService(config)
	if err != nil {
		return err
	}
	uploadHandler = NewUploadHandler(NewUploadService(storageService, config))

	return nil
}

// CreateUploadJob 创建上传任务
// POST /api/v1/jobs
func CreateUploadJob(c *gin.Context) {
	(&UploadHandler{}).HandleCreateUploadJob(c)
}

func (h *UploadHandler) HandleCreateUploadJob(c *gin.Context) {
	var req CreateUploadJobRequest

	// 绑定请求参数
	if err := c.ShouldBindJSON(&req); err != nil {
		errorResponse(c, http.StatusBadRequest, "请求参数错误: "+err.Error())
		return
	}

	// 验证文件类型
	if !ValidateFileType(req.FileType) {
		errorResponse(c, http.StatusBadRequest, "不支持的文件类型: "+req.FileType)
		return
	}

	// 验证文件大小 (默认100MB)
	if !ValidateFileSize(req.FileSize, 100*1024*1024) {
		errorResponse(c, http.StatusBadRequest, "文件大小超出限制")
		return
	}
	if h == nil || h.service == nil {
		if uploadHandler == nil || uploadHandler.service == nil {
			if err := InitStorageService(); err != nil {
				errorResponse(c, http.StatusServiceUnavailable, "存储服务不可用: "+err.Error())
				return
			}
		}
		h = uploadHandler
	}
	response, err := h.service.CreateUploadJob(req)
	if err != nil {
		errorResponse(c, http.StatusInternalServerError, "生成预签名URL失败: "+err.Error())
		return
	}

	// 保存任务信息到数据库 (TODO: 实现数据库存储)
	// saveUploadJobToDatabase(response)

	// 返回成功响应
	successResponse(c, response)
}

// GetUploadJobStatus 获取上传任务状态
// GET /api/v1/jobs/:id
func GetUploadJobStatus(c *gin.Context) {
	jobID := c.Param("id")
	if jobID == "" {
		errorResponse(c, http.StatusBadRequest, "任务ID不能为空")
		return
	}

	// TODO: 从数据库查询任务状态
	// job := getUploadJobFromDatabase(jobID)

	// 模拟任务状态
	job := &UploadJob{
		ID:          jobID,
		DeviceID:    "dev_001",
		FileName:    "audio_sample.wav",
		FileSize:    1024000,
		FileType:    "wav",
		ContentType: "audio/wav",
		Status:      JobStatusPending,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	successResponse(c, job)
}

// ListUploadJobs 列出上传任务
// GET /api/v1/jobs
func ListUploadJobs(c *gin.Context) {
	// 获取查询参数
	_ = c.Query("device_id")              // TODO: 使用deviceID进行过滤
	_ = c.Query("status")                 // TODO: 使用status进行过滤
	_ = c.DefaultQuery("page", "1")       // TODO: 使用page进行分页
	_ = c.DefaultQuery("page_size", "20") // TODO: 使用pageSize进行分页

	// TODO: 从数据库查询任务列表
	// jobs := getUploadJobsFromDatabase(deviceID, status, page, pageSize)

	// 模拟任务列表
	jobs := []UploadJob{
		{
			ID:          "job_abc123",
			DeviceID:    "dev_001",
			FileName:    "audio_sample.wav",
			FileSize:    1024000,
			FileType:    "wav",
			ContentType: "audio/wav",
			Status:      JobStatusCompleted,
			CreatedAt:   time.Now().Add(-1 * time.Hour),
			UpdatedAt:   time.Now().Add(-30 * time.Minute),
		},
		{
			ID:          "job_def456",
			DeviceID:    "dev_002",
			FileName:    "audio_sample.mp3",
			FileSize:    2048000,
			FileType:    "mp3",
			ContentType: "audio/mpeg",
			Status:      JobStatusPending,
			CreatedAt:   time.Now().Add(-2 * time.Hour),
			UpdatedAt:   time.Now().Add(-2 * time.Hour),
		},
	}

	// 分页响应
	response := PaginatedResponse{
		Total:      len(jobs),
		Page:       1,
		PageSize:   20,
		TotalPages: 1,
		Data:       jobs,
	}

	successResponse(c, response)
}

// DeleteUploadJob 删除上传任务
// DELETE /api/v1/jobs/:id
func DeleteUploadJob(c *gin.Context) {
	jobID := c.Param("id")
	if jobID == "" {
		errorResponse(c, http.StatusBadRequest, "任务ID不能为空")
		return
	}

	// TODO: 从数据库删除任务
	// deleteUploadJobFromDatabase(jobID)

	// TODO: 从存储服务删除文件
	// storageService.DeleteFile("pest-detection", storageKey)

	successResponse(c, gin.H{
		"message": "任务删除成功",
		"job_id":  jobID,
	})
}

// UploadCompletionWebhook 上传完成回调
// POST /api/v1/jobs/:id/complete
func UploadCompletionWebhook(c *gin.Context) {
	jobID := c.Param("id")
	if jobID == "" {
		errorResponse(c, http.StatusBadRequest, "任务ID不能为空")
		return
	}

	var notification UploadCompletionNotification
	if err := c.ShouldBindJSON(&notification); err != nil {
		errorResponse(c, http.StatusBadRequest, "回调参数错误: "+err.Error())
		return
	}

	// TODO: 更新任务状态为已完成
	// updateUploadJobStatus(jobID, JobStatusCompleted)

	// TODO: 触发后续处理流程 (音频检测等)
	// triggerAudioDetection(jobID, notification.Bucket, notification.Key)

	successResponse(c, gin.H{
		"message": "上传完成回调处理成功",
		"job_id":  jobID,
		"status":  "completed",
	})
}

// ==================== 辅助函数 ====================

// 注意：errorResponse 和 successResponse 函数已在 handlers.go 中定义
