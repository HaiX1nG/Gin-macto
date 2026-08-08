package database

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/yourorg/livemix/config"
	"github.com/yourorg/livemix/internal/model"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// ============================================
// 线程安全策略说明
// ============================================
// 使用 sync.Once 确保 db 只初始化一次，即使在并发场景下多次调用 InitDB
// GetDB 和 WithContext 在未初始化时返回明确错误，避免 nil pointer panic
// Close 使用 sync.Once 确保只关闭一次
// ============================================

var (
	db          *gorm.DB
	dbInitOnce  sync.Once
	dbCloseOnce sync.Once
	dbInitErr   error
)

// ErrDBNotInitialized 数据库未初始化错误
var ErrDBNotInitialized = errors.New("数据库连接未初始化，请先调用 InitDB")

// InitDB 初始化数据库连接
// 使用 sync.Once 确保只初始化一次，线程安全
// 如果初始化失败，后续调用 GetDB 将返回错误
func InitDB(cfg *config.DatabaseConfig) error {
	dbInitOnce.Do(func() {
		var err error

		gormConfig := &gorm.Config{
			Logger: logger.Default.LogMode(logger.Info),
		}

		db, err = gorm.Open(mysql.Open(cfg.DSN()), gormConfig)
		if err != nil {
			dbInitErr = fmt.Errorf("连接数据库失败: %w", err)
			return
		}

		sqlDB, err := db.DB()
		if err != nil {
			dbInitErr = fmt.Errorf("获取数据库连接池失败: %w", err)
			return
		}

		// 设置连接池参数
		sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
		sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
		sqlDB.SetConnMaxLifetime(time.Hour)

		// 自动迁移
		if err = autoMigrate(); err != nil {
			dbInitErr = fmt.Errorf("数据库迁移失败: %w", err)
			return
		}
	})

	return dbInitErr
}

// autoMigrate 自动迁移数据库表结构
func autoMigrate() error {
	return db.AutoMigrate(
		// 用户域
		&model.User{},
		&model.UserStatus{},
		// 服务器域
		&model.Server{},
		&model.ServerMember{},
		&model.Role{},
		&model.ServerMemberRole{},
		// 频道域
		&model.Channel{},
		// 消息域
		&model.ChannelMessage{},
		&model.MessageAttachment{},
		&model.MessageReaction{},
		// 语音/屏幕共享域
		&model.VoiceParticipant{},
		&model.VoiceSession{},
		&model.ScreenShareSession{},
		// 播放列表域
		&model.PlaylistItem{},
		// 好友域
		&model.FriendRequest{},
		&model.Friendship{},
		&model.PrivateMessage{},
	)
}

// GetDB 获取数据库连接
// 如果数据库未初始化或初始化失败，返回 nil 和 ErrDBNotInitialized 错误
// 调用方必须检查返回的错误
func GetDB() (*gorm.DB, error) {
	if db == nil || dbInitErr != nil {
		if dbInitErr != nil {
			return nil, fmt.Errorf("数据库初始化失败: %w", dbInitErr)
		}
		return nil, ErrDBNotInitialized
	}
	return db, nil
}

// MustGetDB 获取数据库连接，如果未初始化则 panic
// 仅用于启动时必须确保数据库已初始化的场景，运行时请使用 GetDB
func MustGetDB() *gorm.DB {
	if db == nil {
		panic(ErrDBNotInitialized)
	}
	if dbInitErr != nil {
		panic(fmt.Errorf("数据库初始化失败: %w", dbInitErr))
	}
	return db
}

// Close 关闭数据库连接
// 使用 sync.Once 确保只关闭一次，线程安全
func Close() error {
	var closeErr error
	dbCloseOnce.Do(func() {
		if db == nil {
			return
		}
		sqlDB, err := db.DB()
		if err != nil {
			closeErr = err
			return
		}
		closeErr = sqlDB.Close()
	})
	return closeErr
}

// WithContext 返回带有context的数据库连接
// 如果数据库未初始化，返回 nil 和 ErrDBNotInitialized 错误
func WithContext(ctx context.Context) (*gorm.DB, error) {
	if db == nil || dbInitErr != nil {
		if dbInitErr != nil {
			return nil, fmt.Errorf("数据库初始化失败: %w", dbInitErr)
		}
		return nil, ErrDBNotInitialized
	}
	return db.WithContext(ctx), nil
}
