package repository

import (
	"context"

	"github.com/yourorg/livemix/internal/model"
	"gorm.io/gorm"
)

// RoleRepository 角色仓储
type RoleRepository struct {
	db *gorm.DB
}

// NewRoleRepository 创建角色仓储实例
func NewRoleRepository(db *gorm.DB) *RoleRepository {
	return &RoleRepository{db: db}
}

// Create 创建角色
func (r *RoleRepository) Create(ctx context.Context, role *model.Role) error {
	return r.db.WithContext(ctx).Create(role).Error
}

// CreateWithDB 使用指定 DB 创建角色（用于事务）
func (r *RoleRepository) CreateWithDB(db *gorm.DB, role *model.Role) error {
	return db.Create(role).Error
}

// FindByID 根据ID查询角色
func (r *RoleRepository) FindByID(ctx context.Context, id uint64) (*model.Role, error) {
	var role model.Role
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&role).Error
	if err != nil {
		return nil, err
	}
	return &role, nil
}

// FindByServer 查询服务器的所有角色
func (r *RoleRepository) FindByServer(ctx context.Context, serverID uint64) ([]model.Role, error) {
	var roles []model.Role
	err := r.db.WithContext(ctx).
		Where("server_id = ?", serverID).
		Find(&roles).Error
	return roles, err
}

// Update 更新角色
func (r *RoleRepository) Update(ctx context.Context, role *model.Role) error {
	return r.db.WithContext(ctx).Save(role).Error
}

// Delete 删除角色
func (r *RoleRepository) Delete(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Delete(&model.Role{}, id).Error
}

// ListByServerOrdered 按排序权重查询服务器角色（从大到小）
func (r *RoleRepository) ListByServerOrdered(ctx context.Context, serverID uint64) ([]model.Role, error) {
	var roles []model.Role
	err := r.db.WithContext(ctx).
		Where("server_id = ?", serverID).
		Order("position DESC").
		Find(&roles).Error
	return roles, err
}

// AssignRoleToMember 为成员分配角色
func (r *RoleRepository) AssignRoleToMember(ctx context.Context, serverID, memberID, roleID uint64) error {
	mr := &model.ServerMemberRole{
		ServerID: serverID,
		MemberID: memberID,
		RoleID:   roleID,
	}
	return r.db.WithContext(ctx).Create(mr).Error
}

// RemoveRoleFromMember 移除成员的角色
func (r *RoleRepository) RemoveRoleFromMember(ctx context.Context, serverID, memberID, roleID uint64) error {
	return r.db.WithContext(ctx).
		Where("server_id = ? AND member_id = ? AND role_id = ?", serverID, memberID, roleID).
		Delete(&model.ServerMemberRole{}).Error
}

// FindRolesByMember 查询成员的所有角色
func (r *RoleRepository) FindRolesByMember(ctx context.Context, memberID uint64) ([]model.Role, error) {
	var roles []model.Role
	err := r.db.WithContext(ctx).
		Joins("JOIN server_member_roles ON server_member_roles.role_id = roles.id").
		Where("server_member_roles.member_id = ?", memberID).
		Order("roles.position DESC").
		Find(&roles).Error
	return roles, err
}

// SetMemberRoles 设置成员的角色列表（覆盖）
// 先删除旧的角色关联，再创建新的
func (r *RoleRepository) SetMemberRoles(ctx context.Context, serverID, memberID uint64, roleIDs []uint64) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 删除旧的角色关联
		if err := tx.Where("server_id = ? AND member_id = ?", serverID, memberID).
			Delete(&model.ServerMemberRole{}).Error; err != nil {
			return err
		}

		// 创建新的角色关联
		for _, roleID := range roleIDs {
			mr := &model.ServerMemberRole{
				ServerID: serverID,
				MemberID: memberID,
				RoleID:   roleID,
			}
			if err := tx.Create(mr).Error; err != nil {
				return err
			}
		}

		return nil
	})
}

// FindMemberRoleIDs 查询成员的角色ID列表
func (r *RoleRepository) FindMemberRoleIDs(ctx context.Context, memberID uint64) ([]uint64, error) {
	var roleIDs []uint64
	err := r.db.WithContext(ctx).
		Model(&model.ServerMemberRole{}).
		Where("member_id = ?", memberID).
		Pluck("role_id", &roleIDs).Error
	return roleIDs, err
}
