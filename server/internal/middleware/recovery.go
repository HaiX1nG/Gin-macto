// Package middleware 提供 HTTP 中间件功能
package middleware

import (
	"net/http"
	"runtime/debug"

	"github.com/gin-gonic/gin"
	"github.com/yourorg/livemix/pkg/logger"
	"go.uber.org/zap"
)

// Recovery 恢复中间件
// 捕获 panic 并记录完整的堆栈信息，防止服务崩溃
// 返回统一的错误响应格式
func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// 获取堆栈信息
				stack := string(debug.Stack())

				// 获取 TraceID
				traceID := logger.GetTraceID(c.Request.Context())

				// 记录错误日志，包含完整堆栈
				logger.Error("Panic恢复",
					logger.WithTraceID(traceID),
					logger.WithAny("error", err),
					zap.String("stack", stack),
					zap.String("method", c.Request.Method),
					zap.String("path", c.Request.URL.Path),
					zap.String("clientIP", c.ClientIP()),
				)

				// 返回统一错误响应
				c.JSON(http.StatusInternalServerError, gin.H{
					"code":      50000,
					"message":   "服务器内部错误",
					"requestId": traceID,
				})
				c.Abort()
			}
		}()
		c.Next()
	}
}
