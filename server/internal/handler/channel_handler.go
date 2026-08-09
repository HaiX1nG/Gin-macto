package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/yourorg/livemix/internal/dto"
	"github.com/yourorg/livemix/internal/service"
	"github.com/yourorg/livemix/pkg/errcode"
	"github.com/yourorg/livemix/pkg/response"
)

// ChannelHandler 频道处理器
type ChannelHandler struct {
	channelService *service.ChannelService
}

// NewChannelHandler 创建频道处理器实例
func NewChannelHandler(channelService *service.ChannelService) *ChannelHandler {
	return &ChannelHandler{channelService: channelService}
}

// CreateChannel 创建频道
// POST /api/v1/servers/:sid/channels
func (h *ChannelHandler) CreateChannel(c *gin.Context) {
	userID := c.GetUint64("userID")

	serverIDStr := c.Param("id")
	serverID, err := strconv.ParseUint(serverIDStr, 10, 64)
	if err != nil {
		response.Fail(c, errcode.ErrInvalidParam)
		return
	}

	var req dto.CreateChannelRequest
	if err = c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(c, nil, "参数校验失败: "+err.Error())
		return
	}

	resp, err := h.channelService.CreateChannel(c.Request.Context(), serverID, userID, &req)
	if err != nil {
		response.Fail(c, err)
		return
	}

	response.Success(c, resp)
}

// GetChannelTree 获取频道列表
// GET /api/v1/servers/:sid/channels
func (h *ChannelHandler) GetChannelTree(c *gin.Context) {
	userID := c.GetUint64("userID")

	serverIDStr := c.Param("id")
	serverID, err := strconv.ParseUint(serverIDStr, 10, 64)
	if err != nil {
		response.Fail(c, errcode.ErrInvalidParam)
		return
	}

	resp, err := h.channelService.GetChannelTree(c.Request.Context(), serverID, userID)
	if err != nil {
		response.Fail(c, err)
		return
	}

	response.Success(c, resp)
}

// UpdateChannel 更新频道
// PUT /api/v1/servers/:sid/channels/:cid
func (h *ChannelHandler) UpdateChannel(c *gin.Context) {
	userID := c.GetUint64("userID")

	channelIDStr := c.Param("cid")
	channelID, err := strconv.ParseUint(channelIDStr, 10, 64)
	if err != nil {
		response.Fail(c, errcode.ErrInvalidParam)
		return
	}

	var req dto.UpdateChannelRequest
	if err = c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(c, nil, "参数校验失败: "+err.Error())
		return
	}

	resp, err := h.channelService.UpdateChannel(c.Request.Context(), channelID, userID, &req)
	if err != nil {
		response.Fail(c, err)
		return
	}

	response.Success(c, resp)
}

// DeleteChannel 删除频道
// DELETE /api/v1/servers/:sid/channels/:cid
func (h *ChannelHandler) DeleteChannel(c *gin.Context) {
	userID := c.GetUint64("userID")

	channelIDStr := c.Param("cid")
	channelID, err := strconv.ParseUint(channelIDStr, 10, 64)
	if err != nil {
		response.Fail(c, errcode.ErrInvalidParam)
		return
	}

	if err = h.channelService.DeleteChannel(c.Request.Context(), channelID, userID); err != nil {
		response.Fail(c, err)
		return
	}

	response.Success(c, nil)
}

// ReorderChannels 频道排序
// PUT /api/v1/servers/:sid/channels/reorder
func (h *ChannelHandler) ReorderChannels(c *gin.Context) {
	userID := c.GetUint64("userID")

	serverIDStr := c.Param("id")
	serverID, err := strconv.ParseUint(serverIDStr, 10, 64)
	if err != nil {
		response.Fail(c, errcode.ErrInvalidParam)
		return
	}

	var req dto.ReorderChannelsRequest
	if err = c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(c, nil, "参数校验失败: "+err.Error())
		return
	}

	if err = h.channelService.ReorderChannels(c.Request.Context(), serverID, userID, &req); err != nil {
		response.Fail(c, err)
		return
	}

	response.Success(c, nil)
}
