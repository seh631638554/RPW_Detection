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

	if err := container.Provide(func(db *gorm.DB) *dao.Repo {
		return dao.New(db)
	}); err != nil {
		log.Fatalf("Repo初始化失败: %v", err)
	}

	if err := container.Provide(func(repo *dao.Repo, cfg *httpserver.Config) *httpserver.AuthHandler {
		return httpserver.NewAuthHandler(repo, cfg)
	}); err != nil {
		log.Fatalf("AuthHandler初始化失败: %v", err)
	}

	if err := container.Provide(func() *httpserver.ObjectStorageConfig {
		return httpserver.LoadObjectStorageConfig()
	}); err != nil {
		log.Fatalf("对象存储配置初始化失败: %v", err)
	}

	if err := container.Provide(func(cfg *httpserver.ObjectStorageConfig) (httpserver.StorageService, error) {
		return httpserver.NewMinIOStorageService(cfg)
	}); err != nil {
		log.Fatalf("存储服务初始化失败: %v", err)
	}

	if err := container.Provide(func(storage httpserver.StorageService, cfg *httpserver.ObjectStorageConfig) *httpserver.UploadService {
		return httpserver.NewUploadService(storage, cfg)
	}); err != nil {
		log.Fatalf("上传服务初始化失败: %v", err)
	}

	if err := container.Provide(func(svc *httpserver.UploadService, repo *dao.Repo, cfg *httpserver.Config) *httpserver.UploadHandler {
		return httpserver.NewUploadHandler(svc, repo, cfg)
	}); err != nil {
		log.Fatalf("上传处理器初始化失败: %v", err)
	}

	if err := container.Provide(func(repo *dao.Repo) *httpserver.ParkHandler {
		return httpserver.NewParkHandler(repo)
	}); err != nil {
		log.Fatalf("ParkHandler初始化失败: %v", err)
	}

	if err := container.Provide(func(repo *dao.Repo) *httpserver.TreeHandler {
		return httpserver.NewTreeHandler(repo)
	}); err != nil {
		log.Fatalf("TreeHandler初始化失败: %v", err)
	}

	if err := container.Provide(func(repo *dao.Repo) *httpserver.DeviceHandler {
		return httpserver.NewDeviceHandler(repo)
	}); err != nil {
		log.Fatalf("DeviceHandler初始化失败: %v", err)
	}

	if err := container.Provide(func(auth *httpserver.AuthHandler, upload *httpserver.UploadHandler, park *httpserver.ParkHandler, tree *httpserver.TreeHandler, device *httpserver.DeviceHandler) *httpserver.Router {
		return httpserver.NewRouter(auth, upload, park, tree, device)
	}); err != nil {
		log.Fatalf("Router初始化失败: %v", err)
	}

	if err := container.Invoke(func(cfg *httpserver.Config, engine *gin.Engine, router *httpserver.Router, repo *dao.Repo) {
		httpserver.StartClassificationResultConsumer(cfg, repo)
		router.Register(engine)
		_ = httpserver.StartServer(cfg, engine)
	}); err != nil {
		log.Fatalf("容器启动失败: %v", err)
	}
}
