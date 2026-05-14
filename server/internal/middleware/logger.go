// Package middleware 提供 HTTP 中间件功能
package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/yourorg/livemix/pkg/logger"
	"go.uber.org/zap"
)

// Logger 日志中间件
// 记录请求和响应信息，包括请求方法、路径、状态码、响应时间等
// 仅在开发和测试环境启用，生产环境建议关闭以减少日志量
func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		// 处理请求
		c.Next()

		// 计算响应时间
		latency := time.Since(start)

		// 获取 TraceID
		traceID := logger.GetTraceID(c.Request.Context())

		// 记录请求日志
		fields := []zap.Field{
			logger.WithTraceID(traceID),
			zap.String("method", c.Request.Method),
			zap.String("path", path),
			zap.String("query", query),
			zap.Int("status", c.Writer.Status()),
			zap.Duration("latency", latency),
			zap.String("clientIP", c.ClientIP()),
			zap.String("userAgent", c.Request.UserAgent()),
		}

		// 根据状态码选择日志级别
		status := c.Writer.Status()
		switch {
		case status >= 500:
			logger.Error("HTTP请求", fields...)
		case status >= 400:
			logger.Warn("HTTP请求", fields...)
		default:
			logger.Info("HTTP请求", fields...)
		}
	}
}
