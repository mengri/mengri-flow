# RESTful触发器端口配置方案

## 概述

为每个RESTful触发器实例分配独立的HTTP监听端口，通过etcd集群配置统一管理。

## 端口分配

### 端口范围
- **可用范围**: 1024-65535
- **推荐范围**: 18080-18999（RESTful触发器专用）
- **预留范围**: 19000-19999（其他触发器类型）

### 配置结构

```yaml
# 触发器配置（从etcd下发）
triggerId: "webhook-order-created"
path: "/webhooks/order-created"
method: "POST"
port: 18081           # 从etcd配置下发，只读
async: false
auth:
  type: "apiKey"
  apiKey: "secret-key-123"
```

## etcd集成

### 端口注册表
```
/ports/registry/
  ├── 18081: "trigger-webhook-order-created"
  ├── 18082: "trigger-webhook-payment-success"
  └── 18083: "trigger-webhook-user-registered"
```

### 端口分配锁
使用etcd分布式锁确保端口分配的原子性：
```go
mutex := concurrency.NewMutex(session, "/locks/port-allocation")
mutex.Lock(context.Background())
defer mutex.Unlock()
// 分配端口并写入etcd
```

## 配置下发流程

1. **创建触发器**
   - API接收创建请求
   - 触发器管理模块申请端口（自动分配或手动指定）
   - 写入etcd配置（含端口信息）
   - 返回触发器信息（含webhookUrl）

2. **Executor监听**
   ```go
   watchChan := etcdClient.Watch(ctx, "/triggers/", clientv3.WithPrefix())
   for watchResp := range watchChan {
       // 处理配置变更（启动/停止触发器）
   }
   ```

3. **端口冲突处理**
   - 检查端口是否被占用
   - 如果占用者已停止，释放端口
   - 如果真实冲突，报告错误

## 端口管理

### 健康检查
```go
// 每30秒检查端口监听状态
if !isPortListening(server.Port) {
    log.Errorf("Port %d not listening", server.Port)
    restartTrigger(triggerID)
}
```

### 端口回收
```go
// 清理未使用的端口分配
for port, triggerID := range allocatedPorts {
    if !isTriggerActive(triggerID) {
        releasePortFromEtcd(port)
    }
}
```

## 使用示例

### 创建触发器
```bash
curl -X POST http://localhost:8080/api/v1/triggers \
  -d '{
    "name": "Order Webhook",
    "type": "restful",
    "flowId": "order-flow",
    "config": {
      "path": "/webhooks/order",
      "method": "POST",
      "auth": {
        "type": "apiKey",
        "apiKey": "secret-123"
      }
    }
  }'

# 响应包含自动分配的端口
{
  "data": {
    "config": {
      "port": 18081,
      ...
    },
    "webhookUrl": "http://server:18081/webhooks/order"
  }
}
```

### 调用Webhook
```bash
curl -X POST http://server:18081/webhooks/order \
  -H "X-API-Key: secret-123" \
  -d '{"orderId": "ORD-001"}'
```

## 端口分配策略

### 方案A：自动分配（推荐）
```go
func allocatePort(triggerID string) (int, error) {
    usedPorts := getUsedPortsFromEtcd()
    for port := 18080; port <= 18999; port++ {
        if !contains(usedPorts, port) {
            markPortAsUsed(port, triggerID)
            return port, nil
        }
    }
    return 0, errors.New("no available port")
}
```

### 方案B：手动指定
```yaml
# 创建时手动指定端口
triggerId: "webhook-payment"
port: 18095  # 需确保不冲突
```

## 实现状态

✅ 插件支持端口配置（从ConfigSchema读取）
✅ 测试覆盖（82.3%覆盖率）
⏳ etcd集成（待触发器管理模块实现）
⏳ 端口分配算法（待触发器管理模块实现）
⏳ 健康检查（待Executor实现）

## 注意事项

1. **防火墙配置**：确保服务器防火墙允许端口范围18080-18999
2. **负载均衡**：如果使用负载均衡，需要配置端口转发规则
3. **端口冲突**：在启动触发器时严格检查端口冲突
4. **优雅关闭**：触发器停止时释放端口和etcd记录
5. **监控告警**：监控端口监听状态，异常时自动重启
