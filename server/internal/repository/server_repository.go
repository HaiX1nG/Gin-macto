package repository

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"

	"github.com/yourorg/livemix/internal/model"
	"gorm.io/gorm"
)

// ServerRepository 服务器仓储
type ServerRepository struct {
	db *gorm.DB
}

// NewServerRepository 创建服务器仓储实例
func NewServerRepository(db *gorm.DB) *ServerRepository {
	return &ServerRepository{db: db}
}

// Create 创建服务器
func (r *ServerRepository) Create(ctx context.Context, server *model.Server) error {
	return r.db.WithContext(ctx).Create(server).Error
}

// CreateWithDB 使用指定 DB 创建服务器（用于事务）
func (r *ServerRepository) CreateWithDB(db *gorm.DB, server *model.Server) error {
	return db.Create(server).Error
}

// FindByID 根据ID查询服务器
func (r *ServerRepository) FindByID(ctx context.Context, id uint64) (*model.Server, error) {
	var server model.Server
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&server).Error
	if err != nil {
		return nil, err
	}
	return &server, nil
}

// FindByInviteCode 根据邀请码查询服务器
func (r *ServerRepository) FindByInviteCode(ctx context.Context, inviteCode string) (*model.Server, error) {
	var server model.Server
	err := r.db.WithContext(ctx).Where("invite_code = ?", inviteCode).First(&server).Error
	if err != nil {
		return nil, err
	}
	return &server, nil
}

// ListByUserID 查询用户加入的所有服务器
func (r *ServerRepository) ListByUserID(ctx context.Context, userID uint64) ([]model.Server, error) {
	var servers []model.Server
	err := r.db.WithContext(ctx).
		Joins("JOIN server_members ON server_members.server_id = servers.id").
		Where("server_members.user_id = ?", userID).
		Order("servers.created_at DESC").
		Find(&servers).Error
	return servers, err
}

// Update 更新服务器
func (r *ServerRepository) Update(ctx context.Context, server *model.Server) error {
	return r.db.WithContext(ctx).Save(server).Error
}

// Delete 删除服务器
func (r *ServerRepository) Delete(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Delete(&model.Server{}, id).Error
}

// DeleteWithDB 使用指定 DB 删除服务器（用于事务）
func (r *ServerRepository) DeleteWithDB(db *gorm.DB, id uint64) error {
	return db.Delete(&model.Server{}, id).Error
}

// GenerateInviteCode 生成唯一邀请码
func (r *ServerRepository) GenerateInviteCode(ctx context.Context) (string, error) {
	for i := 0; i < 10; i++ {
		bytes := make([]byte, 4)
		if _, err := rand.Read(bytes); err != nil {
			return "", err
		}
		code := hex.EncodeToString(bytes)

		var count int64
		if err := r.db.WithContext(ctx).Model(&model.Server{}).Where("invite_code = ?", code).Count(&count).Error; err != nil {
			return "", err
		}
		if count == 0 {
			return code, nil
		}
	}
	return "", errors.New("生成邀请码失败")
}

// ServerMemberRepository 服务器成员仓储
type ServerMemberRepository struct {
	db *gorm.DB
}

// NewServerMemberRepository 创建服务器成员仓储实例
func NewServerMemberRepository(db *gorm.DB) *ServerMemberRepository {
	return &ServerMemberRepository{db: db}
}

// AddMember 添加服务器成员
func (r *ServerMemberRepository) AddMember(ctx context.Context, member *model.ServerMember) error {
	return r.db.WithContext(ctx).Create(member).Error
}

// AddMemberWithDB 使用指定 DB 添加成员（用于事务）
func (r *ServerMemberRepository) AddMemberWithDB(db *gorm.DB, member *model.ServerMember) error {
	return db.Create(member).Error
}

// RemoveMember 移除服务器成员
func (r *ServerMemberRepository) RemoveMember(ctx context.Context, serverID, userID uint64) error {
	return r.db.WithContext(ctx).
		Where("server_id = ? AND user_id = ?", serverID, userID).
		Delete(&model.ServerMember{}).Error
}

// FindByServerAndUser 查询用户在指定服务器的成员记录
func (r *ServerMemberRepository) FindByServerAndUser(ctx context.Context, serverID, userID uint64) (*model.ServerMember, error) {
	var member model.ServerMember
	err := r.db.WithContext(ctx).
		Where("server_id = ? AND user_id = ?", serverID, userID).
		First(&member).Error
	if err != nil {
		return nil, err
	}
	return &member, nil
}

// ListByServer 查询服务器的所有成员
func (r *ServerMemberRepository) ListByServer(ctx context.Context, serverID uint64) ([]model.ServerMember, error) {
	var members []model.ServerMember
	err := r.db.WithContext(ctx).
		Where("server_id = ?", serverID).
		Order("joined_at ASC").
		Find(&members).Error
	return members, err
}

// ListServersByUser 查询用户加入的所有服务器成员记录
func (r *ServerMemberRepository) ListServersByUser(ctx context.Context, userID uint64) ([]model.ServerMember, error) {
	var members []model.ServerMember
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("joined_at DESC").
		Find(&members).Error
	return members, err
}

// UpdateMember 更新成员记录
func (r *ServerMemberRepository) UpdateMember(ctx context.Context, member *model.ServerMember) error {
	return r.db.WithContext(ctx).Save(member).Error
}

// Exists 检查用户是否是服务器成员
func (r *ServerMemberRepository) Exists(ctx context.Context, serverID, userID uint64) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&model.ServerMember{}).
		Where("server_id = ? AND user_id = ?", serverID, userID).
		Count(&count).Error
	return count > 0, err
}

// CountByServer 统计服务器成员数量
func (r *ServerMemberRepository) CountByServer(ctx context.Context, serverID uint64) (int, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&model.ServerMember{}).
		Where("server_id = ?", serverID).
		Count(&count).Error
	return int(count), err
}
