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

// VoiceService 语音服务（channel_id 维度）
type VoiceService struct {
	voiceRepo   *repository.VoiceRepository
	channelRepo *repository.ChannelRepository
	userRepo    *repository.UserRepository
	serverSvc   *ServerService
}

// NewVoiceService 创建语音服务实例
func NewVoiceService(
	voiceRepo *repository.VoiceRepository,
	channelRepo *repository.ChannelRepository,
	userRepo *repository.UserRepository,
	serverSvc *ServerService,
) *VoiceService {
	return &VoiceService{
		voiceRepo:   voiceRepo,
		channelRepo: channelRepo,
		userRepo:    userRepo,
		serverSvc:   serverSvc,
	}
}

// JoinVoice 加入语音频道
func (s *VoiceService) JoinVoice(ctx context.Context, channelID, userID uint64) (*dto.VoiceParticipantResponse, error) {
	// 校验频道存在且为语音频道
	channel, err := s.channelRepo.FindByID(ctx, channelID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errcode.ErrChannelNotFound
		}
		return nil, errcode.ErrDBError.WithMessage("查询频道失败")
	}
	if channel.Type != model.ChannelTypeVoice {
		return nil, errcode.ErrBadRequest.WithMessage("该频道不是语音频道")
	}

	// 校验权限
	if err := s.serverSvc.CheckPermission(ctx, channel.ServerID, userID, model.PermConnectVoice); err != nil {
		return nil, err
	}

	// 创建/更新语音参与者
	p := &model.VoiceParticipant{
		ChannelID: channelID,
		UserID:    userID,
		JoinedAt:  time.Now(),
	}
	if err = s.voiceRepo.UpsertParticipant(ctx, p); err != nil {
		return nil, errcode.ErrDBError.WithMessage("加入语音失败")
	}

	// 创建语音会话历史记录
	session := &model.VoiceSession{
		ChannelID: channelID,
		UserID:    userID,
		JoinedAt:  time.Now(),
	}
	_ = s.voiceRepo.CreateSession(ctx, session)

	return s.buildParticipantResponse(ctx, p)
}

// LeaveVoice 离开语音频道
func (s *VoiceService) LeaveVoice(ctx context.Context, channelID, userID uint64) error {
	// 删除参与者
	if err := s.voiceRepo.DeleteParticipant(ctx, channelID, userID); err != nil {
		return errcode.ErrDBError.WithMessage("离开语音失败")
	}

	// 结束语音会话历史记录
	_ = s.voiceRepo.EndSession(ctx, channelID, userID)

	return nil
}

// GetVoiceParticipants 获取语音频道参与者列表
func (s *VoiceService) GetVoiceParticipants(ctx context.Context, channelID, userID uint64) ([]dto.VoiceParticipantResponse, error) {
	participants, err := s.voiceRepo.FindParticipantsByChannel(ctx, channelID)
	if err != nil {
		return nil, errcode.ErrDBError.WithMessage("查询语音参与者失败")
	}

	if len(participants) == 0 {
		return []dto.VoiceParticipantResponse{}, nil
	}

	// 批量查询用户信息
	userIDs := make([]uint64, 0, len(participants))
	for _, p := range participants {
		userIDs = append(userIDs, p.UserID)
	}
	userMap, _ := s.userRepo.FindByIDs(ctx, userIDs)

	var responses []dto.VoiceParticipantResponse
	for _, p := range participants {
		userName := ""
		avatarURL := ""
		if user, exists := userMap[p.UserID]; exists {
			userName = user.Username
			avatarURL = user.AvatarURL
		}

		responses = append(responses, dto.VoiceParticipantResponse{
			ID:         p.ID,
			ChannelID:  p.ChannelID,
			UserID:     p.UserID,
			Username:   userName,
			AvatarURL:  avatarURL,
			IsMuted:    p.IsMuted,
			IsDeafened: p.IsDeafened,
			IsSpeaking: p.IsSpeaking,
			Volume:     p.Volume,
			JoinedAt:   p.JoinedAt.Format("2006-01-02 15:04:05"),
		})
	}

	return responses, nil
}

// SetMute 设置静音状态
func (s *VoiceService) SetMute(ctx context.Context, channelID, userID uint64, isMuted bool) error {
	if err := s.voiceRepo.UpdateMute(ctx, channelID, userID, isMuted); err != nil {
		return errcode.ErrDBError.WithMessage("更新静音状态失败")
	}
	return nil
}

