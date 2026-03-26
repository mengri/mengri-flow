package userRepository

import (
	"mengri-flow/internal/domain/repository"
	"mengri-flow/pkg/autowire"
	"mengri-flow/pkg/register"
	"time"
)

func init() {
	autowire.Auto(func() repository.UserRepository {
		register.Register("AutoMigrateOnDebug", &UserModel{})
		return &UserRepositoryImpl{}
	})
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
