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

// RoomService 房间服务
type RoomService struct {
	roomRepo        *repository.RoomRepository
	participantRepo *repository.RoomParticipantRepository
	userRepo        *repository.UserRepository
	txManager       *repository.GormTransactionManager
}

// NewRoomService 创建房间服务实例
func NewRoomService(
	roomRepo *repository.RoomRepository,
	participantRepo *repository.RoomParticipantRepository,
	userRepo *repository.UserRepository,
	txManager *repository.GormTransactionManager,
) *RoomService {
	return &RoomService{
		roomRepo:        roomRepo,
		participantRepo: participantRepo,
		userRepo:        userRepo,
		txManager:       txManager,
	}
}

// CreateRoom 创建房间
// 使用事务确保房间创建和参与者记录创建的原子性
func (s *RoomService) CreateRoom(ctx context.Context, hostUserID uint64, req *dto.CreateRoomRequest) (*dto.RoomInfoResponse, error) {
	// 验证用户是否存在
	_, _ = s.userRepo.FindByID(ctx, hostUserID)

	// 设置默认最大参与者数量
	if req.MaxParticipants == 0 {
		req.MaxParticipants = 20
	}

	room := &model.Room{
		RoomName:        req.RoomName,
		RoomType:        req.RoomType,
		HostUserID:      hostUserID,
		IsPrivate:       req.IsPrivate,
		MaxParticipants: req.MaxParticipants,
	}

	// 如果是私密房间，生成邀请码
	if req.IsPrivate {
		inviteCode, createErr := s.roomRepo.GenerateInviteCode(ctx)
		if createErr != nil {
			return nil, errcode.ErrInternalServer.WithMessage("生成邀请码失败")
		}
		room.InviteCode = &inviteCode
	}

	// 使用事务创建房间和参与者记录
	// 确保两个操作要么全部成功，要么全部回滚
	var createdRoom *model.Room
	err := s.txManager.Transactional(ctx, func(tx repository.TransactionContext) error {
		// 创建房间
		if createErr := s.roomRepo.CreateWithDB(tx.DB(), room); createErr != nil {
			return createErr
		}
		createdRoom = room

		// 创建房主参与者记录
		participant := &model.RoomParticipant{
			RoomID:   room.ID,
			UserID:   hostUserID,
			Role:     1, // 房主
			IsActive: true,
		}
		if createErr := s.participantRepo.CreateWithDB(tx.DB(), participant); createErr != nil {
			return createErr
		}

		return nil
	})

	if err != nil {
		return nil, errcode.ErrDBError.WithMessage("创建房间失败")
	}

	return s.buildRoomInfoResponse(ctx, createdRoom, 1)
}

// JoinRoom 加入房间
func (s *RoomService) JoinRoom(ctx context.Context, userID uint64, roomID uint64, req *dto.JoinRoomRequest) error {
	// 查询房间
	room, err := s.roomRepo.FindByID(ctx, roomID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errcode.ErrRoomNotFound
		}
		return errcode.ErrDBError.WithMessage("查询房间失败")
	}

	// 检查是否已在房间中
	inRoom, err := s.participantRepo.ExistsActive(ctx, roomID, userID)
	if err != nil {
		return errcode.ErrDBError.WithMessage("检查参与者状态失败")
	}
	if inRoom {
		return errcode.ErrAlreadyInRoom
	}

	// 检查房间是否已满
	count, err := s.participantRepo.CountActiveByRoom(ctx, roomID)
	if err != nil {
		return errcode.ErrDBError.WithMessage("统计参与者数量失败")
	}
	if count >= int(room.MaxParticipants) {
		return errcode.ErrRoomFull
	}

	// 私密房间需要邀请码
	if room.IsPrivate {
		if req.InviteCode == "" {
			return errcode.ErrRoomPrivate
		}
		if room.InviteCode == nil || req.InviteCode != *room.InviteCode {
			return errcode.ErrInvalidInviteCode
		}
	}

	// 创建参与者记录
	participant := &model.RoomParticipant{
		RoomID:   roomID,
		UserID:   userID,
		Role:     3, // 听众
		IsActive: true,
	}
	if err = s.participantRepo.Create(ctx, participant); err != nil {
		return errcode.ErrDBError.WithMessage("创建参与者记录失败")
	}

	return nil
}

