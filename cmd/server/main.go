package main

import (
	"fmt"
	"log/slog"
	"mengri-flow/internal/domain/repository"
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
	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		slog.Warn("No .env file found, relying on environment variables")
	}
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

	// --- Database ---
	db, err := mysql.NewDB(&cfg.Database)
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
	slog.Info("redis client created", "addr", cfg.Redis.Addr)

	// --- JWT Manager ---
	jwtMgr := auth.NewJWTManager(&cfg.Auth.JWT)

	// --- TransactionManager ---
	txManager := mysql.NewTransactionManager(db)

	// --- Email Sender ---
	emailSender := external.NewSMTPEmailSender(&cfg.Email)

	// --- Cache Stores ---
	ticketStore := cache.NewSecurityTicketStore(rdb, cfg.Auth.SecurityTicketTTL)

	// --- Autowire Registration ---
	autowire.Auto(func() *gorm.DB { return db })
	autowire.Auto(func() *redis.Client { return rdb })
	autowire.Auto(func() *auth.JWTManager { return jwtMgr })
	autowire.Auto(func() repository.TransactionManager { return txManager })
	autowire.Auto(func() repository.EmailSender { return emailSender })
	autowire.Auto(func() *cache.SecurityTicketStore { return ticketStore })
	autowire.Auto(func() *config.Config { return cfg })

	r := &router.Router{}
	autowire.Autowired(r)
	autowire.Check()

	if cfg.Server.Mode == "debug" {
		mysql.MigrateOnDebug(db)
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
