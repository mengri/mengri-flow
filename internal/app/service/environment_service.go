package service

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"mengri-flow/internal/app/dto"
	"mengri-flow/internal/domain/entity"
	domainErr "mengri-flow/internal/domain/errors"
	"mengri-flow/internal/domain/repository"

	"github.com/google/uuid"
)

type environmentServiceImpl struct {
	envRepo     repository.EnvironmentRepository `autowired:""`
	clusterRepo repository.ClusterRepository     `autowired:""`
}

func (s *environmentServiceImpl) CreateEnvironment(ctx context.Context, req *dto.CreateEnvironmentRequest) (*dto.EnvironmentResponse, error) {
	// 1. 验证输入参数
	if req.Name == "" {
		return nil, domainErr.ErrInvalidInput
	}
	if req.Key == "" {
		return nil, domainErr.ErrInvalidInput
	}

	// 2. 检查key是否已存在
	_, err := s.envRepo.FindByKey(ctx, req.Key)
	if err == nil {
		slog.Error("environment key already exists", "key", req.Key)
		return nil, domainErr.ErrConflict
	}
	if err != domainErr.ErrNotFound {
		slog.Error("failed to check existing environment key", "key", req.Key, "error", err)
		return nil, fmt.Errorf("check environment key: %w", err)
	}

	// 3. 创建Environment领域实体
	env, err := entity.NewEnvironment(req.Name, req.Key, req.Description, req.Color)
	if err != nil {
		slog.Error("failed to create environment entity", "error", err)
		return nil, domainErr.ErrInvalidInput
	}

	// 4. 保存到数据库
	if err := s.envRepo.Create(ctx, env); err != nil {
		slog.Error("failed to create environment", "error", err)
		return nil, fmt.Errorf("create environment: %w", err)
	}

	// 5. 转换为响应DTO
	return s.toEnvironmentResponse(env), nil
}

func (s *environmentServiceImpl) ListEnvironments(ctx context.Context, page int, pageSize int) (*dto.ListEnvironmentsResponse, error) {
	// 1. 获取所有环境
	environments, err := s.envRepo.List(ctx)
	if err != nil {
		slog.Error("failed to list environments", "error", err)
		return nil, fmt.Errorf("list environments: %w", err)
	}

	// 2. 计算分页参数
	total := len(environments)
	offset, limit := normalizePageParams(page, pageSize)
	if offset >= total {
		return &dto.ListEnvironmentsResponse{
			Total:    int64(total),
			Page:     page,
			PageSize: pageSize,
			List:     []dto.EnvironmentResponse{},
		}, nil
	}

	end := offset + limit
	if end > total {
		end = total
	}

	// 3. 为每个环境获取集群数量
	pagedEnvironments := environments[offset:end]
	responses := make([]dto.EnvironmentResponse, len(pagedEnvironments))

	for i, env := range pagedEnvironments {
		// 获取环境的集群数量
		clusters, err := s.clusterRepo.ListWithFilters(ctx, &env.ID, nil)
		clusterCount := 0
		if err == nil {
			clusterCount = len(clusters)
		}

		responses[i] = dto.EnvironmentResponse{
			ID:           env.ID.String(),
			Name:         env.Name,
			Key:          env.Key,
			Description:  env.Description,
			Color:        env.Color,
			ClusterCount: clusterCount,
			CreatedAt:    env.CreatedAt,
			UpdatedAt:    env.UpdatedAt,
		}
	}

	return &dto.ListEnvironmentsResponse{
		Total:    int64(total),
		Page:     page,
		PageSize: pageSize,
		List:     responses,
	}, nil
}

func (s *environmentServiceImpl) GetEnvironment(ctx context.Context, id string) (*dto.EnvironmentResponse, error) {
	// 1. 验证ID格式
	envID, err := uuid.Parse(id)
	if err != nil {
		slog.Error("invalid environment id", "id", id, "error", err)
		return nil, domainErr.ErrInvalidInput
	}

	// 2. 查找环境
	env, err := s.envRepo.FindByID(ctx, envID)
	if err != nil {
		if err == domainErr.ErrNotFound {
			slog.Error("environment not found", "id", id)
			return nil, domainErr.ErrNotFound
		}
		slog.Error("failed to find environment", "id", id, "error", err)
		return nil, fmt.Errorf("find environment: %w", err)
	}

	// 3. 获取环境的集群数量
	clusters, err := s.clusterRepo.ListWithFilters(ctx, &envID, nil)
	clusterCount := 0
	if err == nil {
		clusterCount = len(clusters)
	}

	// 4. 转换为响应DTO
	return &dto.EnvironmentResponse{
		ID:           env.ID.String(),
		Name:         env.Name,
		Key:          env.Key,
		Description:  env.Description,
		Color:        env.Color,
		ClusterCount: clusterCount,
		CreatedAt:    env.CreatedAt,
		UpdatedAt:    env.UpdatedAt,
	}, nil
}

