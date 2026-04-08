package main

import (
	"bufio"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"encoding/json"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
	"mengri-flow/internal/domain/entity"
	"mengri-flow/internal/domain/valueobject"
	"mengri-flow/internal/infra/auth"
	"mengri-flow/internal/infra/cache"
	"mengri-flow/internal/infra/config"
	"mengri-flow/internal/infra/external"
	"mengri-flow/internal/infra/external/oauth"
	"mengri-flow/internal/infra/persistence/mysql"
	accountRepository "mengri-flow/internal/infra/persistence/mysql/account_repository"
	credentialRepository "mengri-flow/internal/infra/persistence/mysql/credential_repository"
	identityRepository "mengri-flow/internal/infra/persistence/mysql/identity_repository"
	"mengri-flow/internal/infra/plugin"
	"mengri-flow/internal/ports/http/router"
	"mengri-flow/pkg/autowire"
	"mengri-flow/pkg/logger"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

func main() {

	cfgPath := "config.yaml"
	godotenv.Load()

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

	// --- 设置启用的插件 ---
	plugin.GlobalRegistry().SetEnabledPlugins(cfg.Plugins.Enabled)
	slog.Info("plugins enabled", "count", len(cfg.Plugins.Enabled), "plugins", cfg.Plugins.Enabled)

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

	// --- SMS Sender ---
	if err := external.GenAliyunSMSSender(&cfg.SMS); err != nil {
		slog.Error("failed to create sms sender", "error", err)
		os.Exit(1)
	}

	// --- Cache Stores ---
	cache.GenSecurityTicketStore(rdb, cfg.Auth.SecurityTicketTTL)

	// --- OAuth Providers ---
	oauth.InitOAuthProviders(&cfg.OAuth)

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

	if err := ensureInitialAdminAccount(db); err != nil {
		slog.Error("failed to initialize admin account", "error", err)
		os.Exit(1)
	}

	autowire.PostEvent(autowire.OnCompleteEvent)

	gin.SetMode(cfg.Server.Mode)
	engine := gin.New()
	// 初始化国际化
	b := i18n.NewBundle(language.English)
	b.RegisterUnmarshalFunc("json", json.Unmarshal)
	b.LoadMessageFile("active.en.json")
	b.LoadMessageFile("active.zh.json")

	// 添加中间件解析语言
	engine.Use(func(c *gin.Context) {
		lang := c.Query("lang") // 默认语言
		if lang == "" {
			lang = "en"
		}
		c.Set("Localizer", i18n.NewLocalizer(b, lang))
		c.Next()
	})

	// 示例路由
	engine.GET("/api/v1/greet", func(c *gin.Context) {
		localizer := c.MustGet("Localizer").(*i18n.Localizer)
		greeting := localizer.MustLocalize(&i18n.LocalizeConfig{
			MessageID: "WelcomeMessage",
		})
		c.JSON(200, gin.H{"message": greeting})
	})

	// 设置路由
	r.Setup(engine)

	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	slog.Info("server starting", "addr", addr, "mode", cfg.Server.Mode)

	if err := engine.Run(addr); err != nil {
		slog.Error("server failed to start", "error", err)
		os.Exit(1)
	}
}

func ensureInitialAdminAccount(db *mysql.MysqlDb) error {
	var adminCount int64
	if err := db.DB.Model(&accountRepository.AccountModel{}).Where("role = ?", "admin").Count(&adminCount).Error; err != nil {
		return fmt.Errorf("count admin accounts: %w", err)
	}

	if adminCount > 0 {
		return nil
	}

	fmt.Println("未检测到管理员账号，开始初始化。")
	reader := bufio.NewReader(os.Stdin)

	for {
		email, err := promptLine(reader, "请输入管理员邮箱: ")
		if err != nil {
			return fmt.Errorf("read admin email: %w", err)
		}
		if _, err := valueobject.NewEmail(email); err != nil {
			fmt.Printf("邮箱格式不正确: %v\n", err)
			continue
		}

		password, err := promptLine(reader, "请输入管理员密码(至少8位且包含大小写字母、数字、特殊字符): ")
		if err != nil {
			return fmt.Errorf("read admin password: %w", err)
		}
		if err := valueobject.ValidatePasswordStrength(password); err != nil {
			fmt.Printf("密码强度不满足要求: %v\n", err)
			continue
		}

		if err := createInitialAdminAccount(db, email, password); err != nil {
			fmt.Printf("创建管理员账号失败: %v\n请重新输入。\n", err)
			continue
		}

		fmt.Println("管理员账号初始化完成。")
		return nil
	}
}

func promptLine(reader *bufio.Reader, label string) (string, error) {
	fmt.Print(label)
	line, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(line), nil
}

func createInitialAdminAccount(db *mysql.MysqlDb, email, password string) error {
	hashedPassword, err := auth.HashPassword(password, 12)
	if err != nil {
		return fmt.Errorf("hash admin password: %w", err)
	}

	now := time.Now()
	return db.DB.Transaction(func(tx *gorm.DB) error {
		var adminCount int64
		if err := tx.Model(&accountRepository.AccountModel{}).Where("role = ?", "admin").Count(&adminCount).Error; err != nil {
			return fmt.Errorf("count admin accounts in transaction: %w", err)
		}
		if adminCount > 0 {
			return nil
		}

		accountID := uuid.New().String()
		activatedAt := now

		accountModel := &accountRepository.AccountModel{
			ID:          accountID,
			Email:       email,
			Username:    "admin_" + strings.ReplaceAll(uuid.New().String()[:8], "-", ""),
			DisplayName: "System Admin",
			Status:      string(entity.AccountStatusActive),
			Role:        "admin",
			ActivatedAt: &activatedAt,
			CreatedAt:   now,
			UpdatedAt:   now,
		}

		credentialModel := &credentialRepository.CredentialModel{
			AccountID:         accountID,
			PasswordHash:      hashedPassword,
			PasswordUpdatedAt: now,
		}

		identityModel := &identityRepository.IdentityModel{
			ID:         uuid.New().String(),
			AccountID:  accountID,
			LoginType:  string(entity.LoginTypePassword),
			ExternalID: email,
			CreatedAt:  now,
		}

		if err := tx.Create(accountModel).Error; err != nil {
			return fmt.Errorf("create admin account: %w", err)
		}
		if err := tx.Create(credentialModel).Error; err != nil {
			return fmt.Errorf("create admin credential: %w", err)
		}
		if err := tx.Create(identityModel).Error; err != nil {
			return fmt.Errorf("create admin identity: %w", err)
		}

		return nil
	})
}
