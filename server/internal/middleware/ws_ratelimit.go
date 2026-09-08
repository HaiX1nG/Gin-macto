// Package middleware 提供 WebSocket 安全中间件
// 包含速率限制、连接数限制、消息大小限制等安全功能
//
// 遵循《阿里巴巴 Java 开发手册》Go 适配版安全规范：
// - 防止 DoS 攻击：限制连接数和消息频率
// - 防止资源耗尽：限制消息大小
// - 防止未授权访问：强化来源验证
package middleware

import (
	"sync"
	"time"
)

// WSRateLimiterConfig WebSocket 速率限制器配置
type WSRateLimiterConfig struct {
	// RateLimitPerSecond 每秒最大消息数，默认 30
	RateLimitPerSecond int
	// BurstSize 桶容量，默认 60（2倍速率）
	BurstSize int
	// CleanupInterval 清理间隔，默认 30s
	CleanupInterval time.Duration
	// ClientExpiration 客户端过期时间，默认 60s
	ClientExpiration time.Duration
}

// DefaultWSRateLimiterConfig 返回默认的 WebSocket 速率限制器配置
func DefaultWSRateLimiterConfig() WSRateLimiterConfig {
	return WSRateLimiterConfig{
		RateLimitPerSecond: 30,
		BurstSize:          60,
		CleanupInterval:    30 * time.Second,
		ClientExpiration:   60 * time.Second,
	}
}

// tokenBucket 令牌桶，用于平滑限流
// 令牌桶算法允许合理突发，避免固定窗口的边界突发问题
type tokenBucket struct {
	mu         sync.Mutex
	tokens     float64
	maxTokens  float64
	refillRate float64
	lastRefill time.Time
	lastActive time.Time
}

// allow 检查是否允许一条消息通过
// 返回 true 表示允许，false 表示被限流
func (tb *tokenBucket) allow() bool {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	now := time.Now()
	// 计算需要补充的令牌数
	elapsed := now.Sub(tb.lastRefill).Seconds()
	tb.tokens = min(tb.tokens+elapsed*tb.refillRate, tb.maxTokens)
	tb.lastRefill = now

	if tb.tokens >= 1 {
		tb.tokens--
		tb.lastActive = now
		return true
	}
	return false
}

// WSRateLimiter WebSocket 速率限制器
// 基于令牌桶算法，每个客户端独立计数
type WSRateLimiter struct {
	mu      sync.RWMutex
	clients map[string]*tokenBucket
	config  WSRateLimiterConfig
	stopCh  chan struct{}
}

// NewWSRateLimiter 创建 WebSocket 速率限制器实例
// 启动后台清理协程，定期移除过期客户端
func NewWSRateLimiter(config WSRateLimiterConfig) *WSRateLimiter {
	rl := &WSRateLimiter{
		clients: make(map[string]*tokenBucket),
		config:  config,
		stopCh:  make(chan struct{}),
	}
	go rl.cleanupLoop()
	return rl
}

// RegisterClient 注册客户端到速率限制器
// 在 WebSocket 连接建立时调用
func (rl *WSRateLimiter) RegisterClient(clientID string) {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	rl.clients[clientID] = &tokenBucket{
		tokens:     float64(rl.config.BurstSize),
		maxTokens:  float64(rl.config.BurstSize),
		refillRate: float64(rl.config.RateLimitPerSecond),
		lastRefill: time.Now(),
		lastActive: time.Now(),
	}
}

// UnregisterClient 从速率限制器移除客户端
// 在 WebSocket 断开连接时调用
func (rl *WSRateLimiter) UnregisterClient(clientID string) {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	delete(rl.clients, clientID)
}

// AllowMessage 检查客户端是否允许发送消息
// 返回 true 表示允许，false 表示被限流
func (rl *WSRateLimiter) AllowMessage(clientID string) bool {
	rl.mu.RLock()
	bucket, exists := rl.clients[clientID]
	rl.mu.RUnlock()

	if !exists {
		return true // 未注册的客户端不限制（首次消息时可能尚未注册）
	}
	return bucket.allow()
}

// cleanupLoop 后台清理协程，定期移除过期客户端
func (rl *WSRateLimiter) cleanupLoop() {
	ticker := time.NewTicker(rl.config.CleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			rl.cleanup()
		case <-rl.stopCh:
			return
		}
	}
}

// cleanup 移除过期未活跃的客户端
func (rl *WSRateLimiter) cleanup() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	for id, bucket := range rl.clients {
		bucket.mu.Lock()
		expired := now.Sub(bucket.lastActive) > rl.config.ClientExpiration
		bucket.mu.Unlock()
		if expired {
			delete(rl.clients, id)
		}
	}
}

// Stop 停止速率限制器的后台清理协程
func (rl *WSRateLimiter) Stop() {
	close(rl.stopCh)
}

// min 返回两个 float64 中的较小值
func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}
