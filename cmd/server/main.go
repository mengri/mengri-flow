package main

import (
	"fmt"
	"log/slog"
	"mengri-flow/internal/app/service"
	"mengri-flow/internal/infra/config"
	"mengri-flow/internal/infra/persistence/mysql"
	"mengri-flow/internal/ports/http/handler"
	"mengri-flow/internal/ports/http/router"
	"mengri-flow/web"
	"os"

	"github.com/gin-gonic/gin"
)

func main() {
	// 1. 加载配置
	cfgPath := "config.yaml"
	if envPath := os.Getenv("CONFIG_PATH"); envPath != "" {
		cfgPath = envPath
	}

	cfg, err := config.Load(cfgPath)
	if err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	// 2. 初始化数据库
	db, err := mysql.NewDB(&cfg.Database)
	if err != nil {
		slog.Error("failed to connect database", "error", err)
		os.Exit(1)
	}

	// 3. 自动迁移（仅开发环境）
	if cfg.Server.Mode == "debug" {
		if err := db.AutoMigrate(&mysql.UserModel{}); err != nil {
			slog.Error("failed to auto migrate", "error", err)
			os.Exit(1)
		}
	}

	// 4. 依赖注入（手动组装）
	userRepo := mysql.NewUserRepository(db)
	userService := service.NewUserService(userRepo)
	userHandler := handler.NewUserHandler(userService)

	// 5. 加载内嵌的前端产物
	frontendFS, err := web.DistFS()
	if err != nil {
		slog.Warn("failed to load embedded frontend", "error", err)
	}

	// 6. 启动 HTTP 服务
	gin.SetMode(cfg.Server.Mode)
	engine := gin.New()
	router.Setup(engine, userHandler, frontendFS)

	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	slog.Info("server starting", "addr", addr, "mode", cfg.Server.Mode)

	if err := engine.Run(addr); err != nil {
		slog.Error("server failed to start", "error", err)
		os.Exit(1)
	}
}
