package service

import (
	"context"
	"errors"

	"github.com/yourorg/livemix/internal/dto"
	"github.com/yourorg/livemix/internal/model"
	"github.com/yourorg/livemix/internal/repository"
	"github.com/yourorg/livemix/pkg/errcode"
	"github.com/yourorg/livemix/pkg/util"
	"gorm.io/gorm"
)

// ServerService 服务器服务
type ServerService struct {
	serverRepo  *repository.ServerRepository
	memberRepo  *repository.ServerMemberRepository
	roleRepo    *repository.RoleRepository
	channelRepo *repository.ChannelRepository
	userRepo    *repository.UserRepository
	txManager   *repository.GormTransactionManager
}

// NewServerService 创建服务器服务实例
func NewServerService(
	serverRepo *repository.ServerRepository,
	memberRepo *repository.ServerMemberRepository,
	roleRepo *repository.RoleRepository,
	channelRepo *repository.ChannelRepository,
	userRepo *repository.UserRepository,
	txManager *repository.GormTransactionManager,
) *ServerService {
	return &ServerService{
		serverRepo:  serverRepo,
		memberRepo:  memberRepo,
		roleRepo:    roleRepo,
		channelRepo: channelRepo,
		userRepo:    userRepo,
		txManager:   txManager,
	}
}

// CreateServer 创建服务器
// 自动生成3个默认角色(owner/admin/member)和默认频道(常规文字频道+general语音频道)
func (s *ServerService) CreateServer(ctx context.Context, ownerID uint64, req *dto.CreateServerRequest) (*dto.ServerDetailResponse, error) {
	// 验证用户是否存在
	_, err := s.userRepo.FindByID(ctx, ownerID)
	if err != nil {
		return nil, errcode.ErrUserNotFound
	}

	// 生成邀请码
	inviteCode, err := s.serverRepo.GenerateInviteCode(ctx)
	if err != nil {
		return nil, errcode.ErrInternalServer.WithMessage("生成邀请码失败")
	}

	server := &model.Server{
		Name:        util.TrimAndEscape(req.Name),
		IconURL:     req.IconURL,
		BannerURL:   req.BannerURL,
		Description: util.TrimAndEscape(req.Description),
		OwnerID:     ownerID,
		InviteCode:  inviteCode,
		IsPrivate:   req.IsPrivate,
		MaxMembers:  500,
	}

	// 使用事务创建服务器、成员记录、默认角色和默认频道
	var createdServer *model.Server
	err = s.txManager.Transactional(ctx, func(tx repository.TransactionContext) error {
		// 创建服务器
		if createErr := s.serverRepo.CreateWithDB(tx.DB(), server); createErr != nil {
			return createErr
		}
		createdServer = server

		// 创建 owner 成员记录
		member := &model.ServerMember{
			ServerID: server.ID,
			UserID:   ownerID,
		}
		if createErr := s.memberRepo.AddMemberWithDB(tx.DB(), member); createErr != nil {
			return createErr
		}

		// 创建3个默认角色
		ownerRole := &model.Role{
			ServerID:      server.ID,
			Name:          "@owner",
			Color:         "#f04747",
			Position:      100,
			Permissions:   model.DefaultOwnerPermissions,
			IsMentionable: false,
		}
		if createErr := s.roleRepo.CreateWithDB(tx.DB(), ownerRole); createErr != nil {
			return createErr
		}

		adminRole := &model.Role{
			ServerID:      server.ID,
			Name:          "@admin",
			Color:         "#faa61a",
			Position:      50,
			Permissions:   model.DefaultAdminPermissions,
			IsMentionable: true,
		}
		if createErr := s.roleRepo.CreateWithDB(tx.DB(), adminRole); createErr != nil {
			return createErr
		}

		memberRole := &model.Role{
			ServerID:      server.ID,
			Name:          "@member",
			Color:         "#99aab5",
			Position:      0,
			Permissions:   model.DefaultMemberPermissions,
			IsMentionable: true,
		}
		if createErr := s.roleRepo.CreateWithDB(tx.DB(), memberRole); createErr != nil {
			return createErr
		}

		// 为 owner 分配 owner 角色
		if assignErr := tx.DB().Create(&model.ServerMemberRole{
			ServerID: server.ID,
			MemberID: member.ID,
			RoleID:   ownerRole.ID,
		}).Error; assignErr != nil {
			return assignErr
		}

		// 创建默认频道
		// 文字频道 "常规"
		textChannel := &model.Channel{
			ServerID: server.ID,
			Name:     "常规",
			Type:     model.ChannelTypeText,
			Position: 0,
		}
		if createErr := s.channelRepo.CreateWithDB(tx.DB(), textChannel); createErr != nil {
			return createErr
		}

		// 语音频道 "general"
		voiceChannel := &model.Channel{
			ServerID: server.ID,
			Name:     "general",
			Type:     model.ChannelTypeVoice,
			Position: 1,
		}
		if createErr := s.channelRepo.CreateWithDB(tx.DB(), voiceChannel); createErr != nil {
			return createErr
		}

		return nil
	})

	if err != nil {
		return nil, errcode.ErrDBError.WithMessage("创建服务器失败")
	}

	return s.buildServerDetailResponse(ctx, createdServer)
}