// StartScreenShare 开始屏幕共享
func (s *VoiceService) StartScreenShare(ctx context.Context, channelID, userID uint64) (*dto.ScreenShareResponse, error) {
	// 校验频道存在
	channel, err := s.channelRepo.FindByID(ctx, channelID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errcode.ErrChannelNotFound
		}
		return nil, errcode.ErrDBError.WithMessage("查询频道失败")
	}

	// 校验权限
	if err := s.serverSvc.CheckPermission(ctx, channel.ServerID, userID, model.PermScreenShare); err != nil {
		return nil, err
	}

	// 检查是否已有正在进行的屏幕共享
	activeShare, err := s.voiceRepo.FindActiveScreenShareByChannel(ctx, channelID)
	if err == nil && activeShare != nil {
		if activeShare.UserID == userID {
			user, _ := s.userRepo.FindByID(ctx, userID)
			return &dto.ScreenShareResponse{
				ID:        activeShare.ID,
				ChannelID: activeShare.ChannelID,
				UserID:    activeShare.UserID,
				Username:  user.Username,
				StartedAt: activeShare.StartedAt.Format("2006-01-02 15:04:05"),
			}, nil
		}
		return nil, errcode.ErrBadRequest.WithMessage("频道内已有其他用户正在共享屏幕")
	}

	// 创建屏幕共享会话
	session := &model.ScreenShareSession{
		ChannelID: channelID,
		UserID:    userID,
		StartedAt: time.Now(),
	}
	if err = s.voiceRepo.CreateScreenShare(ctx, session); err != nil {
		return nil, errcode.ErrDBError.WithMessage("创建屏幕共享会话失败")
	}

	user, _ := s.userRepo.FindByID(ctx, userID)
	return &dto.ScreenShareResponse{
		ID:        session.ID,
		ChannelID: session.ChannelID,
		UserID:    session.UserID,
		Username:  user.Username,
		StartedAt: session.StartedAt.Format("2006-01-02 15:04:05"),
	}, nil
}

// StopScreenShare 停止屏幕共享
func (s *VoiceService) StopScreenShare(ctx context.Context, channelID, userID uint64) error {
	if err := s.voiceRepo.EndScreenShare(ctx, channelID, userID); err != nil {
		return errcode.ErrBadRequest.WithMessage("没有正在进行的屏幕共享")
	}
	return nil
}

// GetActiveScreenShare 获取频道当前屏幕共享
func (s *VoiceService) GetActiveScreenShare(ctx context.Context, channelID uint64) (*dto.ScreenShareResponse, error) {
	session, err := s.voiceRepo.FindActiveScreenShareByChannel(ctx, channelID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // 没有正在进行的屏幕共享
		}
		return nil, errcode.ErrDBError.WithMessage("查询屏幕共享失败")
	}

	user, _ := s.userRepo.FindByID(ctx, session.UserID)
	return &dto.ScreenShareResponse{
		ID:        session.ID,
		ChannelID: session.ChannelID,
		UserID:    session.UserID,
		Username:  user.Username,
		StartedAt: session.StartedAt.Format("2006-01-02 15:04:05"),
	}, nil
}

// SendSignal 发送 WebRTC 信令（REST 备用通道）
func (s *VoiceService) SendSignal(ctx context.Context, channelID, userID uint64, req *dto.WebRTCSignalRequest) error {
	// 校验频道存在
	channel, err := s.channelRepo.FindByID(ctx, channelID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errcode.ErrChannelNotFound
		}
		return errcode.ErrDBError.WithMessage("查询频道失败")
	}

	// 校验权限
	if err := s.serverSvc.CheckPermission(ctx, channel.ServerID, userID, model.PermConnectVoice); err != nil {
		return err
	}

	// 校验信令类型
	if req.Type != "offer" && req.Type != "answer" && req.Type != "ice-candidate" {
		return errcode.ErrInvalidParam.WithMessage("无效的信令类型")
	}

	// 实时信令转发由 WebSocket 层处理，REST 端点仅作校验与备用
	return nil
}

// buildParticipantResponse 构建参与者响应
func (s *VoiceService) buildParticipantResponse(ctx context.Context, p *model.VoiceParticipant) (*dto.VoiceParticipantResponse, error) {
	user, err := s.userRepo.FindByID(ctx, p.UserID)
	userName := ""
	avatarURL := ""
	if err == nil {
		userName = user.Username
		avatarURL = user.AvatarURL
	}

	return &dto.VoiceParticipantResponse{
		ID:         p.ID,
		ChannelID:  p.ChannelID,
		UserID:     p.UserID,
		Username:   userName,
		AvatarURL:  avatarURL,
		IsMuted:    p.IsMuted,
		IsDeafened: p.IsDeafened,
		IsSpeaking: p.IsSpeaking,
		Volume:     p.Volume,
		JoinedAt:   p.JoinedAt.Format("2006-01-02 15:04:05"),
	}, nil
}