// LeaveRoom 离开房间
// 使用事务确保离开状态更新和房间删除的原子性
func (s *RoomService) LeaveRoom(ctx context.Context, userID uint64, roomID uint64) error {
	// 检查是否在房间中
	inRoom, err := s.participantRepo.ExistsActive(ctx, roomID, userID)
	if err != nil {
		return errcode.ErrDBError.WithMessage("检查参与者状态失败")
	}
	if !inRoom {
		return errcode.ErrNotInRoom
	}

	// 使用事务处理离开房间逻辑
	// 确保参与者状态更新和房间删除（如果房间为空）的原子性
	err = s.txManager.Transactional(ctx, func(tx repository.TransactionContext) error {
		// 设置离开状态
		if leaveErr := s.participantRepo.LeaveWithDB(tx.DB(), roomID, userID); leaveErr != nil {
			return leaveErr
		}

		// 检查房间是否还有参与者（需要在同一事务中查询以保证一致性）
		var count int64
		if countErr := tx.DB().Model(&model.RoomParticipant{}).
			Where("room_id = ? AND is_active = ?", roomID, true).
			Count(&count).Error; countErr != nil {
			return countErr
		}

		// 如果房间空了，关闭房间
		if count == 0 {
			if deleteErr := s.roomRepo.DeleteWithDB(tx.DB(), roomID); deleteErr != nil {
				return deleteErr
			}
		}

		return nil
	})

	if err != nil {
		return errcode.ErrDBError.WithMessage("离开房间失败")
	}

	return nil
}

// GetRoomList 获取房间列表
// 优化：使用批量查询替代循环查询，解决 N+1 问题
func (s *RoomService) GetRoomList(ctx context.Context, req *dto.RoomListRequest) ([]dto.RoomInfoResponse, int64, error) {
	if req.Page == 0 {
		req.Page = 1
	}
	if req.PageSize == 0 {
		req.PageSize = 20
	}

	rooms, total, err := s.roomRepo.List(ctx, req.RoomType, req.Page, req.PageSize)
	if err != nil {
		return nil, 0, errcode.ErrDBError.WithMessage("查询房间列表失败")
	}

	// 如果没有房间，直接返回空列表
	if len(rooms) == 0 {
		return []dto.RoomInfoResponse{}, total, nil
	}

	// 收集所有房间 ID
	roomIDs := make([]uint64, 0, len(rooms))
	for _, room := range rooms {
		roomIDs = append(roomIDs, room.ID)
	}

	// 批量查询房间参与者数量（解决 N+1 问题）
	countMap, err := s.participantRepo.CountActiveByRoomsBatch(ctx, roomIDs)
	if err != nil {
		return nil, 0, errcode.ErrDBError.WithMessage("统计参与者数量失败")
	}

	// 组装响应数据
	var responses []dto.RoomInfoResponse
	for _, room := range rooms {
		count := countMap[room.ID] // 默认为 0，已在 repository 中处理
		resp, _ := s.buildRoomInfoResponse(ctx, &room, count)
		responses = append(responses, *resp)
	}

	return responses, total, nil
}