// GetServerList 获取用户的服务器列表
func (s *ServerService) GetServerList(ctx context.Context, userID uint64) ([]dto.ServerInfoResponse, error) {
	servers, err := s.serverRepo.ListByUserID(ctx, userID)
	if err != nil {
		return nil, errcode.ErrDBError.WithMessage("查询服务器列表失败")
	}

	if len(servers) == 0 {
		return []dto.ServerInfoResponse{}, nil
	}

	var responses []dto.ServerInfoResponse
	for _, srv := range servers {
		count, _ := s.memberRepo.CountByServer(ctx, srv.ID)
		responses = append(responses, dto.ServerInfoResponse{
			ID:          srv.ID,
			Name:        srv.Name,
			IconURL:     srv.IconURL,
			BannerURL:   srv.BannerURL,
			Description: srv.Description,
			OwnerID:     srv.OwnerID,
			InviteCode:  srv.InviteCode,
			IsPrivate:   srv.IsPrivate,
			MaxMembers:  srv.MaxMembers,
			MemberCount: count,
			CreatedAt:   srv.CreatedAt.Format("2006-01-02 15:04:05"),
		})
	}

	return responses, nil
}

// GetServerDetail 获取服务器详情（含频道列表）
func (s *ServerService) GetServerDetail(ctx context.Context, serverID uint64) (*dto.ServerDetailResponse, error) {
	server, err := s.serverRepo.FindByID(ctx, serverID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errcode.ErrServerNotFound
		}
		return nil, errcode.ErrDBError.WithMessage("查询服务器失败")
	}

	return s.buildServerDetailResponse(ctx, server)
}

// UpdateServer 更新服务器
func (s *ServerService) UpdateServer(ctx context.Context, userID, serverID uint64, req *dto.UpdateServerRequest) (*dto.ServerInfoResponse, error) {
	server, err := s.serverRepo.FindByID(ctx, serverID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errcode.ErrServerNotFound
		}
		return nil, errcode.ErrDBError.WithMessage("查询服务器失败")
	}

	// 校验权限：owner 或拥有管理服务器权限
	if err := s.CheckPermission(ctx, serverID, userID, model.PermManageServer); err != nil {
		return nil, err
	}

	if req.Name != "" {
		server.Name = util.TrimAndEscape(req.Name)
	}
	if req.IconURL != "" {
		server.IconURL = req.IconURL
	}
	if req.BannerURL != "" {
		server.BannerURL = req.BannerURL
	}
	if req.Description != "" {
		server.Description = util.TrimAndEscape(req.Description)
	}

	if err = s.serverRepo.Update(ctx, server); err != nil {
		return nil, errcode.ErrDBError.WithMessage("更新服务器失败")
	}

	count, _ := s.memberRepo.CountByServer(ctx, serverID)
	return &dto.ServerInfoResponse{
		ID:          server.ID,
		Name:        server.Name,
		IconURL:     server.IconURL,
		BannerURL:   server.BannerURL,
		Description: server.Description,
		OwnerID:     server.OwnerID,
		InviteCode:  server.InviteCode,
		IsPrivate:   server.IsPrivate,
		MaxMembers:  server.MaxMembers,
		MemberCount: count,
		CreatedAt:   server.CreatedAt.Format("2006-01-02 15:04:05"),
	}, nil
}

