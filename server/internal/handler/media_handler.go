package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/yourorg/livemix/internal/dto"
	"github.com/yourorg/livemix/internal/service"
	"github.com/yourorg/livemix/pkg/errcode"
	"github.com/yourorg/livemix/pkg/response"
)

// PlaylistHandler 播放列表处理器
type PlaylistHandler struct {
	playlistService *service.PlaylistService
}

// NewPlaylistHandler 创建播放列表处理器实例
func NewPlaylistHandler(playlistService *service.PlaylistService) *PlaylistHandler {
	return &PlaylistHandler{playlistService: playlistService}
}

// AddItem 添加播放项
func (h *PlaylistHandler) AddItem(c *gin.Context) {
	roomIDStr := c.Param("id")
	roomID, err := strconv.ParseUint(roomIDStr, 10, 64)
	if err != nil {
		response.Fail(c, errcode.ErrInvalidParam)
		return
	}

	userID := c.GetUint64("userID")

	var req dto.AddPlaylistItemRequest
	if err = c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(c, errcode.ErrInvalidParam, "参数校验失败: "+err.Error())
		return
	}

	resp, err := h.playlistService.AddItem(c.Request.Context(), roomID, userID, &req)
	if err != nil {
		response.Fail(c, err)
		return
	}

	response.Success(c, resp)
}

// RemoveItem 删除播放项
func (h *PlaylistHandler) RemoveItem(c *gin.Context) {
	roomIDStr := c.Param("id")
	roomID, err := strconv.ParseUint(roomIDStr, 10, 64)
	if err != nil {
		response.Fail(c, errcode.ErrInvalidParam)
		return
	}

	itemIDStr := c.Param("itemId")
	itemID, err := strconv.ParseUint(itemIDStr, 10, 64)
	if err != nil {
		response.Fail(c, errcode.ErrInvalidParam)
		return
	}

	userID := c.GetUint64("userID")

	if err = h.playlistService.RemoveItem(c.Request.Context(), roomID, userID, itemID); err != nil {
		response.Fail(c, err)
		return
	}

	response.Success(c, nil)
}

// GetPlaylist 获取播放列表
func (h *PlaylistHandler) GetPlaylist(c *gin.Context) {
	roomIDStr := c.Param("id")
	roomID, err := strconv.ParseUint(roomIDStr, 10, 64)
	if err != nil {
		response.Fail(c, errcode.ErrInvalidParam)
		return
	}

	userID := c.GetUint64("userID")

	resp, err := h.playlistService.GetPlaylist(c.Request.Context(), roomID, userID)
	if err != nil {
		response.Fail(c, err)
		return
	}

	response.Success(c, resp)
}

// Play 播放
func (h *PlaylistHandler) Play(c *gin.Context) {
	roomIDStr := c.Param("id")
	roomID, err := strconv.ParseUint(roomIDStr, 10, 64)
	if err != nil {
		response.Fail(c, errcode.ErrInvalidParam)
		return
	}

	userID := c.GetUint64("userID")

	resp, err := h.playlistService.Play(c.Request.Context(), roomID, userID)
	if err != nil {
		response.Fail(c, err)
		return
	}

	response.Success(c, resp)
}

// Pause 暂停
func (h *PlaylistHandler) Pause(c *gin.Context) {
	roomIDStr := c.Param("id")
	roomID, err := strconv.ParseUint(roomIDStr, 10, 64)
	if err != nil {
		response.Fail(c, errcode.ErrInvalidParam)
		return
	}

	userID := c.GetUint64("userID")

	if err = h.playlistService.Pause(c.Request.Context(), roomID, userID); err != nil {
		response.Fail(c, err)
		return
	}

	response.Success(c, nil)
}

// Skip 跳过
func (h *PlaylistHandler) Skip(c *gin.Context) {
	roomIDStr := c.Param("id")
	roomID, err := strconv.ParseUint(roomIDStr, 10, 64)
	if err != nil {
		response.Fail(c, errcode.ErrInvalidParam)
		return
	}

	userID := c.GetUint64("userID")

	resp, err := h.playlistService.Skip(c.Request.Context(), roomID, userID)
	if err != nil {
		response.Fail(c, err)
		return
	}

	response.Success(c, resp)
}

// ChatHandler 聊天处理器
type ChatHandler struct {
	chatService *service.ChatService
}

// NewChatHandler 创建聊天处理器实例
func NewChatHandler(chatService *service.ChatService) *ChatHandler {
	return &ChatHandler{chatService: chatService}
}

// SendMessage 发送消息
func (h *ChatHandler) SendMessage(c *gin.Context) {
	roomIDStr := c.Param("id")
	roomID, err := strconv.ParseUint(roomIDStr, 10, 64)
	if err != nil {
		response.Fail(c, errcode.ErrInvalidParam)
		return
	}

	userID := c.GetUint64("userID")

	var req dto.SendMessageRequest
	if err = c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(c, errcode.ErrInvalidParam, "参数校验失败: "+err.Error())
		return
	}

	resp, err := h.chatService.SendMessage(c.Request.Context(), roomID, userID, &req)
	if err != nil {
		response.Fail(c, err)
		return
	}

	response.Success(c, resp)
}

// GetMessages 获取消息列表
func (h *ChatHandler) GetMessages(c *gin.Context) {
	roomIDStr := c.Param("id")
	roomID, err := strconv.ParseUint(roomIDStr, 10, 64)
	if err != nil {
		response.Fail(c, errcode.ErrInvalidParam)
		return
	}

	userID := c.GetUint64("userID")

	var req dto.MessageListRequest
	if err = c.ShouldBindQuery(&req); err != nil {
		response.FailWithMessage(c, errcode.ErrInvalidParam, "参数校验失败: "+err.Error())
		return
	}

	resp, total, err := h.chatService.GetMessages(c.Request.Context(), roomID, userID, &req)
	if err != nil {
		response.Fail(c, err)
		return
	}

	response.Page(c, resp, total, req.Page, req.PageSize)
}

// GetUserHistoryMessages 获取用户历史消息
func (h *ChatHandler) GetUserHistoryMessages(c *gin.Context) {
	userID := c.GetUint64("userID")

	var req dto.UserHistoryMessagesRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.FailWithMessage(c, errcode.ErrInvalidParam, "参数校验失败: "+err.Error())
		return
	}

	resp, total, err := h.chatService.GetUserHistoryMessages(c.Request.Context(), userID, req.Page, req.PageSize)
	if err != nil {
		response.Fail(c, err)
		return
	}

	if req.Page == 0 {
		req.Page = 1
	}
	if req.PageSize == 0 {
		req.PageSize = 50
	}

	response.Page(c, resp, total, req.Page, req.PageSize)
}

// SearchMessages 搜索消息
func (h *ChatHandler) SearchMessages(c *gin.Context) {
	userID := c.GetUint64("userID")

	var req dto.SearchMessagesRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.FailWithMessage(c, errcode.ErrInvalidParam, "参数校验失败: "+err.Error())
		return
	}

	resp, err := h.chatService.SearchMessages(c.Request.Context(), &req, userID)
	if err != nil {
		response.Fail(c, err)
		return
	}

	response.Success(c, resp)
}
