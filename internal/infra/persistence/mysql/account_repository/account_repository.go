package accountRepository

import (
	"context"
	"mengri-flow/internal/domain/entity"
	domainErr "mengri-flow/internal/domain/errors"
	"mengri-flow/internal/domain/repository"
	"mengri-flow/internal/domain/valueobject"

	"gorm.io/gorm"
)

// AccountRepositoryImpl 是 AccountRepository 的 GORM 实现。
type AccountRepositoryImpl struct {
	db *gorm.DB `autowired:""`
}

var _ repository.AccountRepository = (*AccountRepositoryImpl)(nil)

func (r *AccountRepositoryImpl) Create(ctx context.Context, account *entity.Account) error {
	model := toModel(account)
	result := r.db.WithContext(ctx).Create(model)
	if result.Error != nil {
		return result.Error
	}
	account.ID = model.ID
	return nil
}

func (r *AccountRepositoryImpl) GetByID(ctx context.Context, id string) (*entity.Account, error) {
	var model AccountModel
	result := r.db.WithContext(ctx).Where("id = ?", id).First(&model)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, domainErr.ErrAccountNotFound
		}
		return nil, result.Error
	}
	return toEntity(&model), nil
}

func (r *AccountRepositoryImpl) GetByEmail(ctx context.Context, email string) (*entity.Account, error) {
	var model AccountModel
	result := r.db.WithContext(ctx).Where("email = ?", email).First(&model)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, domainErr.ErrAccountNotFound
		}
		return nil, result.Error
	}
	return toEntity(&model), nil
}

func (r *AccountRepositoryImpl) GetByUsername(ctx context.Context, username string) (*entity.Account, error) {
	var model AccountModel
	result := r.db.WithContext(ctx).Where("username = ?", username).First(&model)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, domainErr.ErrAccountNotFound
		}
		return nil, result.Error
	}
	return toEntity(&model), nil
}

func (r *AccountRepositoryImpl) Update(ctx context.Context, account *entity.Account) error {
	model := toModel(account)
	return r.db.WithContext(ctx).Save(model).Error
}

func (r *AccountRepositoryImpl) List(ctx context.Context, offset, limit int, status *entity.AccountStatus, keyword string) ([]*entity.Account, int64, error) {
	var models []AccountModel
	var total int64

	query := r.db.WithContext(ctx).Model(&AccountModel{})
	if status != nil {
		query = query.Where("status = ?", string(*status))
	}
	if keyword != "" {
		like := "%" + keyword + "%"
		query = query.Where("email LIKE ? OR username LIKE ? OR display_name LIKE ?", like, like, like)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Offset(offset).Limit(limit).Order("created_at DESC").Find(&models).Error; err != nil {
		return nil, 0, err
	}

	accounts := make([]*entity.Account, 0, len(models))
	for _, m := range models {
		accounts = append(accounts, toEntity(&m))
	}
	return accounts, total, nil
}

func toModel(account *entity.Account) *AccountModel {
	return &AccountModel{
		ID:          account.ID,
		Email:       account.Email.String(),
		Username:    account.Username,
		DisplayName: account.DisplayName,
		Status:      string(account.Status),
		Role:        account.Role,
		ActivatedAt: account.ActivatedAt,
		CreatedAt:   account.CreatedAt,
		UpdatedAt:   account.UpdatedAt,
	}
}

func toEntity(model *AccountModel) *entity.Account {
	email, _ := valueobject.NewEmail(model.Email)
	return &entity.Account{
		ID:          model.ID,
		Email:       email,
		Username:    model.Username,
		DisplayName: model.DisplayName,
		Status:      entity.AccountStatus(model.Status),
		Role:        model.Role,
		ActivatedAt: model.ActivatedAt,
		CreatedAt:   model.CreatedAt,
		UpdatedAt:   model.UpdatedAt,
	}
}
