package main

import (
	"bufio"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"mengri-flow/internal/domain/entity"
	"mengri-flow/internal/domain/valueobject"
	"mengri-flow/internal/executor"
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

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/redis/go-redis/v9"
	"golang.org/x/text/language"
	"gorm.io/gorm"
)

func main() {
	// 加载环境变量
	godotenv.Load()

	// 解析子命令
	if len(os.Args) < 2 {
		// 默认启动console
		runConsole(getConfigPath())
		return
	}

	// 检查第一个参数是否是子命令
	switch os.Args[1] {
	case "node":
		// 启动executor
		if err := runExecutorCommand(); err != nil {
			log.Fatal(err)
		}
	case "-h", "--help", "help":
		printHelp()
	case "-v", "--version", "version":
		printVersion()
	default:
		// 如果第一个参数不是node，则认为是console的启动参数
		// 需要重新解析console的参数
		if strings.HasPrefix(os.Args[1], "-") {
			// 是参数，不是子命令，启动console
			runConsole(parseConsoleFlags())
		} else {
			log.Fatalf("Unknown command: %s. Use 'node' to start executor or no command to start console", os.Args[1])
		}
	}
}

// parseConsoleFlags 解析console的参数
func parseConsoleFlags() string {
	cfgPath := flag.String("config", "config.yaml", "配置文件路径")
	flag.Parse()

	if envPath := os.Getenv("CONFIG_PATH"); envPath != "" {
		*cfgPath = envPath
	}
	return *cfgPath
}

// getConfigPath 获取配置文件路径
func getConfigPath() string {
	// 检查是否有 --config 参数
	for i, arg := range os.Args {
		if arg == "--config" && i+1 < len(os.Args) {
			return os.Args[i+1]
		}
		if strings.HasPrefix(arg, "--config=") {
			return strings.TrimPrefix(arg, "--config=")
		}
	}

	// 检查环境变量
	if envPath := os.Getenv("CONFIG_PATH"); envPath != "" {
		return envPath
	}

	return "config.yaml"
}

// runExecutorCommand 运行executor子命令
func runExecutorCommand() error {
	// 为executor创建新的flag set
	executorFlags := flag.NewFlagSet("node", flag.ExitOnError)

	var (
		etcdEndpoints = executorFlags.String("etcd-endpoints", "", "etcd endpoints (必需)")
		etcdUsername  = executorFlags.String("etcd-username", "", "etcd username")
		etcdPassword  = executorFlags.String("etcd-password", "", "etcd password")
		clusterID     = executorFlags.String("cluster-id", "", "cluster ID (必需)")
		nodeID        = executorFlags.String("node-id", "", "executor node ID (可选，自动生成)")
		port          = executorFlags.Int("port", 0, "RESTful触发器端口 (可选)")
		logLevel      = executorFlags.String("log-level", "info", "日志级别 (debug|info|warn|error)")
	)

	// 解析executor的参数（跳过第一个参数"node"）
	if err := executorFlags.Parse(os.Args[2:]); err != nil {
		return err
	}

	// 验证必需参数
	if *etcdEndpoints == "" {
		return fmt.Errorf("--etcd-endpoints is required for executor")
	}
	if *clusterID == "" {
		return fmt.Errorf("--cluster-id is required for executor")
	}

	// 启动executor
	runExecutor(&ExecutorCLIConfig{
		EtcdEndpoints: *etcdEndpoints,
		EtcdUsername:  *etcdUsername,
		EtcdPassword:  *etcdPassword,
		ClusterID:     *clusterID,
		NodeID:        *nodeID,
		Port:          *port,
		LogLevel:      *logLevel,
	})
	return nil
}

