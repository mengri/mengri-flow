package entity

import (
	"fmt"
	"regexp"
	"time"

	"github.com/google/uuid"
)

// Environment 表示一个环境（如 dev/staging/prod）
type Environment struct {
	ID          uuid.UUID
	Name        string
	Key         string // 唯一标识，如 dev/staging/prod
	Description string
	Color       string // Hex颜色
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// NewEnvironment 创建一个新的环境
func NewEnvironment(name, key, description, color string) (*Environment, error) {
	if err := validateEnvironmentKey(key); err != nil {
		return nil, err
	}

	if name == "" {
		return nil, fmt.Errorf("environment name cannot be empty")
	}

	if err := validateHexColor(color); err != nil {
		return nil, err
	}

	now := time.Now()
	return &Environment{
		ID:          uuid.New(),
		Name:        name,
		Key:         key,
		Description: description,
		Color:       color,
		CreatedAt:   now,
		UpdatedAt:   now,
	}, nil
}

// Update 更新环境信息
func (e *Environment) Update(name, description, color string) error {
	if name == "" {
		return fmt.Errorf("environment name cannot be empty")
	}

	if err := validateHexColor(color); err != nil {
		return err
	}

	e.Name = name
	e.Description = description
	e.Color = color
	e.UpdatedAt = time.Now()
	return nil
}

// validateEnvironmentKey 验证环境Key格式
func validateEnvironmentKey(key string) error {
	if key == "" {
		return fmt.Errorf("environment key cannot be empty")
	}

	// Key只能包含小写字母、数字和连字符
	pattern := `^[a-z0-9-]+$`
	matched, err := regexp.MatchString(pattern, key)
	if err != nil {
		return fmt.Errorf("failed to validate key format: %w", err)
	}
	if !matched {
		return fmt.Errorf("environment key must contain only lowercase letters, numbers, and hyphens")
	}

	return nil
}

// validateHexColor 验证Hex颜色格式
func validateHexColor(color string) error {
	if color == "" {
		return nil // 可选字段
	}

	pattern := `^#[0-9A-Fa-f]{6}$`
	matched, err := regexp.MatchString(pattern, color)
	if err != nil {
		return fmt.Errorf("failed to validate color format: %w", err)
	}
	if !matched {
		return fmt.Errorf("color must be a valid hex color (e.g., #FF5733)")
	}

	return nil
}
