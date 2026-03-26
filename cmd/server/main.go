package main

import (
	"fmt"
	"log/slog"
	"mengri-flow/internal/infra/config"
	"mengri-flow/internal/infra/persistence/mysql"
	"mengri-flow/internal/ports/http/router"
	"mengri-flow/pkg/autowire"
	"mengri-flow/pkg/logger"
	"mengri-flow/pkg/register"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"gorm.io/gorm"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		slog.Warn("No .env file found, relying on environment variables")
	}
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
	logger.Setup(cfg.Log.Level, cfg.Log.Format)
	slog.Info("config loaded successfully", "path", cfgPath)
	slog.Info("config details", "config", fmt.Sprintf("%+v", cfg))
	// 2. 初始化数据库
	db, err := mysql.NewDB(&cfg.Database)
	if err != nil {
		slog.Error("failed to connect database", "error", err)
		os.Exit(1)
	}

	autowire.Auto(func() *gorm.DB { return db })
	router := &router.Router{}
	autowire.Autowired(router)
	autowire.Check()
	// 3. 自动迁移（仅开发环境）
	if cfg.Server.Mode == "debug" {
		register.Do("AutoMigrateOnDebug", func(model any) error {

			if err := db.AutoMigrate(model); err != nil {
				return err
			}

			return nil
		})
	}

	autowire.PostEvent(autowire.OnCompleteEvent)

	// 6. 启动 HTTP 服务
	gin.SetMode(cfg.Server.Mode)
	engine := gin.New()
	router.Setup(engine)

	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	slog.Info("server starting", "addr", addr, "mode", cfg.Server.Mode)

	if err := engine.Run(addr); err != nil {
		slog.Error("server failed to start", "error", err)
		os.Exit(1)
	}
}
