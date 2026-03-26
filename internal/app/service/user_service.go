package service

import (
	"context"
	"mengri-flow/internal/app/dto"
	"mengri-flow/internal/domain/entity"
	domainErr "mengri-flow/internal/domain/errors"
	"mengri-flow/internal/domain/repository"
	"time"
)

// UserServiceImpl 应用服务层实现，编排领域逻辑，不包含业务规则。
type UserServiceImpl struct {
	userRepo repository.UserRepository `autowired:""`
	// 可注入密码哈希器等接口
}

// 编译期接口合规检查
var _ UserService = (*UserServiceImpl)(nil)

// NewUserService 通过构造函数注入依赖，返回 UserService 接口。
func NewUserService(userRepo repository.UserRepository) UserService {
	return &UserServiceImpl{
		userRepo: userRepo,
	}
}

// CreateUser 创建用户用例
func (s *UserServiceImpl) CreateUser(ctx context.Context, req *dto.CreateUserRequest) (*dto.UserResponse, error) {
	// 检查邮箱唯一性
	existing, _ := s.userRepo.GetByEmail(ctx, req.Email)
	if existing != nil {
		return nil, domainErr.ErrEmailTaken
	}

	// TODO: 实际项目中应使用 bcrypt 哈希密码
	hashedPassword := req.Password

	// 创建领域实体（业务校验在 Entity 内部完成）
	user, err := entity.NewUser(req.Username, req.Email, hashedPassword)
	if err != nil {
		return nil, err
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}

	return toUserResponse(user), nil
}

// GetUser 查询单个用户用例
func (s *UserServiceImpl) GetUser(ctx context.Context, id uint64) (*dto.UserResponse, error) {
	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return toUserResponse(user), nil
}

// ListUsers 查询用户列表用例
func (s *UserServiceImpl) ListUsers(ctx context.Context, req *dto.ListUsersRequest) (*dto.ListUsersResponse, error) {
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 20
	}

	offset := (req.Page - 1) * req.PageSize
	users, total, err := s.userRepo.List(ctx, offset, req.PageSize)
	if err != nil {
		return nil, err
	}

	items := make([]*dto.UserResponse, 0, len(users))
	for _, u := range users {
		items = append(items, toUserResponse(u))
	}

	return &dto.ListUsersResponse{
		Items:    items,
		Total:    total,
		Page:     req.Page,
		PageSize: req.PageSize,
	}, nil
}

// UpdateUser 更新用户用例
func (s *UserServiceImpl) UpdateUser(ctx context.Context, id uint64, req *dto.UpdateUserRequest) (*dto.UserResponse, error) {
	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if req.Email != "" {
		if err := user.ChangeEmail(req.Email); err != nil {
			return nil, err
		}
	}
	if req.Username != "" {
		user.Username = req.Username
		user.UpdatedAt = time.Now()
	}

	if err := s.userRepo.Update(ctx, user); err != nil {
		return nil, err
	}

	return toUserResponse(user), nil
}

// DeleteUser 删除用户用例
func (s *UserServiceImpl) DeleteUser(ctx context.Context, id uint64) error {
	return s.userRepo.Delete(ctx, id)
}

// toUserResponse 将领域实体转换为响应 DTO
func toUserResponse(user *entity.User) *dto.UserResponse {
	return &dto.UserResponse{
		ID:        user.ID,
		Username:  user.Username,
		Email:     user.Email.String(),
		Status:    int(user.Status),
		CreatedAt: user.CreatedAt.Format(time.RFC3339),
		UpdatedAt: user.UpdatedAt.Format(time.RFC3339),
	}
}
