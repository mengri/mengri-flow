package userRepository

import (
	"context"
	"mengri-flow/internal/domain/entity"
	domainErr "mengri-flow/internal/domain/errors"
	"mengri-flow/internal/domain/repository"
	"mengri-flow/internal/domain/valueobject"

	"gorm.io/gorm"
)

// UserRepositoryImpl 是 UserRepository 接口的 GORM 实现。
// 实现在 Infra 层，接口定义在 Domain 层 — 依赖倒置。
type UserRepositoryImpl struct {
	db *gorm.DB `autowired:""`
}

// 编译期接口合规检查
var _ repository.UserRepository = (*UserRepositoryImpl)(nil)

func (r *UserRepositoryImpl) Create(ctx context.Context, user *entity.User) error {
	model := toModel(user)
	result := r.db.WithContext(ctx).Create(model)
	if result.Error != nil {
		return result.Error
	}
	user.ID = model.ID
	return nil
}

func (r *UserRepositoryImpl) GetByID(ctx context.Context, id uint64) (*entity.User, error) {
	var model UserModel
	result := r.db.WithContext(ctx).First(&model, id)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, domainErr.ErrUserNotFound
		}
		return nil, result.Error
	}
	return toEntity(&model), nil
}

func (r *UserRepositoryImpl) GetByEmail(ctx context.Context, email string) (*entity.User, error) {
	var model UserModel
	result := r.db.WithContext(ctx).Where("email = ?", email).First(&model)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, domainErr.ErrUserNotFound
		}
		return nil, result.Error
	}
	return toEntity(&model), nil
}

func (r *UserRepositoryImpl) Update(ctx context.Context, user *entity.User) error {
	model := toModel(user)
	model.ID = user.ID
	return r.db.WithContext(ctx).Save(model).Error
}

func (r *UserRepositoryImpl) Delete(ctx context.Context, id uint64) error {
	result := r.db.WithContext(ctx).Delete(&UserModel{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return domainErr.ErrUserNotFound
	}
	return nil
}

func (r *UserRepositoryImpl) List(ctx context.Context, offset, limit int) ([]*entity.User, int64, error) {
	var models []UserModel
	var total int64

	if err := r.db.WithContext(ctx).Model(&UserModel{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := r.db.WithContext(ctx).Offset(offset).Limit(limit).Order("id DESC").Find(&models).Error; err != nil {
		return nil, 0, err
	}

	users := make([]*entity.User, 0, len(models))
	for _, m := range models {
		users = append(users, toEntity(&m))
	}
	return users, total, nil
}

// toModel 领域实体 -> 数据模型
func toModel(user *entity.User) *UserModel {
	return &UserModel{
		Username:  user.Username,
		Email:     user.Email.String(),
		Password:  user.Password,
		Status:    int(user.Status),
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}
}

// toEntity 数据模型 -> 领域实体
func toEntity(model *UserModel) *entity.User {
	email, _ := valueobject.NewEmail(model.Email)
	return &entity.User{
		ID:        model.ID,
		Username:  model.Username,
		Email:     email,
		Password:  model.Password,
		Status:    entity.UserStatus(model.Status),
		CreatedAt: model.CreatedAt,
		UpdatedAt: model.UpdatedAt,
	}
}
