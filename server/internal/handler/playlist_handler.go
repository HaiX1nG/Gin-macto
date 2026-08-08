package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/yourorg/livemix/internal/dto"
	"github.com/yourorg/livemix/internal/service"
	"github.com/yourorg/livemix/pkg/errcode"
	"github.com/yourorg/livemix/pkg/response"
)

// PlaylistHandler 播放列表处理器（channel_id 维度）
type PlaylistHandler struct {
	playlistService *service.PlaylistService
}

// NewPlaylistHandler 创建播放列表处理器实例
func NewPlaylistHandler(playlistService *service.PlaylistService) *PlaylistHandler {
	return &PlaylistHandler{playlistService: playlistService}
}

// GetPlaylist 获取播放列表
// GET /api/v1/channels/:cid/playlist
func (h *PlaylistHandler) GetPlaylist(c *gin.Context) {
	userID := c.GetUint64("userID")

	channelIDStr := c.Param("cid")
	channelID, err := strconv.ParseUint(channelIDStr, 10, 64)
	if err != nil {
		response.Fail(c, errcode.ErrInvalidParam)
		return
	}

	resp, err := h.playlistService.GetPlaylist(c.Request.Context(), channelID, userID)
	if err != nil {
		response.Fail(c, err)
		return
	}

	response.Success(c, resp)
}

// AddItem 添加播放项
// POST /api/v1/channels/:cid/playlist
func (h *PlaylistHandler) AddItem(c *gin.Context) {
	userID := c.GetUint64("userID")

	channelIDStr := c.Param("cid")
	channelID, err := strconv.ParseUint(channelIDStr, 10, 64)
	if err != nil {
		response.Fail(c, errcode.ErrInvalidParam)
		return
	}

	var req dto.AddPlaylistItemRequest
	if err = c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(c, nil, "参数校验失败: "+err.Error())
		return
	}

	resp, err := h.playlistService.AddItem(c.Request.Context(), channelID, userID, &req)
	if err != nil {
		response.Fail(c, err)
		return
	}

	response.Success(c, resp)
}

// RemoveItem 删除播放项
// DELETE /api/v1/channels/:cid/playlist/:itemId
func (h *PlaylistHandler) RemoveItem(c *gin.Context) {
	userID := c.GetUint64("userID")

	channelIDStr := c.Param("cid")
	channelID, err := strconv.ParseUint(channelIDStr, 10, 64)
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

	if err = h.playlistService.RemoveItem(c.Request.Context(), channelID, userID, itemID); err != nil {
		response.Fail(c, err)
		return
	}

	response.Success(c, nil)
}

// Play 播放
// POST /api/v1/channels/:cid/playlist/play
func (h *PlaylistHandler) Play(c *gin.Context) {
	userID := c.GetUint64("userID")

	channelIDStr := c.Param("cid")
	channelID, err := strconv.ParseUint(channelIDStr, 10, 64)
	if err != nil {
		response.Fail(c, errcode.ErrInvalidParam)
		return
	}

	resp, err := h.playlistService.Play(c.Request.Context(), channelID, userID)
	if err != nil {
		response.Fail(c, err)
		return
	}

	response.Success(c, resp)
}

// Pause 暂停
// POST /api/v1/channels/:cid/playlist/pause
func (h *PlaylistHandler) Pause(c *gin.Context) {
	userID := c.GetUint64("userID")

	channelIDStr := c.Param("cid")
	channelID, err := strconv.ParseUint(channelIDStr, 10, 64)
	if err != nil {
		response.Fail(c, errcode.ErrInvalidParam)
		return
	}

	if err = h.playlistService.Pause(c.Request.Context(), channelID, userID); err != nil {
		response.Fail(c, err)
		return
	}

	response.Success(c, nil)
}

// Skip 跳过
// POST /api/v1/channels/:cid/playlist/skip
func (h *PlaylistHandler) Skip(c *gin.Context) {
	userID := c.GetUint64("userID")

	channelIDStr := c.Param("cid")
	channelID, err := strconv.ParseUint(channelIDStr, 10, 64)
	if err != nil {
		response.Fail(c, errcode.ErrInvalidParam)
		return
	}

	resp, err := h.playlistService.Skip(c.Request.Context(), channelID, userID)
	if err != nil {
		response.Fail(c, err)
		return
	}

	response.Success(c, resp)
}

// Reorder 排序
// POST /api/v1/channels/:cid/playlist/reorder
func (h *PlaylistHandler) Reorder(c *gin.Context) {
	userID := c.GetUint64("userID")

	channelIDStr := c.Param("cid")
	channelID, err := strconv.ParseUint(channelIDStr, 10, 64)
	if err != nil {
		response.Fail(c, errcode.ErrInvalidParam)
		return
	}

	var req dto.ReorderPlaylistRequest
	if err = c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(c, nil, "参数校验失败: "+err.Error())
		return
	}

	if err = h.playlistService.Reorder(c.Request.Context(), channelID, userID, &req); err != nil {
		response.Fail(c, err)
		return
	}

	response.Success(c, nil)
}
