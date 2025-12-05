package httpserver

import (
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// 创建Gin引擎
func NewGinEngine(config *Config) *gin.Engine {
	// 设置运行模式
	gin.SetMode(config.Server.Mode)

	// 创建引擎
	engine := gin.New()

	// 使用中间件
	engine.Use(RequestLoggerMiddleware()) // 自定义日志中间件
	engine.Use(ErrorHandlerMiddleware())  // 错误处理中间件
	engine.Use(RequestIDMiddleware())     // 请求ID中间件
	engine.Use(PerformanceMiddleware())   // 性能监控中间件
	engine.Use(CORSMiddleware())          // CORS中间件
	// engine.Use(CheckLegal())              //检测用户合法性
	return engine
}

// 设置路由
func SetupRoutes(engine *gin.Engine) {
	// API版本组
	api := engine.Group("/api/v1")

	// 健康检查
	api.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "ok",
			"message": "病虫害检测服务器运行正常",
			"time":    time.Now().Format("2006-01-02 15:04:05"),
		})
	})

	// 用户认证相关路由
	auth := api.Group("/auth")
	{
		auth.POST("/login", handleLogin)
		auth.POST("/register", handleRegister)
		auth.GET("/verify", handleTokenVerify)
	}

	// 音频检测相关路由
	detection := api.Group("/detection")
	{
		detection.POST("/upload", handleAudioUpload)
		detection.GET("/result/:id", handleGetResult)
		detection.GET("/status/:id", handleGetStatus)
	}

	// 文件上传任务管理路由
	jobs := api.Group("/jobs")
	{
		jobs.POST("", CreateUploadJob)                      // 创建上传任务
		jobs.GET("", ListUploadJobs)                        // 列出上传任务
		jobs.GET("/:id", GetUploadJobStatus)                // 获取任务状态
		jobs.DELETE("/:id", DeleteUploadJob)                // 删除任务
		jobs.POST("/:id/complete", UploadCompletionWebhook) // 上传完成回调
	}

	// 设备管理相关路由
	device := api.Group("/device")
	{
		device.GET("/list", handleDeviceList)
		device.GET("/:id", handleDeviceInfo)
		device.POST("/register", handleDeviceRegister)
	}
}

// 启动服务器
func StartServer(config *Config, engine *gin.Engine) error {
	// 初始化存储服务
	if err := InitStorageService(); err != nil {
		log.Printf("初始化存储服务失败: %v", err)
		// 注意：这里不返回错误，因为存储服务不是必需的
	}

	port := ":" + config.Server.Port
	log.Printf("启动病虫害检测服务器，监听端口: %s", port)
	log.Printf("服务器模式: %s", config.Server.Mode)
	log.Printf("数据库地址: %s:%d", config.Database.Host, config.Database.Port)
	log.Printf("Redis地址: %s:%d", config.Redis.Host, config.Redis.Port)
	log.Printf("Kafka brokers: %s", config.Kafka.GetBrokersString())

	return engine.Run(port)
}

// func main() {
// 	// 加载配置
// 	config := LoadConfig()

// 	// 创建Gin引擎
// 	engine := NewGinEngine(config)

// 	// 设置路由
// 	SetupRoutes(engine)

// 	// 启动服务器
// 	if err := StartServer(config, engine); err != nil {
// 		log.Fatalf("服务器启动失败: %v", err)
// 	}
// }
