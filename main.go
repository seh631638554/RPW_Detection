package main

import (
	httpserver "RPW_Detection/Http"
	"log"

	"go.uber.org/dig"
)

func main() {
	container := dig.New()

	cfg := httpserver.LoadConfig()
	container.Provide(cfg)
	engine := httpserver.NewGinEngine(cfg)
	container.Provide(engine)
	httpserver.SetupRoutes(engine)
	if err := httpserver.StartServer(cfg, engine); err != nil {
		log.Fatalf("服务器启动失败", err)
	}
}
