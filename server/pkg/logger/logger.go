// Package logger 提供基于 zap 的结构化日志功能
// 支持日志级别控制、开发/生产环境格式切换、TraceID 追踪
package logger

import (
	"context"
	"os"
	"sync"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Level 日志级别类型
type Level string

const (
	// DebugLevel 调试级别
	DebugLevel Level = "debug"
	// InfoLevel 信息级别
	InfoLevel Level = "info"
	// WarnLevel 警告级别
	WarnLevel Level = "warn"
	// ErrorLevel 错误级别
	ErrorLevel Level = "error"
)

// Config 日志配置
type Config struct {
	// Level 日志级别 (debug/info/warn/error)
	Level Level `mapstructure:"level"`
	// Format 输出格式 (console/json)
	Format string `mapstructure:"format"`
	// OutputPaths 输出路径，默认为 ["stdout"]
	OutputPaths []string `mapstructure:"output_paths"`
	// EnableCaller 是否启用调用者信息（文件名和行号）
	EnableCaller bool `mapstructure:"enable_caller"`
}

// DefaultConfig 返回默认日志配置
func DefaultConfig() *Config {
	return &Config{
		Level:        InfoLevel,
		Format:       "console",
		OutputPaths:  []string{"stdout"},
		EnableCaller: true,
	}
}

// contextKey 上下文键类型
type contextKey string

const (
	// TraceIDKey TraceID 在上下文中的键
	TraceIDKey contextKey = "traceId"
)

var (
	// globalLogger 全局日志实例
	globalLogger *zap.Logger
	// sugarLogger SugarLogger 实例，提供更便捷的 API
	sugarLogger *zap.SugaredLogger
	// once 确保只初始化一次
	once sync.Once
	// globalLevel 全局日志级别
	globalLevel zapcore.Level
)

// Init 初始化全局日志实例
// 参数:
//   - cfg: 日志配置，如果为 nil 则使用默认配置
//
// 返回:
//   - error: 初始化失败时返回错误
func Init(cfg *Config) error {
	var initErr error
	once.Do(func() {
		if cfg == nil {
			cfg = DefaultConfig()
		}

		// 解析日志级别
		globalLevel = parseLevel(cfg.Level)

		// 构建核心配置
		core := buildCore(cfg, globalLevel)

		// 构建日志选项
		opts := buildOptions(cfg)

		// 创建日志实例
		globalLogger = zap.New(core, opts...)
		sugarLogger = globalLogger.Sugar()
	})

	return initErr
}

// parseLevel 解析日志级别字符串
func parseLevel(level Level) zapcore.Level {
	switch level {
	case DebugLevel:
		return zapcore.DebugLevel
	case InfoLevel:
		return zapcore.InfoLevel
	case WarnLevel:
		return zapcore.WarnLevel
	case ErrorLevel:
		return zapcore.ErrorLevel
	default:
		return zapcore.InfoLevel
	}
}

// buildCore 构建 zap core
func buildCore(cfg *Config, level zapcore.Level) zapcore.Core {
	// 创建编码器
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:       "time",
		LevelKey:      "level",
		NameKey:       "logger",
		CallerKey:     "caller",
		MessageKey:    "msg",
		StacktraceKey: "stacktrace",
		LineEnding:    zapcore.DefaultLineEnding,
		EncodeLevel:   zapcore.CapitalLevelEncoder,
		EncodeTime:    zapcore.ISO8601TimeEncoder,
		EncodeCaller:  zapcore.ShortCallerEncoder,
	}

	// 开发环境使用更友好的格式
	if cfg.Format == "console" {
		encoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}

	// 选择编码器
	var encoder zapcore.Encoder
	if cfg.Format == "json" {
		encoder = zapcore.NewJSONEncoder(encoderConfig)
	} else {
		encoder = zapcore.NewConsoleEncoder(encoderConfig)
	}

	// 构建输出写入器
	writers := make([]zapcore.WriteSyncer, 0, len(cfg.OutputPaths))
	for _, path := range cfg.OutputPaths {
		if path == "stdout" {
			writers = append(writers, zapcore.AddSync(os.Stdout))
		} else if path == "stderr" {
			writers = append(writers, zapcore.AddSync(os.Stderr))
		} else {
			// 文件输出
			file, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
			if err == nil {
				writers = append(writers, zapcore.AddSync(file))
			}
		}
	}

	// 如果没有配置输出路径，默认使用 stdout
	if len(writers) == 0 {
		writers = append(writers, zapcore.AddSync(os.Stdout))
	}

	// 合并多个写入器
	writeSyncer := zapcore.NewMultiWriteSyncer(writers...)

	return zapcore.NewCore(encoder, writeSyncer, level)
}

