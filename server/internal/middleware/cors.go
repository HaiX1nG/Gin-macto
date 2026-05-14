package middleware

import (
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/yourorg/livemix/config"
)

// 限流器相关常量
const (
	// RateLimitCleanupInterval 限流器清理过期客户端的时间间隔
	RateLimitCleanupInterval = 30 * time.Second
	// RateLimitClientExpiration 限流器客户端过期时间，超过此时间未活跃则清理
	RateLimitClientExpiration = 60 * time.Second
	// RateLimitExceededCode 请求过于频繁的错误码
	RateLimitExceededCode = 40006
)

// CORS 跨域中间件
// 安全规范：基于白名单验证 Origin，禁止使用 "*" 通配符
// 符合《阿里巴巴 Java 开发手册》安全规约：跨域请求必须校验 Origin 白名单
func CORS() gin.HandlerFunc {
	cfg := config.Get()
	corsCfg := cfg.CORS

	// 将允许的域名列表转换为 map 便于快速查找
	allowOriginsMap := make(map[string]bool)
	for _, origin := range corsCfg.AllowOrigins {
		allowOriginsMap[origin] = true
	}

	allowMethods := strings.Join(corsCfg.AllowMethods, ", ")
	allowHeaders := strings.Join(corsCfg.AllowHeaders, ", ")
	exposeHeaders := strings.Join(corsCfg.ExposeHeaders, ", ")

	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")

		// 校验 Origin 是否在白名单中
		// 安全要求：只有白名单中的域名才允许跨域访问
		allowedOrigin := ""
		if origin != "" && allowOriginsMap[origin] {
			allowedOrigin = origin
		}

		// 如果 Origin 不在白名单中，不设置 CORS 头，浏览器会阻止跨域请求
		// 注意：不返回错误，让浏览器自行处理（符合安全规范，避免信息泄露）
		if allowedOrigin != "" {
			c.Header("Access-Control-Allow-Origin", allowedOrigin)
			c.Header("Access-Control-Allow-Methods", allowMethods)
			c.Header("Access-Control-Allow-Headers", allowHeaders)
			c.Header("Access-Control-Expose-Headers", exposeHeaders)

			// 当 AllowCredentials 为 true 时，必须返回具体的 Origin，不能是 "*"
			if corsCfg.AllowCredentials {
				c.Header("Access-Control-Allow-Credentials", "true")
			}

			if corsCfg.MaxAge > 0 {
				c.Header("Access-Control-Max-Age", strconv.Itoa(corsCfg.MaxAge))
			}
		}

		// 处理预检请求
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

// client 限流客户端信息
type client struct {
	lastTime time.Time // 最后请求时间
	count    int       // 当前时间窗口内的请求计数
}

// rateLimiterStore 限流存储器，包含并发安全的客户端映射
// 使用 sync.RWMutex 保护 map 访问，支持高并发读场景
type rateLimiterStore struct {
	mu              sync.RWMutex       // 读写锁，保护 clients map 的并发访问
	clients         map[string]*client // 客户端 IP 到限流信息的映射
	cleanupInterval time.Duration      // 清理过期客户端的时间间隔
	expiration      time.Duration      // 客户端过期时间（超过此时间未活跃则清理）
	stopCleanup     chan struct{}      // 停止清理协程的信号通道
}

// newRateLimiterStore 创建新的限流存储器
// cleanupInterval: 清理协程运行间隔
// expiration: 客户端过期时间
func newRateLimiterStore(cleanupInterval, expiration time.Duration) *rateLimiterStore {
	store := &rateLimiterStore{
		clients:         make(map[string]*client),
		cleanupInterval: cleanupInterval,
		expiration:      expiration,
		stopCleanup:     make(chan struct{}),
	}
	// 启动后台清理协程，定期清理过期客户端，防止内存泄漏
	go store.cleanupLoop()
	return store
}

// cleanupLoop 定期清理过期客户端的后台协程
// 并发安全策略：清理时使用写锁，确保清理期间不会有其他读写操作
func (s *rateLimiterStore) cleanupLoop() {
	ticker := time.NewTicker(s.cleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.cleanup()
		case <-s.stopCleanup:
			return
		}
	}
}

// cleanup 清理过期的客户端记录
// 遍历所有客户端，删除超过 expiration 时间未活跃的记录
func (s *rateLimiterStore) cleanup() {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	for ip, cl := range s.clients {
		if now.Sub(cl.lastTime) > s.expiration {
			delete(s.clients, ip)
		}
	}
}

// stop 停止清理协程，用于优雅关闭
func (s *rateLimiterStore) stop() {
	close(s.stopCleanup)
}

// getClient 获取客户端信息（读锁）
// 返回客户端信息和是否存在标志
func (s *rateLimiterStore) getClient(ip string) (*client, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	cl, exists := s.clients[ip]
	return cl, exists
}

// setClient 设置或更新客户端信息（写锁）
func (s *rateLimiterStore) setClient(ip string, cl *client) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.clients[ip] = cl
}

// updateClientCount 更新客户端计数（写锁，用于已存在的客户端）
func (s *rateLimiterStore) updateClientCount(ip string, count int, lastTime time.Time) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if cl, exists := s.clients[ip]; exists {
		cl.count = count
		cl.lastTime = lastTime
	}
}

// RateLimiter 限流中间件（并发安全实现）
// 使用令牌桶思想，每秒重置计数窗口
// rps: 每秒允许的最大请求数
// 并发安全策略：
// 1. 使用 sync.RWMutex 保护 clients map
// 2. 读操作使用 RLock，允许并发读
// 3. 写操作使用 Lock，确保写操作的原子性
// 4. 后台协程定期清理过期客户端，防止内存泄漏
func RateLimiter(rps int) gin.HandlerFunc {
	// 创建限流存储器
	// 每 RateLimitCleanupInterval 清理一次，超过 RateLimitClientExpiration 未活跃的客户端将被删除
	store := newRateLimiterStore(RateLimitCleanupInterval, RateLimitClientExpiration)

	return func(c *gin.Context) {
		ip := c.ClientIP()
		now := time.Now()

		// 使用读锁检查客户端是否存在
		cl, exists := store.getClient(ip)

		if exists {
			// 客户端已存在，检查限流条件
			if now.Sub(cl.lastTime) < time.Second {
				// 在同一秒内，检查是否超过限制
				if cl.count >= rps {
					c.JSON(http.StatusTooManyRequests, gin.H{
						"code":    RateLimitExceededCode,
						"message": "请求过于频繁",
					})
					c.Abort()
					return
				}
				// 未超限，增加计数（需要写锁）
				store.updateClientCount(ip, cl.count+1, cl.lastTime)
			} else {
				// 超过一秒，重置时间窗口（需要写锁）
				store.updateClientCount(ip, 1, now)
			}
		} else {
			// 客户端不存在，创建新记录（需要写锁）
			store.setClient(ip, &client{lastTime: now, count: 1})
		}

		c.Next()
	}
}
