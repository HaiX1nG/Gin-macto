package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/yourorg/livemix/internal/dto"
	"github.com/yourorg/livemix/internal/service"
	"github.com/yourorg/livemix/pkg/errcode"
	"github.com/yourorg/livemix/pkg/response"
)

// MessageHandler 频道消息处理器
type MessageHandler struct {
	messageService *service.MessageService
}

// NewMessageHandler 创建频道消息处理器实例
func NewMessageHandler(messageService *service.MessageService) *MessageHandler {
	return &MessageHandler{messageService: messageService}
}

// SendMessage 发送消息
// POST /api/v1/channels/:cid/messages
func (h *MessageHandler) SendMessage(c *gin.Context) {
	userID := c.GetUint64("userID")

	channelIDStr := c.Param("cid")
	channelID, err := strconv.ParseUint(channelIDStr, 10, 64)
	if err != nil {
		response.Fail(c, errcode.ErrInvalidParam)
		return
	}

	var req dto.SendMessageRequest
	if err = c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(c, nil, "参数校验失败: "+err.Error())
		return
	}

	resp, err := h.messageService.SendMessage(c.Request.Context(), channelID, userID, &req)
	if err != nil {
		response.Fail(c, err)
		return
	}

	response.Success(c, resp)
}

// GetMessages 获取消息列表
// GET /api/v1/channels/:cid/messages
func (h *MessageHandler) GetMessages(c *gin.Context) {
	userID := c.GetUint64("userID")

	channelIDStr := c.Param("cid")
	channelID, err := strconv.ParseUint(channelIDStr, 10, 64)
	if err != nil {
		response.Fail(c, errcode.ErrInvalidParam)
		return
	}

	var req dto.MessageListRequest
	if err = c.ShouldBindQuery(&req); err != nil {
		response.FailWithMessage(c, nil, "参数校验失败: "+err.Error())
		return
	}

	resp, total, err := h.messageService.GetMessages(c.Request.Context(), channelID, userID, &req)
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

// UpdateMessage 编辑消息
// PUT /api/v1/channels/:cid/messages/:mid
func (h *MessageHandler) UpdateMessage(c *gin.Context) {
	userID := c.GetUint64("userID")

	channelIDStr := c.Param("cid")
	channelID, err := strconv.ParseUint(channelIDStr, 10, 64)
	if err != nil {
		response.Fail(c, errcode.ErrInvalidParam)
		return
	}

	messageIDStr := c.Param("mid")
	messageID, err := strconv.ParseUint(messageIDStr, 10, 64)
	if err != nil {
		response.Fail(c, errcode.ErrInvalidParam)
		return
	}

	var req dto.UpdateMessageRequest
	if err = c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(c, nil, "参数校验失败: "+err.Error())
		return
	}

	resp, err := h.messageService.UpdateMessage(c.Request.Context(), channelID, userID, messageID, &req)
	if err != nil {
		response.Fail(c, err)
		return
	}

	response.Success(c, resp)
}

// DeleteMessage 删除消息
// DELETE /api/v1/channels/:cid/messages/:mid
func (h *MessageHandler) DeleteMessage(c *gin.Context) {
	userID := c.GetUint64("userID")

	channelIDStr := c.Param("cid")
	channelID, err := strconv.ParseUint(channelIDStr, 10, 64)
	if err != nil {
		response.Fail(c, errcode.ErrInvalidParam)
		return
	}

	messageIDStr := c.Param("mid")
	messageID, err := strconv.ParseUint(messageIDStr, 10, 64)
	if err != nil {
		response.Fail(c, errcode.ErrInvalidParam)
		return
	}

	if err = h.messageService.DeleteMessage(c.Request.Context(), channelID, userID, messageID); err != nil {
		response.Fail(c, err)
		return
	}

	response.Success(c, nil)
}

// PinMessage 置顶/取消置顶消息
// POST /api/v1/channels/:cid/messages/:mid/pin
func (h *MessageHandler) PinMessage(c *gin.Context) {
	userID := c.GetUint64("userID")

	channelIDStr := c.Param("cid")
	channelID, err := strconv.ParseUint(channelIDStr, 10, 64)
	if err != nil {
		response.Fail(c, errcode.ErrInvalidParam)
		return
	}

	messageIDStr := c.Param("mid")
	messageID, err := strconv.ParseUint(messageIDStr, 10, 64)
	if err != nil {
		response.Fail(c, errcode.ErrInvalidParam)
		return
	}

	if err = h.messageService.PinMessage(c.Request.Context(), channelID, userID, messageID); err != nil {
		response.Fail(c, err)
		return
	}

	response.Success(c, nil)
}

// AddReaction 添加表情反应
// POST /api/v1/channels/:cid/messages/:mid/reactions
func (h *MessageHandler) AddReaction(c *gin.Context) {
	userID := c.GetUint64("userID")

	channelIDStr := c.Param("cid")
	channelID, err := strconv.ParseUint(channelIDStr, 10, 64)
	if err != nil {
		response.Fail(c, errcode.ErrInvalidParam)
		return
	}

	messageIDStr := c.Param("mid")
	messageID, err := strconv.ParseUint(messageIDStr, 10, 64)
	if err != nil {
		response.Fail(c, errcode.ErrInvalidParam)
		return
	}

	var req dto.ReactionRequest
	if err = c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(c, nil, "参数校验失败: "+err.Error())
		return
	}

	if err = h.messageService.AddReaction(c.Request.Context(), channelID, userID, messageID, &req); err != nil {
		response.Fail(c, err)
		return
	}

	response.Success(c, nil)
}

// RemoveReaction 移除表情反应
// DELETE /api/v1/channels/:cid/messages/:mid/reactions/:emoji
func (h *MessageHandler) RemoveReaction(c *gin.Context) {
	userID := c.GetUint64("userID")

	channelIDStr := c.Param("cid")
	channelID, err := strconv.ParseUint(channelIDStr, 10, 64)
	if err != nil {
		response.Fail(c, errcode.ErrInvalidParam)
		return
	}

	messageIDStr := c.Param("mid")
	messageID, err := strconv.ParseUint(messageIDStr, 10, 64)
	if err != nil {
		response.Fail(c, errcode.ErrInvalidParam)
		return
	}

	emoji := c.Param("emoji")

	if err = h.messageService.RemoveReaction(c.Request.Context(), channelID, userID, messageID, emoji); err != nil {
		response.Fail(c, err)
		return
	}

	response.Success(c, nil)
}

// SearchMessages 搜索消息
// GET /api/v1/messages/search
func (h *MessageHandler) SearchMessages(c *gin.Context) {
	userID := c.GetUint64("userID")

	var req dto.SearchMessagesRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.FailWithMessage(c, nil, "参数校验失败: "+err.Error())
		return
	}

	resp, err := h.messageService.SearchMessages(c.Request.Context(), &req, userID)
	if err != nil {
		response.Fail(c, err)
		return
	}

	response.Success(c, resp)
}
