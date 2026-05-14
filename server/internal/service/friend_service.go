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

// FriendService 好友服务
type FriendService struct {
	friendRepo *repository.FriendRepository
	userRepo   *repository.UserRepository
	statusRepo *repository.UserStatusRepository
	txManager  *repository.GormTransactionManager
}

// NewFriendService 创建好友服务实例
func NewFriendService(
	friendRepo *repository.FriendRepository,
	userRepo *repository.UserRepository,
	statusRepo *repository.UserStatusRepository,
	txManager *repository.GormTransactionManager,
) *FriendService {
	return &FriendService{
		friendRepo: friendRepo,
		userRepo:   userRepo,
		statusRepo: statusRepo,
		txManager:  txManager,
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
	// XSS过滤：对请求附言进行HTML转义，防止存储型XSS攻击
	friendReq := &model.FriendRequest{
		SenderID:   senderID,
		ReceiverID: req.ReceiverID,
		Status:     model.FriendRequestPending,
		Message:    util.TrimAndEscape(req.Message),
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
// 使用事务确保更新请求状态和创建好友关系的原子性
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

	// 如果拒绝请求，直接更新状态即可（单表操作，无需事务）
	if !accept {
		if err = s.friendRepo.UpdateFriendRequest(ctx, friendReq); err != nil {
			return errcode.ErrDBError.WithMessage("更新请求失败")
		}
		return nil
	}

	// 接受请求时，使用事务确保更新请求状态和创建好友关系的原子性
	// 避免请求状态更新成功但好友关系创建失败的数据不一致问题
	err = s.txManager.Transactional(ctx, func(tx repository.TransactionContext) error {
		// 更新请求状态
		if updateErr := s.friendRepo.UpdateFriendRequestWithDB(tx.DB(), friendReq); updateErr != nil {
			return updateErr
		}

		// 创建好友关系（双向）
		if createErr := s.friendRepo.CreateFriendshipWithDB(tx.DB(), friendReq.SenderID, friendReq.ReceiverID); createErr != nil {
			return createErr
		}

		return nil
	})

	if err != nil {
		return errcode.ErrDBError.WithMessage("处理好友请求失败")
	}

	return nil
}

// GetFriendList 获取好友列表
// 优化：使用批量查询替代循环查询，解决 N+1 问题
func (s *FriendService) GetFriendList(ctx context.Context, userID uint64, page, pageSize int) (*dto.FriendListResponse, error) {
	friendships, total, err := s.friendRepo.FindFriendsByUserID(ctx, userID, page, pageSize)
	if err != nil {
		return nil, errcode.ErrDBError.WithMessage("查询好友列表失败")
	}

	// 如果没有好友，直接返回空列表
	if len(friendships) == 0 {
		return &dto.FriendListResponse{
			Friends: []dto.FriendResponse{},
			Total:   total,
		}, nil
	}

	// 收集所有好友 ID
	friendIDs := make([]uint64, 0, len(friendships))
	for _, f := range friendships {
		friendIDs = append(friendIDs, f.FriendID)
	}

	// 批量查询好友用户信息（解决 N+1 问题）
	userMap, err := s.userRepo.FindByIDs(ctx, friendIDs)
	if err != nil {
		return nil, errcode.ErrDBError.WithMessage("查询好友信息失败")
	}

	// 批量查询好友在线状态（解决 N+1 问题）
	statusMap, err := s.statusRepo.FindByUserIDsMap(ctx, friendIDs)
	if err != nil {
		return nil, errcode.ErrDBError.WithMessage("查询好友状态失败")
	}

	// 组装响应数据
	var friends []dto.FriendResponse
	for _, f := range friendships {
		friend, exists := userMap[f.FriendID]
		if !exists {
			continue
		}

		// 获取在线状态，默认为离线
		isOnline := false
		customStatus := ""
		if status, ok := statusMap[f.FriendID]; ok {
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
// 优化：使用批量查询替代循环查询，解决 N+1 问题
func (s *FriendService) GetPendingRequests(ctx context.Context, userID uint64, page, pageSize int) (*dto.PendingRequestsResponse, error) {
	requests, total, err := s.friendRepo.FindPendingRequestsByReceiver(ctx, userID, page, pageSize)
	if err != nil {
		return nil, errcode.ErrDBError.WithMessage("查询好友请求失败")
	}

	// 如果没有请求，直接返回空列表
	if len(requests) == 0 {
		return &dto.PendingRequestsResponse{
			Requests: []dto.FriendRequestResponse{},
			Total:    total,
		}, nil
	}

	// 收集所有发送者 ID
	senderIDs := make([]uint64, 0, len(requests))
	for _, req := range requests {
		senderIDs = append(senderIDs, req.SenderID)
	}

	// 批量查询发送者用户信息（解决 N+1 问题）
	userMap, err := s.userRepo.FindByIDs(ctx, senderIDs)
	if err != nil {
		return nil, errcode.ErrDBError.WithMessage("查询发送者信息失败")
	}

	// 组装响应数据
	var reqResponses []dto.FriendRequestResponse
	for _, req := range requests {
		sender, exists := userMap[req.SenderID]
		if !exists {
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
	// XSS过滤：对私聊消息内容进行HTML转义，防止存储型XSS攻击
	msg := &model.PrivateMessage{
		SenderID:   senderID,
		ReceiverID: req.ReceiverID,
		Content:    util.TrimAndEscape(req.Content),
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
// 优化：使用批量查询替代循环查询，解决 N+1 问题
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

	// 如果没有消息，直接返回空列表
	if len(messages) == 0 {
		return &dto.PrivateMessageListResponse{
			Messages: []dto.PrivateMessageResponse{},
			Total:    total,
		}, nil
	}

	// 收集所有发送者 ID
	senderIDs := make([]uint64, 0, len(messages))
	for _, msg := range messages {
		senderIDs = append(senderIDs, msg.SenderID)
	}

	// 批量查询发送者用户信息（解决 N+1 问题）
	userMap, err := s.userRepo.FindByIDs(ctx, senderIDs)
	if err != nil {
		return nil, errcode.ErrDBError.WithMessage("查询发送者信息失败")
	}

	// 组装响应数据
	var msgResponses []dto.PrivateMessageResponse
	for _, msg := range messages {
		sender, exists := userMap[msg.SenderID]
		if !exists {
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
// 优化：使用批量查询替代循环查询，解决 N+1 问题
func (s *FriendService) GetConversations(ctx context.Context, userID uint64) (*dto.ConversationListResponse, error) {
	userIDs, err := s.friendRepo.FindConversations(ctx, userID)
	if err != nil {
		return nil, errcode.ErrDBError.WithMessage("查询会话失败")
	}

	// 如果没有会话，直接返回空列表
	if len(userIDs) == 0 {
		return &dto.ConversationListResponse{
			Conversations: []dto.ConversationResponse{},
			Total:         0,
		}, nil
	}

	// 批量查询用户信息（解决 N+1 问题）
	userMap, err := s.userRepo.FindByIDs(ctx, userIDs)
	if err != nil {
		return nil, errcode.ErrDBError.WithMessage("查询用户信息失败")
	}

	// 批量查询用户在线状态（解决 N+1 问题）
	statusMap, err := s.statusRepo.FindByUserIDsMap(ctx, userIDs)
	if err != nil {
		return nil, errcode.ErrDBError.WithMessage("查询用户状态失败")
	}

	// 批量获取最后一条消息（解决 N+1 问题）
	lastMsgMap, err := s.friendRepo.GetLastMessagesBatch(ctx, userID, userIDs)
	if err != nil {
		return nil, errcode.ErrDBError.WithMessage("查询最后消息失败")
	}

	// 批量获取未读消息数（解决 N+1 问题）
	unreadCountMap, err := s.friendRepo.GetUnreadCountBatch(ctx, userID, userIDs)
	if err != nil {
		return nil, errcode.ErrDBError.WithMessage("查询未读数失败")
	}

	// 组装响应数据
	var conversations []dto.ConversationResponse
	for _, otherUserID := range userIDs {
		user, exists := userMap[otherUserID]
		if !exists {
			continue
		}

		// 获取在线状态，默认为离线
		isOnline := false
		customStatus := ""
		if status, ok := statusMap[otherUserID]; ok {
			isOnline = status.IsOnline
			customStatus = status.CustomStatus
		}

		// 获取最后一条消息内容
		lastMessage := ""
		if lastMsg, ok := lastMsgMap[otherUserID]; ok && lastMsg != nil {
			lastMessage = lastMsg.Content
		}

		// 获取未读消息数，默认为 0
		unreadCount := int64(0)
		if count, ok := unreadCountMap[otherUserID]; ok {
			unreadCount = count
		}

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
