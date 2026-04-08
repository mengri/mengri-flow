package timer

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/robfig/cron/v3"
	clientv3 "go.etcd.io/etcd/client/v3"
	concurrency "go.etcd.io/etcd/client/v3/concurrency"
	"mengri-flow/internal/infra/plugin"
)

func init() {
	registry := plugin.GlobalRegistry()
	registry.RegisterTrigger(&TimerTriggerPlugin{})
}

// 确保接口实现
var _ plugin.TriggerPlugin = (*TimerTriggerPlugin)(nil)

type TimerTriggerPlugin struct {
	mu         sync.RWMutex
	crons      map[string]*cron.Cron
	configs    map[string]map[string]interface{}
	handlers   map[string]plugin.TriggerHandler
	etcdClient *clientv3.Client
}

func (p *TimerTriggerPlugin) PluginMeta() plugin.PluginMeta {
	return plugin.PluginMeta{
		Name:        "timer",
		Type:        plugin.PluginTypeTrigger,
		Version:     "1.0.0",
		Description: "定时任务触发器插件（Cron表达式），支持分布式锁",
		Author:      "Platform Team",
	}
}

func (p *TimerTriggerPlugin) ConfigSchema() plugin.JSONSchema {
	return plugin.JSONSchema{
		"type": "object",
		"required": []string{"cronExpression"},
		"properties": map[string]interface{}{
			"cronExpression": map[string]interface{}{
				"type":        "string",
				"title":       "Cron表达式",
				"description": "标准的Cron表达式，如 0 2 * * *（每天凌晨2点）",
				"placeholder": "0 2 * * *",
				"pattern":     "^([\\*0-9,\\-/]+\\s){4,5}[\\*0-9,\\-/]+$",
			},
			"timezone": map[string]interface{}{
				"type":        "string",
				"title":       "时区",
				"description": "Cron表达式使用的时区",
				"default":     "Asia/Shanghai",
				"enum":        []string{"Asia/Shanghai", "UTC", "America/New_York"},
			},
			"distributedLock": map[string]interface{}{
				"type":        "boolean",
				"title":       "启用分布式锁",
				"description": "确保集群中只有一个实例执行定时任务",
				"default":     true,
			},
			"lockKey": map[string]interface{}{
				"type":        "string",
				"title":       "分布式锁键",
				"description": "etcd中的锁路径，如 /locks/timer-{trigger-id}",
				"condition":   map[string]interface{}{"distributedLock": true},
			},
			"lockTTL": map[string]interface{}{
				"type":        "integer",
				"title":       "锁TTL（秒）",
				"description": "分布式锁的过期时间，应大于任务执行时间",
				"default":     60,
				"minimum":     10,
				"maximum":     3600,
				"condition":   map[string]interface{}{"distributedLock": true},
			},
		},
	}
}

func (p *TimerTriggerPlugin) InputSchema() plugin.JSONSchema {
	return plugin.JSONSchema{
		"type": "object",
		"properties": map[string]interface{}{
			"triggerTime": map[string]interface{}{
				"type":        "string",
				"format":      "date-time",
				"description": "触发时间（ISO8601格式）",
			},
			"executionId": map[string]interface{}{
				"type":        "string",
				"description": "执行唯一标识",
			},
			"triggerId": map[string]interface{}{
				"type":        "string",
				"description": "触发器ID",
			},
		},
	}
}

func (p *TimerTriggerPlugin) OutputSchema() plugin.JSONSchema {
	return nil
}

