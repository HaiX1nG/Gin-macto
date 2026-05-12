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

// FriendService 好友服务
type FriendService struct {
	friendRepo *repository.FriendRepository
	userRepo   *repository.UserRepository
	statusRepo *repository.UserStatusRepository
}

// NewFriendService 创建好友服务实例
func NewFriendService(
	friendRepo *repository.FriendRepository,
	userRepo *repository.UserRepository,
	statusRepo *repository.UserStatusRepository,
) *FriendService {
	return &FriendService{
		friendRepo: friendRepo,
		userRepo:   userRepo,
		statusRepo: statusRepo,
	}
}

// SendFriendRequest 发送好友请求
func (s *FriendService) SendFriendRequest(ctx context.Context, senderID uint64, req *dto.SendFriendRequestRequest) (*dto.FriendRequestResponse, error) {
	// 检查接收者是否存在
	_, err := s.userRepo.FindByID(ctx, req.ReceiverID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errcode.ErrUserNotFound
		}
		return nil, errcode.ErrDBError.WithMessage("查询用户失败")
	}

	// 不能添加自己为好友
	if senderID == req.ReceiverID {
		return nil, errcode.ErrBadRequest.WithMessage("不能添加自己为好友")
	}

	// 检查是否已经是好友
	isFriend, err := s.friendRepo.IsFriend(ctx, senderID, req.ReceiverID)
	if err != nil {
		return nil, errcode.ErrDBError.WithMessage("检查好友关系失败")
	}
	if isFriend {
		return nil, errcode.ErrBadRequest.WithMessage("已经是好友")
	}

	// 检查是否已经有待处理的请求
	existingReq, err := s.friendRepo.FindAnyPendingRequestByUsers(ctx, senderID, req.ReceiverID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errcode.ErrDBError.WithMessage("检查请求失败")
	}
	if existingReq != nil {
		if existingReq.SenderID == senderID {
			return nil, errcode.ErrBadRequest.WithMessage("已经发送过请求，请等待对方处理")
		}
		return nil, errcode.ErrBadRequest.WithMessage("对方已经发送过请求，请先处理")
	}

	// 创建好友请求
	friendReq := &model.FriendRequest{
		SenderID:   senderID,
		ReceiverID: req.ReceiverID,
		Status:     model.FriendRequestPending,
		Message:    req.Message,
	}

	if err = s.friendRepo.CreateFriendRequest(ctx, friendReq); err != nil {
		return nil, errcode.ErrDBError.WithMessage("创建好友请求失败")
	}

	// 获取发送者信息
	sender, err := s.userRepo.FindByID(ctx, senderID)
	if err != nil {
		return nil, errcode.ErrDBError.WithMessage("查询发送者失败")
	}

	return &dto.FriendRequestResponse{
		ID:         friendReq.ID,
		SenderID:   friendReq.SenderID,
		SenderName: sender.Username,
		ReceiverID: friendReq.ReceiverID,
		Status:     int(friendReq.Status),
		Message:    friendReq.Message,
		CreatedAt:  friendReq.CreatedAt.Format("2006-01-02 15:04:05"),
	}, nil
}

// HandleFriendRequest 处理好友请求
func (s *FriendService) HandleFriendRequest(ctx context.Context, userID, requestID uint64, accept bool) error {
	// 查询请求
	friendReq, err := s.friendRepo.FindFriendRequestByID(ctx, requestID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errcode.ErrNotFound.WithMessage("好友请求不存在")
		}
		return errcode.ErrDBError.WithMessage("查询请求失败")
	}

	// 验证请求状态
	if friendReq.Status != model.FriendRequestPending {
		return errcode.ErrBadRequest.WithMessage("请求已经处理过")
	}

	// 验证接收者
	if friendReq.ReceiverID != userID {
		return errcode.ErrForbidden.WithMessage("无权处理此请求")
	}

	// 更新请求状态
	if accept {
		friendReq.Status = model.FriendRequestAccepted
	} else {
		friendReq.Status = model.FriendRequestRejected
	}

	if err = s.friendRepo.UpdateFriendRequest(ctx, friendReq); err != nil {
		return errcode.ErrDBError.WithMessage("更新请求失败")
	}

	// 如果接受，创建好友关系
	if accept {
		if err = s.friendRepo.CreateFriendship(ctx, friendReq.SenderID, friendReq.ReceiverID); err != nil {
			return errcode.ErrDBError.WithMessage("创建好友关系失败")
		}
	}

	return nil
}

