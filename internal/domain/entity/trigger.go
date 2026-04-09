package entity

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// TriggerType 触发器类型
type TriggerType string

const (
	TriggerTypeRESTful TriggerType = "restful"
	TriggerTypeTimer   TriggerType = "timer"
	TriggerTypeMQ      TriggerType = "mq"
)

// TriggerStatus 触发器状态
type TriggerStatus string

const (
	TriggerStatusActive   TriggerStatus = "active"
	TriggerStatusInactive TriggerStatus = "inactive"
	TriggerStatusError    TriggerStatus = "error"
)

// Trigger 表示一个触发器
type Trigger struct {
	ID             uuid.UUID
	Name           string
	Type           TriggerType
	FlowID         uuid.UUID
	FlowVersion    int
	ClusterID      uuid.UUID
	InputMapping   map[string]string
	OutputMapping  map[string]string
	ErrorHandling  ErrorHandling
	Config         map[string]interface{}
	WorkspaceID    uuid.UUID
	Status         TriggerStatus
	LastExecutedAt *time.Time
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

// ErrorHandling 错误处理配置
type ErrorHandling struct {
	Strategy         string
	CustomErrorFormat map[string]interface{}
	RetryOnFailure   bool
}

// NewTrigger 创建一个新的触发器
func NewTrigger(name string, triggerType TriggerType, flowID uuid.UUID, config map[string]interface{}) (*Trigger, error) {
	return NewTriggerWithOptions(name, triggerType, flowID, config, nil, nil, nil)
}

// NewTriggerWithOptions 创建一个新的触发器（带完整选项）
func NewTriggerWithOptions(
	name string,
	triggerType TriggerType,
	flowID uuid.UUID,
	config map[string]interface{},
	clusterID *uuid.UUID,
	workspaceID *uuid.UUID,
	errorHandling *ErrorHandling,
) (*Trigger, error) {
	if name == "" {
		return nil, fmt.Errorf("trigger name cannot be empty")
	}

	if err := validateTriggerType(triggerType); err != nil {
		return nil, err
	}

	now := time.Now()
	trigger := &Trigger{
		ID:        uuid.New(),
		Name:      name,
		Type:      triggerType,
		FlowID:    flowID,
		Config:    config,
		Status:    TriggerStatusActive,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if clusterID != nil {
		trigger.ClusterID = *clusterID
	}
	if workspaceID != nil {
		trigger.WorkspaceID = *workspaceID
	}
	if errorHandling != nil {
		trigger.ErrorHandling = *errorHandling
	} else {
		// 默认错误处理配置
		trigger.ErrorHandling = ErrorHandling{
			Strategy:       "default",
			RetryOnFailure: false,
		}
	}

	return trigger, nil
}

// Update 更新触发器信息
func (t *Trigger) Update(name string, config map[string]interface{}) error {
	if name == "" {
		return fmt.Errorf("trigger name cannot be empty")
	}

	t.Name = name
	t.Config = config
	t.UpdatedAt = time.Now()
	return nil
}

// UpdateStatus 更新触发器状态
func (t *Trigger) UpdateStatus(status TriggerStatus) error {
	if err := validateTriggerStatus(status); err != nil {
		return err
	}
	t.Status = status
	t.UpdatedAt = time.Now()
	return nil
}

// MarkExecuted 标记触发器已执行
func (t *Trigger) MarkExecuted() {
	now := time.Now()
	t.LastExecutedAt = &now
	t.UpdatedAt = now
}

func validateTriggerType(triggerType TriggerType) error {
	switch triggerType {
	case TriggerTypeRESTful, TriggerTypeTimer, TriggerTypeMQ:
		return nil
	default:
		return fmt.Errorf("invalid trigger type: %s", triggerType)
	}
}

// ValidateTriggerType 验证触发器类型（公共函数）
func ValidateTriggerType(triggerType TriggerType) error {
	return validateTriggerType(triggerType)
}

func validateTriggerStatus(status TriggerStatus) error {
	switch status {
	case TriggerStatusActive, TriggerStatusInactive, TriggerStatusError:
		return nil
	default:
		return fmt.Errorf("invalid trigger status: %s", status)
	}
}