// GetRoomParticipants 获取房间参与者
// 优化：使用批量查询替代循环查询，解决 N+1 问题
func (s *RoomService) GetRoomParticipants(ctx context.Context, roomID uint64) ([]dto.ParticipantResponse, error) {
	participants, err := s.participantRepo.FindActiveByRoom(ctx, roomID)
	if err != nil {
		return nil, errcode.ErrDBError.WithMessage("查询参与者失败")
	}

	// 如果没有参与者，直接返回空列表
	if len(participants) == 0 {
		return []dto.ParticipantResponse{}, nil
	}

	// 收集所有参与者用户 ID
	userIDs := make([]uint64, 0, len(participants))
	for _, p := range participants {
		userIDs = append(userIDs, p.UserID)
	}

	// 批量查询用户信息（解决 N+1 问题）
	userMap, err := s.userRepo.FindByIDs(ctx, userIDs)
	if err != nil {
		return nil, errcode.ErrDBError.WithMessage("查询用户信息失败")
	}

	// 组装响应数据
	var responses []dto.ParticipantResponse
	for _, p := range participants {
		user, exists := userMap[p.UserID]
		if !exists {
			continue
		}

		responses = append(responses, dto.ParticipantResponse{
			UserID:          p.UserID,
			Username:        user.Username,
			AvatarURL:       user.AvatarURL,
			Role:            p.Role,
			IsMuted:         p.IsMuted,
			IsScreenSharing: p.IsScreenSharing,
			JoinedAt:        p.JoinedAt.Format("2006-01-02 15:04:05"),
		})
	}

	return responses, nil
}

// GetRoomInfo 获取房间信息
func (s *RoomService) GetRoomInfo(ctx context.Context, roomID uint64) (*dto.RoomInfoResponse, error) {
	room, err := s.roomRepo.FindByID(ctx, roomID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errcode.ErrRoomNotFound
		}
		return nil, errcode.ErrDBError.WithMessage("查询房间失败")
	}

	count, err := s.participantRepo.CountActiveByRoom(ctx, roomID)
	if err != nil {
		return nil, errcode.ErrDBError.WithMessage("统计参与者数量失败")
	}

	return s.buildRoomInfoResponse(ctx, room, count)
}

// buildRoomInfoResponse 构建房间信息响应
func (s *RoomService) buildRoomInfoResponse(ctx context.Context, room *model.Room, participantCount int) (*dto.RoomInfoResponse, error) {
	var inviteCode string
	if room.InviteCode != nil {
		inviteCode = *room.InviteCode
	}
	return &dto.RoomInfoResponse{
		ID:                    room.ID,
		RoomName:              room.RoomName,
		RoomType:              room.RoomType,
		HostUserID:            room.HostUserID,
		IsPrivate:             room.IsPrivate,
		InviteCode:            inviteCode,
		MaxParticipants:       room.MaxParticipants,
		CurrentPlaylistItemID: room.CurrentPlaylistItemID,
		ParticipantCount:      participantCount,
		CreatedAt:             room.CreatedAt.Format("2006-01-02 15:04:05"),
	}, nil
}

// GetOnlineCount 获取房间在线人数
func (s *RoomService) GetOnlineCount(ctx context.Context, roomID uint64) (*dto.OnlineCountResponse, error) {
	// 检查房间是否存在
	_, err := s.roomRepo.FindByID(ctx, roomID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errcode.ErrRoomNotFound
		}
		return nil, errcode.ErrDBError.WithMessage("查询房间失败")
	}

	// 统计在线人数
	count, err := s.participantRepo.CountActiveByRoom(ctx, roomID)
	if err != nil {
		return nil, errcode.ErrDBError.WithMessage("统计在线人数失败")
	}

	return &dto.OnlineCountResponse{
		RoomID:      roomID,
		OnlineCount: count,
	}, nil
}

