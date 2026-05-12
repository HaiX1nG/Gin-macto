package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/yourorg/livemix/internal/dto"
	"github.com/yourorg/livemix/internal/service"
	"github.com/yourorg/livemix/pkg/response"
)

// FriendHandler 好友处理器
type FriendHandler struct {
	friendService *service.FriendService
}

// NewFriendHandler 创建好友处理器实例
func NewFriendHandler(friendService *service.FriendService) *FriendHandler {
	return &FriendHandler{friendService: friendService}
}

// SendFriendRequest 发送好友请求
// POST /api/v1/friends/request
func (h *FriendHandler) SendFriendRequest(c *gin.Context) {
	userID := c.GetUint64("userID")

	var req dto.SendFriendRequestRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(c, nil, "参数校验失败: "+err.Error())
		return
	}

	resp, err := h.friendService.SendFriendRequest(c.Request.Context(), userID, &req)
	if err != nil {
		response.Fail(c, err)
		return
	}

	response.Success(c, resp)
}

// HandleFriendRequest 处理好友请求（接受/拒绝）
// POST /api/v1/friends/request/:id/handle
func (h *FriendHandler) HandleFriendRequest(c *gin.Context) {
	userID := c.GetUint64("userID")

	requestIDStr := c.Param("id")
	requestID, err := strconv.ParseUint(requestIDStr, 10, 64)
	if err != nil {
		response.FailWithMessage(c, nil, "无效的请求ID")
		return
	}

	var req dto.HandleFriendRequestRequest
	if err = c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(c, nil, "参数校验失败: "+err.Error())
		return
	}

	if err = h.friendService.HandleFriendRequest(c.Request.Context(), userID, requestID, req.Accept); err != nil {
		response.Fail(c, err)
		return
	}

	response.Success(c, nil)
}

// GetFriendList 获取好友列表
// GET /api/v1/friends
func (h *FriendHandler) GetFriendList(c *gin.Context) {
	userID := c.GetUint64("userID")

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	resp, err := h.friendService.GetFriendList(c.Request.Context(), userID, page, pageSize)
	if err != nil {
		response.Fail(c, err)
		return
	}

	response.Success(c, resp)
}

// GetPendingRequests 获取待处理的好友请求
// GET /api/v1/friends/requests
func (h *FriendHandler) GetPendingRequests(c *gin.Context) {
	userID := c.GetUint64("userID")

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	resp, err := h.friendService.GetPendingRequests(c.Request.Context(), userID, page, pageSize)
	if err != nil {
		response.Fail(c, err)
		return
	}

	response.Success(c, resp)
}

// DeleteFriend 删除好友
// DELETE /api/v1/friends/:id
func (h *FriendHandler) DeleteFriend(c *gin.Context) {
	userID := c.GetUint64("userID")

	friendIDStr := c.Param("id")
	friendID, err := strconv.ParseUint(friendIDStr, 10, 64)
	if err != nil {
		response.FailWithMessage(c, nil, "无效的好友ID")
		return
	}

	if err = h.friendService.DeleteFriend(c.Request.Context(), userID, friendID); err != nil {
		response.Fail(c, err)
		return
	}

	response.Success(c, nil)
}

// SendPrivateMessage 发送私聊消息
// POST /api/v1/friends/messages
func (h *FriendHandler) SendPrivateMessage(c *gin.Context) {
	userID := c.GetUint64("userID")

	var req dto.SendPrivateMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(c, nil, "参数校验失败: "+err.Error())
		return
	}

	resp, err := h.friendService.SendPrivateMessage(c.Request.Context(), userID, &req)
	if err != nil {
		response.Fail(c, err)
		return
	}

	response.Success(c, resp)
}

// GetPrivateMessages 获取私聊消息
// GET /api/v1/friends/:id/messages
func (h *FriendHandler) GetPrivateMessages(c *gin.Context) {
	userID := c.GetUint64("userID")

	friendIDStr := c.Param("id")
	friendID, err := strconv.ParseUint(friendIDStr, 10, 64)
	if err != nil {
		response.FailWithMessage(c, nil, "无效的好友ID")
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "50"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 50
	}

	resp, err := h.friendService.GetPrivateMessages(c.Request.Context(), userID, friendID, page, pageSize)
	if err != nil {
		response.Fail(c, err)
		return
	}

	response.Success(c, resp)
}

// GetConversations 获取会话列表
// GET /api/v1/friends/conversations
func (h *FriendHandler) GetConversations(c *gin.Context) {
	userID := c.GetUint64("userID")

	resp, err := h.friendService.GetConversations(c.Request.Context(), userID)
	if err != nil {
		response.Fail(c, err)
		return
	}

	response.Success(c, resp)
}

// SearchUser 搜索用户
// GET /api/v1/friends/search
func (h *FriendHandler) SearchUser(c *gin.Context) {
	userID := c.GetUint64("userID")

	username := c.Query("username")
	if username == "" {
		response.FailWithMessage(c, nil, "用户名不能为空")
		return
	}

	resp, err := h.friendService.SearchUser(c.Request.Context(), userID, username)
	if err != nil {
		response.Fail(c, err)
		return
	}

	response.Success(c, resp)
}

// GetUnreadCount 获取未读消息数
// GET /api/v1/friends/messages/unread
func (h *FriendHandler) GetUnreadCount(c *gin.Context) {
	userID := c.GetUint64("userID")

	count, err := h.friendService.GetUnreadCount(c.Request.Context(), userID)
	if err != nil {
		response.Fail(c, err)
		return
	}

	response.Success(c, gin.H{
		"unreadCount": count,
	})
}
