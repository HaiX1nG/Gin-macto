package config

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/viper"
)

// Config 应用配置结构体
type Config struct {
	Server    ServerConfig    `mapstructure:"server"`
	Database  DatabaseConfig  `mapstructure:"database"`
	Redis     RedisConfig     `mapstructure:"redis"`
	JWT       JWTConfig       `mapstructure:"jwt"`
	CORS      CORSConfig      `mapstructure:"cors"`
	Log       LogConfig       `mapstructure:"log"`
	WebSocket WebSocketConfig `mapstructure:"websocket"`
}

// ServerConfig HTTP服务器配置
type ServerConfig struct {
	Port         int           `mapstructure:"port"`
	ReadTimeout  time.Duration `mapstructure:"read_timeout"`
	WriteTimeout time.Duration `mapstructure:"write_timeout"`
}

// DatabaseConfig MySQL数据库配置
type DatabaseConfig struct {
	Host         string `mapstructure:"host"`
	Port         int    `mapstructure:"port"`
	Username     string `mapstructure:"username"`
	Password     string `mapstructure:"password"`
	Database     string `mapstructure:"database"`
	MaxIdleConns int    `mapstructure:"max_idle_conns"`
	MaxOpenConns int    `mapstructure:"max_open_conns"`
}

// RedisConfig Redis配置
// 遵循阿里巴巴Redis开发规范：合理配置连接池、超时、重试策略
type RedisConfig struct {
	// Host Redis服务器地址
	Host string `mapstructure:"host"`
	// Port Redis服务器端口
	Port int `mapstructure:"port"`
	// Password Redis认证密码，生产环境必须设置
	// 通过环境变量 APP_REDIS_PASSWORD 配置
	Password string `mapstructure:"password"`
	// DB Redis数据库编号，默认0
	DB int `mapstructure:"db"`
	// PoolSize 连接池最大连接数
	// 默认为CPU核心数*10，高并发场景建议根据业务调整
	// 阿里云Redis建议：单实例连接数上限为20000，合理分配到各服务
	PoolSize int `mapstructure:"pool_size"`
	// MinIdleConns 最小空闲连接数
	// 保持一定数量的空闲连接，避免频繁建立连接的开销
	// 建议设置为PoolSize的1/4到1/2
	MinIdleConns int `mapstructure:"min_idle_conns"`
	// MaxRetries 命令执行最大重试次数
	// 网络抖动时自动重试，建议3次
	MaxRetries int `mapstructure:"max_retries"`
	// DialTimeout 连接建立超时时间
	// 默认5秒，网络不稳定时可适当增加
	DialTimeout time.Duration `mapstructure:"dial_timeout"`
	// ReadTimeout 读操作超时时间
	// 默认3秒，大Value场景可适当增加
	ReadTimeout time.Duration `mapstructure:"read_timeout"`
	// WriteTimeout 写操作超时时间
	// 默认3秒
	WriteTimeout time.Duration `mapstructure:"write_timeout"`
	// PoolTimeout 连接池获取连接超时时间
	// 当连接池耗尽时等待空闲连接的最长时间
	// 默认4秒，与ReadTimeout配合
	PoolTimeout time.Duration `mapstructure:"pool_timeout"`
	// ConnMaxIdleTime 连接最大空闲时间
	// 超过此时间的空闲连接会被关闭，防止使用过期连接
	// 建议设置为5分钟
	ConnMaxIdleTime time.Duration `mapstructure:"conn_max_idle_time"`
}

// JWTConfig JWT配置
type JWTConfig struct {
	AccessTokenSecret  string        `mapstructure:"access_token_secret"`
	RefreshTokenSecret string        `mapstructure:"refresh_token_secret"`
	AccessTokenExpire  time.Duration `mapstructure:"access_token_expire"`
	RefreshTokenExpire time.Duration `mapstructure:"refresh_token_expire"`
	Issuer             string        `mapstructure:"issuer"`
}

// CORSConfig CORS跨域配置
// 安全规范：生产环境必须配置明确的允许域名列表，禁止使用 "*"
type CORSConfig struct {
	// AllowOrigins 允许的域名白名单
	// 示例: ["https://example.com", "https://admin.example.com"]
	// 开发环境可包含: ["http://localhost:3000", "http://127.0.0.1:3000"]
	AllowOrigins []string `mapstructure:"allow_origins"`

	// AllowMethods 允许的HTTP方法
	AllowMethods []string `mapstructure:"allow_methods"`

	// AllowHeaders 允许的请求头
	AllowHeaders []string `mapstructure:"allow_headers"`

	// ExposeHeaders 暴露给客户端的响应头
	ExposeHeaders []string `mapstructure:"expose_headers"`

	// AllowCredentials 是否允许携带凭证（cookies等）
	// 注意：当 AllowCredentials 为 true 时，AllowOrigins 不能包含 "*"
	AllowCredentials bool `mapstructure:"allow_credentials"`

	// MaxAge 预检请求缓存时间（秒）
	MaxAge int `mapstructure:"max_age"`
}