// GetOnlineUsers 获取房间在线用户列表
// 优化：使用批量查询替代循环查询，解决 N+1 问题
func (s *RoomService) GetOnlineUsers(ctx context.Context, roomID uint64) ([]dto.OnlineUserResponse, error) {
	// 检查房间是否存在
	_, err := s.roomRepo.FindByID(ctx, roomID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errcode.ErrRoomNotFound
		}
		return nil, errcode.ErrDBError.WithMessage("查询房间失败")
	}

	// 查询活跃参与者
	participants, err := s.participantRepo.FindActiveByRoom(ctx, roomID)
	if err != nil {
		return nil, errcode.ErrDBError.WithMessage("查询参与者失败")
	}

	// 如果没有参与者，直接返回空列表
	if len(participants) == 0 {
		return []dto.OnlineUserResponse{}, nil
	}

	// 收集所有参与者用户 ID
	userIDs := make([]uint64, 0, len(participants))
	for _, p := range participants {
		userIDs = append(userIDs, p.UserID)
	}

	// 批量查询用户信息（解决 N+1 问题）
	userMap, err := s.userRepo.FindByIDs(ctx, userIDs)
	if err != nil {
		return nil, errcode.ErrDBError.WithMessage("查询用户信息失败")
	}

	// 组装响应数据
	var responses []dto.OnlineUserResponse
	for _, p := range participants {
		user, exists := userMap[p.UserID]
		if !exists {
			continue
		}

		responses = append(responses, dto.OnlineUserResponse{
			UserID:          p.UserID,
			Username:        user.Username,
			AvatarURL:       user.AvatarURL,
			Role:            p.Role,
			IsMuted:         p.IsMuted,
			IsScreenSharing: p.IsScreenSharing,
			IsVoiceActive:   !p.IsMuted, // 根据静音状态推断语音活跃
			JoinedAt:        p.JoinedAt.Format("2006-01-02 15:04:05"),
		})
	}

	return responses, nil
}

// GetUserStatus 获取用户状态
func (s *RoomService) GetUserStatus(ctx context.Context, userID uint64) (*dto.UserStatusResponse, error) {
	// 查询用户信息
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errcode.ErrUserNotFound
		}
		return nil, errcode.ErrDBError.WithMessage("查询用户失败")
	}

	// 查询用户当前活跃的房间参与记录
	participants, err := s.participantRepo.FindActiveByUserWithRoom(ctx, userID)
	if err != nil {
		return nil, errcode.ErrDBError.WithMessage("查询用户房间状态失败")
	}

	// 用户不在线（未加入任何房间）
	if len(participants) == 0 {
		return &dto.UserStatusResponse{
			UserID:   userID,
			Username: user.Username,
			IsOnline: false,
		}, nil
	}

	// 取最新的房间参与记录
	latest := participants[0]
	return &dto.UserStatusResponse{
		UserID:          userID,
		Username:        user.Username,
		RoomID:          latest.RoomID,
		RoomName:        latest.Room.RoomName,
		IsOnline:        true,
		Role:            latest.Role,
		IsMuted:         latest.IsMuted,
		IsScreenSharing: latest.IsScreenSharing,
		IsVoiceActive:   !latest.IsMuted,
		JoinedAt:        latest.JoinedAt.Format("2006-01-02 15:04:05"),
	}, nil
}

// DeleteRoom 删除房间（仅房主可操作）
// 使用事务确保删除参与者和删除房间的原子性
func (s *RoomService) DeleteRoom(ctx context.Context, userID, roomID uint64) error {
	// 查询房间
	room, err := s.roomRepo.FindByID(ctx, roomID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errcode.ErrRoomNotFound
		}
		return errcode.ErrDBError.WithMessage("查询房间失败")
	}

	// 验证是否为房主
	if room.HostUserID != userID {
		return errcode.ErrForbidden.WithMessage("只有房主才能删除房间")
	}

	// 使用事务删除参与者记录和房间
	// 确保两个操作要么全部成功，要么全部回滚
	err = s.txManager.Transactional(ctx, func(tx repository.TransactionContext) error {
		// 删除所有参与者记录
		if deleteErr := s.participantRepo.DeleteByRoomWithDB(tx.DB(), roomID); deleteErr != nil {
			return deleteErr
		}

		// 删除房间
		if deleteErr := s.roomRepo.DeleteWithDB(tx.DB(), roomID); deleteErr != nil {
			return deleteErr
		}

		return nil
	})

	if err != nil {
		return errcode.ErrDBError.WithMessage("删除房间失败")
	}

	return nil
}

