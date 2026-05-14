// Package middleware 提供 HTTP 中间件功能
package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/yourorg/livemix/pkg/logger"
)

// RequestID 请求ID中间件
// 为每个请求生成唯一的 TraceID，用于日志追踪和问题定位
// TraceID 会添加到响应头 X-Request-Id 中，方便客户端关联问题
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 优先从请求头获取已有的 RequestID（支持分布式追踪）
		traceID := c.GetHeader("X-Request-Id")
		if traceID == "" {
			// 生成新的 TraceID
			traceID = uuid.New().String()
		}

		// 设置到上下文
		ctx := logger.SetTraceID(c.Request.Context(), traceID)
		c.Request = c.Request.WithContext(ctx)

		// 设置到响应头
		c.Header("X-Request-Id", traceID)

		c.Next()
	}
}
