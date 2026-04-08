package common

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"fmt"
	"sync"
	"time"
)

// DBConnection 包装数据库连接，包含元数据
type DBConnection struct {
	DB          *sql.DB
	DSN         string
	CreatedAt   time.Time
	LastUsedAt  time.Time
	MaxLifetime time.Duration
}

// IsExpired 检查连接是否过期
func (c *DBConnection) IsExpired() bool {
	return time.Since(c.LastUsedAt) > c.MaxLifetime
}

// ConnectionPool 数据库连接池管理器
type ConnectionPool struct {
	mu           sync.RWMutex
	connections  map[string]*DBConnection
	maxLifetime  time.Duration
	cleanupTimer *time.Timer
}

// GlobalConnectionPool 全局连接池实例
var (
	GlobalConnectionPool *ConnectionPool
	poolOnce             sync.Once
)

// GetConnectionPool 获取全局连接池实例
func GetConnectionPool() *ConnectionPool {
	poolOnce.Do(func() {
		GlobalConnectionPool = &ConnectionPool{
			connections: make(map[string]*DBConnection),
			maxLifetime: 30 * time.Minute,
		}
		// 启动清理goroutine
		go GlobalConnectionPool.cleanupRoutine()
	})
	return GlobalConnectionPool
}

// GetOrCreateConnection 获取或创建数据库连接
func (p *ConnectionPool) GetOrCreateConnection(dsn string, openDBFunc func(string) (*sql.DB, error)) (*sql.DB, error) {
	// 生成连接唯一标识
	connID := p.generateConnID(dsn)

	p.mu.RLock()
	conn, exists := p.connections[connID]
	p.mu.RUnlock()

	if exists && !conn.IsExpired() {
		// 更新最后使用时间
		conn.LastUsedAt = time.Now()
		return conn.DB, nil
	}

	// 连接不存在或已过期，创建新连接
	return p.createConnection(connID, dsn, openDBFunc)
}

// createConnection 创建新连接
func (p *ConnectionPool) createConnection(connID, dsn string, openDBFunc func(string) (*sql.DB, error)) (*sql.DB, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	// 双重检查，防止重复创建
	if conn, exists := p.connections[connID]; exists && !conn.IsExpired() {
		conn.LastUsedAt = time.Now()
		return conn.DB, nil
	}

	// 创建新连接
	db, err := openDBFunc(dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// 测试连接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// 存储连接
	p.connections[connID] = &DBConnection{
		DB:          db,
		DSN:         dsn,
		CreatedAt:   time.Now(),
		LastUsedAt:  time.Now(),
		MaxLifetime: p.maxLifetime,
	}

	return db, nil
}

// generateConnID 生成连接唯一标识（DSN的SHA256哈希）
func (p *ConnectionPool) generateConnID(dsn string) string {
	hash := sha256.Sum256([]byte(dsn))
	return fmt.Sprintf("%x", hash[:8])
}

// cleanupRoutine 定期清理过期连接
func (p *ConnectionPool) cleanupRoutine() {
	ticker := time.NewTicker(10 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		p.cleanupExpiredConnections()
	}
}

// cleanupExpiredConnections 清理过期连接
func (p *ConnectionPool) cleanupExpiredConnections() {
	p.mu.Lock()
	defer p.mu.Unlock()

	for connID, conn := range p.connections {
		if conn.IsExpired() {
			conn.DB.Close()
			delete(p.connections, connID)
		}
	}
}

// Close 关闭所有连接（程序退出时调用）
func (p *ConnectionPool) Close() {
	p.mu.Lock()
	defer p.mu.Unlock()

	for connID, conn := range p.connections {
		conn.DB.Close()
		delete(p.connections, connID)
	}
}

// GetStats 获取连接池统计信息
func (p *ConnectionPool) GetStats() map[string]interface{} {
	p.mu.RLock()
	defer p.mu.RUnlock()

	stats := make(map[string]interface{})
	stats["total_connections"] = len(p.connections)

	for connID, conn := range p.connections {
		stats[connID] = map[string]interface{}{
			"created_at":   conn.CreatedAt,
			"last_used_at": conn.LastUsedAt,
			"is_expired":   conn.IsExpired(),
		}
	}

	return stats
}
