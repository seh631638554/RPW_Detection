package httpserver

import (
	dao "RPW_Detection/Dao"
	models "RPW_Detection/Models"
	"errors"
	"net/http"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// ==================== 文件上传处理器 ====================

// 全局存储服务实例
var storageService StorageService
var uploadHandler *UploadHandler

// UploadHandler handles upload APIs.
type UploadHandler struct {
	service *UploadService
	repo    *dao.Repo
	cfg     *Config
}

func NewUploadHandler(service *UploadService, repo *dao.Repo, cfg *Config) *UploadHandler {
	return &UploadHandler{service: service, repo: repo, cfg: cfg}
}

// InitStorageService 初始化存储服务
func InitStorageService() error {
	config := LoadObjectStorageConfig()

	var err error
	storageService, err = NewMinIOStorageService(config)
	if err != nil {
		return err
	}
	uploadHandler = NewUploadHandler(NewUploadService(storageService, config), nil, nil)

	return nil
}

func ensureUploadInfrastructure() error {
	if storageService == nil {
		if err := InitStorageService(); err != nil {
			return err
		}
	}
	if err := ensureUploadEventPublisher(); err != nil {
		return err
	}
	return nil
}

// CreateUploadJob 创建上传任务
// POST /api/v1/jobs
func CreateUploadJob(c *gin.Context) {
	(&UploadHandler{}).HandleCreateUploadJob(c)
}

func GetUploadJobStatus(c *gin.Context) {
	(&UploadHandler{}).HandleGetUploadJobStatus(c)
}

func ListUploadJobs(c *gin.Context) {
	(&UploadHandler{}).HandleListUploadJobs(c)
}

func DeleteUploadJob(c *gin.Context) {
	(&UploadHandler{}).HandleDeleteUploadJob(c)
}

func UploadCompletionWebhook(c *gin.Context) {
	(&UploadHandler{}).HandleUploadCompletionWebhook(c)
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

	if err := h.createDetectionRecord(c, req, response); err != nil {
		errorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	successResponse(c, response)
}

func (h *UploadHandler) HandleGetUploadJobStatus(c *gin.Context) {
	jobID := c.Param("id")
	if jobID == "" {
		errorResponse(c, http.StatusBadRequest, "任务ID不能为空")
		return
	}

	user, err := getCurrentUser(c, h.repo)
	if err != nil {
		errorResponse(c, http.StatusUnauthorized, err.Error())
		return
	}

	record, err := h.repo.FindDetectionRecordByJobID(jobID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			errorResponse(c, http.StatusNotFound, "任务不存在")
			return
		}
		errorResponse(c, http.StatusInternalServerError, "查询任务失败")
		return
	}
	if user.IsAdmin != 1 && record.UserID != user.ID {
		errorResponse(c, http.StatusForbidden, "无权查看该任务")
		return
	}

	successResponse(c, buildJobResponse(record))
}

func (h *UploadHandler) HandleListUploadJobs(c *gin.Context) {
	user, err := getCurrentUser(c, h.repo)
	if err != nil {
		errorResponse(c, http.StatusUnauthorized, err.Error())
		return
	}

	page := parsePositiveIntQuery(c, "page", 1)
	pageSize := parsePositiveIntQuery(c, "page_size", 20)
	deviceCode := c.Query("device_id")
	status := c.Query("status")
	parkID := c.Query("park_id")
	treeID := c.Query("tree_id")

	baseOpts := []dao.Option{}
	if user.IsAdmin != 1 {
		baseOpts = append(baseOpts, dao.WithWhere("user_id = ?", user.ID))
	}
	if strings.TrimSpace(status) != "" {
		baseOpts = append(baseOpts, dao.WithWhere("status = ?", strings.TrimSpace(status)))
	}
	if strings.TrimSpace(parkID) != "" {
		baseOpts = append(baseOpts, dao.WithWhere("park_id = ?", strings.TrimSpace(parkID)))
	}
	if strings.TrimSpace(treeID) != "" {
		baseOpts = append(baseOpts, dao.WithWhere("tree_id = ?", strings.TrimSpace(treeID)))
	}
	if strings.TrimSpace(deviceCode) != "" {
		device, err := h.repo.FindDeviceByCode(strings.TrimSpace(deviceCode))
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				successResponse(c, PaginatedResponse{
					Total:      0,
					Page:       page,
					PageSize:   pageSize,
					TotalPages: 0,
					Data:       []gin.H{},
				})
				return
			}
			errorResponse(c, http.StatusInternalServerError, "查询设备失败")
			return
		}
		baseOpts = append(baseOpts, dao.WithWhere("device_id = ?", device.ID))
	}

	total, err := h.repo.Count(&models.DetectionRecord{}, baseOpts...)
	if err != nil {
		errorResponse(c, http.StatusInternalServerError, "统计任务失败")
		return
	}

	listOpts := append([]dao.Option{}, baseOpts...)
	listOpts = append(listOpts,
		dao.WithPreload("Park", "Tree", "Device"),
		dao.WithOrder("id DESC"),
		dao.WithLimit(pageSize),
		dao.WithOffset((page-1)*pageSize),
	)
	records, err := h.repo.ListDetectionRecords(listOpts...)
	if err != nil {
		errorResponse(c, http.StatusInternalServerError, "查询任务列表失败")
		return
	}

	items := make([]gin.H, 0, len(records))
	for _, record := range records {
		items = append(items, buildJobResponse(&record))
	}

	totalPages := 0
	if total > 0 {
		totalPages = int((total + int64(pageSize) - 1) / int64(pageSize))
	}
	response := PaginatedResponse{
		Total:      int(total),
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
		Data:       items,
	}
	successResponse(c, response)
}