// DeleteServer 删除服务器（仅 owner 可操作）
func (s *ServerService) DeleteServer(ctx context.Context, userID, serverID uint64) error {
	server, err := s.serverRepo.FindByID(ctx, serverID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errcode.ErrServerNotFound
		}
		return errcode.ErrDBError.WithMessage("查询服务器失败")
	}

	if server.OwnerID != userID {
		return errcode.ErrNotServerOwner
	}

	if err = s.serverRepo.Delete(ctx, serverID); err != nil {
		return errcode.ErrDBError.WithMessage("删除服务器失败")
	}

	return nil
}

// JoinServer 加入服务器（通过邀请码）
func (s *ServerService) JoinServer(ctx context.Context, userID, serverID uint64, req *dto.JoinServerRequest) error {
	server, err := s.serverRepo.FindByID(ctx, serverID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errcode.ErrServerNotFound
		}
		return errcode.ErrDBError.WithMessage("查询服务器失败")
	}

	// 验证邀请码
	if server.InviteCode != req.InviteCode {
		return errcode.ErrInvalidServerInviteCode
	}

	// 检查是否已是成员
	isMember, err := s.memberRepo.Exists(ctx, serverID, userID)
	if err != nil {
		return errcode.ErrDBError.WithMessage("检查成员状态失败")
	}
	if isMember {
		return errcode.ErrAlreadyInServer
	}

	// 检查人数上限
	count, err := s.memberRepo.CountByServer(ctx, serverID)
	if err != nil {
		return errcode.ErrDBError.WithMessage("统计成员数量失败")
	}
	if count >= server.MaxMembers {
		return errcode.ErrServerFull
	}

	// 添加成员
	member := &model.ServerMember{
		ServerID: serverID,
		UserID:   userID,
	}
	if err = s.memberRepo.AddMember(ctx, member); err != nil {
		return errcode.ErrDBError.WithMessage("加入服务器失败")
	}

	// 自动分配 @member 默认角色
	roles, _ := s.roleRepo.FindByServer(ctx, serverID)
	for _, role := range roles {
		if role.Name == "@member" {
			_ = s.roleRepo.AssignRoleToMember(ctx, serverID, member.ID, role.ID)
			break
		}
	}

	return nil
}

// LeaveServer 离开服务器
func (s *ServerService) LeaveServer(ctx context.Context, userID, serverID uint64) error {
	// 检查是否是成员
	isMember, err := s.memberRepo.Exists(ctx, serverID, userID)
	if err != nil {
		return errcode.ErrDBError.WithMessage("检查成员状态失败")
	}
	if !isMember {
		return errcode.ErrNotInServer
	}

	// 检查是否是 owner
	server, err := s.serverRepo.FindByID(ctx, serverID)
	if err != nil {
		return errcode.ErrDBError.WithMessage("查询服务器失败")
	}
	if server.OwnerID == userID {
		return errcode.ErrBadRequest.WithMessage("服务器拥有者不能离开，请先转让或删除服务器")
	}

	if err = s.memberRepo.RemoveMember(ctx, serverID, userID); err != nil {
		return errcode.ErrDBError.WithMessage("离开服务器失败")
	}

	return nil
}

// KickMember 踢出服务器成员
func (s *ServerService) KickMember(ctx context.Context, callerID, serverID, targetUserID uint64) error {
	// 校验权限
	if err := s.CheckPermission(ctx, serverID, callerID, model.PermKickMembers); err != nil {
		return err
	}

	// 不能踢自己
	if callerID == targetUserID {
		return errcode.ErrBadRequest.WithMessage("不能踢出自己")
	}

	// 检查目标是否是成员
	isMember, err := s.memberRepo.Exists(ctx, serverID, targetUserID)
	if err != nil {
		return errcode.ErrDBError.WithMessage("检查成员状态失败")
	}
	if !isMember {
		return errcode.ErrServerMemberNotFound
	}

	// 不能踢 owner
	server, _ := s.serverRepo.FindByID(ctx, serverID)
	if server.OwnerID == targetUserID {
		return errcode.ErrForbidden.WithMessage("不能踢出服务器拥有者")
	}

	if err = s.memberRepo.RemoveMember(ctx, serverID, targetUserID); err != nil {
		return errcode.ErrDBError.WithMessage("踢出成员失败")
	}

	return nil
}