// LogConfig 日志配置
type LogConfig struct {
	// Level 日志级别 (debug/info/warn/error)
	Level string `mapstructure:"level"`
	// Format 输出格式 (console/json)，开发环境用console，生产环境用json
	Format string `mapstructure:"format"`
	// OutputPaths 输出路径，支持 stdout、stderr 或文件路径
	OutputPaths []string `mapstructure:"output_paths"`
	// EnableCaller 是否启用调用者信息（文件名和行号）
	EnableCaller bool `mapstructure:"enable_caller"`
}

// WebSocketConfig WebSocket配置
type WebSocketConfig struct {
	// HeartbeatTimeout 心跳超时时间，超过此时间未收到客户端响应则断开连接
	// 默认60秒，与前端心跳间隔（通常30秒）配合使用，允许丢失一次心跳
	HeartbeatTimeout time.Duration `mapstructure:"heartbeat_timeout"`
	// PingInterval 服务端发送Ping消息的间隔
	// 默认30秒，客户端需要在超时前回复Pong
	PingInterval time.Duration `mapstructure:"ping_interval"`
	// WriteTimeout 写操作超时时间
	WriteTimeout time.Duration `mapstructure:"write_timeout"`
	// ReadTimeout 读操作超时时间
	ReadTimeout time.Duration `mapstructure:"read_timeout"`
	// SendBufferSize 发送缓冲区大小
	SendBufferSize int `mapstructure:"send_buffer_size"`
}

// DSN 返回MySQL连接字符串
func (c *DatabaseConfig) DSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		c.Username, c.Password, c.Host, c.Port, c.Database)
}

// Addr 返回Redis地址
func (c *RedisConfig) Addr() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

var globalConfig *Config

// Load 加载配置文件
// 配置优先级：环境变量 > 配置文件 > 默认值
// 敏感配置项（JWT Secret、数据库密码、Redis密码）支持环境变量覆盖
func Load(configPath string) error {
	viper.SetConfigFile(configPath)
	viper.SetConfigType("yaml")

	// 设置默认值
	setDefaults()

	// 读取配置文件
	if err := viper.ReadInConfig(); err != nil {
		return fmt.Errorf("读取配置文件失败: %w", err)
	}

	// 启用环境变量支持
	// 环境变量命名规范：APP_<配置路径>，如 APP_JWT_ACCESS_TOKEN_SECRET
	viper.SetEnvPrefix("APP")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	// 绑定敏感配置项的环境变量
	bindSensitiveEnvVars()

	// 解析配置到结构体
	if err := viper.Unmarshal(&globalConfig); err != nil {
		return fmt.Errorf("解析配置文件失败: %w", err)
	}

	// 验证敏感配置（生产环境必须通过环境变量设置）
	if err := validateSensitiveConfig(); err != nil {
		return fmt.Errorf("配置验证失败: %w", err)
	}

	return nil
}

// bindSensitiveEnvVars 绑定敏感配置项的环境变量
// 环境变量命名规范：
// - APP_JWT_ACCESS_TOKEN_SECRET: JWT访问令牌密钥
// - APP_JWT_REFRESH_TOKEN_SECRET: JWT刷新令牌密钥
// - APP_DB_PASSWORD: 数据库密码
// - APP_REDIS_PASSWORD: Redis密码
func bindSensitiveEnvVars() {
	// JWT密钥
	_ = viper.BindEnv("jwt.access_token_secret", "APP_JWT_ACCESS_TOKEN_SECRET")
	_ = viper.BindEnv("jwt.refresh_token_secret", "APP_JWT_REFRESH_TOKEN_SECRET")

	// 数据库密码
	_ = viper.BindEnv("database.password", "APP_DB_PASSWORD")

	// Redis密码
	_ = viper.BindEnv("redis.password", "APP_REDIS_PASSWORD")
}

