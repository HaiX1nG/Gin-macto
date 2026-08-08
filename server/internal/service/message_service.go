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

// MessageService 频道消息服务
type MessageService struct {
	msgRepo      *repository.MessageRepository
	reactionRepo *repository.ReactionRepository
	channelRepo  *repository.ChannelRepository
	userRepo     *repository.UserRepository
	serverSvc    *ServerService
}

// NewMessageService 创建频道消息服务实例
func NewMessageService(
	msgRepo *repository.MessageRepository,
	reactionRepo *repository.ReactionRepository,
	channelRepo *repository.ChannelRepository,
	userRepo *repository.UserRepository,
	serverSvc *ServerService,
) *MessageService {
	return &MessageService{
		msgRepo:      msgRepo,
		reactionRepo: reactionRepo,
		channelRepo:  channelRepo,
		userRepo:     userRepo,
		serverSvc:    serverSvc,
	}
}

// SendMessage 发送频道消息
func (s *MessageService) SendMessage(ctx context.Context, channelID, userID uint64, req *dto.SendMessageRequest) (*dto.MessageResponse, error) {
	// 校验频道存在
	channel, err := s.channelRepo.FindByID(ctx, channelID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errcode.ErrChannelNotFound
		}
		return nil, errcode.ErrDBError.WithMessage("查询频道失败")
	}

	// 校验权限：查看频道 + 发送消息
	if err := s.serverSvc.CheckPermission(ctx, channel.ServerID, userID, model.PermViewChannel); err != nil {
		return nil, err
	}
	if err := s.serverSvc.CheckPermission(ctx, channel.ServerID, userID, model.PermSendMessages); err != nil {
		return nil, err
	}

	// XSS 过滤
	content := util.TrimAndEscape(req.Content)

	msg := &model.ChannelMessage{
		ChannelID:    channelID,
		SenderUserID: userID,
		Type:         req.Type,
		Content:      content,
		ReplyToID:    req.ReplyToID,
	}

	if err = s.msgRepo.Create(ctx, msg); err != nil {
		return nil, errcode.ErrDBError.WithMessage("发送消息失败")
	}

	return s.buildMessageResponse(ctx, msg)
}

// GetMessages 分页获取频道消息
func (s *MessageService) GetMessages(ctx context.Context, channelID, userID uint64, req *dto.MessageListRequest) ([]dto.MessageResponse, int64, error) {
	// 校验频道存在
	channel, err := s.channelRepo.FindByID(ctx, channelID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, 0, errcode.ErrChannelNotFound
		}
		return nil, 0, errcode.ErrDBError.WithMessage("查询频道失败")
	}

	// 校验权限：查看频道 + 读取历史
	if err := s.serverSvc.CheckPermission(ctx, channel.ServerID, userID, model.PermViewChannel); err != nil {
		return nil, 0, err
	}
	if err := s.serverSvc.CheckPermission(ctx, channel.ServerID, userID, model.PermReadHistory); err != nil {
		return nil, 0, err
	}

	if req.Page == 0 {
		req.Page = 1
	}
	if req.PageSize == 0 {
		req.PageSize = 50
	}

	messages, total, err := s.msgRepo.FindByChannel(ctx, channelID, req.Page, req.PageSize)
	if err != nil {
		return nil, 0, errcode.ErrDBError.WithMessage("查询消息失败")
	}

	// 批量收集发送者 ID
	userIDs := make([]uint64, 0, len(messages))
	for _, msg := range messages {
		userIDs = append(userIDs, msg.SenderUserID)
	}
	userMap, _ := s.userRepo.FindByIDs(ctx, userIDs)

	var responses []dto.MessageResponse
	for _, msg := range messages {
		resp := s.buildMessageResponseWithUserMap(ctx, &msg, userMap)
		responses = append(responses, resp)
	}

	return responses, total, nil
}

// UpdateMessage 编辑消息
func (s *MessageService) UpdateMessage(ctx context.Context, channelID, userID, messageID uint64, req *dto.UpdateMessageRequest) (*dto.MessageResponse, error) {
	msg, err := s.msgRepo.FindByID(ctx, messageID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errcode.ErrMessageNotFound
		}
		return nil, errcode.ErrDBError.WithMessage("查询消息失败")
	}

	if msg.ChannelID != channelID {
		return nil, errcode.ErrMessageNotFound
	}

	// 仅发送者可编辑，或拥有管理消息权限
	if msg.SenderUserID != userID {
		channel, _ := s.channelRepo.FindByID(ctx, channelID)
		if channel != nil {
			if err := s.serverSvc.CheckPermission(ctx, channel.ServerID, userID, model.PermManageMessages); err != nil {
				return nil, errcode.ErrForbidden.WithMessage("只能编辑自己的消息")
			}
		} else {
			return nil, errcode.ErrForbidden.WithMessage("只能编辑自己的消息")
		}
	}

	content := util.TrimAndEscape(req.Content)
	if err = s.msgRepo.UpdateContent(ctx, messageID, content); err != nil {
		return nil, errcode.ErrDBError.WithMessage("更新消息失败")
	}

	// 重新查询以获取更新后的消息
	msg, _ = s.msgRepo.FindByID(ctx, messageID)
	return s.buildMessageResponse(ctx, msg)
}

