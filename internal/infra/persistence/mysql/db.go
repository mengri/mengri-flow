package mysql

import (
	"fmt"
	"mengri-flow/internal/infra/config"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// NewDB 初始化 GORM 数据库连接
func NewDB(cfg *config.DatabaseConfig) (*gorm.DB, error) {
	var logLevel logger.LogLevel
	logLevel = logger.Info

	db, err := gorm.Open(mysql.Open(cfg.DSN), &gorm.Config{
		Logger: logger.Default.LogMode(logLevel),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetConnMaxLifetime(time.Duration(cfg.ConnMaxLifetime) * time.Second)

	return db, nil
}

// UserModel GORM 数据库模型 — 仅在 Infra 层使用，与 Domain Entity 解耦。
type UserModel struct {
	ID        uint64 `gorm:"primaryKey;autoIncrement"`
	Username  string `gorm:"type:varchar(50);not null"`
	Email     string `gorm:"type:varchar(100);uniqueIndex;not null"`
	Password  string `gorm:"type:varchar(255);not null"`
	Status    int    `gorm:"type:tinyint;default:1;not null"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

// TableName 指定表名
func (UserModel) TableName() string {
	return "users"
}
