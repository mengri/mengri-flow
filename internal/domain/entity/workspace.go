package entity

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// Workspace 表示一个工作空间
type Workspace struct {
	ID          uuid.UUID
	Name        string
	Description string
	OwnerID     string // 所有者账户ID
	MemberCount int    // 成员数量，创建时默认为1
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// NewWorkspace 创建一个新的工作空间
func NewWorkspace(name, description string, ownerID string) (*Workspace, error) {
	if name == "" {
		return nil, fmt.Errorf("workspace name cannot be empty")
	}

	if len(name) > 100 {
		return nil, fmt.Errorf("workspace name cannot exceed 100 characters")
	}

	if len(description) > 500 {
		return nil, fmt.Errorf("workspace description cannot exceed 500 characters")
	}

	now := time.Now()
	return &Workspace{
		ID:          uuid.New(),
		Name:        name,
		Description: description,
		OwnerID:     ownerID,
		MemberCount: 1, // 创建时默认包含所有者
		CreatedAt:   now,
		UpdatedAt:   now,
	}, nil
}

// Update 更新工作空间信息
func (w *Workspace) Update(name, description string) error {
	if name == "" {
		return fmt.Errorf("workspace name cannot be empty")
	}

	if len(name) > 100 {
		return fmt.Errorf("workspace name cannot exceed 100 characters")
	}

	if len(description) > 500 {
		return fmt.Errorf("workspace description cannot exceed 500 characters")
	}

	w.Name = name
	w.Description = description
	w.UpdatedAt = time.Now()
	return nil
}

// IncrementMemberCount 增加成员数量
func (w *Workspace) IncrementMemberCount() {
	w.MemberCount++
	w.UpdatedAt = time.Now()
}

// DecrementMemberCount 减少成员数量
func (w *Workspace) DecrementMemberCount() {
	if w.MemberCount > 1 {
		w.MemberCount--
		w.UpdatedAt = time.Now()
	}
}