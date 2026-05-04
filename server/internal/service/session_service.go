package service

import (
	"context"
	"errors"
	"time"

	"github.com/yourorg/livemix/internal/dto"
	"github.com/yourorg/livemix/internal/model"
	"github.com/yourorg/livemix/internal/repository"
	"github.com/yourorg/livemix/pkg/errcode"
	"gorm.io/gorm"
)

// ScreenShareService 屏幕共享服务
type ScreenShareService struct {
	screenShareRepo *repository.ScreenShareRepository
	roomRepo        *repository.RoomRepository
	userRepo        *repository.UserRepository
	participantRepo *repository.RoomParticipantRepository
}

// NewScreenShareService 创建屏幕共享服务实例
func NewScreenShareService(
	screenShareRepo *repository.ScreenShareRepository,
	roomRepo *repository.RoomRepository,
	userRepo *repository.UserRepository,
	participantRepo *repository.RoomParticipantRepository,
) *ScreenShareService {
	return &ScreenShareService{
		screenShareRepo: screenShareRepo,
		roomRepo:        roomRepo,
		userRepo:        userRepo,
		participantRepo: participantRepo,
	}
}

// StartScreenShare 开始屏幕共享
func (s *ScreenShareService) StartScreenShare(ctx context.Context, roomID, userID uint64) (*dto.ScreenShareResponse, error) {
	// 检查房间是否存在
	room, err := s.roomRepo.FindByID(ctx, roomID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errcode.ErrRoomNotFound
		}
		return nil, errcode.ErrDBError.WithMessage("查询房间失败")
	}

	// 检查用户是否在房间中
	inRoom, err := s.participantRepo.ExistsActive(ctx, roomID, userID)
	if err != nil {
		return nil, errcode.ErrDBError.WithMessage("检查参与者状态失败")
	}
	if !inRoom {
		return nil, errcode.ErrNotInRoom
	}

	// 检查房间类型是否支持屏幕共享（仅语音房支持）
	if room.RoomType != 2 {
		return nil, errcode.ErrBadRequest.WithMessage("该房间类型不支持屏幕共享")
	}

	// 检查是否已有正在进行的屏幕共享
	activeShare, err := s.screenShareRepo.FindActiveByRoom(ctx, roomID)
	if err == nil && activeShare != nil {
		// 如果是同一个用户，返回现有会话
		if activeShare.UserID == userID {
			user, _ := s.userRepo.FindByID(ctx, userID)
			return &dto.ScreenShareResponse{
				ID:        activeShare.ID,
				RoomID:    activeShare.RoomID,
				UserID:    activeShare.UserID,
				Username:  user.Username,
				StartedAt: activeShare.StartedAt.Format("2006-01-02 15:04:05"),
			}, nil
		}
		return nil, errcode.ErrBadRequest.WithMessage("房间内已有其他用户正在共享屏幕")
	}

	// 创建屏幕共享会话
	session := &model.ScreenShareSession{
		RoomID:    roomID,
		UserID:    userID,
		StartedAt: time.Now(),
	}

	if err = s.screenShareRepo.Create(ctx, session); err != nil {
		return nil, errcode.ErrDBError.WithMessage("创建屏幕共享会话失败")
	}

	// 更新参与者状态
	participant, err := s.participantRepo.FindByRoomAndUser(ctx, roomID, userID)
	if err == nil {
		participant.IsScreenSharing = true
		s.participantRepo.Update(ctx, participant)
	}

	// 获取用户信息
	user, _ := s.userRepo.FindByID(ctx, userID)

	return &dto.ScreenShareResponse{
		ID:        session.ID,
		RoomID:    session.RoomID,
		UserID:    session.UserID,
		Username:  user.Username,
		StartedAt: session.StartedAt.Format("2006-01-02 15:04:05"),
	}, nil
}

// StopScreenShare 停止屏幕共享
func (s *ScreenShareService) StopScreenShare(ctx context.Context, roomID, userID uint64) error {
	// 查找用户正在进行的屏幕共享
	session, err := s.screenShareRepo.FindActiveByUser(ctx, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errcode.ErrBadRequest.WithMessage("没有正在进行的屏幕共享")
		}
		return errcode.ErrDBError.WithMessage("查询屏幕共享会话失败")
	}

	// 验证会话属于该房间
	if session.RoomID != roomID {
		return errcode.ErrBadRequest.WithMessage("屏幕共享会话不属于该房间")
	}

	// 结束会话
	if err = s.screenShareRepo.EndSession(ctx, session.ID); err != nil {
		return errcode.ErrDBError.WithMessage("结束屏幕共享会话失败")
	}

	// 更新参与者状态
	participant, err := s.participantRepo.FindByRoomAndUser(ctx, roomID, userID)
	if err == nil {
		participant.IsScreenSharing = false
		s.participantRepo.Update(ctx, participant)
	}

	return nil
}

