package config

import (
	"fmt"
	"log/slog"
	"os"

	"gopkg.in/yaml.v3"
)

// Config 应用配置
type Config struct {
	Server   ServerConfig   `yaml:"server"`
	Database DatabaseConfig `yaml:"database"`
	Redis    RedisConfig    `yaml:"redis"`
	Log      LogConfig      `yaml:"log"`
	Auth     AuthConfig     `yaml:"auth"`
	OAuth    OAuthConfig    `yaml:"oauth"`
	SMS      SMSConfig      `yaml:"sms"`
	Email    EmailConfig    `yaml:"email"`
}

// ServerConfig HTTP 服务配置
type ServerConfig struct {
	Port int    `yaml:"port"`
	Mode string `yaml:"mode"` // debug, release, test
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Driver          string `yaml:"driver"`
	DSN             string `yaml:"dsn"`
	MaxIdleConns    int    `yaml:"max_idle_conns"`
	MaxOpenConns    int    `yaml:"max_open_conns"`
	ConnMaxLifetime int    `yaml:"conn_max_lifetime"` // 秒
}

// RedisConfig Redis 配置
type RedisConfig struct {
	Addr     string `yaml:"addr"`
	Password string `yaml:"password"`
	DB       int    `yaml:"db"`
}

// LogConfig 日志配置
type LogConfig struct {
	Level  string `yaml:"level"`  // debug, info, warn, error
	Format string `yaml:"format"` // json, text
}

// AuthConfig 认证相关配置
type AuthConfig struct {
	JWT               JWTConfig      `yaml:"jwt"`
	Activation        ActivationConf `yaml:"activation"`
	Password          PasswordConf   `yaml:"password"`
	Lockout           LockoutConf    `yaml:"lockout"`
	SecurityTicketTTL int            `yaml:"security_ticket_ttl"`
}

// JWTConfig JWT 配置
type JWTConfig struct {
	Secret             string `yaml:"secret"`
	AccessTokenExpiry  int    `yaml:"access_token_expiry"`
	RefreshTokenExpiry int    `yaml:"refresh_token_expiry"`
}

// ActivationConf 激活配置
type ActivationConf struct {
	TokenExpiry    int `yaml:"token_expiry"`
	ResendCooldown int `yaml:"resend_cooldown"`
}

// PasswordConf 密码配置
type PasswordConf struct {
	MinLength  int `yaml:"min_length"`
	BcryptCost int `yaml:"bcrypt_cost"`
}

// LockoutConf 锁定配置
type LockoutConf struct {
	MaxAttempts  int `yaml:"max_attempts"`
	LockDuration int `yaml:"lock_duration"`
}

// OAuthConfig OAuth 第三方登录配置
type OAuthConfig struct {
	GitHub OAuthProviderConf `yaml:"github"`
	WeChat OAuthProviderConf `yaml:"wechat"`
	Lark   OAuthProviderConf `yaml:"lark"`
}

// OAuthProviderConf 单个 OAuth 提供方配置
type OAuthProviderConf struct {
	ClientID     string `yaml:"client_id"`
	ClientSecret string `yaml:"client_secret"`
	RedirectURI  string `yaml:"redirect_uri"`
}

// SMSConfig 短信配置
type SMSConfig struct {
	Provider        string `yaml:"provider"`
	AccessKeyID     string `yaml:"access_key_id"`
	AccessKeySecret string `yaml:"access_key_secret"`
	SignName        string `yaml:"sign_name"`
	TemplateCode    string `yaml:"template_code"`
	OTPTTL          int    `yaml:"otp_ttl"`
	OTPCooldown     int    `yaml:"otp_cooldown"`
	OTPLength       int    `yaml:"otp_length"`
}

// EmailConfig 邮件配置
type EmailConfig struct {
	SMTP       SMTPConfig      `yaml:"smtp"`
	Activation ActivationEmail `yaml:"activation"`
}

// SMTPConfig SMTP 配置
type SMTPConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	From     string `yaml:"from"`
}

// ActivationEmail 激活邮件配置
type ActivationEmail struct {
	Subject string `yaml:"subject"`
	BaseURL string `yaml:"base_url"`
}

// Load 从 YAML 文件加载配置
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}
	slog.Info("Loaded config content", "content", string(data))

	// 展开环境变量
	expanded := os.ExpandEnv(string(data))
	slog.Info("Expanded config content", "content", expanded)
	cfg := &Config{}
	if err := yaml.Unmarshal([]byte(expanded), cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	setDefaults(cfg)
	return cfg, nil
}

func setDefaults(cfg *Config) {
	if cfg.Server.Port == 0 {
		cfg.Server.Port = 8080
	}
	if cfg.Server.Mode == "" {
		cfg.Server.Mode = "debug"
	}
	if cfg.Database.MaxIdleConns == 0 {
		cfg.Database.MaxIdleConns = 10
	}
	if cfg.Database.MaxOpenConns == 0 {
		cfg.Database.MaxOpenConns = 100
	}
	if cfg.Database.ConnMaxLifetime == 0 {
		cfg.Database.ConnMaxLifetime = 3600
	}
	if cfg.Log.Level == "" {
		cfg.Log.Level = "info"
	}
	if cfg.Log.Format == "" {
		cfg.Log.Format = "json"
	}

	// Auth defaults
	if cfg.Auth.JWT.AccessTokenExpiry == 0 {
		cfg.Auth.JWT.AccessTokenExpiry = 7200
	}
	if cfg.Auth.JWT.RefreshTokenExpiry == 0 {
		cfg.Auth.JWT.RefreshTokenExpiry = 604800
	}
	if cfg.Auth.Activation.TokenExpiry == 0 {
		cfg.Auth.Activation.TokenExpiry = 86400
	}
	if cfg.Auth.Activation.ResendCooldown == 0 {
		cfg.Auth.Activation.ResendCooldown = 60
	}
	if cfg.Auth.Password.MinLength == 0 {
		cfg.Auth.Password.MinLength = 8
	}
	if cfg.Auth.Password.BcryptCost == 0 {
		cfg.Auth.Password.BcryptCost = 12
	}
	if cfg.Auth.Lockout.MaxAttempts == 0 {
		cfg.Auth.Lockout.MaxAttempts = 5
	}
	if cfg.Auth.Lockout.LockDuration == 0 {
		cfg.Auth.Lockout.LockDuration = 1800
	}
	if cfg.Auth.SecurityTicketTTL == 0 {
		cfg.Auth.SecurityTicketTTL = 300
	}

	// SMS defaults
	if cfg.SMS.OTPTTL == 0 {
		cfg.SMS.OTPTTL = 300
	}
	if cfg.SMS.OTPCooldown == 0 {
		cfg.SMS.OTPCooldown = 60
	}
	if cfg.SMS.OTPLength == 0 {
		cfg.SMS.OTPLength = 6
	}

	// Email defaults
	if cfg.Email.SMTP.Port == 0 {
		cfg.Email.SMTP.Port = 587
	}
	if cfg.Email.Activation.Subject == "" {
		cfg.Email.Activation.Subject = "激活您的账号"
	}
}
