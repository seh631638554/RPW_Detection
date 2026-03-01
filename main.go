package main

import (
	httpserver "RPW_Detection/Http"
	"RPW_Detection/db"
	"log"

	"github.com/gin-gonic/gin"
	"go.uber.org/dig"
	"gorm.io/gorm"
)

func main() {
	container := dig.New()

	if err := container.Provide(func() *httpserver.Config {
		return httpserver.LoadConfig()
	}); err != nil {
		log.Fatalf("配置初始化失败: %v", err)
	}

	if err := container.Provide(func(cfg *httpserver.Config) (*gorm.DB, error) {
		return db.New(
			cfg.Database.GetDSN(),
			cfg.Database.MaxOpenConns,
			cfg.Database.MaxIdleConns,
			cfg.Database.ConnLifetime,
		)
	}); err != nil {
		log.Fatalf("数据库初始化失败: %v", err)
	}

	if err := container.Provide(func(cfg *httpserver.Config) *gin.Engine {
		return httpserver.NewGinEngine(cfg)
	}); err != nil {
		log.Fatalf("Gin引擎初始化失败: %v", err)
	}

	if err := container.Invoke(func(cfg *httpserver.Config, engine *gin.Engine) {
		httpserver.SetupRoutes(engine)
		if err := httpserver.StartServer(cfg, engine); err != nil {
			log.Fatalf("服务器启动失败: %v", err)
		}
	}); err != nil {
		log.Fatalf("容器启动失败: %v", err)
	}
}
