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

// ChannelService 频道服务
type ChannelService struct {
	channelRepo *repository.ChannelRepository
	serverRepo  *repository.ServerRepository
	serverSvc   *ServerService
}

// NewChannelService 创建频道服务实例
func NewChannelService(
	channelRepo *repository.ChannelRepository,
	serverRepo *repository.ServerRepository,
	serverSvc *ServerService,
) *ChannelService {
	return &ChannelService{
		channelRepo: channelRepo,
		serverRepo:  serverRepo,
		serverSvc:   serverSvc,
	}
}

// CreateChannel 创建频道
func (s *ChannelService) CreateChannel(ctx context.Context, serverID, userID uint64, req *dto.CreateChannelRequest) (*dto.ChannelResponse, error) {
	// 校验权限
	if err := s.serverSvc.CheckPermission(ctx, serverID, userID, model.PermManageChannels); err != nil {
		return nil, err
	}

	// 设置默认值
	bitrate := req.Bitrate
	if bitrate == 0 {
		bitrate = 64000
	}

	channel := &model.Channel{
		ServerID:  serverID,
		Name:      util.TrimAndEscape(req.Name),
		Type:      req.Type,
		Topic:     util.TrimAndEscape(req.Topic),
		ParentID:  req.ParentID,
		Bitrate:   bitrate,
		UserLimit: req.UserLimit,
		SlowMode:  req.SlowMode,
	}

	if err := s.channelRepo.Create(ctx, channel); err != nil {
		return nil, errcode.ErrDBError.WithMessage("创建频道失败")
	}

	return s.buildChannelResponse(channel), nil
}

// GetChannelTree 获取频道的树形列表
func (s *ChannelService) GetChannelTree(ctx context.Context, serverID, userID uint64) ([]dto.ChannelResponse, error) {
	// 检查用户是否是服务器成员
	isMember, err := s.serverSvc.memberRepo.Exists(ctx, serverID, userID)
	if err != nil {
		return nil, errcode.ErrDBError.WithMessage("检查成员状态失败")
	}
	if !isMember {
		return nil, errcode.ErrNotInServer
	}

	channels, err := s.channelRepo.FindByServer(ctx, serverID)
	if err != nil {
		return nil, errcode.ErrDBError.WithMessage("查询频道失败")
	}

	var responses []dto.ChannelResponse
	for _, ch := range channels {
		responses = append(responses, *s.buildChannelResponse(&ch))
	}

	return responses, nil
}

// UpdateChannel 更新频道
func (s *ChannelService) UpdateChannel(ctx context.Context, channelID, userID uint64, req *dto.UpdateChannelRequest) (*dto.ChannelResponse, error) {
	channel, err := s.channelRepo.FindByID(ctx, channelID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errcode.ErrChannelNotFound
		}
		return nil, errcode.ErrDBError.WithMessage("查询频道失败")
	}

	// 校验权限
	if err := s.serverSvc.CheckPermission(ctx, channel.ServerID, userID, model.PermManageChannels); err != nil {
		return nil, err
	}

	if req.Name != "" {
		channel.Name = util.TrimAndEscape(req.Name)
	}
	if req.Topic != "" {
		channel.Topic = util.TrimAndEscape(req.Topic)
	}
	if req.ParentID != nil {
		channel.ParentID = req.ParentID
	}
	if req.Position != nil {
		channel.Position = *req.Position
	}
	if req.Bitrate != nil {
		channel.Bitrate = *req.Bitrate
	}
	if req.UserLimit != nil {
		channel.UserLimit = *req.UserLimit
	}
	if req.SlowMode != nil {
		channel.SlowMode = *req.SlowMode
	}

	if err = s.channelRepo.Update(ctx, channel); err != nil {
		return nil, errcode.ErrDBError.WithMessage("更新频道失败")
	}

	return s.buildChannelResponse(channel), nil
}

// DeleteChannel 删除频道
func (s *ChannelService) DeleteChannel(ctx context.Context, channelID, userID uint64) error {
	channel, err := s.channelRepo.FindByID(ctx, channelID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errcode.ErrChannelNotFound
		}
		return errcode.ErrDBError.WithMessage("查询频道失败")
	}

	// 校验权限
	if err := s.serverSvc.CheckPermission(ctx, channel.ServerID, userID, model.PermManageChannels); err != nil {
		return err
	}

	if err = s.channelRepo.Delete(ctx, channelID); err != nil {
		return errcode.ErrDBError.WithMessage("删除频道失败")
	}

	return nil
}

// ReorderChannels 批量更新频道排序
func (s *ChannelService) ReorderChannels(ctx context.Context, serverID, userID uint64, req *dto.ReorderChannelsRequest) error {
	// 校验权限
	if err := s.serverSvc.CheckPermission(ctx, serverID, userID, model.PermManageChannels); err != nil {
		return err
	}

	orders := make([]struct {
		ID       uint64
		Position int
	}, len(req.Orders))
	for i, o := range req.Orders {
		orders[i] = struct {
			ID       uint64
			Position int
		}{ID: o.ID, Position: o.Position}
	}

	if err := s.channelRepo.Reorder(ctx, serverID, orders); err != nil {
		return errcode.ErrDBError.WithMessage("更新频道排序失败")
	}

	return nil
}

// buildChannelResponse 构建频道响应
func (s *ChannelService) buildChannelResponse(ch *model.Channel) *dto.ChannelResponse {
	return &dto.ChannelResponse{
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
	}
}
