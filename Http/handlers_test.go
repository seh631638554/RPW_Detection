package httpserver

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// 设置测试环境
func setupTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// 设置测试路由
	api := router.Group("/api/v1")
	api.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "ok",
			"message": "病虫害检测服务器运行正常",
		})
	})

	return router
}

// 测试健康检查接口
func TestHealthCheck(t *testing.T) {
	router := setupTestRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/health", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "病虫害检测服务器运行正常")
}

// 测试登录接口
func TestLoginHandler(t *testing.T) {
	router := setupTestRouter()

	// 添加登录路由
	api := router.Group("/api/v1")
	api.POST("/auth/login", handleLogin)

	// 测试有效登录
	validLogin := `{"username":"testuser","password":"testpass"}`
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/auth/login", strings.NewReader(validLogin))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "操作成功")
}

// 测试无效JSON登录
func TestLoginInvalidJSON(t *testing.T) {
	router := setupTestRouter()

	// 添加登录路由
	api := router.Group("/api/v1")
	api.POST("/auth/login", handleLogin)

	// 测试无效JSON
	invalidJSON := `{"username":"testuser","password":}`
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/auth/login", strings.NewReader(invalidJSON))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "请求参数错误")
}

// 测试设备列表接口
func TestDeviceListHandler(t *testing.T) {
	router := setupTestRouter()

	// 添加设备列表路由
	api := router.Group("/api/v1")
	api.GET("/device/list", handleDeviceList)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/device/list", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "操作成功")
	assert.Contains(t, w.Body.String(), "devices")
}

// 测试设备信息接口
func TestDeviceInfoHandler(t *testing.T) {
	router := setupTestRouter()

	// 添加设备信息路由
	api := router.Group("/api/v1")
	api.GET("/device/:id", handleDeviceInfo)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/device/dev_001", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "操作成功")
	assert.Contains(t, w.Body.String(), "dev_001")
}

// 测试检测结果接口
func TestDetectionResultHandler(t *testing.T) {
	router := setupTestRouter()

	// 添加检测结果路由
	api := router.Group("/api/v1")
	api.GET("/detection/result/:id", handleGetResult)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/detection/result/task_123", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "操作成功")
	assert.Contains(t, w.Body.String(), "task_123")
}

// 测试检测状态接口
func TestDetectionStatusHandler(t *testing.T) {
	router := setupTestRouter()

	// 添加检测状态路由
	api := router.Group("/api/v1")
	api.GET("/detection/status/:id", handleGetStatus)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/detection/status/task_123", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "操作成功")
	assert.Contains(t, w.Body.String(), "task_123")
}

// 测试响应结构体
func TestResponseStruct(t *testing.T) {
	response := Response{
		Code:    200,
		Message: "测试消息",
		Data:    gin.H{"key": "value"},
		Time:    "2024-01-01 12:00:00",
	}

	assert.Equal(t, 200, response.Code)
	assert.Equal(t, "测试消息", response.Message)
	assert.NotNil(t, response.Data)
	assert.Equal(t, "2024-01-01 12:00:00", response.Time)
}

// 测试请求结构体
func TestRequestStructs(t *testing.T) {
	// 测试登录请求
	loginReq := LoginRequest{
		Username: "testuser",
		Password: "testpass",
	}
	assert.Equal(t, "testuser", loginReq.Username)
	assert.Equal(t, "testpass", loginReq.Password)

	// 测试注册请求
	registerReq := RegisterRequest{
		Username: "testuser",
		Password: "testpass",
		Email:    "test@example.com",
	}
	assert.Equal(t, "testuser", registerReq.Username)
	assert.Equal(t, "testpass", registerReq.Password)
	assert.Equal(t, "test@example.com", registerReq.Email)

	// 测试音频上传请求
	audioReq := AudioUploadRequest{
		DeviceID:  "dev_001",
		AudioType: "wav",
		Timestamp: 1640995200,
	}
	assert.Equal(t, "dev_001", audioReq.DeviceID)
	assert.Equal(t, "wav", audioReq.AudioType)
	assert.Equal(t, int64(1640995200), audioReq.Timestamp)
}
