package httpserver

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type Router struct {
	auth   *AuthHandler
	upload *UploadHandler
	park   *ParkHandler
	tree   *TreeHandler
	device *DeviceHandler
}

func NewRouter(auth *AuthHandler, upload *UploadHandler, park *ParkHandler, tree *TreeHandler, device *DeviceHandler) *Router {
	return &Router{
		auth:   auth,
		upload: upload,
		park:   park,
		tree:   tree,
		device: device,
	}
}

func (r *Router) Register(engine *gin.Engine) {
	api := engine.Group("/api/v1")
	jwtConfig := NewJWTConfig()
	if r.auth != nil && r.auth.cfg != nil {
		jwtConfig = &r.auth.cfg.JWT
	}

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
	jobs.Use(JWTAuthMiddleware(jwtConfig))
	{
		if r.upload != nil {
			jobs.POST("", r.upload.HandleCreateUploadJob)
			jobs.GET("", r.upload.HandleListUploadJobs)
			jobs.GET("/:id", r.upload.HandleGetUploadJobStatus)
			jobs.DELETE("/:id", r.upload.HandleDeleteUploadJob)
			jobs.POST("/:id/complete", r.upload.HandleUploadCompletionWebhook)
		} else {
			jobs.POST("", CreateUploadJob)
			jobs.GET("", ListUploadJobs)
			jobs.GET("/:id", GetUploadJobStatus)
			jobs.DELETE("/:id", DeleteUploadJob)
			jobs.POST("/:id/complete", UploadCompletionWebhook)
		}
	}

	parks := api.Group("/parks")
	parks.Use(JWTAuthMiddleware(jwtConfig))
	{
		if r.park != nil {
			parks.GET("", r.park.HandleListParks)
			parks.GET("/:id", r.park.HandleGetPark)
			parks.POST("", r.park.HandleCreatePark)
			parks.PUT("/:id", r.park.HandleUpdatePark)
			parks.DELETE("/:id", r.park.HandleDeletePark)
		}
	}

	trees := api.Group("/trees")
	trees.Use(JWTAuthMiddleware(jwtConfig))
	{
		if r.tree != nil {
			trees.GET("", r.tree.HandleListTrees)
			trees.GET("/:id", r.tree.HandleGetTree)
			trees.POST("", r.tree.HandleCreateTree)
			trees.PUT("/:id", r.tree.HandleUpdateTree)
			trees.DELETE("/:id", r.tree.HandleDeleteTree)
		}
	}

	devices := api.Group("/devices")
	devices.Use(JWTAuthMiddleware(jwtConfig))
	{
		if r.device != nil {
			devices.GET("", r.device.HandleListDevices)
			devices.GET("/:id", r.device.HandleGetDevice)
			devices.POST("", r.device.HandleCreateDevice)
			devices.PUT("/:id", r.device.HandleUpdateDevice)
			devices.DELETE("/:id", r.device.HandleDeleteDevice)
		}
	}

	deviceCompat := api.Group("/device")
	{
		if r.device != nil {
			deviceCompat.GET("/list", r.device.HandleListDevices)
			deviceCompat.GET("/:id", r.device.HandleGetDevice)
			deviceCompat.POST("/register", JWTAuthMiddleware(jwtConfig), r.device.HandleCreateDevice)
		} else {
			deviceCompat.GET("/list", handleDeviceList)
			deviceCompat.GET("/:id", handleDeviceInfo)
			deviceCompat.POST("/register", handleDeviceRegister)
		}
	}
}