func (h *UploadHandler) HandleDeleteUploadJob(c *gin.Context) {
	jobID := c.Param("id")
	if jobID == "" {
		errorResponse(c, http.StatusBadRequest, "任务ID不能为空")
		return
	}

	user, err := getCurrentUser(c, h.repo)
	if err != nil {
		errorResponse(c, http.StatusUnauthorized, err.Error())
		return
	}

	record, err := h.repo.FindDetectionRecordByJobID(jobID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			errorResponse(c, http.StatusNotFound, "任务不存在")
			return
		}
		errorResponse(c, http.StatusInternalServerError, "查询任务失败")
		return
	}
	if user.IsAdmin != 1 && record.UserID != user.ID {
		errorResponse(c, http.StatusForbidden, "无权删除该任务")
		return
	}

	if storageService != nil {
		_ = storageService.DeleteFile(record.AudioBucket, record.AudioKey)
		if strings.TrimSpace(record.FeatureBucket) != "" && strings.TrimSpace(record.FeatureKey) != "" {
			_ = storageService.DeleteFile(record.FeatureBucket, record.FeatureKey)
		}
	}
	if err := h.repo.DeleteByID(&models.DetectionRecord{}, record.ID); err != nil {
		errorResponse(c, http.StatusInternalServerError, "删除任务失败")
		return
	}

	successResponse(c, gin.H{
		"message": "任务删除成功",
		"job_id":  jobID,
	})
}