// GetActiveScreenShare 获取房间内正在进行的屏幕共享
func (s *ScreenShareService) GetActiveScreenShare(ctx context.Context, roomID uint64) (*dto.ScreenShareResponse, error) {
	session, err := s.screenShareRepo.FindActiveByRoom(ctx, roomID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // 没有正在进行的屏幕共享
		}
		return nil, errcode.ErrDBError.WithMessage("查询屏幕共享会话失败")
	}

	user, _ := s.userRepo.FindByID(ctx, session.UserID)

	return &dto.ScreenShareResponse{
		ID:        session.ID,
		RoomID:    session.RoomID,
		UserID:    session.UserID,
		Username:  user.Username,
		StartedAt: session.StartedAt.Format("2006-01-02 15:04:05"),
	}, nil
}

// VoiceService 语音服务
type VoiceService struct {
	voiceRepo       *repository.VoiceSessionRepository
	roomRepo        *repository.RoomRepository
	userRepo        *repository.UserRepository
	participantRepo *repository.RoomParticipantRepository
}

// NewVoiceService 创建语音服务实例
func NewVoiceService(
	voiceRepo *repository.VoiceSessionRepository,
	roomRepo *repository.RoomRepository,
	userRepo *repository.UserRepository,
	participantRepo *repository.RoomParticipantRepository,
) *VoiceService {
	return &VoiceService{
		voiceRepo:       voiceRepo,
		roomRepo:        roomRepo,
		userRepo:        userRepo,
		participantRepo: participantRepo,
	}
}

// JoinVoice 加入语音
func (s *VoiceService) JoinVoice(ctx context.Context, roomID, userID uint64) (*dto.VoiceSessionResponse, error) {
	// 检查房间是否存在
	room, err := s.roomRepo.FindByID(ctx, roomID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errcode.ErrRoomNotFound
		}
		return nil, errcode.ErrDBError.WithMessage("查询房间失败")
	}

	// 检查房间类型是否支持语音（仅语音房支持）
	if room.RoomType != 2 {
		return nil, errcode.ErrBadRequest.WithMessage("该房间类型不支持语音")
	}

	// 检查用户是否在房间中
	inRoom, err := s.participantRepo.ExistsActive(ctx, roomID, userID)
	if err != nil {
		return nil, errcode.ErrDBError.WithMessage("检查参与者状态失败")
	}
	if !inRoom {
		return nil, errcode.ErrNotInRoom
	}

	// 检查是否已在语音中
	existingSession, err := s.voiceRepo.FindActiveByRoomAndUser(ctx, roomID, userID)
	if err == nil && existingSession != nil {
		user, _ := s.userRepo.FindByID(ctx, userID)
		return &dto.VoiceSessionResponse{
			ID:       existingSession.ID,
			RoomID:   existingSession.RoomID,
			UserID:   existingSession.UserID,
			Username: user.Username,
			JoinedAt: existingSession.JoinedAt.Format("2006-01-02 15:04:05"),
		}, nil
	}

	// 创建语音会话
	session := &model.VoiceSession{
		RoomID:   roomID,
		UserID:   userID,
		JoinedAt: time.Now(),
	}

	if err = s.voiceRepo.Create(ctx, session); err != nil {
		return nil, errcode.ErrDBError.WithMessage("创建语音会话失败")
	}

	user, _ := s.userRepo.FindByID(ctx, userID)

	return &dto.VoiceSessionResponse{
		ID:       session.ID,
		RoomID:   session.RoomID,
		UserID:   session.UserID,
		Username: user.Username,
		JoinedAt: session.JoinedAt.Format("2006-01-02 15:04:05"),
	}, nil
}

// LeaveVoice 离开语音
func (s *VoiceService) LeaveVoice(ctx context.Context, roomID, userID uint64) error {
	if err := s.voiceRepo.Leave(ctx, roomID, userID); err != nil {
		return errcode.ErrDBError.WithMessage("离开语音会话失败")
	}
	return nil
}

// GetVoiceParticipants 获取语音参与者列表
func (s *VoiceService) GetVoiceParticipants(ctx context.Context, roomID uint64) ([]dto.VoiceSessionResponse, error) {
	sessions, err := s.voiceRepo.FindActiveByRoom(ctx, roomID)
	if err != nil {
		return nil, errcode.ErrDBError.WithMessage("查询语音会话失败")
	}

	var responses []dto.VoiceSessionResponse
	for _, session := range sessions {
		user, err := s.userRepo.FindByID(ctx, session.UserID)
		if err != nil {
			continue
		}
		responses = append(responses, dto.VoiceSessionResponse{
			ID:       session.ID,
			RoomID:   session.RoomID,
			UserID:   session.UserID,
			Username: user.Username,
			JoinedAt: session.JoinedAt.Format("2006-01-02 15:04:05"),
		})
	}

	return responses, nil
}

// SetMute 设置静音状态
func (s *VoiceService) SetMute(ctx context.Context, roomID, userID uint64, muted bool) error {
	participant, err := s.participantRepo.FindByRoomAndUser(ctx, roomID, userID)
	if err != nil {
		return errcode.ErrNotInRoom
	}

	participant.IsMuted = muted
	if err = s.participantRepo.Update(ctx, participant); err != nil {
		return errcode.ErrDBError.WithMessage("更新静音状态失败")
	}

	return nil
}