// GetFriendList 获取好友列表
func (s *FriendService) GetFriendList(ctx context.Context, userID uint64, page, pageSize int) (*dto.FriendListResponse, error) {
	friendships, total, err := s.friendRepo.FindFriendsByUserID(ctx, userID, page, pageSize)
	if err != nil {
		return nil, errcode.ErrDBError.WithMessage("查询好友列表失败")
	}

	var friends []dto.FriendResponse
	for _, f := range friendships {
		friend, err := s.userRepo.FindByID(ctx, f.FriendID)
		if err != nil {
			continue
		}

		status, err := s.statusRepo.FindByUserID(ctx, f.FriendID)
		isOnline := false
		customStatus := ""
		if err == nil {
			isOnline = status.IsOnline
			customStatus = status.CustomStatus
		}

		friends = append(friends, dto.FriendResponse{
			FriendID:     friend.ID,
			FriendName:   friend.Username,
			AvatarURL:    friend.AvatarURL,
			IsOnline:     isOnline,
			CustomStatus: customStatus,
			CreatedAt:    f.CreatedAt.Format("2006-01-02 15:04:05"),
		})
	}

	return &dto.FriendListResponse{
		Friends: friends,
		Total:   total,
	}, nil
}

// GetPendingRequests 获取待处理的好友请求
func (s *FriendService) GetPendingRequests(ctx context.Context, userID uint64, page, pageSize int) (*dto.PendingRequestsResponse, error) {
	requests, total, err := s.friendRepo.FindPendingRequestsByReceiver(ctx, userID, page, pageSize)
	if err != nil {
		return nil, errcode.ErrDBError.WithMessage("查询好友请求失败")
	}

	var reqResponses []dto.FriendRequestResponse
	for _, req := range requests {
		sender, err := s.userRepo.FindByID(ctx, req.SenderID)
		if err != nil {
			continue
		}

		reqResponses = append(reqResponses, dto.FriendRequestResponse{
			ID:         req.ID,
			SenderID:   req.SenderID,
			SenderName: sender.Username,
			ReceiverID: req.ReceiverID,
			Status:     int(req.Status),
			Message:    req.Message,
			CreatedAt:  req.CreatedAt.Format("2006-01-02 15:04:05"),
		})
	}

	return &dto.PendingRequestsResponse{
		Requests: reqResponses,
		Total:    total,
	}, nil
}

// DeleteFriend 删除好友
func (s *FriendService) DeleteFriend(ctx context.Context, userID, friendID uint64) error {
	// 检查是否为好友
	isFriend, err := s.friendRepo.IsFriend(ctx, userID, friendID)
	if err != nil {
		return errcode.ErrDBError.WithMessage("检查好友关系失败")
	}
	if !isFriend {
		return errcode.ErrBadRequest.WithMessage("不是好友关系")
	}

	// 删除好友关系
	if err = s.friendRepo.DeleteFriendship(ctx, userID, friendID); err != nil {
		return errcode.ErrDBError.WithMessage("删除好友失败")
	}

	return nil
}

// SendPrivateMessage 发送私聊消息
func (s *FriendService) SendPrivateMessage(ctx context.Context, senderID uint64, req *dto.SendPrivateMessageRequest) (*dto.PrivateMessageResponse, error) {
	// 检查接收者是否存在
	_, err := s.userRepo.FindByID(ctx, req.ReceiverID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errcode.ErrUserNotFound
		}
		return nil, errcode.ErrDBError.WithMessage("查询用户失败")
	}

	// 不能发送消息给自己
	if senderID == req.ReceiverID {
		return nil, errcode.ErrBadRequest.WithMessage("不能发送消息给自己")
	}

	// 检查是否为好友（可选：也可以允许非好友发送消息）
	isFriend, err := s.friendRepo.IsFriend(ctx, senderID, req.ReceiverID)
	if err != nil {
		return nil, errcode.ErrDBError.WithMessage("检查好友关系失败")
	}
	if !isFriend {
		return nil, errcode.ErrForbidden.WithMessage("只能给好友发送消息")
	}

	// 创建消息
	msg := &model.PrivateMessage{
		SenderID:   senderID,
		ReceiverID: req.ReceiverID,
		Content:    req.Content,
		IsRead:     false,
	}

	if err = s.friendRepo.CreatePrivateMessage(ctx, msg); err != nil {
		return nil, errcode.ErrDBError.WithMessage("发送消息失败")
	}

	// 获取发送者信息
	sender, err := s.userRepo.FindByID(ctx, senderID)
	if err != nil {
		return nil, errcode.ErrDBError.WithMessage("查询发送者失败")
	}

	return &dto.PrivateMessageResponse{
		ID:         msg.ID,
		SenderID:   msg.SenderID,
		SenderName: sender.Username,
		ReceiverID: msg.ReceiverID,
		Content:    msg.Content,
		IsRead:     msg.IsRead,
		CreatedAt:  msg.CreatedAt.Format("2006-01-02 15:04:05"),
	}, nil
}