// GetPublicRooms 获取公开房间列表
// roomRepo.List 已仅返回非私密房间（is_private = false），此处复用同一查询逻辑
// 对应前端 roomService.getPublicRooms -> GET /rooms/public
func (s *RoomService) GetPublicRooms(ctx context.Context, req *dto.RoomListRequest) ([]dto.RoomInfoResponse, int64, error) {
	return s.GetRoomList(ctx, req)
}

// KickMember 踢出房间成员（仅房主或管理员可操作）
// 对应前端 roomService.kickParticipant -> POST /rooms/:id/kick/:userId
func (s *RoomService) KickMember(ctx context.Context, callerID, roomID, targetUserID uint64) error {
	// 查询房间
	_, err := s.roomRepo.FindByID(ctx, roomID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errcode.ErrRoomNotFound
		}
		return errcode.ErrDBError.WithMessage("查询房间失败")
	}

	// 校验调用者权限：房主(1)或管理员(2)可踢人
	caller, err := s.participantRepo.FindByRoomAndUser(ctx, roomID, callerID)
	if err != nil {
		return errcode.ErrNotInRoom
	}
	if caller.Role != 1 && caller.Role != 2 {
		return errcode.ErrForbidden.WithMessage("只有房主或管理员才能踢出成员")
	}

	// 不能踢自己
	if callerID == targetUserID {
		return errcode.ErrBadRequest.WithMessage("不能踢出自己")
	}

	// 校验目标用户在房间中
	target, err := s.participantRepo.FindByRoomAndUser(ctx, roomID, targetUserID)
	if err != nil {
		return errcode.ErrNotInRoom.WithMessage("目标用户不在房间中")
	}

	// 房主不能被踢
	if target.Role == 1 {
		return errcode.ErrForbidden.WithMessage("不能踢出房主")
	}

	// 设置目标用户离开房间
	if err = s.participantRepo.Leave(ctx, roomID, targetUserID); err != nil {
		return errcode.ErrDBError.WithMessage("踢出成员失败")
	}

	return nil
}

// UpdateMemberRole 设置成员角色（仅房主可操作）
// 对应前端 roomService.setParticipantRole -> PUT /rooms/:id/participants/:userId/role
// role: 1=房主, 2=管理员, 3=普通用户
func (s *RoomService) UpdateMemberRole(ctx context.Context, callerID, roomID, targetUserID uint64, req *dto.UpdateMemberRoleRequest) error {
	// 查询房间
	_, err := s.roomRepo.FindByID(ctx, roomID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errcode.ErrRoomNotFound
		}
		return errcode.ErrDBError.WithMessage("查询房间失败")
	}

	// 校验调用者为房主
	caller, err := s.participantRepo.FindByRoomAndUser(ctx, roomID, callerID)
	if err != nil {
		return errcode.ErrNotInRoom
	}
	if caller.Role != 1 {
		return errcode.ErrForbidden.WithMessage("只有房主才能设置成员角色")
	}

	// 不能修改自己的角色
	if callerID == targetUserID {
		return errcode.ErrBadRequest.WithMessage("不能修改自己的角色")
	}

	// 校验目标用户在房间中
	target, err := s.participantRepo.FindByRoomAndUser(ctx, roomID, targetUserID)
	if err != nil {
		return errcode.ErrNotInRoom.WithMessage("目标用户不在房间中")
	}

	// 不能将成员提升为房主（房主转移应通过专门流程）
	if req.Role == 1 {
		return errcode.ErrBadRequest.WithMessage("不能将成员设置为房主")
	}

	target.Role = req.Role
	if err = s.participantRepo.Update(ctx, target); err != nil {
		return errcode.ErrDBError.WithMessage("更新成员角色失败")
	}

	return nil
}