func (h *UploadHandler) HandleUploadCompletionWebhook(c *gin.Context) {
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
	if notification.JobID != "" && notification.JobID != jobID {
		errorResponse(c, http.StatusBadRequest, "任务ID不匹配")
		return
	}
	if notification.Bucket == "" || notification.Key == "" {
		errorResponse(c, http.StatusBadRequest, "bucket和key不能为空")
		return
	}
	user, err := getCurrentUser(c, h.repo)
	if err != nil {
		errorResponse(c, http.StatusUnauthorized, err.Error())
		return
	}

	record, err := h.repo.FindDetectionRecordByJobID(jobID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			errorResponse(c, http.StatusNotFound, "任务不存在")
			return
		}
		errorResponse(c, http.StatusInternalServerError, "查询任务失败")
		return
	}
	if user.IsAdmin != 1 && record.UserID != user.ID {
		errorResponse(c, http.StatusForbidden, "无权操作该任务")
		return
	}

	if err := ensureUploadInfrastructure(); err != nil {
		errorResponse(c, http.StatusServiceUnavailable, "上传基础设施不可用: "+err.Error())
		return
	}

	exists, err := storageService.FileExists(notification.Bucket, notification.Key)
	if err != nil {
		errorResponse(c, http.StatusInternalServerError, "校验上传文件失败: "+err.Error())
		return
	}
	if !exists {
		errorResponse(c, http.StatusNotFound, "上传文件不存在")
		return
	}

	fileInfo, err := storageService.GetFileInfo(notification.Bucket, notification.Key)
	if err != nil {
		errorResponse(c, http.StatusInternalServerError, "读取上传文件信息失败: "+err.Error())
		return
	}

	completedAt := notification.CompletedAt
	if completedAt.IsZero() {
		completedAt = time.Now()
	}
	size := notification.Size
	if size <= 0 {
		size = fileInfo.Size
	}
	contentType := fileInfo.ContentType
	if contentType == "" {
		contentType = GetContentType(notification.Key)
	}

	event := AudioUploadCompletedEvent{
		JobID:       jobID,
		DeviceID:    extractDeviceIDFromObjectKey(notification.Key),
		Bucket:      notification.Bucket,
		Key:         notification.Key,
		Size:        size,
		ContentType: contentType,
		UploadedAt:  completedAt,
	}

	if err := uploadEventPublisher.PublishAudioUploadCompleted(c.Request.Context(), event); err != nil {
		errorResponse(c, http.StatusInternalServerError, "发送上传完成事件失败: "+err.Error())
		return
	}

	if err := h.repo.UpdateDetectionRecordByJobID(jobID, map[string]interface{}{
		"status":       "processing",
		"audio_bucket": notification.Bucket,
		"audio_key":    notification.Key,
	}); err != nil {
		errorResponse(c, http.StatusInternalServerError, "更新任务状态失败")
		return
	}

	successResponse(c, gin.H{
		"message":         "上传完成回调处理成功",
		"job_id":          jobID,
		"status":          "completed",
		"event_published": true,
		"device_id":       event.DeviceID,
	})
}

// ==================== 辅助函数 ====================

func (h *UploadHandler) createDetectionRecord(c *gin.Context, req CreateUploadJobRequest, response *CreateUploadJobResponse) error {
	if h == nil || h.repo == nil {
		return errors.New("上传服务未初始化")
	}

	user, err := getCurrentUser(c, h.repo)
	if err != nil {
		return err
	}

	device, err := h.repo.FindDeviceByCode(strings.TrimSpace(req.DeviceID))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("所选设备不存在")
		}
		return errors.New("查询设备失败")
	}
	if device.TreeID == nil {
		return errors.New("设备未绑定树木，无法上传音频")
	}

	record := &models.DetectionRecord{
		JobID:        response.JobID,
		UserID:       user.ID,
		ParkID:       device.ParkID,
		TreeID:       *device.TreeID,
		DeviceID:     device.ID,
		AudioBucket:  response.Bucket,
		AudioKey:     response.Key,
		Status:       "pending",
		ErrorMessage: "",
	}
	if err := h.repo.Create(record); err != nil {
		return errors.New("创建检测记录失败")
	}
	return nil
}

func buildJobResponse(record *models.DetectionRecord) gin.H {
	if record == nil {
		return gin.H{}
	}

	var resultScore interface{}
	if record.ResultScore != nil {
		resultScore = *record.ResultScore
	}

	return gin.H{
		"id":             record.JobID,
		"job_id":         record.JobID,
		"user_id":        record.UserID,
		"park_id":        record.ParkID,
		"park_name":      record.Park.Name,
		"tree_id":        record.TreeID,
		"tree_code":      record.Tree.TreeCode,
		"device_id":      record.Device.DeviceCode,
		"device_code":    record.Device.DeviceCode,
		"device_name":    record.Device.Name,
		"file_name":      path.Base(record.AudioKey),
		"audio_bucket":   record.AudioBucket,
		"audio_key":      record.AudioKey,
		"feature_bucket": record.FeatureBucket,
		"feature_key":    record.FeatureKey,
		"status":         record.Status,
		"result_label":   record.ResultLabel,
		"result_score":   resultScore,
		"error_message":  record.ErrorMessage,
		"created_at":     record.CreatedAt,
		"updated_at":     record.UpdatedAt,
	}
}

func parsePositiveIntQuery(c *gin.Context, key string, fallback int) int {
	raw := strings.TrimSpace(c.DefaultQuery(key, ""))
	if raw == "" {
		return fallback
	}
	n, err := strconv.Atoi(raw)
	if err != nil || n <= 0 {
		return fallback
	}
	return n
}
