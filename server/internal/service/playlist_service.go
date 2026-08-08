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

// PlaylistService 播放列表服务（channel_id 维度）
type PlaylistService struct {
	playlistRepo *repository.PlaylistRepository
	channelRepo  *repository.ChannelRepository
	serverSvc    *ServerService
}

// NewPlaylistService 创建播放列表服务实例
func NewPlaylistService(
	playlistRepo *repository.PlaylistRepository,
	channelRepo *repository.ChannelRepository,
	serverSvc *ServerService,
) *PlaylistService {
	return &PlaylistService{
		playlistRepo: playlistRepo,
		channelRepo:  channelRepo,
		serverSvc:    serverSvc,
	}
}

// checkChannelType 检查频道类型是否支持播放列表（仅语音频道支持）
func (s *PlaylistService) checkChannelType(ctx context.Context, channelID uint64) error {
	channel, err := s.channelRepo.FindByID(ctx, channelID)
	if err != nil {
		return errcode.ErrChannelNotFound
	}
	if channel.Type != model.ChannelTypeVoice {
		return errcode.ErrBadRequest.WithMessage("该频道类型不支持音乐播放")
	}
	return nil
}

// AddItem 添加播放项
func (s *PlaylistService) AddItem(ctx context.Context, channelID, userID uint64, req *dto.AddPlaylistItemRequest) (*dto.PlaylistItemResponse, error) {
	if err := s.checkChannelType(ctx, channelID); err != nil {
		return nil, err
	}

	maxOrder, err := s.playlistRepo.GetMaxOrder(ctx, channelID)
	if err != nil {
		return nil, errcode.ErrDBError.WithMessage("获取播放顺序失败")
	}

	item := &model.PlaylistItem{
		ChannelID: channelID,
		AddedBy:   userID,
		Title:     util.TrimAndEscape(req.Title),
		Artist:    util.TrimAndEscape(req.Artist),
		MusicURL:  req.MusicURL,
		Duration:  req.Duration,
		PlayOrder: maxOrder + 1,
		Status:    0,
	}

	if err = s.playlistRepo.Create(ctx, item); err != nil {
		return nil, errcode.ErrDBError.WithMessage("添加播放项失败")
	}

	return &dto.PlaylistItemResponse{
		ID:        item.ID,
		Title:     item.Title,
		Artist:    item.Artist,
		MusicURL:  item.MusicURL,
		Duration:  item.Duration,
		PlayOrder: item.PlayOrder,
		Status:    item.Status,
		AddedBy:   item.AddedBy,
	}, nil
}

// RemoveItem 删除播放项
func (s *PlaylistService) RemoveItem(ctx context.Context, channelID, userID, itemID uint64) error {
	item, err := s.playlistRepo.FindByID(ctx, itemID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errcode.ErrPlaylistItemNotFound
		}
		return errcode.ErrDBError.WithMessage("查询播放项失败")
	}

	if item.ChannelID != channelID {
		return errcode.ErrPlaylistItemNotFound
	}

	if item.Status == 1 {
		return errcode.ErrBadRequest.WithMessage("正在播放的歌曲不能删除")
	}

	if err = s.playlistRepo.Delete(ctx, itemID); err != nil {
		return errcode.ErrDBError.WithMessage("删除播放项失败")
	}

	return nil
}

// GetPlaylist 获取播放列表
func (s *PlaylistService) GetPlaylist(ctx context.Context, channelID, userID uint64) ([]dto.PlaylistItemResponse, error) {
	if err := s.checkChannelType(ctx, channelID); err != nil {
		return nil, err
	}

	items, err := s.playlistRepo.FindByChannel(ctx, channelID)
	if err != nil {
		return nil, errcode.ErrDBError.WithMessage("查询播放列表失败")
	}

	var responses []dto.PlaylistItemResponse
	for _, item := range items {
		responses = append(responses, dto.PlaylistItemResponse{
			ID:        item.ID,
			Title:     item.Title,
			Artist:    item.Artist,
			MusicURL:  item.MusicURL,
			Duration:  item.Duration,
			PlayOrder: item.PlayOrder,
			Status:    item.Status,
			AddedBy:   item.AddedBy,
		})
	}

	return responses, nil
}

// Play 播放
func (s *PlaylistService) Play(ctx context.Context, channelID, userID uint64) (*dto.PlaylistItemResponse, error) {
	if err := s.checkChannelType(ctx, channelID); err != nil {
		return nil, err
	}

	playing, err := s.playlistRepo.FindPlayingByChannel(ctx, channelID)
	if err == nil {
		return &dto.PlaylistItemResponse{
			ID:        playing.ID,
			Title:     playing.Title,
			Artist:    playing.Artist,
			MusicURL:  playing.MusicURL,
			Duration:  playing.Duration,
			PlayOrder: playing.PlayOrder,
			Status:    playing.Status,
			AddedBy:   playing.AddedBy,
		}, nil
	}

	items, err := s.playlistRepo.FindWaitingByChannel(ctx, channelID)
	if err != nil {
		return nil, errcode.ErrDBError.WithMessage("查询播放列表失败")
	}
	if len(items) == 0 {
		return nil, errcode.ErrPlaylistEmpty
	}

	firstItem := items[0]
	firstItem.Status = 1
	if err = s.playlistRepo.Update(ctx, &firstItem); err != nil {
		return nil, errcode.ErrDBError.WithMessage("更新播放状态失败")
	}

	return &dto.PlaylistItemResponse{
		ID:        firstItem.ID,
		Title:     firstItem.Title,
		Artist:    firstItem.Artist,
		MusicURL:  firstItem.MusicURL,
		Duration:  firstItem.Duration,
		PlayOrder: firstItem.PlayOrder,
		Status:    firstItem.Status,
		AddedBy:   firstItem.AddedBy,
	}, nil
}

// Pause 暂停
func (s *PlaylistService) Pause(ctx context.Context, channelID, userID uint64) error {
	playing, err := s.playlistRepo.FindPlayingByChannel(ctx, channelID)
	if err != nil {
		return errcode.ErrBadRequest.WithMessage("没有正在播放的项目")
	}

	playing.Status = 0
	if err = s.playlistRepo.Update(ctx, playing); err != nil {
		return errcode.ErrDBError.WithMessage("更新播放状态失败")
	}

	return nil
}

// Skip 跳过
func (s *PlaylistService) Skip(ctx context.Context, channelID, userID uint64) (*dto.PlaylistItemResponse, error) {
	playing, err := s.playlistRepo.FindPlayingByChannel(ctx, channelID)
	if err != nil {
		return nil, errcode.ErrBadRequest.WithMessage("没有正在播放的项目")
	}

	playing.Status = 2
	if err = s.playlistRepo.Update(ctx, playing); err != nil {
		return nil, errcode.ErrDBError.WithMessage("更新播放状态失败")
	}

	return s.Play(ctx, channelID, userID)
}

// Reorder 调整播放列表顺序
func (s *PlaylistService) Reorder(ctx context.Context, channelID, userID uint64, req *dto.ReorderPlaylistRequest) error {
	if err := s.checkChannelType(ctx, channelID); err != nil {
		return err
	}

	if err := s.playlistRepo.Reorder(ctx, channelID, req.ItemIDs); err != nil {
		return errcode.ErrDBError.WithMessage("重排序播放列表失败")
	}

	return nil
}
