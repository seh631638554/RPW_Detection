package httpserver

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type Router struct {
	auth   *AuthHandler
	upload *UploadHandler
}

func NewRouter(auth *AuthHandler, upload *UploadHandler) *Router {
	return &Router{
		auth:   auth,
		upload: upload,
	}
}

func (r *Router) Register(engine *gin.Engine) {
	api := engine.Group("/api/v1")

	api.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "ok",
			"message": "病虫害检测服务器运行正常",
			"time":    time.Now().Format("2006-01-02 15:04:05"),
		})
	})

	auth := api.Group("/auth")
	{
		auth.POST("/login", r.auth.handleLogin)
		auth.POST("/register", r.auth.handleRegister)
		auth.GET("/verify", r.auth.handleTokenVerify)
	}

	detection := api.Group("/detection")
	{
		detection.GET("/result/:id", handleGetResult)
		detection.GET("/status/:id", handleGetStatus)
	}

	jobs := api.Group("/jobs")
	{
		if r.upload != nil {
			jobs.POST("", r.upload.HandleCreateUploadJob)
		} else {
			jobs.POST("", CreateUploadJob)
		}
		jobs.GET("", ListUploadJobs)
		jobs.GET("/:id", GetUploadJobStatus)
		jobs.DELETE("/:id", DeleteUploadJob)
		jobs.POST("/:id/complete", UploadCompletionWebhook)
	}

	device := api.Group("/device")
	{
		device.GET("/list", handleDeviceList)
		device.GET("/:id", handleDeviceInfo)
		device.POST("/register", handleDeviceRegister)
	}
}
