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

// PlaylistService 播放列表服务
type PlaylistService struct {
	playlistRepo    *repository.PlaylistRepository
	roomRepo        *repository.RoomRepository
	participantRepo *repository.RoomParticipantRepository
}

// NewPlaylistService 创建播放列表服务实例
func NewPlaylistService(playlistRepo *repository.PlaylistRepository, roomRepo *repository.RoomRepository, participantRepo *repository.RoomParticipantRepository) *PlaylistService {
	return &PlaylistService{
		playlistRepo:    playlistRepo,
		roomRepo:        roomRepo,
		participantRepo: participantRepo,
	}
}

// checkUserInRoom 检查用户是否在房间中
func (s *PlaylistService) checkUserInRoom(ctx context.Context, roomID, userID uint64) error {
	inRoom, err := s.participantRepo.ExistsActive(ctx, roomID, userID)
	if err != nil {
		return errcode.ErrDBError.WithMessage("检查参与者状态失败")
	}
	if !inRoom {
		return errcode.ErrNotInRoom
	}
	return nil
}

// AddItem 添加播放项
func (s *PlaylistService) AddItem(ctx context.Context, roomID, userID uint64, req *dto.AddPlaylistItemRequest) (*dto.PlaylistItemResponse, error) {
	// 检查用户是否在房间中
	if err := s.checkUserInRoom(ctx, roomID, userID); err != nil {
		return nil, err
	}

	// 获取最大顺序号
	maxOrder, err := s.playlistRepo.GetMaxOrder(ctx, roomID)
	if err != nil {
		return nil, errcode.ErrDBError.WithMessage("获取播放顺序失败")
	}

	item := &model.PlaylistItem{
		RoomID:    roomID,
		AddedBy:   userID,
		Title:     util.TrimAndEscape(req.Title),
		Artist:    util.TrimAndEscape(req.Artist),
		MusicURL:  req.MusicURL,
		Duration:  req.Duration,
		PlayOrder: maxOrder + 1,
		Status:    0, // 等待播放
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
func (s *PlaylistService) RemoveItem(ctx context.Context, roomID, userID, itemID uint64) error {
	// 检查用户是否在房间中
	if err := s.checkUserInRoom(ctx, roomID, userID); err != nil {
		return err
	}

	item, err := s.playlistRepo.FindByID(ctx, itemID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errcode.ErrPlaylistItemNotFound
		}
		return errcode.ErrDBError.WithMessage("查询播放项失败")
	}

	if item.RoomID != roomID {
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
func (s *PlaylistService) GetPlaylist(ctx context.Context, roomID, userID uint64) ([]dto.PlaylistItemResponse, error) {
	// 检查用户是否在房间中
	if err := s.checkUserInRoom(ctx, roomID, userID); err != nil {
		return nil, err
	}

	items, err := s.playlistRepo.FindByRoom(ctx, roomID)
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
func (s *PlaylistService) Play(ctx context.Context, roomID, userID uint64) (*dto.PlaylistItemResponse, error) {
	// 检查用户是否在房间中
	if err := s.checkUserInRoom(ctx, roomID, userID); err != nil {
		return nil, err
	}

	// 检查是否有正在播放的项目
	playing, err := s.playlistRepo.FindPlayingByRoom(ctx, roomID)
	if err == nil {
		// 已有正在播放的项目
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

	// 获取等待播放的项目
	items, err := s.playlistRepo.FindWaitingByRoom(ctx, roomID)
	if err != nil {
		return nil, errcode.ErrDBError.WithMessage("查询播放列表失败")
	}
	if len(items) == 0 {
		return nil, errcode.ErrPlaylistEmpty
	}

	// 设置第一个项目为播放中
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
func (s *PlaylistService) Pause(ctx context.Context, roomID, userID uint64) error {
	// 检查用户是否在房间中
	if err := s.checkUserInRoom(ctx, roomID, userID); err != nil {
		return err
	}

	playing, err := s.playlistRepo.FindPlayingByRoom(ctx, roomID)
	if err != nil {
		return errcode.ErrBadRequest.WithMessage("没有正在播放的项目")
	}

	playing.Status = 0 // 恢复为等待状态
	if err = s.playlistRepo.Update(ctx, playing); err != nil {
		return errcode.ErrDBError.WithMessage("更新播放状态失败")
	}

	return nil
}

// Skip 跳过
func (s *PlaylistService) Skip(ctx context.Context, roomID, userID uint64) (*dto.PlaylistItemResponse, error) {
	// 检查用户是否在房间中
	if err := s.checkUserInRoom(ctx, roomID, userID); err != nil {
		return nil, err
	}

	// 获取正在播放的项目
	playing, err := s.playlistRepo.FindPlayingByRoom(ctx, roomID)
	if err != nil {
		return nil, errcode.ErrBadRequest.WithMessage("没有正在播放的项目")
	}

	// 标记为已播
	playing.Status = 2
	if err = s.playlistRepo.Update(ctx, playing); err != nil {
		return nil, errcode.ErrDBError.WithMessage("更新播放状态失败")
	}

	// 播放下一首
	return s.Play(ctx, roomID, userID)
}

// ChatService 聊天服务
type ChatService struct {
	msgRepo         *repository.ChatMessageRepository
	userRepo        *repository.UserRepository
	participantRepo *repository.RoomParticipantRepository
}

// NewChatService 创建聊天服务实例
func NewChatService(msgRepo *repository.ChatMessageRepository, userRepo *repository.UserRepository, participantRepo *repository.RoomParticipantRepository) *ChatService {
	return &ChatService{
		msgRepo:         msgRepo,
		userRepo:        userRepo,
		participantRepo: participantRepo,
	}
}

// SendMessage 发送消息
func (s *ChatService) SendMessage(ctx context.Context, roomID, userID uint64, req *dto.SendMessageRequest) (*dto.MessageResponse, error) {
	// 检查用户是否在房间中
	inRoom, err := s.participantRepo.ExistsActive(ctx, roomID, userID)
	if err != nil {
		return nil, errcode.ErrDBError.WithMessage("检查参与者状态失败")
	}
	if !inRoom {
		return nil, errcode.ErrNotInRoom
	}

	// XSS过滤
	content := util.TrimAndEscape(req.Content)

	msg := &model.ChatMessage{
		RoomID:       roomID,
		SenderUserID: userID,
		MessageType:  req.MessageType,
		Content:      content,
	}

	if err = s.msgRepo.Create(ctx, msg); err != nil {
		return nil, errcode.ErrDBError.WithMessage("发送消息失败")
	}

	// 获取发送者信息
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, errcode.ErrDBError.WithMessage("查询用户失败")
	}

	return &dto.MessageResponse{
		ID:           msg.ID,
		RoomID:       msg.RoomID,
		SenderUserID: msg.SenderUserID,
		SenderName:   user.Username,
		MessageType:  msg.MessageType,
		Content:      msg.Content,
		CreatedAt:    msg.CreatedAt.Format("2006-01-02 15:04:05"),
	}, nil
}

// GetMessages 获取消息列表
func (s *ChatService) GetMessages(ctx context.Context, roomID, userID uint64, page, pageSize int) ([]dto.MessageResponse, int64, error) {
	// 检查用户是否在房间中
	inRoom, err := s.participantRepo.ExistsActive(ctx, roomID, userID)
	if err != nil {
		return nil, 0, errcode.ErrDBError.WithMessage("检查参与者状态失败")
	}
	if !inRoom {
		return nil, 0, errcode.ErrNotInRoom
	}

	if page == 0 {
		page = 1
	}
	if pageSize == 0 {
		pageSize = 50
	}

	messages, total, err := s.msgRepo.FindByRoom(ctx, roomID, page, pageSize)
	if err != nil {
		return nil, 0, errcode.ErrDBError.WithMessage("查询消息失败")
	}

	// 获取所有发送者ID
	userMap := make(map[uint64]string)
	for _, msg := range messages {
		if _, ok := userMap[msg.SenderUserID]; !ok {
			user, err := s.userRepo.FindByID(ctx, msg.SenderUserID)
			if err == nil {
				userMap[msg.SenderUserID] = user.Username
			}
		}
	}

	var responses []dto.MessageResponse
	for _, msg := range messages {
		responses = append(responses, dto.MessageResponse{
			ID:           msg.ID,
			RoomID:       msg.RoomID,
			SenderUserID: msg.SenderUserID,
			SenderName:   userMap[msg.SenderUserID],
			MessageType:  msg.MessageType,
			Content:      msg.Content,
			CreatedAt:    msg.CreatedAt.Format("2006-01-02 15:04:05"),
		})
	}

	return responses, total, nil
}