func (p *TimerTriggerPlugin) Start(
	ctx context.Context,
	config map[string]interface{},
	handler plugin.TriggerHandler,
) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	triggerID, ok := config["triggerId"].(string)
	if !ok || triggerID == "" {
		return plugin.NewPluginError("config_error", "missing or invalid triggerId", nil)
	}

	if p.crons == nil {
		p.crons = make(map[string]*cron.Cron)
		p.configs = make(map[string]map[string]interface{})
		p.handlers = make(map[string]plugin.TriggerHandler)
	}

	p.configs[triggerID] = config
	p.handlers[triggerID] = handler

	timezone := "Asia/Shanghai"
	if tz, ok := config["timezone"].(string); ok && tz != "" {
		timezone = tz
	}

	loc, err := time.LoadLocation(timezone)
	if err != nil {
		return plugin.NewPluginError("config_error", fmt.Sprintf("invalid timezone: %s", timezone), err)
	}

	c := cron.New(cron.WithLocation(loc))

	cronExpr, ok := config["cronExpression"].(string)
	if !ok || cronExpr == "" {
		return plugin.NewPluginError("config_error", "missing or invalid cronExpression", nil)
	}

	_, err = c.AddFunc(cronExpr, func() {
		p.executeWithLock(ctx, triggerID)
	})
	if err != nil {
		return plugin.NewPluginError("config_error", fmt.Sprintf("invalid cron expression: %s", cronExpr), err)
	}

	p.crons[triggerID] = c

	c.Start()

	fmt.Printf("Timer trigger %s started with cron: %s, timezone: %s\n", triggerID, cronExpr, timezone)

	return nil
}

func (p *TimerTriggerPlugin) Stop() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	for triggerID, cron := range p.crons {
		cron.Stop()
		fmt.Printf("Timer trigger %s stopped\n", triggerID)
	}

	p.crons = nil
	p.configs = nil
	p.handlers = nil

	return nil
}

// SetEtcdClient 设置etcd客户端（由Executor注入）
func (p *TimerTriggerPlugin) SetEtcdClient(client *clientv3.Client) {
	p.etcdClient = client
}

func (p *TimerTriggerPlugin) executeWithLock(ctx context.Context, triggerID string) {
	p.mu.RLock()
	handler, handlerExists := p.handlers[triggerID]
	config, configExists := p.configs[triggerID]
	p.mu.RUnlock()

	if !handlerExists || !configExists {
		fmt.Printf("Handler or config not found for trigger %s\n", triggerID)
		return
	}

	useLock := true
	if val, ok := config["distributedLock"].(bool); ok {
		useLock = val
	}

	if !useLock || p.etcdClient == nil {
		p.execute(ctx, triggerID, handler, config)
		return
	}

	lockKey := fmt.Sprintf("/locks/timer/%s", triggerID)
	if val, ok := config["lockKey"].(string); ok && val != "" {
		lockKey = val
	}

	lockTTL := 60
	if val, ok := config["lockTTL"].(int); ok && val > 0 {
		lockTTL = val
	}

	session, err := concurrency.NewSession(p.etcdClient, concurrency.WithTTL(lockTTL))
	if err != nil {
		fmt.Printf("Failed to create etcd session for trigger %s: %v\n", triggerID, err)
		return
	}
	defer session.Close()

	mutex := concurrency.NewMutex(session, lockKey)

	lockCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	if err := mutex.Lock(lockCtx); err != nil {
		fmt.Printf("Failed to acquire lock for trigger %s: %v\n", triggerID, err)
		return
	}

	stopRenewal := make(chan struct{})
	go p.renewLock(session, stopRenewal)

	p.execute(ctx, triggerID, handler, config)

	close(stopRenewal)

	if err := mutex.Unlock(ctx); err != nil {
		fmt.Printf("Failed to release lock for trigger %s: %v\n", triggerID, err)
	}
}

func (p *TimerTriggerPlugin) renewLock(session *concurrency.Session, stop <-chan struct{}) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-stop:
			return
		case <-ticker.C:
			// etcd会话的KeepAlive会自动续期
			// 这里只是保持goroutine活跃
		}
	}
}

func (p *TimerTriggerPlugin) execute(
	ctx context.Context,
	triggerID string,
	handler plugin.TriggerHandler,
	config map[string]interface{},
) {
	executionID := generateExecutionID()

	input := map[string]interface{}{
		"triggerTime": time.Now().Format(time.RFC3339),
		"executionId": executionID,
		"triggerId":   triggerID,
	}

	result, err := handler(ctx, input)
	if err != nil {
		fmt.Printf("Timer trigger %s execution failed: %v\n", triggerID, err)
		return
	}

	if !result.Success {
		fmt.Printf("Timer trigger %s execution failed: %s\n", triggerID, result.Error)
		return
	}

	fmt.Printf("Timer trigger %s executed successfully (executionId: %s)\n", triggerID, executionID)
}

func generateExecutionID() string {
	return fmt.Sprintf("timer-%d", time.Now().UnixNano())
}