// DeleteMessage 删除消息
func (s *MessageService) DeleteMessage(ctx context.Context, channelID, userID, messageID uint64) error {
	msg, err := s.msgRepo.FindByID(ctx, messageID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errcode.ErrMessageNotFound
		}
		return errcode.ErrDBError.WithMessage("查询消息失败")
	}

	if msg.ChannelID != channelID {
		return errcode.ErrMessageNotFound
	}

	// 仅发送者可删除，或拥有管理消息权限
	if msg.SenderUserID != userID {
		channel, _ := s.channelRepo.FindByID(ctx, channelID)
		if channel != nil {
			if err := s.serverSvc.CheckPermission(ctx, channel.ServerID, userID, model.PermManageMessages); err != nil {
				return errcode.ErrForbidden.WithMessage("只能删除自己的消息")
			}
		} else {
			return errcode.ErrForbidden.WithMessage("只能删除自己的消息")
		}
	}

	if err = s.msgRepo.Delete(ctx, messageID); err != nil {
		return errcode.ErrDBError.WithMessage("删除消息失败")
	}

	return nil
}

// PinMessage 置顶/取消置顶消息
func (s *MessageService) PinMessage(ctx context.Context, channelID, userID, messageID uint64) error {
	msg, err := s.msgRepo.FindByID(ctx, messageID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errcode.ErrMessageNotFound
		}
		return errcode.ErrDBError.WithMessage("查询消息失败")
	}

	if msg.ChannelID != channelID {
		return errcode.ErrMessageNotFound
	}

	// 校验权限
	channel, _ := s.channelRepo.FindByID(ctx, channelID)
	if channel != nil {
		if err := s.serverSvc.CheckPermission(ctx, channel.ServerID, userID, model.PermManageMessages); err != nil {
			return err
		}
	}

	// 切换置顶状态
	newPinned := !msg.IsPinned
	if err = s.msgRepo.UpdatePinStatus(ctx, messageID, newPinned); err != nil {
		return errcode.ErrDBError.WithMessage("更新置顶状态失败")
	}

	return nil
}

// AddReaction 添加表情反应
func (s *MessageService) AddReaction(ctx context.Context, channelID, userID, messageID uint64, req *dto.ReactionRequest) error {
	msg, err := s.msgRepo.FindByID(ctx, messageID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errcode.ErrMessageNotFound
		}
		return errcode.ErrDBError.WithMessage("查询消息失败")
	}

	if msg.ChannelID != channelID {
		return errcode.ErrMessageNotFound
	}

	// 校验权限
	channel, _ := s.channelRepo.FindByID(ctx, channelID)
	if channel != nil {
		if err := s.serverSvc.CheckPermission(ctx, channel.ServerID, userID, model.PermAddReactions); err != nil {
			return err
		}
	}

	reaction := &model.MessageReaction{
		MessageID: messageID,
		UserID:    userID,
		Emoji:     req.Emoji,
	}
	if err = s.reactionRepo.Add(ctx, reaction); err != nil {
		return errcode.ErrDBError.WithMessage("添加反应失败")
	}

	return nil
}

// RemoveReaction 移除表情反应
func (s *MessageService) RemoveReaction(ctx context.Context, channelID, userID, messageID uint64, emoji string) error {
	msg, err := s.msgRepo.FindByID(ctx, messageID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errcode.ErrMessageNotFound
		}
		return errcode.ErrDBError.WithMessage("查询消息失败")
	}

	if msg.ChannelID != channelID {
		return errcode.ErrMessageNotFound
	}

	if err = s.reactionRepo.Remove(ctx, messageID, userID, emoji); err != nil {
		return errcode.ErrDBError.WithMessage("移除反应失败")
	}

	return nil
}

