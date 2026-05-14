package repository

import (
	"context"

	"github.com/yourorg/livemix/internal/model"
	"gorm.io/gorm"
)

// TransactionManager 事务管理器接口
// 提供事务的开启、提交、回滚能力
// 使用方式：
//
//	tx := tm.Begin(ctx)
//	defer tx.Rollback() // 安全回滚，已提交时无操作
//	// ... 执行业务操作
//	if err := tx.Commit(); err != nil {
//	    return err
//	}
type TransactionManager interface {
	// Begin 开始事务，返回事务上下文
	Begin(ctx context.Context) TransactionContext
}

// TransactionContext 事务上下文接口
// 封装事务的生命周期管理
type TransactionContext interface {
	// Commit 提交事务
	Commit() error
	// Rollback 回滚事务，已提交或已回滚时为空操作
	Rollback() error
	// Context 获取带有事务的 context.Context
	Context() context.Context
	// DB 获取事务 DB 实例
	DB() *gorm.DB
}

// UserRepositoryInterface 用户仓储接口
type UserRepositoryInterface interface {
	Create(ctx context.Context, user *model.User) error
	FindByID(ctx context.Context, id uint64) (*model.User, error)
	FindByUsername(ctx context.Context, username string) (*model.User, error)
	FindByEmail(ctx context.Context, email string) (*model.User, error)
	Update(ctx context.Context, user *model.User) error
	ExistsByUsername(ctx context.Context, username string) (bool, error)
	ExistsByEmail(ctx context.Context, email string) (bool, error)
	Delete(ctx context.Context, id uint64) error
	// DeleteWithDB 使用指定 DB 删除用户（用于事务）
	// 注意：事务 DB 应通过 TransactionContext.DB() 获取，内部已包含 context
	DeleteWithDB(db *gorm.DB, id uint64) error
	// FindByIDs 批量查询用户
	FindByIDs(ctx context.Context, userIDs []uint64) (map[uint64]*model.User, error)
}

// RoomRepositoryInterface 房间仓储接口
type RoomRepositoryInterface interface {
	Create(ctx context.Context, room *model.Room) error
	FindByID(ctx context.Context, id uint64) (*model.Room, error)
	FindByInviteCode(ctx context.Context, inviteCode string) (*model.Room, error)
	List(ctx context.Context, roomType int8, page, pageSize int) ([]model.Room, int64, error)
	Update(ctx context.Context, room *model.Room) error
	Delete(ctx context.Context, id uint64) error
	GenerateInviteCode(ctx context.Context) (string, error)
	// CreateWithDB 使用指定 DB 创建房间（用于事务）
	// 注意：事务 DB 应通过 TransactionContext.DB() 获取，内部已包含 context
	CreateWithDB(db *gorm.DB, room *model.Room) error
	// DeleteWithDB 使用指定 DB 删除房间（用于事务）
	DeleteWithDB(db *gorm.DB, id uint64) error
}

// RoomParticipantRepositoryInterface 房间参与者仓储接口
type RoomParticipantRepositoryInterface interface {
	Create(ctx context.Context, participant *model.RoomParticipant) error
	FindByRoomAndUser(ctx context.Context, roomID, userID uint64) (*model.RoomParticipant, error)
	FindActiveByRoom(ctx context.Context, roomID uint64) ([]model.RoomParticipant, error)
	CountActiveByRoom(ctx context.Context, roomID uint64) (int, error)
	Update(ctx context.Context, participant *model.RoomParticipant) error
	Leave(ctx context.Context, roomID, userID uint64) error
	ExistsActive(ctx context.Context, roomID, userID uint64) (bool, error)
	// CreateWithDB 使用指定 DB 创建参与者记录（用于事务）
	CreateWithDB(db *gorm.DB, participant *model.RoomParticipant) error
	// LeaveWithDB 使用指定 DB 设置参与者离开（用于事务）
	LeaveWithDB(db *gorm.DB, roomID, userID uint64) error
	// DeleteByRoomWithDB 使用指定 DB 删除房间内所有参与者记录（用于事务）
	DeleteByRoomWithDB(db *gorm.DB, roomID uint64) error
	// FindActiveByUser 查询用户当前活跃的房间参与记录
	FindActiveByUser(ctx context.Context, userID uint64) ([]model.RoomParticipant, error)
	// FindActiveByUserWithRoom 查询用户当前活跃的房间参与记录（含房间信息）
	FindActiveByUserWithRoom(ctx context.Context, userID uint64) ([]struct {
		model.RoomParticipant
		Room model.Room
	}, error)
	// DeleteByRoom 删除房间内所有参与者记录
	DeleteByRoom(ctx context.Context, roomID uint64) error
	// CountActiveByRoomsBatch 批量统计多个房间的活跃参与者数量
	CountActiveByRoomsBatch(ctx context.Context, roomIDs []uint64) (map[uint64]int, error)
}

// PlaylistRepositoryInterface 播放列表仓储接口
type PlaylistRepositoryInterface interface {
	Create(ctx context.Context, item *model.PlaylistItem) error
	FindByID(ctx context.Context, id uint64) (*model.PlaylistItem, error)
	FindByRoom(ctx context.Context, roomID uint64) ([]model.PlaylistItem, error)
	FindWaitingByRoom(ctx context.Context, roomID uint64) ([]model.PlaylistItem, error)
	FindPlayingByRoom(ctx context.Context, roomID uint64) (*model.PlaylistItem, error)
	Update(ctx context.Context, item *model.PlaylistItem) error
	UpdateStatus(ctx context.Context, id uint64, status int8) error
	Delete(ctx context.Context, id uint64) error
	GetMaxOrder(ctx context.Context, roomID uint64) (uint32, error)
}

// ChatMessageRepositoryInterface 聊天消息仓储接口
type ChatMessageRepositoryInterface interface {
	Create(ctx context.Context, msg *model.ChatMessage) error
	FindByRoom(ctx context.Context, roomID uint64, page, pageSize int) ([]model.ChatMessage, int64, error)
	FindRecentByRoom(ctx context.Context, roomID uint64, limit int) ([]model.ChatMessage, error)
	SearchMessages(ctx context.Context, query string, roomID uint64, userID uint64, page, pageSize int) ([]model.ChatMessage, int64, error)
}
