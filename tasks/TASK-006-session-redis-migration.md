# TASK-006 Session 存储迁移至 Redis

**优先级**: 中
**模块**: auth
**涉及层**: Infra

---

## 背景

当前 Session（Refresh Token）存储在 MySQL 数据库中。每次 token 验证都需要查询数据库，在高并发场景下会成为性能瓶颈。Session 数据天然适合 Redis 的 KV + TTL 模型，无需手动清理过期记录。

## 问题定位

```
internal/infra/persistence/mysql/session_repository/session_repository.go
// 基于 MySQL 的 session 存储，每次 ParseToken 后需要查询 DB 验证 session
```

## 实现方案

### 1. 实现 Redis 版本的 SessionStore

```go
package cache

type SessionStoreImpl struct {
    rdb *redis.Client `autowired:""`
}

func (s *SessionStoreImpl) SaveRefreshToken(
    ctx context.Context, sessionID, accountID, refreshHash, deviceJSON, ip string, ttl time.Duration,
) error {
    key := fmt.Sprintf("session:%s", sessionID)
    data := map[string]string{
        "accountID":    accountID,
        "refreshHash":  refreshHash,
        "deviceJSON":   deviceJSON,
        "ip":           ip,
        "createdAt":    time.Now().Format(time.RFC3339),
    }
    return s.rdb.HSet(ctx, key, data).Err()
}

func (s *SessionStoreImpl) GetRefreshToken(ctx context.Context, sessionID string) (*SessionData, error) {
    key := fmt.Sprintf("session:%s", sessionID)
    result, err := s.rdb.HGetAll(ctx, key).Result()
    // ...
}
```

### 2. 使用 autowire 切换实现

通过 `autowire.Auto` 注册 Redis 实现替换 MySQL 实现。根据配置中 `redis.addr` 是否为空决定使用哪个实现，或直接在 `config.yaml` 中增加 `auth.session_store: redis` 开关。

### 3. 迁移策略

1. 新写入走 Redis
2. 读取时先查 Redis，miss 后查 MySQL（双读过渡期）
3. 过渡期结束后移除 MySQL 实现

## 验收标准

- [ ] `SessionStore` Redis 实现通过 `autowire` 注册
- [ ] Session 写入 Redis 并设置 TTL，过期自动清理
- [ ] 现有 MySQL 实现可作为 fallback
- [ ] 切换实现无需修改 App Service 层代码
- [ ] 性能基准测试：Redis 读写延迟 < 1ms

## 相关文件

- `internal/domain/repository/session_repository.go`（接口）
- `internal/infra/persistence/mysql/session_repository/`（当前实现）
- `internal/infra/cache/`（新增 Redis 实现）
- `internal/infra/config/config.go`（配置项）