// GetPrivateMessages 获取私聊消息
func (s *FriendService) GetPrivateMessages(ctx context.Context, userID, friendID uint64, page, pageSize int) (*dto.PrivateMessageListResponse, error) {
	// 检查是否为好友
	isFriend, err := s.friendRepo.IsFriend(ctx, userID, friendID)
	if err != nil {
		return nil, errcode.ErrDBError.WithMessage("检查好友关系失败")
	}
	if !isFriend {
		return nil, errcode.ErrForbidden.WithMessage("只能查看好友的消息")
	}

	messages, total, err := s.friendRepo.FindPrivateMessages(ctx, userID, friendID, page, pageSize)
	if err != nil {
		return nil, errcode.ErrDBError.WithMessage("查询消息失败")
	}

	// 标记消息为已读
	if err = s.friendRepo.MarkMessagesAsRead(ctx, friendID, userID); err != nil {
		// 记录错误但不影响返回结果
	}

	var msgResponses []dto.PrivateMessageResponse
	for _, msg := range messages {
		sender, err := s.userRepo.FindByID(ctx, msg.SenderID)
		if err != nil {
			continue
		}

		msgResponses = append(msgResponses, dto.PrivateMessageResponse{
			ID:         msg.ID,
			SenderID:   msg.SenderID,
			SenderName: sender.Username,
			ReceiverID: msg.ReceiverID,
			Content:    msg.Content,
			IsRead:     msg.IsRead,
			CreatedAt:  msg.CreatedAt.Format("2006-01-02 15:04:05"),
		})
	}

	return &dto.PrivateMessageListResponse{
		Messages: msgResponses,
		Total:    total,
	}, nil
}

// GetConversations 获取会话列表
func (s *FriendService) GetConversations(ctx context.Context, userID uint64) (*dto.ConversationListResponse, error) {
	userIDs, err := s.friendRepo.FindConversations(ctx, userID)
	if err != nil {
		return nil, errcode.ErrDBError.WithMessage("查询会话失败")
	}

	var conversations []dto.ConversationResponse
	for _, otherUserID := range userIDs {
		user, err := s.userRepo.FindByID(ctx, otherUserID)
		if err != nil {
			continue
		}

		status, err := s.statusRepo.FindByUserID(ctx, otherUserID)
		isOnline := false
		customStatus := ""
		if err == nil {
			isOnline = status.IsOnline
			customStatus = status.CustomStatus
		}

		lastMsg, err := s.friendRepo.GetLastMessage(ctx, userID, otherUserID)
		lastMessage := ""
		if err == nil && lastMsg != nil {
			lastMessage = lastMsg.Content
		}

		unreadCount, _ := s.friendRepo.GetUnreadCountFromUser(ctx, otherUserID, userID)

		conversations = append(conversations, dto.ConversationResponse{
			UserID:       user.ID,
			Username:     user.Username,
			AvatarURL:    user.AvatarURL,
			IsOnline:     isOnline,
			CustomStatus: customStatus,
			LastMessage:  lastMessage,
			UnreadCount:  unreadCount,
		})
	}

	return &dto.ConversationListResponse{
		Conversations: conversations,
		Total:         int64(len(conversations)),
	}, nil
}

// SearchUser 搜索用户
func (s *FriendService) SearchUser(ctx context.Context, userID uint64, username string) (*dto.SearchUserResponse, error) {
	user, err := s.userRepo.FindByUsername(ctx, username)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errcode.ErrUserNotFound
		}
		return nil, errcode.ErrDBError.WithMessage("查询用户失败")
	}

	// 检查是否为好友
	isFriend, _ := s.friendRepo.IsFriend(ctx, userID, user.ID)

	// 检查是否有待处理的请求
	hasPendingRequest := false
	if !isFriend {
		_, err := s.friendRepo.FindAnyPendingRequestByUsers(ctx, userID, user.ID)
		if err == nil {
			hasPendingRequest = true
		}
	}

	status, err := s.statusRepo.FindByUserID(ctx, user.ID)
	isOnline := false
	customStatus := ""
	if err == nil {
		isOnline = status.IsOnline
		customStatus = status.CustomStatus
	}

	return &dto.SearchUserResponse{
		UserID:            user.ID,
		Username:          user.Username,
		AvatarURL:         user.AvatarURL,
		IsOnline:          isOnline,
		CustomStatus:      customStatus,
		IsFriend:          isFriend,
		HasPendingRequest: hasPendingRequest,
	}, nil
}

// GetUnreadCount 获取总未读消息数
func (s *FriendService) GetUnreadCount(ctx context.Context, userID uint64) (int64, error) {
	return s.friendRepo.GetUnreadCount(ctx, userID)
}