// printHelp 打印帮助信息
func printHelp() {
	help := `Mengri Flow - API编排平台

使用方式:
  mengri-flow [命令] [选项]

命令:
  <无命令>        启动Console（控制台），默认行为
  node            启动Executor（执行器）
  help, --help    显示帮助信息
  version, --version 显示版本信息

Console选项:
  --config string    配置文件路径 (默认 "config.yaml")

Executor选项（node命令）:
  --etcd-endpoints string    etcd集群地址，多个用逗号分隔 (必需)
  --cluster-id string        集群ID (必需)
  --node-id string           执行器节点ID (可选，自动生成)
  --port int                RESTful触发器端口 (可选)
  --etcd-username string     etcd用户名 (可选)
  --etcd-password string     etcd密码 (可选)
  --log-level string         日志级别: debug|info|warn|error (默认 "info")

示例:
  # 启动Console（默认）
  ./mengri-flow
  ./mengri-flow --config=config.yaml

  # 启动Executor
  ./mengri-flow node --etcd-endpoints=etcd:2379 --cluster-id=cluster-prod-001
  ./mengri-flow node --etcd-endpoints=etcd:2379 --cluster-id=cluster-prod-001 --node-id=executor-1

更多信息请参考: docs/dual-role-deployment.md
`
	fmt.Println(help)
	os.Exit(0)
}

// printVersion 打印版本信息
func printVersion() {
	fmt.Println("Mengri Flow v1.0.0")
	os.Exit(0)
}

// ExecutorCLIConfig 执行器命令行配置
type ExecutorCLIConfig struct {
	EtcdEndpoints string
	EtcdUsername  string
	EtcdPassword  string
	ClusterID     string
	NodeID        string
	Port          int
	LogLevel      string
}

// runConsole 启动控制台服务
func runConsole(cfgPath string) {
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
	cache.GenSecurityTicketStore(cfg.Auth.SecurityTicketTTL)

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
	slog.Info("server starting", "addr", addr, "mode", cfg.Server.Mode, "role", "console")

	if err := engine.Run(addr); err != nil {
		slog.Error("server failed to start", "error", err)
		os.Exit(1)
	}
}

// runExecutor 启动执行器服务
func runExecutor(cliCfg *ExecutorCLIConfig) {
	// 验证必需参数
	if cliCfg.ClusterID == "" {
		log.Fatal("cluster-id is required for executor role")
	}

	// 生成Node ID（如果不指定）
	if cliCfg.NodeID == "" {
		cliCfg.NodeID = generateNodeID()
		log.Printf("Auto-generated node ID: %s", cliCfg.NodeID)
	}

	// 加载配置
	executorCfg := &config.ExecutorConfig{
		EtcdEndpoints: cliCfg.EtcdEndpoints,
		EtcdUsername:  cliCfg.EtcdUsername,
		EtcdPassword:  cliCfg.EtcdPassword,
		ClusterID:     cliCfg.ClusterID,
		NodeID:        cliCfg.NodeID,
		LogLevel:      "info",
	}

	if cliCfg.EtcdEndpoints == "" {
		log.Fatal("etcd-endpoints is required for executor role")
	}

	// 设置日志级别
	logger.Setup(executorCfg.LogLevel, "json")

	// 初始化Executor
	exec := executor.NewExecutor(executorCfg)

	// 启动
	ctx := context.Background()
	if err := exec.Start(ctx); err != nil {
		log.Fatal("Failed to start executor:", err)
	}

	log.Printf("Executor %s started successfully for cluster %s", cliCfg.NodeID, cliCfg.ClusterID)

	// 等待信号
	waitForShutdownSignal()

	// 优雅关闭
	log.Printf("Shutting down executor...")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := exec.Stop(shutdownCtx); err != nil {
		log.Fatal("Failed to stop executor:", err)
	}

	log.Printf("Executor stopped gracefully")
}

// generateNodeID 生成执行器节点ID
func generateNodeID() string {
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "unknown"
	}
	return fmt.Sprintf("%s-%d", hostname, time.Now().Unix())
}

// waitForShutdownSignal 等待中断信号
func waitForShutdownSignal() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan
	log.Printf("Received shutdown signal")
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
