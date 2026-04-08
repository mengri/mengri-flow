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
	Config         map[string]interface{}
	InputMapping   map[string]interface{}
	OutputMapping  map[string]interface{}
	ErrorHandling  map[string]interface{}
	Status         TriggerStatus
	LastExecutedAt *time.Time
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

// NewTrigger 创建一个新的触发器
func NewTrigger(name string, triggerType TriggerType, flowID uuid.UUID, config map[string]interface{}) (*Trigger, error) {
	if name == "" {
		return nil, fmt.Errorf("trigger name cannot be empty")
	}

	if err := validateTriggerType(triggerType); err != nil {
		return nil, err
	}

	now := time.Now()
	return &Trigger{
		ID:        uuid.New(),
		Name:      name,
		Type:      triggerType,
		FlowID:    flowID,
		Config:    config,
		Status:    TriggerStatusActive,
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
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

func validateTriggerStatus(status TriggerStatus) error {
	switch status {
	case TriggerStatusActive, TriggerStatusInactive, TriggerStatusError:
		return nil
	default:
		return fmt.Errorf("invalid trigger status: %s", status)
	}
}