func (s *environmentServiceImpl) UpdateEnvironment(ctx context.Context, id string, req *dto.UpdateEnvironmentRequest) (*dto.EnvironmentResponse, error) {
	// 1. 验证ID格式
	envID, err := uuid.Parse(id)
	if err != nil {
		slog.Error("invalid environment id", "id", id, "error", err)
		return nil, domainErr.ErrInvalidInput
	}

	// 2. 查找现有环境
	env, err := s.envRepo.FindByID(ctx, envID)
	if err != nil {
		if err == domainErr.ErrNotFound {
			slog.Error("environment not found", "id", id)
			return nil, domainErr.ErrNotFound
		}
		slog.Error("failed to find environment", "id", id, "error", err)
		return nil, fmt.Errorf("find environment: %w", err)
	}

	// 3. 更新环境信息（只有非空字段才更新）
	if req.Name != "" {
		env.Name = req.Name
	}
	if req.Description != "" {
		env.Description = req.Description
	}
	if req.Color != "" {
		env.Color = req.Color
	}
	env.UpdatedAt = time.Now()

	// 4. 保存到数据库
	if err := s.envRepo.Update(ctx, env); err != nil {
		slog.Error("failed to update environment", "id", id, "error", err)
		return nil, fmt.Errorf("update environment: %w", err)
	}

	// 5. 获取更新的集群数量
	clusters, err := s.clusterRepo.ListWithFilters(ctx, &envID, nil)
	clusterCount := 0
	if err == nil {
		clusterCount = len(clusters)
	}

	// 6. 转换为响应DTO
	return &dto.EnvironmentResponse{
		ID:           env.ID.String(),
		Name:         env.Name,
		Key:          env.Key,
		Description:  env.Description,
		Color:        env.Color,
		ClusterCount: clusterCount,
		CreatedAt:    env.CreatedAt,
		UpdatedAt:    env.UpdatedAt,
	}, nil
}

func (s *environmentServiceImpl) DeleteEnvironment(ctx context.Context, id string) error {
	// 1. 验证ID格式
	envID, err := uuid.Parse(id)
	if err != nil {
		slog.Error("invalid environment id", "id", id, "error", err)
		return domainErr.ErrInvalidInput
	}

	// 2. 查找环境
	_, err = s.envRepo.FindByID(ctx, envID)
	if err != nil {
		if err == domainErr.ErrNotFound {
			slog.Error("environment not found", "id", id)
			return domainErr.ErrNotFound
		}
		slog.Error("failed to find environment", "id", id, "error", err)
		return fmt.Errorf("find environment: %w", err)
	}

	// 3. 检查是否有集群关联
	clusters, err := s.clusterRepo.ListWithFilters(ctx, &envID, nil)
	if err != nil {
		slog.Error("failed to check clusters for environment", "id", id, "error", err)
		return fmt.Errorf("check clusters: %w", err)
	}

	if len(clusters) > 0 {
		slog.Error("cannot delete environment with associated clusters", "id", id, "clusterCount", len(clusters))
		return domainErr.ErrInvalidOperation
	}

	// 4. 删除环境
	if err := s.envRepo.Delete(ctx, envID); err != nil {
		slog.Error("failed to delete environment", "id", id, "error", err)
		return fmt.Errorf("delete environment: %w", err)
	}

	return nil
}

// toEnvironmentResponse 将Environment实体转换为响应DTO
func (s *environmentServiceImpl) toEnvironmentResponse(env *entity.Environment) *dto.EnvironmentResponse {
	// 获取环境的集群数量
	clusterCount := 0
	if env.ID != uuid.Nil {
		clusters, err := s.clusterRepo.ListWithFilters(context.Background(), &env.ID, nil)
		if err == nil {
			clusterCount = len(clusters)
		}
	}

	return &dto.EnvironmentResponse{
		ID:           env.ID.String(),
		Name:         env.Name,
		Key:          env.Key,
		Description:  env.Description,
		Color:        env.Color,
		ClusterCount: clusterCount,
		CreatedAt:    env.CreatedAt,
		UpdatedAt:    env.UpdatedAt,
	}
}
