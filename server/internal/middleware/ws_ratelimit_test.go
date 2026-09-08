package middleware

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestWSRateLimiter_AllowMessage 测试令牌桶限流正常通过
func TestWSRateLimiter_AllowMessage(t *testing.T) {
	config := WSRateLimiterConfig{
		RateLimitPerSecond: 10,
		BurstSize:          20,
		CleanupInterval:    1 * time.Second,
		ClientExpiration:   5 * time.Second,
	}
	rl := NewWSRateLimiter(config)
	defer rl.Stop()

	clientID := "test-client-1"
	rl.RegisterClient(clientID)

	// 突发 20 条消息应该全部通过
	for i := 0; i < 20; i++ {
		assert.True(t, rl.AllowMessage(clientID), "第%d条消息应该通过", i+1)
	}

	// 第 21 条应该被限流
	assert.False(t, rl.AllowMessage(clientID), "第21条消息应该被限流")
}

// TestWSRateLimiter_RateRecovery 测试令牌桶恢复速率
func TestWSRateLimiter_RateRecovery(t *testing.T) {
	config := WSRateLimiterConfig{
		RateLimitPerSecond: 10,
		BurstSize:          10,
		CleanupInterval:    1 * time.Second,
		ClientExpiration:   5 * time.Second,
	}
	rl := NewWSRateLimiter(config)
	defer rl.Stop()

	clientID := "test-client-2"
	rl.RegisterClient(clientID)

	// 耗尽令牌
	for i := 0; i < 10; i++ {
		rl.AllowMessage(clientID)
	}
	assert.False(t, rl.AllowMessage(clientID), "令牌耗尽后应该被限流")

	// 等待 100ms 恢复约 1 个令牌
	time.Sleep(110 * time.Millisecond)
	assert.True(t, rl.AllowMessage(clientID), "等待恢复后应该通过")
}

// TestWSRateLimiter_UnregisteredClient 测试未注册客户端不限制
func TestWSRateLimiter_UnregisteredClient(t *testing.T) {
	config := DefaultWSRateLimiterConfig()
	rl := NewWSRateLimiter(config)
	defer rl.Stop()

	// 未注册的客户端应该不限制
	for i := 0; i < 100; i++ {
		assert.True(t, rl.AllowMessage("unregistered-client"))
	}
}

// TestWSRateLimiter_UnregisterClient 测试注销客户端
func TestWSRateLimiter_UnregisterClient(t *testing.T) {
	config := DefaultWSRateLimiterConfig()
	rl := NewWSRateLimiter(config)
	defer rl.Stop()

	clientID := "test-client-3"
	rl.RegisterClient(clientID)

	// 注销后再次注册应该重置令牌
	rl.UnregisterClient(clientID)
	rl.RegisterClient(clientID)

	// 应该能通过 BurstSize 条消息
	for i := 0; i < config.BurstSize; i++ {
		assert.True(t, rl.AllowMessage(clientID))
	}
}
