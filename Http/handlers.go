package httpserver

import (
	dao "RPW_Detection/Dao"
	models "RPW_Detection/Models"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type AuthHandler struct {
	repo *dao.Repo
	cfg  *Config
}

// 响应结构体
type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Time    string      `json:"time"`
}

// 用户登录请求
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// 用户注册请求
type RegisterRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
}

// 音频上传请求
type AudioUploadRequest struct {
	DeviceID  string `json:"device_id" binding:"required"`
	AudioType string `json:"audio_type" binding:"required"`
	Timestamp int64  `json:"timestamp"`
}

// 设备注册请求
type DeviceRegisterRequest struct {
	DeviceID   string `json:"device_id" binding:"required"`
	DeviceName string `json:"device_name" binding:"required"`
	Location   string `json:"location"`
}

// 成功响应
func successResponse(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code:    200,
		Message: "操作成功",
		Data:    data,
		Time:    time.Now().Format("2006-01-02 15:04:05"),
	})
}

// 错误响应
func errorResponse(c *gin.Context, code int, message string) {
	c.JSON(code, Response{
		Code:    code,
		Message: message,
		Time:    time.Now().Format("2006-01-02 15:04:05"),
	})
}

// ==================== 认证相关处理函数 ====================
func NewAuthHandler(repo *dao.Repo, cfg *Config) *AuthHandler {
	return &AuthHandler{repo: repo, cfg: cfg}
}

// 用户登录
func (h *AuthHandler) handleLogin(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errorResponse(c, http.StatusBadRequest, "请求参数错误: "+err.Error())
		return
	}

	user, err := h.repo.FindUserByUsername(req.Username)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			errorResponse(c, http.StatusUnauthorized, "用户名或密码错误")
			return
		}
		errorResponse(c, http.StatusInternalServerError, "查询用户失败")
		return
	}
	if user.Status != 1 {
		errorResponse(c, http.StatusForbidden, "用户已被禁用")
		return
	}
	if bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)) != nil {
		errorResponse(c, http.StatusUnauthorized, "用户名或密码错误")
		return
	}
	jwtConfig := NewJWTConfig()
	if h.cfg != nil {
		jwtConfig = &h.cfg.JWT
	}
	token, err := GenerateJWT(strconv.FormatUint(uint64(user.ID), 10), user.Username, jwtConfig)
	if err != nil {
		errorResponse(c, http.StatusInternalServerError, "生成token失败")
		return
	}

	successResponse(c, gin.H{
		"token": token,
		"user": gin.H{
			"id":         user.ID,
			"username":   user.Username,
			"login_time": time.Now().Format("2006-01-02 15:04:05"),
		},
	})
}

// 用户注册
func (h *AuthHandler) handleRegister(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errorResponse(c, http.StatusBadRequest, "请求参数错误: "+err.Error())
		return
	}
	if h == nil || h.repo == nil {
		errorResponse(c, http.StatusInternalServerError, "注册服务未初始化")
		return
	}

	_, err := h.repo.FindUserByUsername(req.Username)
	if err == nil {
		errorResponse(c, http.StatusConflict, "用户名已存在")
		return
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		errorResponse(c, http.StatusInternalServerError, "查询用户失败")
		return
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		errorResponse(c, http.StatusInternalServerError, "密码加密失败")
		return
	}

	user := &models.User{
		Username:     req.Username,
		PasswordHash: string(passwordHash),
		Status:       1,
	}
	if err := h.repo.Create(user); err != nil {
		errorResponse(c, http.StatusInternalServerError, "创建用户失败")
		return
	}

	successResponse(c, gin.H{
		"message": "用户注册成功",
		"user": gin.H{
			"id":            user.ID,
			"username":      req.Username,
			"email":         req.Email,
			"register_time": time.Now().Format("2006-01-02 15:04:05"),
		},
	})
}

// Token验证
func (h *AuthHandler) handleTokenVerify(c *gin.Context) {
	token := c.GetHeader("Authorization")
	if token == "" {
		errorResponse(c, http.StatusUnauthorized, "缺少认证token")
		return
	}

	// TODO: 实现实际的token验证逻辑
	// 1. 解析JWT token
	// 2. 验证token有效性
	// 3. 检查token是否过期

	successResponse(c, gin.H{
		"message": "Token验证成功",
		"valid":   true,
	})
}