// validateSensitiveConfig 验证敏感配置
// 生产环境必须通过环境变量设置敏感配置，禁止使用默认值
func validateSensitiveConfig() error {
	env := os.Getenv("APP_ENV")
	if env == "" {
		env = "development"
	}

	// 生产环境验证
	if env == "production" {
		// 检查JWT密钥是否为默认值
		defaultAccessTokenSecret := "your-access-token-secret-key-change-in-production"
		defaultRefreshTokenSecret := "your-refresh-token-secret-key-change-in-production"

		if globalConfig.JWT.AccessTokenSecret == "" ||
			globalConfig.JWT.AccessTokenSecret == defaultAccessTokenSecret {
			return fmt.Errorf("生产环境必须通过环境变量 APP_JWT_ACCESS_TOKEN_SECRET 设置JWT访问令牌密钥")
		}

		if globalConfig.JWT.RefreshTokenSecret == "" ||
			globalConfig.JWT.RefreshTokenSecret == defaultRefreshTokenSecret {
			return fmt.Errorf("生产环境必须通过环境变量 APP_JWT_REFRESH_TOKEN_SECRET 设置JWT刷新令牌密钥")
		}

		// 检查数据库密码
		if globalConfig.Database.Password == "" || globalConfig.Database.Password == "123456" {
			return fmt.Errorf("生产环境必须通过环境变量 APP_DB_PASSWORD 设置数据库密码")
		}

		// 检查Redis密码（如果Redis需要认证）
		if globalConfig.Redis.Password == "" {
			// 生产环境Redis应该设置密码，但这里只给出警告
			fmt.Println("[WARN] 生产环境建议通过环境变量 APP_REDIS_PASSWORD 设置Redis密码")
		}
	}

	return nil
}

// setDefaults 设置配置默认值
// 注意：敏感配置项不设置默认值，必须通过配置文件或环境变量提供
func setDefaults() {
	viper.SetDefault("server.port", 8080)
	viper.SetDefault("server.read_timeout", "10s")
	viper.SetDefault("server.write_timeout", "10s")
	viper.SetDefault("database.host", "127.0.0.1")
	viper.SetDefault("database.port", 3306)
	viper.SetDefault("database.max_idle_conns", 10)
	viper.SetDefault("database.max_open_conns", 100)
	viper.SetDefault("redis.host", "127.0.0.1")
	viper.SetDefault("redis.port", 6379)
	viper.SetDefault("redis.db", 0)
	// Redis连接池配置（遵循阿里云Redis最佳实践）
	// PoolSize: 默认CPU核心数*10，单机开发环境设为10
	viper.SetDefault("redis.pool_size", 10)
	// MinIdleConns: 保持2个空闲连接，避免频繁建连
	viper.SetDefault("redis.min_idle_conns", 2)
	// MaxRetries: 网络抖动时重试3次
	viper.SetDefault("redis.max_retries", 3)
	// DialTimeout: 连接建立超时5秒
	viper.SetDefault("redis.dial_timeout", "5s")
	// ReadTimeout: 读操作超时3秒
	viper.SetDefault("redis.read_timeout", "3s")
	// WriteTimeout: 写操作超时3秒
	viper.SetDefault("redis.write_timeout", "3s")
	// PoolTimeout: 连接池等待超时4秒
	viper.SetDefault("redis.pool_timeout", "4s")
	// ConnMaxIdleTime: 空闲连接最大存活5分钟
	viper.SetDefault("redis.conn_max_idle_time", "5m")
	// CORS 默认配置：开发环境允许 localhost，生产环境必须显式配置
	viper.SetDefault("cors.allow_origins", []string{
		"http://localhost:3000",
		"http://localhost:5173",
		"http://127.0.0.1:3000",
		"http://127.0.0.1:5173",
	})
	viper.SetDefault("cors.allow_methods", []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"})
	viper.SetDefault("cors.allow_headers", []string{
		"Origin", "Content-Type", "Content-Length",
		"Accept-Encoding", "X-Requested-With", "Authorization",
	})
	viper.SetDefault("cors.expose_headers", []string{
		"Content-Length", "Access-Control-Allow-Origin", "Access-Control-Allow-Headers",
	})
	viper.SetDefault("cors.allow_credentials", true)
	viper.SetDefault("cors.max_age", 86400) // 预检请求缓存24小时
	viper.SetDefault("jwt.access_token_expire", "15m")
	viper.SetDefault("jwt.refresh_token_expire", "168h")
	viper.SetDefault("jwt.issuer", "livemix")
	// 日志默认配置：开发环境使用console格式，生产环境应切换为json
	viper.SetDefault("log.level", "info")
	viper.SetDefault("log.format", "console")
	viper.SetDefault("log.output_paths", []string{"stdout"})
	viper.SetDefault("log.enable_caller", true)
	// WebSocket默认配置
	// 心跳超时60秒，允许客户端丢失一次心跳（前端通常30秒发送一次）
	viper.SetDefault("websocket.heartbeat_timeout", "60s")
	// Ping间隔30秒，与前端心跳配合
	viper.SetDefault("websocket.ping_interval", "30s")
	// 写超时10秒
	viper.SetDefault("websocket.write_timeout", "10s")
	// 读超时60秒，与心跳超时一致
	viper.SetDefault("websocket.read_timeout", "60s")
	// 发送缓冲区大小256条消息
	viper.SetDefault("websocket.send_buffer_size", 256)
}

// Get 获取全局配置
func Get() *Config {
	return globalConfig
}

// Init 初始化配置（用于测试）
func Init(cfg *Config) {
	globalConfig = cfg
}
