package mysql

import (
	"fmt"
	"log/slog"
	"mengri-flow/internal/domain/repository"
	"mengri-flow/internal/infra/config"
	"mengri-flow/pkg/autowire"
	"mengri-flow/pkg/register"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type MysqlDb struct {
	DB *gorm.DB
}

// GenDB 初始化 GORM 数据库连接
func GenDB(cfg *config.DatabaseConfig) (*MysqlDb, error) {
	var logLevel logger.LogLevel
	logLevel = logger.Info
	slog.Debug("Database DSN:", "dsn", cfg.DSN)

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
	autowire.Auto(func() *gorm.DB { return db })

	// --- TransactionManager ---
	autowire.Auto(func() repository.TransactionManager { return &TransactionManagerImpl{db: db} })
	return &MysqlDb{DB: db}, nil
}

func (m *MysqlDb) MigrateOnDebug() error {
	return register.Do("AutoMigrateOnDebug", func(model any) error {

		if err := m.DB.AutoMigrate(model); err != nil {
			return err
		}

		return nil
	})
}