// GetMembers 获取服务器成员列表
func (s *ServerService) GetMembers(ctx context.Context, serverID uint64) ([]dto.ServerMemberResponse, error) {
	members, err := s.memberRepo.ListByServer(ctx, serverID)
	if err != nil {
		return nil, errcode.ErrDBError.WithMessage("查询成员列表失败")
	}

	if len(members) == 0 {
		return []dto.ServerMemberResponse{}, nil
	}

	// 收集用户 ID
	userIDs := make([]uint64, 0, len(members))
	for _, m := range members {
		userIDs = append(userIDs, m.UserID)
	}

	// 批量查询用户信息
	userMap, err := s.userRepo.FindByIDs(ctx, userIDs)
	if err != nil {
		return nil, errcode.ErrDBError.WithMessage("查询用户信息失败")
	}

	var responses []dto.ServerMemberResponse
	for _, m := range members {
		user, exists := userMap[m.UserID]
		if !exists {
			continue
		}

		// 查询成员角色
		roles, _ := s.roleRepo.FindRolesByMember(ctx, m.ID)
		var roleResponses []dto.RoleResponse
		for _, role := range roles {
			roleResponses = append(roleResponses, dto.RoleResponse{
				ID:            role.ID,
				ServerID:      role.ServerID,
				Name:          role.Name,
				Color:         role.Color,
				Position:      role.Position,
				Permissions:   role.Permissions,
				IsMentionable: role.IsMentionable,
				CreatedAt:     role.CreatedAt.Format("2006-01-02 15:04:05"),
			})
		}

		responses = append(responses, dto.ServerMemberResponse{
			ID:        m.ID,
			ServerID:  m.ServerID,
			UserID:    m.UserID,
			Username:  user.Username,
			AvatarURL: user.AvatarURL,
			Nickname:  m.Nickname,
			JoinedAt:  m.JoinedAt.Format("2006-01-02 15:04:05"),
			Roles:     roleResponses,
		})
	}

	return responses, nil
}

// GetMember 获取单个成员详情
func (s *ServerService) GetMember(ctx context.Context, serverID, userID uint64) (*dto.ServerMemberResponse, error) {
	member, err := s.memberRepo.FindByServerAndUser(ctx, serverID, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errcode.ErrServerMemberNotFound
		}
		return nil, errcode.ErrDBError.WithMessage("查询成员失败")
	}

	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, errcode.ErrDBError.WithMessage("查询用户失败")
	}

	roles, _ := s.roleRepo.FindRolesByMember(ctx, member.ID)
	var roleResponses []dto.RoleResponse
	for _, role := range roles {
		roleResponses = append(roleResponses, dto.RoleResponse{
			ID:            role.ID,
			ServerID:      role.ServerID,
			Name:          role.Name,
			Color:         role.Color,
			Position:      role.Position,
			Permissions:   role.Permissions,
			IsMentionable: role.IsMentionable,
			CreatedAt:     role.CreatedAt.Format("2006-01-02 15:04:05"),
		})
	}

	return &dto.ServerMemberResponse{
		ID:        member.ID,
		ServerID:  member.ServerID,
		UserID:    member.UserID,
		Username:  user.Username,
		AvatarURL: user.AvatarURL,
		Nickname:  member.Nickname,
		JoinedAt:  member.JoinedAt.Format("2006-01-02 15:04:05"),
		Roles:     roleResponses,
	}, nil
}

// UpdateMember 更新成员昵称和角色
func (s *ServerService) UpdateMember(ctx context.Context, callerID, serverID, targetUserID uint64, req *dto.UpdateServerMemberRequest) error {
	member, err := s.memberRepo.FindByServerAndUser(ctx, serverID, targetUserID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errcode.ErrServerMemberNotFound
		}
		return errcode.ErrDBError.WithMessage("查询成员失败")
	}

	// 修改昵称：自己可以修改自己，或拥有 ManageNicknames 权限
	if req.Nickname != "" {
		if callerID == targetUserID {
			// 自己修改自己，检查 ChangeNickname 权限
			if err := s.CheckPermission(ctx, serverID, callerID, model.PermChangeNickname); err != nil {
				return err
			}
		} else {
			// 修改他人，检查 ManageNicknames 权限
			if err := s.CheckPermission(ctx, serverID, callerID, model.PermManageNicknames); err != nil {
				return err
			}
		}
		member.Nickname = req.Nickname
		if err = s.memberRepo.UpdateMember(ctx, member); err != nil {
			return errcode.ErrDBError.WithMessage("更新成员失败")
		}
	}

	// 修改角色：需要 ManageRoles 权限
	if req.RoleIDs != nil {
		if err := s.CheckPermission(ctx, serverID, callerID, model.PermManageRoles); err != nil {
			return err
		}
		if err = s.roleRepo.SetMemberRoles(ctx, serverID, member.ID, req.RoleIDs); err != nil {
			return errcode.ErrDBError.WithMessage("更新成员角色失败")
		}
	}

	return nil
}

