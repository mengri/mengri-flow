package entity

import (
	"fmt"
	"time"
)

// MemberRole 工作空间成员角色枚举
type MemberRole string

const (
	MemberRoleOwner MemberRole = "owner"
	MemberRoleAdmin MemberRole = "admin"
	MemberRoleMember MemberRole = "member"
)

// IsValid 检查角色是否合法
func (r MemberRole) IsValid() bool {
	switch r {
	case MemberRoleOwner, MemberRoleAdmin, MemberRoleMember:
		return true
	}
	return false
}

// WorkspaceMember 表示工作空间成员关系
type WorkspaceMember struct {
	WorkspaceID string
	AccountID   string
	Role        MemberRole
	JoinedAt    time.Time
}

// NewWorkspaceMember 创建一个新的工作空间成员关系
func NewWorkspaceMember(workspaceID, accountID string, role MemberRole) (*WorkspaceMember, error) {
	if workspaceID == "" {
		return nil, fmt.Errorf("workspace id cannot be empty")
	}
	if accountID == "" {
		return nil, fmt.Errorf("account id cannot be empty")
	}
	if !role.IsValid() {
		return nil, fmt.Errorf("invalid member role: %s", role)
	}

	return &WorkspaceMember{
		WorkspaceID: workspaceID,
		AccountID:   accountID,
		Role:        role,
		JoinedAt:    time.Now(),
	}, nil
}
