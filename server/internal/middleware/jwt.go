package middleware

import (
	"context"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/yourorg/livemix/pkg/errcode"
	"github.com/yourorg/livemix/pkg/jwt"
	"github.com/yourorg/livemix/pkg/logger"
	"go.uber.org/zap"
)

// JWTAuth JWT认证中间件
// 认证流程：
// 1. 解析Authorization头获取Bearer Token
// 2. 验证Token签名和有效期
// 3. 检查Token是否在黑名单中（退出登录、修改密码等场景）
// 4. 将用户信息存入上下文供后续处理使用
func JWTAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(401, gin.H{
				"code":    errcode.ErrUnauthorized.Code,
				"message": errcode.ErrUnauthorized.Message,
				"data":    nil,
			})
			return
		}

		// 解析Bearer Token
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.AbortWithStatusJSON(401, gin.H{
				"code":    errcode.ErrTokenInvalid.Code,
				"message": errcode.ErrTokenInvalid.Message,
				"data":    nil,
			})
			return
		}

		tokenString := parts[1]

		// 解析并验证Token
		claims, err := jwt.ParseAccessToken(tokenString)
		if err != nil {
			statusCode := 401
			errResp := errcode.ErrTokenInvalid
			if err == jwt.ErrTokenExpired {
				errResp = errcode.ErrTokenExpired
			}
			c.AbortWithStatusJSON(statusCode, gin.H{
				"code":    errResp.Code,
				"message": errResp.Message,
				"data":    nil,
			})
			return
		}

		// 检查Token是否在黑名单中
		// 使用短超时避免阻塞请求
		ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
		defer cancel()

		isBlacklisted, err := jwt.IsBlacklisted(ctx, tokenString)
		if err != nil {
			// Redis不可用时，降级处理，不阻止请求
			// 记录警告日志但不中断认证流程
			logger.Warn("检查Token黑名单失败，降级处理",
				zap.String("traceID", c.GetString("traceID")),
				zap.Error(err),
			)
		}

		if isBlacklisted {
			c.AbortWithStatusJSON(401, gin.H{
				"code":    errcode.ErrTokenBlacklisted.Code,
				"message": errcode.ErrTokenBlacklisted.Message,
				"data":    nil,
			})
			return
		}

		// 将用户信息存入上下文
		c.Set("userID", claims.UserID)
		c.Set("username", claims.Username)

		// 将原始Token存入上下文，供退出登录等场景使用
		c.Set("accessToken", tokenString)

		c.Next()
	}
}

// GetUserID 从上下文获取用户ID
func GetUserID(c *gin.Context) uint64 {
	userID, exists := c.Get("userID")
	if !exists {
		return 0
	}
	return userID.(uint64)
}

// GetUsername 从上下文获取用户名
func GetUsername(c *gin.Context) string {
	username, exists := c.Get("username")
	if !exists {
		return ""
	}
	return username.(string)
}

// GetAccessToken 从上下文获取原始Access Token
// 用于退出登录时将Token加入黑名单
func GetAccessToken(c *gin.Context) string {
	token, exists := c.Get("accessToken")
	if !exists {
		return ""
	}
	return token.(string)
}