// CheckPermission 校验用户在服务器中的权限
// owner 用户恒为全权限
func (s *ServerService) CheckPermission(ctx context.Context, serverID, userID uint64, perm int64) error {
	server, err := s.serverRepo.FindByID(ctx, serverID)
	if err != nil {
		return errcode.ErrServerNotFound
	}

	// owner 恒为全权限
	if server.OwnerID == userID {
		return nil
	}

	// 检查是否是成员
	member, err := s.memberRepo.FindByServerAndUser(ctx, serverID, userID)
	if err != nil {
		return errcode.ErrNotInServer
	}

	// 查询成员的所有角色
	roles, err := s.roleRepo.FindRolesByMember(ctx, member.ID)
	if err != nil {
		return errcode.ErrDBError.WithMessage("查询角色失败")
	}

	// 合并权限
	var rolePerms []int64
	for _, role := range roles {
		rolePerms = append(rolePerms, role.Permissions)
	}
	mergedPerms := model.MergePermissions(rolePerms)

	if !model.HasPermission(mergedPerms, perm) {
		return errcode.ErrNoPermission
	}

	return nil
}

// ComputeMemberPermissions 计算成员的最终权限
func (s *ServerService) ComputeMemberPermissions(ctx context.Context, serverID, userID uint64) (int64, error) {
	server, err := s.serverRepo.FindByID(ctx, serverID)
	if err != nil {
		return 0, errcode.ErrServerNotFound
	}

	// owner 恒为全权限
	if server.OwnerID == userID {
		return model.DefaultOwnerPermissions, nil
	}

	member, err := s.memberRepo.FindByServerAndUser(ctx, serverID, userID)
	if err != nil {
		return 0, errcode.ErrNotInServer
	}

	roles, err := s.roleRepo.FindRolesByMember(ctx, member.ID)
	if err != nil {
		return 0, errcode.ErrDBError.WithMessage("查询角色失败")
	}

	var rolePerms []int64
	for _, role := range roles {
		rolePerms = append(rolePerms, role.Permissions)
	}
	return model.MergePermissions(rolePerms), nil
}

// buildServerDetailResponse 构建服务器详情响应
func (s *ServerService) buildServerDetailResponse(ctx context.Context, server *model.Server) (*dto.ServerDetailResponse, error) {
	count, _ := s.memberRepo.CountByServer(ctx, server.ID)
	channels, err := s.channelRepo.FindByServer(ctx, server.ID)
	if err != nil {
		return nil, errcode.ErrDBError.WithMessage("查询频道失败")
	}

	var channelResponses []dto.ChannelResponse
	for _, ch := range channels {
		channelResponses = append(channelResponses, dto.ChannelResponse{
			ID:        ch.ID,
			ServerID:  ch.ServerID,
			Name:      ch.Name,
			Type:      ch.Type,
			Topic:     ch.Topic,
			ParentID:  ch.ParentID,
			Position:  ch.Position,
			Bitrate:   ch.Bitrate,
			UserLimit: ch.UserLimit,
			SlowMode:  ch.SlowMode,
		})
	}

	return &dto.ServerDetailResponse{
		ServerInfoResponse: dto.ServerInfoResponse{
			ID:          server.ID,
			Name:        server.Name,
			IconURL:     server.IconURL,
			BannerURL:   server.BannerURL,
			Description: server.Description,
			OwnerID:     server.OwnerID,
			InviteCode:  server.InviteCode,
			IsPrivate:   server.IsPrivate,
			MaxMembers:  server.MaxMembers,
			MemberCount: count,
			CreatedAt:   server.CreatedAt.Format("2006-01-02 15:04:05"),
		},
		Channels: channelResponses,
	}, nil
}
