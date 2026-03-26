package service

import (
	"context"
	"mengri-flow/internal/app/dto"
	"mengri-flow/pkg/autowire"
)

// UserService 定义用户应用服务接口。
// 接口定义在 App 层（调用方），Handler 层依赖此接口而非具体实现。
type UserService interface {
	CreateUser(ctx context.Context, req *dto.CreateUserRequest) (*dto.UserResponse, error)
	GetUser(ctx context.Context, id uint64) (*dto.UserResponse, error)
	ListUsers(ctx context.Context, req *dto.ListUsersRequest) (*dto.ListUsersResponse, error)
	UpdateUser(ctx context.Context, id uint64, req *dto.UpdateUserRequest) (*dto.UserResponse, error)
	DeleteUser(ctx context.Context, id uint64) error
}

func init() {
	autowire.Auto(func() UserService {
		return &UserServiceImpl{}
	})
}