// buildOptions 构建日志选项
func buildOptions(cfg *Config) []zap.Option {
	opts := make([]zap.Option, 0, 2)

	// 启用调用者信息
	if cfg.EnableCaller {
		opts = append(opts, zap.AddCaller(), zap.AddCallerSkip(1))
	}

	// 错误级别及以上自动添加堆栈信息
	opts = append(opts, zap.AddStacktrace(zapcore.ErrorLevel))

	return opts
}

// L 获取全局日志实例
func L() *zap.Logger {
	if globalLogger == nil {
		// 如果未初始化，使用默认配置初始化
		_ = Init(nil)
	}
	return globalLogger
}

// S 获取 Sugar 日志实例
func S() *zap.SugaredLogger {
	if sugarLogger == nil {
		_ = Init(nil)
	}
	return sugarLogger
}

// WithTraceID 添加 TraceID 到日志字段
func WithTraceID(traceID string) zap.Field {
	return zap.String("traceId", traceID)
}

// WithUserID 添加用户ID到日志字段
func WithUserID(userID uint64) zap.Field {
	return zap.Uint64("userId", userID)
}

// WithRoomID 添加房间ID到日志字段
func WithRoomID(roomID uint64) zap.Field {
	return zap.Uint64("roomId", roomID)
}

// WithError 添加错误到日志字段
func WithError(err error) zap.Field {
	return zap.Error(err)
}

// WithString 添加字符串字段
func WithString(key, value string) zap.Field {
	return zap.String(key, value)
}

// WithInt 添加整数字段
func WithInt(key string, value int) zap.Field {
	return zap.Int(key, value)
}

// WithAny 添加任意类型字段
func WithAny(key string, value any) zap.Field {
	return zap.Any(key, value)
}

// --- 上下文相关方法 ---

// Ctx 从上下文中获取 TraceID 并记录日志
// 使用方式: logger.Ctx(ctx).Info("消息", zap.String("key", "value"))
func Ctx(ctx context.Context) *zap.Logger {
	traceID, ok := ctx.Value(TraceIDKey).(string)
	if ok && traceID != "" {
		return L().With(WithTraceID(traceID))
	}
	return L()
}

// GetTraceID 从上下文获取 TraceID
func GetTraceID(ctx context.Context) string {
	traceID, _ := ctx.Value(TraceIDKey).(string)
	return traceID
}

// SetTraceID 设置 TraceID 到上下文
func SetTraceID(ctx context.Context, traceID string) context.Context {
	return context.WithValue(ctx, TraceIDKey, traceID)
}

// --- 便捷日志方法 ---

// Debug 记录调试级别日志
func Debug(msg string, fields ...zap.Field) {
	L().Debug(msg, fields...)
}

// Info 记录信息级别日志
func Info(msg string, fields ...zap.Field) {
	L().Info(msg, fields...)
}

// Warn 记录警告级别日志
func Warn(msg string, fields ...zap.Field) {
	L().Warn(msg, fields...)
}

// Error 记录错误级别日志
func Error(msg string, fields ...zap.Field) {
	L().Error(msg, fields...)
}

// Fatal 记录致命错误日志并退出程序
func Fatal(msg string, fields ...zap.Field) {
	L().Fatal(msg, fields...)
}

// Panic 记录恐慌日志并触发 panic
func Panic(msg string, fields ...zap.Field) {
	L().Panic(msg, fields...)
}

// Debugf 使用格式化记录调试日志
func Debugf(template string, args ...any) {
	S().Debugf(template, args...)
}

// Infof 使用格式化记录信息日志
func Infof(template string, args ...any) {
	S().Infof(template, args...)
}

// Warnf 使用格式化记录警告日志
func Warnf(template string, args ...any) {
	S().Warnf(template, args...)
}

// Errorf 使用格式化记录错误日志
func Errorf(template string, args ...any) {
	S().Errorf(template, args...)
}

// Fatalf 使用格式化记录致命错误日志并退出程序
func Fatalf(template string, args ...any) {
	S().Fatalf(template, args...)
}

// Sync 刷新日志缓冲区
func Sync() error {
	if globalLogger != nil {
		return globalLogger.Sync()
	}
	return nil
}
