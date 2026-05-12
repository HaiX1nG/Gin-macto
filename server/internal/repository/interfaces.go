package repository

import (
	"context"

	"github.com/yourorg/livemix/internal/model"
)

// UserRepositoryInterface 用户仓储接口
type UserRepositoryInterface interface {
	Create(ctx context.Context, user *model.User) error
	FindByID(ctx context.Context, id uint64) (*model.User, error)
	FindByUsername(ctx context.Context, username string) (*model.User, error)
	FindByEmail(ctx context.Context, email string) (*model.User, error)
	Update(ctx context.Context, user *model.User) error
	ExistsByUsername(ctx context.Context, username string) (bool, error)
	ExistsByEmail(ctx context.Context, email string) (bool, error)
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
