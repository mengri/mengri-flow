package main

import (
	"fmt"
	"log/slog"
	"mengri-flow/internal/infra/auth"
	"mengri-flow/internal/infra/cache"
	"mengri-flow/internal/infra/config"
	"mengri-flow/internal/infra/external"
	"mengri-flow/internal/infra/persistence/mysql"
	"mengri-flow/internal/ports/http/router"
	"mengri-flow/pkg/autowire"
	"mengri-flow/pkg/logger"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

func main() {

	cfgPath := "config.yaml"
	if envPath := os.Getenv("CONFIG_PATH"); envPath != "" {
		cfgPath = envPath
	}

	cfg, err := config.Load(cfgPath)
	if err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}
	cfg.Autowired()
	logger.Setup(cfg.Log.Level, cfg.Log.Format)
	slog.Info("config loaded successfully", "path", cfgPath)
	slog.Info("config details", "config", fmt.Sprintf("%+v", cfg))

	// --- Database ---
	db, err := mysql.GenDB(&cfg.Database)
	if err != nil {
		slog.Error("failed to connect database", "error", err)
		os.Exit(1)
	}

	// --- Redis ---
	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Addr,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})
	autowire.Auto(func() *redis.Client { return rdb })

	slog.Info("redis client created", "addr", cfg.Redis.Addr)

	// --- JWT Manager ---
	auth.GenerateJWTManager(&cfg.Auth.JWT)

	// --- Email Sender ---
	external.GenSMTPEmailSender(&cfg.Email)

	// --- Cache Stores ---
	cache.GenSecurityTicketStore(rdb, cfg.Auth.SecurityTicketTTL)

	r := &router.Router{}
	autowire.Autowired(r)
	autowire.Check()
	if cfg.Server.Mode == "debug" {
		if err := db.MigrateOnDebug(); err != nil {
			slog.Error("failed to auto-migrate database", "error", err)
			os.Exit(1)
		}
		slog.Info("database auto-migration completed")
	}
	autowire.PostEvent(autowire.OnCompleteEvent)

	gin.SetMode(cfg.Server.Mode)
	engine := gin.New()
	r.Setup(engine)

	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	slog.Info("server starting", "addr", addr, "mode", cfg.Server.Mode)

	if err := engine.Run(addr); err != nil {
		slog.Error("server failed to start", "error", err)
		os.Exit(1)
	}
}
