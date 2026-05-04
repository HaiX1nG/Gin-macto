package middleware

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// CORS 跨域中间件
func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Content-Length, Accept-Encoding, X-Requested-With, Authorization")
		c.Header("Access-Control-Expose-Headers", "Content-Length, Access-Control-Allow-Origin, Access-Control-Allow-Headers")
		c.Header("Access-Control-Allow-Credentials", "true")

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

// RateLimiter 限流中间件（简单实现）
func RateLimiter(rps int) gin.HandlerFunc {
	type client struct {
		lastTime time.Time
		count    int
	}

	clients := make(map[string]*client)

	return func(c *gin.Context) {
		ip := c.ClientIP()
		now := time.Now()

		if cl, exists := clients[ip]; exists {
			if now.Sub(cl.lastTime) < time.Second {
				if cl.count >= rps {
					c.JSON(http.StatusTooManyRequests, gin.H{
						"code":    40006,
						"message": "请求过于频繁",
					})
					c.Abort()
					return
				}
				cl.count++
			} else {
				cl.lastTime = now
				cl.count = 1
			}
		} else {
			clients[ip] = &client{lastTime: now, count: 1}
		}

		c.Next()
	}
}
