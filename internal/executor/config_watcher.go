package executor

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"mengri-flow/internal/domain/entity"

	clientv3 "go.etcd.io/etcd/client/v3"
)

// ConfigWatcher 配置监听器
type ConfigWatcher struct {
	etcdClient     *clientv3.Client
	clusterID      string
	triggerManager *TriggerManager
	watchChan      clientv3.WatchChan
	stopChan       chan struct{}
}

// NewConfigWatcher 创建新的配置监听器
func NewConfigWatcher(client *clientv3.Client, clusterID string, triggerManager *TriggerManager) *ConfigWatcher {
	return &ConfigWatcher{
		etcdClient:     client,
		clusterID:      clusterID,
		triggerManager: triggerManager,
		stopChan:       make(chan struct{}),
	}
}

// Start 启动配置监听
func (w *ConfigWatcher) Start(ctx context.Context) {
	// 监听前缀
	prefix := fmt.Sprintf("/clusters/%s/", w.clusterID)

	// 创建watch
	w.watchChan = w.etcdClient.Watch(ctx, prefix, clientv3.WithPrefix())

	// 处理变更
	go func() {
		for {
			select {
			case resp, ok := <-w.watchChan:
				if !ok {
					log.Printf("Watch channel closed")
					return
				}
				if resp.Err() != nil {
					log.Printf("Watch error: %v", resp.Err())
					continue
				}
				for _, event := range resp.Events {
					w.handleEtcdEvent(event)
				}
			case <-w.stopChan:
				log.Printf("Config watcher stopped")
				return
			}
		}
	}()

	log.Printf("Config watcher started for %s", prefix)
}

// Stop 停止配置监听
func (w *ConfigWatcher) Stop() {
	close(w.stopChan)
}

// handleEtcdEvent 处理etcd事件
func (w *ConfigWatcher) handleEtcdEvent(event *clientv3.Event) {
	key := string(event.Kv.Key)

	// 解析key格式: /clusters/{cluster-id}/{type}/{id}
	parts := strings.Split(key, "/")
	if len(parts) < 5 {
		log.Printf("Invalid key format: %s", key)
		return
	}

	resourceType := parts[3]
	resourceID := parts[4]

	switch event.Type {
	case clientv3.EventTypePut:
		// 创建或更新
		switch resourceType {
		case "triggers":
			var trigger entity.Trigger
			if err := json.Unmarshal(event.Kv.Value, &trigger); err != nil {
				log.Printf("Failed to unmarshal trigger: %v", err)
				return
			}
			if err := w.triggerManager.AddOrUpdateTrigger(&trigger); err != nil {
				log.Printf("Failed to add/update trigger: %v", err)
			}
			log.Printf("Trigger %s added/updated", resourceID)

		case "flows":
			// 更新流程缓存
			log.Printf("Flow %s updated", resourceID)

		case "resources":
			// 更新资源缓存
			log.Printf("Resource %s updated", resourceID)

		case "tools":
			// 更新工具缓存
			log.Printf("Tool %s updated", resourceID)
		}

	case clientv3.EventTypeDelete:
		// 删除
		switch resourceType {
		case "triggers":
			w.triggerManager.RemoveTrigger(resourceID)
			log.Printf("Trigger %s removed", resourceID)
		}
	}
}
