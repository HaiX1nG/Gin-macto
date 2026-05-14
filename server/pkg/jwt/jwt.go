package jwt

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/yourorg/livemix/config"
	"github.com/yourorg/livemix/pkg/database"
)

// Claims JWT声明结构体
type Claims struct {
	UserID   uint64 `json:"user_id"`
	Username string `json:"username"`
	jwt.RegisteredClaims
}

// TokenPair Token对
type TokenPair struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
	ExpiresIn    int64  `json:"expiresIn"` // AccessToken过期时间（秒）
}

var (
	ErrTokenInvalid     = errors.New("token无效")
	ErrTokenExpired     = errors.New("token已过期")
	ErrTokenBlacklisted = errors.New("token已被加入黑名单")
)

// ============================================
// JWT黑名单机制
// ============================================
// 使用Redis存储黑名单，支持以下场景：
// 1. 用户主动退出登录
// 2. 修改密码后使旧Token失效
// 3. 管理员强制下线用户
//
// 黑名单Key设计：
// - 使用token的SHA256哈希值作为key，避免敏感信息泄露
// - Key格式: jwt:blacklist:<token_hash>
// - TTL设置为token的剩余有效期，自动过期清理
// ============================================

const (
	// blacklistKeyPrefix Redis黑名单key前缀
	blacklistKeyPrefix = "jwt:blacklist:"
)

// tokenHash 计算token的SHA256哈希值
// 使用哈希值作为Redis key，避免在日志或监控中暴露完整token
func tokenHash(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}

// blacklistKey 生成Redis黑名单key
func blacklistKey(token string) string {
	return blacklistKeyPrefix + tokenHash(token)
}

// AddToBlacklist 将token加入黑名单
// 参数：
//   - ctx: 上下文，用于超时控制
//   - token: 需要加入黑名单的JWT token字符串
//   - expiration: 黑名单有效期，通常设置为token的剩余有效期
//
// 返回：
//   - error: 操作失败时返回错误
//
// 使用场景：
//   - 用户退出登录
//   - 修改密码后使旧token失效
//   - 账户被强制下线
func AddToBlacklist(ctx context.Context, token string, expiration time.Duration) error {
	rdb, err := database.GetRedis()
	if err != nil {
		// Redis不可用时记录日志但不阻止业务流程
		// 黑名单检查会降级为仅验证token有效性
		return fmt.Errorf("获取Redis连接失败: %w", err)
	}

	key := blacklistKey(token)
	return rdb.Set(ctx, key, "1", expiration).Err()
}

// IsBlacklisted 检查token是否在黑名单中
// 参数：
//   - ctx: 上下文，用于超时控制
//   - token: 需要检查的JWT token字符串
//
// 返回：
//   - bool: true表示token在黑名单中，false表示不在黑名单中
//   - error: Redis操作失败时返回错误
//
// 注意：
//   - 当Redis不可用时，返回false（不阻止请求）
//   - 这是降级策略，确保Redis故障不影响正常业务
func IsBlacklisted(ctx context.Context, token string) (bool, error) {
	rdb, err := database.GetRedis()
	if err != nil {
		// Redis不可用时，降级处理，不阻止请求
		// 日志记录在GetRedis内部处理
		return false, err
	}

	key := blacklistKey(token)
	exists, err := rdb.Exists(ctx, key).Result()
	if err != nil {
		return false, fmt.Errorf("检查黑名单失败: %w", err)
	}

	return exists > 0, nil
}

// AddAccessTokenToBlacklist 将Access Token加入黑名单
// 自动计算token的剩余有效期作为TTL
// 参数：
//   - ctx: 上下文
//   - token: Access Token字符串
//
// 返回：
//   - error: 操作失败时返回错误
func AddAccessTokenToBlacklist(ctx context.Context, token string) error {
	claims, err := parseToken(token, config.Get().JWT.AccessTokenSecret)
	if err != nil {
		// token无效或已过期，无需加入黑名单
		return nil
	}

	// 计算剩余有效期
	now := time.Now()
	if claims.ExpiresAt != nil {
		expiration := claims.ExpiresAt.Time.Sub(now)
		if expiration <= 0 {
			// token已过期，无需加入黑名单
			return nil
		}
		return AddToBlacklist(ctx, token, expiration)
	}

	// 没有过期时间，使用默认配置的token有效期
	cfg := config.Get().JWT
	return AddToBlacklist(ctx, token, cfg.AccessTokenExpire)
}

// AddRefreshTokenToBlacklist 将Refresh Token加入黑名单
// 自动计算token的剩余有效期作为TTL
func AddRefreshTokenToBlacklist(ctx context.Context, token string) error {
	claims, err := parseToken(token, config.Get().JWT.RefreshTokenSecret)
	if err != nil {
		return nil
	}

	now := time.Now()
	if claims.ExpiresAt != nil {
		expiration := claims.ExpiresAt.Time.Sub(now)
		if expiration <= 0 {
			return nil
		}
		return AddToBlacklist(ctx, token, expiration)
	}

	cfg := config.Get().JWT
	return AddToBlacklist(ctx, token, cfg.RefreshTokenExpire)
}

// AddAllUserTokensToBlacklist 将用户的所有token加入黑名单
// 用于修改密码、删除账户等需要使所有token失效的场景
// 参数：
//   - ctx: 上下文
//   - accessToken: Access Token
//   - refreshToken: Refresh Token（可选，传空字符串则跳过）
//
// 返回：
//   - error: 操作失败时返回错误
func AddAllUserTokensToBlacklist(ctx context.Context, accessToken, refreshToken string) error {
	// 加入Access Token黑名单
	if err := AddAccessTokenToBlacklist(ctx, accessToken); err != nil {
		return fmt.Errorf("加入Access Token黑名单失败: %w", err)
	}

	// 加入Refresh Token黑名单
	if refreshToken != "" {
		if err := AddRefreshTokenToBlacklist(ctx, refreshToken); err != nil {
			return fmt.Errorf("加入Refresh Token黑名单失败: %w", err)
		}
	}

	return nil
}

// GenerateToken 生成Token对
func GenerateToken(userID uint64, username string) (*TokenPair, error) {
	cfg := config.Get().JWT

	now := time.Now()
	accessExpire := now.Add(cfg.AccessTokenExpire)
	refreshExpire := now.Add(cfg.RefreshTokenExpire)

	// Access Token
	accessClaims := Claims{
		UserID:   userID,
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(accessExpire),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    cfg.Issuer,
		},
	}
	accessToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims).
		SignedString([]byte(cfg.AccessTokenSecret))
	if err != nil {
		return nil, err
	}

	// Refresh Token
	refreshClaims := Claims{
		UserID:   userID,
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(refreshExpire),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    cfg.Issuer,
		},
	}
	refreshToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims).
		SignedString([]byte(cfg.RefreshTokenSecret))
	if err != nil {
		return nil, err
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int64(cfg.AccessTokenExpire.Seconds()),
	}, nil
}

// ParseAccessToken 解析Access Token
func ParseAccessToken(tokenString string) (*Claims, error) {
	return parseToken(tokenString, config.Get().JWT.AccessTokenSecret)
}

// ParseRefreshToken 解析Refresh Token
func ParseRefreshToken(tokenString string) (*Claims, error) {
	return parseToken(tokenString, config.Get().JWT.RefreshTokenSecret)
}

// parseToken 解析Token
func parseToken(tokenString, secret string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (any, error) {
		return []byte(secret), nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrTokenExpired
		}
		return nil, ErrTokenInvalid
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, ErrTokenInvalid
}