// ==================== 音频检测相关处理函数 ====================

// 获取检测结果
func handleGetResult(c *gin.Context) {
	taskID := c.Param("id")
	if taskID == "" {
		errorResponse(c, http.StatusBadRequest, "任务ID不能为空")
		return
	}

	// TODO: 实现实际的查询逻辑
	// 1. 从数据库查询检测结果
	// 2. 检查缓存
	// 3. 返回结果

	// 模拟检测结果
	result := gin.H{
		"task_id":        taskID,
		"status":         "completed",
		"result":         "检测到虫害",
		"confidence":     0.85,
		"detection_time": time.Now().Format("2006-01-02 15:04:05"),
		"details": gin.H{
			"pest_type":      "天牛",
			"severity":       "中等",
			"recommendation": "建议及时处理",
		},
	}

	successResponse(c, result)
}

// 获取检测状态
func handleGetStatus(c *gin.Context) {
	taskID := c.Param("id")
	if taskID == "" {
		errorResponse(c, http.StatusBadRequest, "任务ID不能为空")
		return
	}

	// TODO: 实现实际的状态查询逻辑
	// 1. 查询任务状态
	// 2. 返回进度信息

	status := gin.H{
		"task_id":        taskID,
		"status":         "processing",
		"progress":       75,
		"estimated_time": "2分钟",
		"update_time":    time.Now().Format("2006-01-02 15:04:05"),
	}

	successResponse(c, status)
}

// ==================== 设备管理相关处理函数 ====================

// 设备列表
func handleDeviceList(c *gin.Context) {
	// TODO: 实现实际的设备列表查询逻辑
	// 1. 分页查询
	// 2. 过滤条件
	// 3. 返回设备列表

	devices := []gin.H{
		{
			"device_id":   "dev_001",
			"device_name": "检测设备A",
			"location":    "果园A区",
			"status":      "online",
			"last_active": time.Now().Add(-5 * time.Minute).Format("2006-01-02 15:04:05"),
		},
		{
			"device_id":   "dev_002",
			"device_name": "检测设备B",
			"location":    "果园B区",
			"status":      "offline",
			"last_active": time.Now().Add(-2 * time.Hour).Format("2006-01-02 15:04:05"),
		},
	}

	successResponse(c, gin.H{
		"total":   len(devices),
		"devices": devices,
	})
}

// 设备信息
func handleDeviceInfo(c *gin.Context) {
	deviceID := c.Param("id")
	if deviceID == "" {
		errorResponse(c, http.StatusBadRequest, "设备ID不能为空")
		return
	}

	// TODO: 实现实际的设备信息查询逻辑
	// 1. 查询设备详细信息
	// 2. 查询设备历史记录
	// 3. 返回设备状态

	deviceInfo := gin.H{
		"device_id":        deviceID,
		"device_name":      "检测设备A",
		"location":         "果园A区",
		"status":           "online",
		"firmware_version": "v1.2.3",
		"last_maintenance": "2024-01-15",
		"total_detections": 156,
		"last_active":      time.Now().Add(-5 * time.Minute).Format("2006-01-02 15:04:05"),
	}

	successResponse(c, deviceInfo)
}

// 设备注册
func handleDeviceRegister(c *gin.Context) {
	var req DeviceRegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errorResponse(c, http.StatusBadRequest, "请求参数错误: "+err.Error())
		return
	}

	// TODO: 实现实际的设备注册逻辑
	// 1. 检查设备ID是否已存在
	// 2. 创建设备记录
	// 3. 分配初始配置

	successResponse(c, gin.H{
		"message": "设备注册成功",
		"device": gin.H{
			"device_id":     req.DeviceID,
			"device_name":   req.DeviceName,
			"location":      req.Location,
			"register_time": time.Now().Format("2006-01-02 15:04:05"),
			"status":        "registered",
		},
	})
}
