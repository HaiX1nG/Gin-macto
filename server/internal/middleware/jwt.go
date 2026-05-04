package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/yourorg/livemix/pkg/errcode"
	"github.com/yourorg/livemix/pkg/jwt"
	"github.com/yourorg/livemix/pkg/response"
)

// JWTAuth JWT认证中间件
func JWTAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			response.Fail(c, errcode.ErrUnauthorized)
			c.Abort()
			return
		}

		// 解析Bearer Token
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			response.Fail(c, errcode.ErrTokenInvalid)
			c.Abort()
			return
		}

		tokenString := parts[1]
		claims, err := jwt.ParseAccessToken(tokenString)
		if err != nil {
			if err == jwt.ErrTokenExpired {
				response.Fail(c, errcode.ErrTokenExpired)
			} else {
				response.Fail(c, errcode.ErrTokenInvalid)
			}
			c.Abort()
			return
		}

		// 将用户信息存入上下文
		c.Set("userID", claims.UserID)
		c.Set("username", claims.Username)
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