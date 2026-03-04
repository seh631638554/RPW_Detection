package main

import (
	dao "RPW_Detection/Dao"
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

	container.Provide(func(db *gorm.DB) *dao.Repo { return dao.New(db) })
	container.Provide(func(repo *dao.Repo, cfg *httpserver.Config) *httpserver.AuthHandler {
		return httpserver.NewAuthHandler(repo, cfg)
	})
	container.Provide(func(auth *httpserver.AuthHandler) *httpserver.Router {
		return httpserver.NewRouter(auth)
	})
	if err := container.Invoke(func(cfg *httpserver.Config, engine *gin.Engine, router *httpserver.Router) {
		router.Register(engine)
		_ = httpserver.StartServer(cfg, engine)
	}); err != nil {
		log.Fatalf("容器启动失败: %v", err)
	}
}
