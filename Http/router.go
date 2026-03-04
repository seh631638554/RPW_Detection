package httpserver

import "github.com/gin-gonic/gin"

type Router struct {
	auth *AuthHandler
}

func NewRouter(auth *AuthHandler) *Router {
	return &Router{auth: auth}
}
func (r *Router) Register(engine *gin.Engine) {
	api := engine.Group("/api/v1")
	// health...
	auth := api.Group("/auth")
	{
		auth.POST("/login", r.auth.handleLogin)
		auth.POST("/register", r.auth.handleRegister) // 先保留原全局
		auth.GET("/verify", r.auth.handleTokenVerify) // 先保留原全局
	}
	// detection/jobs/device 先照旧
}
