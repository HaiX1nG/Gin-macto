package database

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/yourorg/livemix/config"
)

// ============================================
// 线程安全策略说明
// ============================================
// 使用 sync.Once 确保 rdb 只初始化一次，即使在并发场景下多次调用 InitRedis
// GetRedis 在未初始化时返回明确错误，避免 nil pointer panic
// CloseRedis 使用 sync.Once 确保只关闭一次
// ============================================

// ============================================
// Redis连接安全与性能配置说明
// ============================================
// 1. 密码认证：生产环境必须配置密码，防止未授权访问
// 2. 连接池：合理配置PoolSize/MinIdleConns，避免连接耗尽或频繁建连
// 3. 超时控制：DialTimeout/ReadTimeout/WriteTimeout/PoolTimeout
//    防止慢查询阻塞连接池，遵循阿里云Redis慢查询治理建议
// 4. 重试策略：MaxRetries配置命令重试，应对网络抖动
// 5. 空闲连接回收：ConnMaxIdleTime防止使用过期连接
// 6. 健康检查：初始化时Ping验证连接可用性
// ============================================

var (
	rdb            *redis.Client
	redisInitOnce  sync.Once
	redisCloseOnce sync.Once
	redisInitErr   error
)

// ErrRedisNotInitialized Redis未初始化错误
var ErrRedisNotInitialized = errors.New("Redis连接未初始化，请先调用 InitRedis")

// buildRedisOptions 根据配置构建Redis连接选项
// 遵循阿里云Redis最佳实践：合理配置连接池、超时、重试策略
func buildRedisOptions(cfg *config.RedisConfig) *redis.Options {
	opts := &redis.Options{
		// 网络地址
		Addr: cfg.Addr(),

		// 密码认证：生产环境必须配置，防止未授权访问
		// 通过环境变量 APP_REDIS_PASSWORD 设置
		Password: cfg.Password,

		// 数据库编号
		DB: cfg.DB,

		// 连接池配置
		// PoolSize: 最大连接数，默认CPU*10
		// 高并发场景建议根据QPS评估：连接数 ≈ QPS * 平均耗时(s)
		// 阿里云Redis单实例连接上限20000，需合理分配到各服务实例
		PoolSize: cfg.PoolSize,

		// MinIdleConns: 最小空闲连接数
		// 保持一定数量空闲连接，避免突发流量时频繁建立TCP连接
		// 建议设置为PoolSize的1/4到1/2
		MinIdleConns: cfg.MinIdleConns,

		// MaxRetries: 命令执行最大重试次数
		// 网络抖动时自动重试，但需注意幂等性
		// 非幂等命令（如INCR）重试可能导致数据不一致，业务层需自行保证
		MaxRetries: cfg.MaxRetries,

		// 超时配置（遵循阿里云Redis慢查询治理建议）
		// DialTimeout: 连接建立超时，默认5秒
		// 网络不稳定时可适当增加，但不宜超过10秒
		DialTimeout: cfg.DialTimeout,

		// ReadTimeout: 读操作超时，默认3秒
		// 大Value场景（>10KB）可适当增加，但需警惕慢查询
		// 阿里云建议：慢查询阈值10ms，超时应配合业务容忍度设置
		ReadTimeout: cfg.ReadTimeout,

		// WriteTimeout: 写操作超时，默认3秒
		WriteTimeout: cfg.WriteTimeout,

		// PoolTimeout: 连接池获取连接超时，默认4秒
		// 当所有连接被占用时，等待空闲连接的最长时间
		// 超过此时间返回ErrPoolExhausted，业务层应做降级处理
		PoolTimeout: cfg.PoolTimeout,

		// ConnMaxIdleTime: 连接最大空闲时间，默认5分钟
		// 超过此时间的空闲连接会被自动关闭
		// 防止使用已被服务端关闭的连接（如阿里云Redis的freeClient触发）
		ConnMaxIdleTime: cfg.ConnMaxIdleTime,
	}

	return opts
}

// InitRedis 初始化Redis连接
// 使用 sync.Once 确保只初始化一次，线程安全
// 初始化时会执行Ping健康检查，验证连接和认证是否正常
func InitRedis(cfg *config.RedisConfig) error {
	redisInitOnce.Do(func() {
		opts := buildRedisOptions(cfg)
		rdb = redis.NewClient(opts)

		// 健康检查：验证连接可用性和认证正确性
		// 使用独立超时上下文，避免依赖全局context
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := rdb.Ping(ctx).Err(); err != nil {
			redisInitErr = fmt.Errorf("连接Redis失败（请检查地址、密码及网络）: %w", err)
			return
		}
	})

	return redisInitErr
}

// GetRedis 获取Redis连接
// 如果Redis未初始化或初始化失败，返回 nil 和明确错误
// 调用方必须检查返回的错误
func GetRedis() (*redis.Client, error) {
	if rdb == nil || redisInitErr != nil {
		if redisInitErr != nil {
			return nil, fmt.Errorf("Redis初始化失败: %w", redisInitErr)
		}
		return nil, ErrRedisNotInitialized
	}
	return rdb, nil
}

// MustGetRedis 获取Redis连接，如果未初始化则 panic
// 仅用于启动时必须确保Redis已初始化的场景，运行时请使用 GetRedis
func MustGetRedis() *redis.Client {
	if rdb == nil {
		panic(ErrRedisNotInitialized)
	}
	if redisInitErr != nil {
		panic(fmt.Errorf("Redis初始化失败: %w", redisInitErr))
	}
	return rdb
}

// CloseRedis 关闭Redis连接
// 使用 sync.Once 确保只关闭一次，线程安全
func CloseRedis() error {
	var closeErr error
	redisCloseOnce.Do(func() {
		if rdb == nil {
			return
		}
		closeErr = rdb.Close()
	})
	return closeErr
}
