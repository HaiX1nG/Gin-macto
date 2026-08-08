package repository

import (
	"context"

	"github.com/yourorg/livemix/internal/model"
	"gorm.io/gorm"
)

// TransactionManager 事务管理器接口
// 提供事务的开启、提交、回滚能力
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
	DeleteWithDB(db *gorm.DB, id uint64) error
	FindByIDs(ctx context.Context, userIDs []uint64) (map[uint64]*model.User, error)
}

// ServerRepositoryInterface 服务器仓储接口
type ServerRepositoryInterface interface {
	Create(ctx context.Context, server *model.Server) error
	CreateWithDB(db *gorm.DB, server *model.Server) error
	FindByID(ctx context.Context, id uint64) (*model.Server, error)
	FindByInviteCode(ctx context.Context, inviteCode string) (*model.Server, error)
	ListByUserID(ctx context.Context, userID uint64) ([]model.Server, error)
	Update(ctx context.Context, server *model.Server) error
	Delete(ctx context.Context, id uint64) error
	DeleteWithDB(db *gorm.DB, id uint64) error
	GenerateInviteCode(ctx context.Context) (string, error)
}

// ServerMemberRepositoryInterface 服务器成员仓储接口
type ServerMemberRepositoryInterface interface {
	AddMember(ctx context.Context, member *model.ServerMember) error
	AddMemberWithDB(db *gorm.DB, member *model.ServerMember) error
	RemoveMember(ctx context.Context, serverID, userID uint64) error
	FindByServerAndUser(ctx context.Context, serverID, userID uint64) (*model.ServerMember, error)
	ListByServer(ctx context.Context, serverID uint64) ([]model.ServerMember, error)
	ListServersByUser(ctx context.Context, userID uint64) ([]model.ServerMember, error)
	UpdateMember(ctx context.Context, member *model.ServerMember) error
	Exists(ctx context.Context, serverID, userID uint64) (bool, error)
	CountByServer(ctx context.Context, serverID uint64) (int, error)
}

// RoleRepositoryInterface 角色仓储接口
type RoleRepositoryInterface interface {
	Create(ctx context.Context, role *model.Role) error
	CreateWithDB(db *gorm.DB, role *model.Role) error
	FindByID(ctx context.Context, id uint64) (*model.Role, error)
	FindByServer(ctx context.Context, serverID uint64) ([]model.Role, error)
	Update(ctx context.Context, role *model.Role) error
	Delete(ctx context.Context, id uint64) error
	ListByServerOrdered(ctx context.Context, serverID uint64) ([]model.Role, error)
	// AssignRoleToMember 为成员分配角色
	AssignRoleToMember(ctx context.Context, serverID, memberID, roleID uint64) error
	// RemoveRoleFromMember 移除成员的角色
	RemoveRoleFromMember(ctx context.Context, serverID, memberID, roleID uint64) error
	// FindRolesByMember 查询成员的所有角色
	FindRolesByMember(ctx context.Context, memberID uint64) ([]model.Role, error)
	// SetMemberRoles 设置成员的角色列表（覆盖）
	SetMemberRoles(ctx context.Context, serverID, memberID uint64, roleIDs []uint64) error
	// FindMemberRoleIDs 查询成员的角色ID列表
	FindMemberRoleIDs(ctx context.Context, memberID uint64) ([]uint64, error)
}

// ChannelRepositoryInterface 频道仓储接口
type ChannelRepositoryInterface interface {
	Create(ctx context.Context, channel *model.Channel) error
	CreateWithDB(db *gorm.DB, channel *model.Channel) error
	FindByID(ctx context.Context, id uint64) (*model.Channel, error)
	FindByServer(ctx context.Context, serverID uint64) ([]model.Channel, error)
	Update(ctx context.Context, channel *model.Channel) error
	Delete(ctx context.Context, id uint64) error
	Reorder(ctx context.Context, serverID uint64, orders []struct {
		ID       uint64
		Position int
	}) error
	DeleteByServer(ctx context.Context, serverID uint64) error
}

// MessageRepositoryInterface 频道消息仓储接口
type MessageRepositoryInterface interface {
	Create(ctx context.Context, msg *model.ChannelMessage) error
	FindByID(ctx context.Context, id uint64) (*model.ChannelMessage, error)
	FindByChannel(ctx context.Context, channelID uint64, page, pageSize int) ([]model.ChannelMessage, int64, error)
	FindRecent(ctx context.Context, channelID uint64, limit int) ([]model.ChannelMessage, error)
	Update(ctx context.Context, msg *model.ChannelMessage) error
	UpdateContent(ctx context.Context, messageID uint64, content string) error
	Delete(ctx context.Context, id uint64) error
	FindPinned(ctx context.Context, channelID uint64) ([]model.ChannelMessage, error)
	SearchByChannel(ctx context.Context, query string, channelID uint64, page, pageSize int) ([]model.ChannelMessage, int64, error)
	UpdatePinStatus(ctx context.Context, messageID uint64, isPinned bool) error
}

// ReactionRepositoryInterface 表情反应仓储接口
type ReactionRepositoryInterface interface {
	Add(ctx context.Context, reaction *model.MessageReaction) error
	Remove(ctx context.Context, messageID, userID uint64, emoji string) error
	ListByMessage(ctx context.Context, messageID uint64) ([]model.MessageReaction, error)
}

// VoiceRepositoryInterface 语音仓储接口
type VoiceRepositoryInterface interface {
	// VoiceParticipant 实时状态
	UpsertParticipant(ctx context.Context, p *model.VoiceParticipant) error
	DeleteParticipant(ctx context.Context, channelID, userID uint64) error
	FindParticipantsByChannel(ctx context.Context, channelID uint64) ([]model.VoiceParticipant, error)
	UpdateMute(ctx context.Context, channelID, userID uint64, isMuted bool) error
	// VoiceSession 历史记录
	CreateSession(ctx context.Context, session *model.VoiceSession) error
	EndSession(ctx context.Context, channelID, userID uint64) error
	// ScreenShareSession
	CreateScreenShare(ctx context.Context, session *model.ScreenShareSession) error
	FindActiveScreenShareByChannel(ctx context.Context, channelID uint64) (*model.ScreenShareSession, error)
	EndScreenShare(ctx context.Context, channelID, userID uint64) error
}

// PlaylistRepositoryInterface 播放列表仓储接口
type PlaylistRepositoryInterface interface {
	Create(ctx context.Context, item *model.PlaylistItem) error
	FindByID(ctx context.Context, id uint64) (*model.PlaylistItem, error)
	FindByChannel(ctx context.Context, channelID uint64) ([]model.PlaylistItem, error)
	FindWaitingByChannel(ctx context.Context, channelID uint64) ([]model.PlaylistItem, error)
	FindPlayingByChannel(ctx context.Context, channelID uint64) (*model.PlaylistItem, error)
	Update(ctx context.Context, item *model.PlaylistItem) error
	UpdateStatus(ctx context.Context, id uint64, status int8) error
	Delete(ctx context.Context, id uint64) error
	GetMaxOrder(ctx context.Context, channelID uint64) (uint32, error)
	Reorder(ctx context.Context, channelID uint64, itemIDs []uint64) error
}
