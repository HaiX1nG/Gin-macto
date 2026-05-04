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
}

// NewRoomService 创建房间服务实例
func NewRoomService(roomRepo *repository.RoomRepository, participantRepo *repository.RoomParticipantRepository, userRepo *repository.UserRepository) *RoomService {
	return &RoomService{
		roomRepo:        roomRepo,
		participantRepo: participantRepo,
		userRepo:        userRepo,
	}
}

// CreateRoom 创建房间
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

	if createErr := s.roomRepo.Create(ctx, room); createErr != nil {
		return nil, errcode.ErrDBError.WithMessage("创建房间失败")
	}

	// 创建房主参与者记录
	participant := &model.RoomParticipant{
		RoomID:   room.ID,
		UserID:   hostUserID,
		Role:     1, // 房主
		IsActive: true,
	}
	if createErr := s.participantRepo.Create(ctx, participant); createErr != nil {
		return nil, errcode.ErrDBError.WithMessage("创建参与者记录失败")
	}

	return s.buildRoomInfoResponse(ctx, room, 1)
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
func (s *RoomService) LeaveRoom(ctx context.Context, userID uint64, roomID uint64) error {
	// 检查是否在房间中
	inRoom, err := s.participantRepo.ExistsActive(ctx, roomID, userID)
	if err != nil {
		return errcode.ErrDBError.WithMessage("检查参与者状态失败")
	}
	if !inRoom {
		return errcode.ErrNotInRoom
	}

	// 设置离开状态
	if err = s.participantRepo.Leave(ctx, roomID, userID); err != nil {
		return errcode.ErrDBError.WithMessage("离开房间失败")
	}

	// 检查房间是否还有参与者
	count, err := s.participantRepo.CountActiveByRoom(ctx, roomID)
	if err != nil {
		return errcode.ErrDBError.WithMessage("统计参与者数量失败")
	}

	// 如果房间空了，关闭房间
	if count == 0 {
		if err = s.roomRepo.Delete(ctx, roomID); err != nil {
			return errcode.ErrDBError.WithMessage("关闭房间失败")
		}
	}

	return nil
}

// GetRoomList 获取房间列表
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

	var responses []dto.RoomInfoResponse
	for _, room := range rooms {
		count, _ := s.participantRepo.CountActiveByRoom(ctx, room.ID)
		resp, _ := s.buildRoomInfoResponse(ctx, &room, count)
		responses = append(responses, *resp)
	}

	return responses, total, nil
}

// GetRoomParticipants 获取房间参与者
func (s *RoomService) GetRoomParticipants(ctx context.Context, roomID uint64) ([]dto.ParticipantResponse, error) {
	participants, err := s.participantRepo.FindActiveByRoom(ctx, roomID)
	if err != nil {
		return nil, errcode.ErrDBError.WithMessage("查询参与者失败")
	}

	var responses []dto.ParticipantResponse
	for _, p := range participants {
		user, err := s.userRepo.FindByID(ctx, p.UserID)
		if err != nil {
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

	var responses []dto.OnlineUserResponse
	for _, p := range participants {
		user, err := s.userRepo.FindByID(ctx, p.UserID)
		if err != nil {
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

	// 删除所有参与者记录
	if err = s.participantRepo.DeleteByRoom(ctx, roomID); err != nil {
		return errcode.ErrDBError.WithMessage("删除参与者记录失败")
	}

	// 删除房间
	if err = s.roomRepo.Delete(ctx, roomID); err != nil {
		return errcode.ErrDBError.WithMessage("删除房间失败")
	}

	return nil
}