// SearchMessages 在频道中搜索消息
func (s *MessageService) SearchMessages(ctx context.Context, req *dto.SearchMessagesRequest, userID uint64) (*dto.SearchMessagesResponse, error) {
	if req.Page == 0 {
		req.Page = 1
	}
	if req.PageSize == 0 {
		req.PageSize = 50
	}

	// 如果指定了频道ID，校验权限
	if req.ChannelID != nil {
		channel, err := s.channelRepo.FindByID(ctx, *req.ChannelID)
		if err != nil {
			return nil, errcode.ErrChannelNotFound
		}
		if err := s.serverSvc.CheckPermission(ctx, channel.ServerID, userID, model.PermReadHistory); err != nil {
			return nil, err
		}
	}

	channelID := uint64(0)
	if req.ChannelID != nil {
		channelID = *req.ChannelID
	}

	messages, total, err := s.msgRepo.SearchByChannel(ctx, req.Query, channelID, req.Page, req.PageSize)
	if err != nil {
		return nil, errcode.ErrDBError.WithMessage("搜索消息失败")
	}

	// 批量收集发送者 ID
	userIDs := make([]uint64, 0, len(messages))
	for _, msg := range messages {
		userIDs = append(userIDs, msg.SenderUserID)
	}
	userMap, _ := s.userRepo.FindByIDs(ctx, userIDs)

	var responses []dto.MessageResponse
	for _, msg := range messages {
		resp := s.buildMessageResponseWithUserMap(ctx, &msg, userMap)
		responses = append(responses, resp)
	}

	return &dto.SearchMessagesResponse{
		Messages: responses,
		Total:    total,
	}, nil
}

// buildMessageResponse 构建消息响应（完整查询）
func (s *MessageService) buildMessageResponse(ctx context.Context, msg *model.ChannelMessage) (*dto.MessageResponse, error) {
	user, err := s.userRepo.FindByID(ctx, msg.SenderUserID)
	userName := ""
	avatarURL := ""
	if err == nil {
		userName = user.Username
		avatarURL = user.AvatarURL
	}

	// 查询反应
	reactions, _ := s.reactionRepo.ListByMessage(ctx, msg.ID)
	reactionMap := make(map[string][]uint64)
	for _, r := range reactions {
		reactionMap[r.Emoji] = append(reactionMap[r.Emoji], r.UserID)
	}
	var reactionResponses []dto.ReactionResponse
	for emoji, userIDs := range reactionMap {
		reactionResponses = append(reactionResponses, dto.ReactionResponse{
			Emoji:   emoji,
			Count:   len(userIDs),
			UserIDs: userIDs,
		})
	}

	return &dto.MessageResponse{
		ID:           msg.ID,
		ChannelID:    msg.ChannelID,
		SenderUserID: msg.SenderUserID,
		SenderName:   userName,
		SenderAvatar: avatarURL,
		Type:         msg.Type,
		Content:      msg.Content,
		ReplyToID:    msg.ReplyToID,
		EditedAt:     msg.EditedAt,
		IsPinned:     msg.IsPinned,
		Reactions:    reactionResponses,
		Attachments:  []dto.AttachmentResponse{},
		CreatedAt:    msg.CreatedAt.Format("2006-01-02 15:04:05"),
	}, nil
}

// buildMessageResponseWithUserMap 构建消息响应（使用预查询的 userMap，避免 N+1）
func (s *MessageService) buildMessageResponseWithUserMap(ctx context.Context, msg *model.ChannelMessage, userMap map[uint64]*model.User) dto.MessageResponse {
	userName := ""
	avatarURL := ""
	if user, exists := userMap[msg.SenderUserID]; exists {
		userName = user.Username
		avatarURL = user.AvatarURL
	}

	// 查询反应
	reactions, _ := s.reactionRepo.ListByMessage(ctx, msg.ID)
	reactionMap := make(map[string][]uint64)
	for _, r := range reactions {
		reactionMap[r.Emoji] = append(reactionMap[r.Emoji], r.UserID)
	}
	var reactionResponses []dto.ReactionResponse
	for emoji, userIDs := range reactionMap {
		reactionResponses = append(reactionResponses, dto.ReactionResponse{
			Emoji:   emoji,
			Count:   len(userIDs),
			UserIDs: userIDs,
		})
	}

	return dto.MessageResponse{
		ID:           msg.ID,
		ChannelID:    msg.ChannelID,
		SenderUserID: msg.SenderUserID,
		SenderName:   userName,
		SenderAvatar: avatarURL,
		Type:         msg.Type,
		Content:      msg.Content,
		ReplyToID:    msg.ReplyToID,
		EditedAt:     msg.EditedAt,
		IsPinned:     msg.IsPinned,
		Reactions:    reactionResponses,
		Attachments:  []dto.AttachmentResponse{},
		CreatedAt:    msg.CreatedAt.Format("2006-01-02 15:04:05"),
	}
}
