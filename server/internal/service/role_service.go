package service

import (
	"context"
	"errors"

	"github.com/yourorg/livemix/internal/dto"
	"github.com/yourorg/livemix/internal/model"
	"github.com/yourorg/livemix/internal/repository"
	"github.com/yourorg/livemix/pkg/errcode"
	"gorm.io/gorm"
)

// RoleService 角色服务
type RoleService struct {
	roleRepo   *repository.RoleRepository
	serverRepo *repository.ServerRepository
	serverSvc  *ServerService
}

// NewRoleService 创建角色服务实例
func NewRoleService(
	roleRepo *repository.RoleRepository,
	serverRepo *repository.ServerRepository,
	serverSvc *ServerService,
) *RoleService {
	return &RoleService{
		roleRepo:   roleRepo,
		serverRepo: serverRepo,
		serverSvc:  serverSvc,
	}
}

// CreateRole 创建角色
func (s *RoleService) CreateRole(ctx context.Context, serverID, userID uint64, req *dto.CreateRoleRequest) (*dto.RoleResponse, error) {
	// 校验权限
	if err := s.serverSvc.CheckPermission(ctx, serverID, userID, model.PermManageRoles); err != nil {
		return nil, err
	}

	color := req.Color
	if color == "" {
		color = "#99aab5"
	}
	isMentionable := true
	if req.IsMentionable != nil {
		isMentionable = *req.IsMentionable
	}

	// 获取当前最大 position
	roles, _ := s.roleRepo.FindByServer(ctx, serverID)
	maxPosition := 0
	for _, r := range roles {
		if r.Name != "@owner" && r.Position > maxPosition && r.Position < 100 {
			maxPosition = r.Position
		}
	}

	role := &model.Role{
		ServerID:      serverID,
		Name:          req.Name,
		Color:         color,
		Position:      maxPosition + 1,
		Permissions:   req.Permissions,
		IsMentionable: isMentionable,
	}

	if err := s.roleRepo.Create(ctx, role); err != nil {
		return nil, errcode.ErrDBError.WithMessage("创建角色失败")
	}

	return s.buildRoleResponse(role), nil
}

// GetRoles 获取服务器角色列表
func (s *RoleService) GetRoles(ctx context.Context, serverID, userID uint64) ([]dto.RoleResponse, error) {
	// 检查用户是否是成员
	isMember, err := s.serverSvc.memberRepo.Exists(ctx, serverID, userID)
	if err != nil {
		return nil, errcode.ErrDBError.WithMessage("检查成员状态失败")
	}
	if !isMember {
		return nil, errcode.ErrNotInServer
	}

	roles, err := s.roleRepo.ListByServerOrdered(ctx, serverID)
	if err != nil {
		return nil, errcode.ErrDBError.WithMessage("查询角色失败")
	}

	var responses []dto.RoleResponse
	for _, role := range roles {
		responses = append(responses, *s.buildRoleResponse(&role))
	}

	return responses, nil
}

// UpdateRole 更新角色
func (s *RoleService) UpdateRole(ctx context.Context, serverID, roleID, userID uint64, req *dto.UpdateRoleRequest) (*dto.RoleResponse, error) {
	// 校验权限
	if err := s.serverSvc.CheckPermission(ctx, serverID, userID, model.PermManageRoles); err != nil {
		return nil, err
	}

	role, err := s.roleRepo.FindByID(ctx, roleID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errcode.ErrRoleNotFound
		}
		return nil, errcode.ErrDBError.WithMessage("查询角色失败")
	}

	if role.ServerID != serverID {
		return nil, errcode.ErrRoleNotFound
	}

	// 不能修改 @owner 角色的权限
	if role.Name == "@owner" && req.Permissions != nil {
		return nil, errcode.ErrCannotDeleteDefaultRole.WithMessage("不能修改 owner 角色的权限")
	}

	if req.Name != "" {
		role.Name = req.Name
	}
	if req.Color != "" {
		role.Color = req.Color
	}
	if req.Permissions != nil {
		role.Permissions = *req.Permissions
	}
	if req.Position != nil {
		role.Position = *req.Position
	}
	if req.IsMentionable != nil {
		role.IsMentionable = *req.IsMentionable
	}

	if err = s.roleRepo.Update(ctx, role); err != nil {
		return nil, errcode.ErrDBError.WithMessage("更新角色失败")
	}

	return s.buildRoleResponse(role), nil
}

// DeleteRole 删除角色
func (s *RoleService) DeleteRole(ctx context.Context, serverID, roleID, userID uint64) error {
	// 校验权限
	if err := s.serverSvc.CheckPermission(ctx, serverID, userID, model.PermManageRoles); err != nil {
		return err
	}

	role, err := s.roleRepo.FindByID(ctx, roleID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errcode.ErrRoleNotFound
		}
		return errcode.ErrDBError.WithMessage("查询角色失败")
	}

	if role.ServerID != serverID {
		return errcode.ErrRoleNotFound
	}

	// 不能删除默认角色
	if role.Name == "@owner" || role.Name == "@admin" || role.Name == "@member" {
		return errcode.ErrCannotDeleteDefaultRole
	}

	if err = s.roleRepo.Delete(ctx, roleID); err != nil {
		return errcode.ErrDBError.WithMessage("删除角色失败")
	}

	return nil
}

// AssignRoleToMember 为成员分配角色
func (s *RoleService) AssignRoleToMember(ctx context.Context, serverID, memberID, roleID, callerID uint64) error {
	// 校验权限
	if err := s.serverSvc.CheckPermission(ctx, serverID, callerID, model.PermManageRoles); err != nil {
		return err
	}

	return s.roleRepo.AssignRoleToMember(ctx, serverID, memberID, roleID)
}

// buildRoleResponse 构建角色响应
func (s *RoleService) buildRoleResponse(role *model.Role) *dto.RoleResponse {
	return &dto.RoleResponse{
		ID:            role.ID,
		ServerID:      role.ServerID,
		Name:          role.Name,
		Color:         role.Color,
		Position:      role.Position,
		Permissions:   role.Permissions,
		IsMentionable: role.IsMentionable,
		CreatedAt:     role.CreatedAt.Format("2006-01-02 15:04:05"),
	}
}
