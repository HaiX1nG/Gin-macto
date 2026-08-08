package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/yourorg/livemix/internal/dto"
	"github.com/yourorg/livemix/internal/service"
	"github.com/yourorg/livemix/pkg/errcode"
	"github.com/yourorg/livemix/pkg/response"
)

// VoiceHandler 语音处理器（channel_id 维度）
type VoiceHandler struct {
	voiceService *service.VoiceService
}

// NewVoiceHandler 创建语音处理器实例
func NewVoiceHandler(voiceService *service.VoiceService) *VoiceHandler {
	return &VoiceHandler{voiceService: voiceService}
}

// JoinVoice 加入语音频道
// POST /api/v1/channels/:cid/voice/join
func (h *VoiceHandler) JoinVoice(c *gin.Context) {
	userID := c.GetUint64("userID")

	channelIDStr := c.Param("cid")
	channelID, err := strconv.ParseUint(channelIDStr, 10, 64)
	if err != nil {
		response.Fail(c, errcode.ErrInvalidParam)
		return
	}

	resp, err := h.voiceService.JoinVoice(c.Request.Context(), channelID, userID)
	if err != nil {
		response.Fail(c, err)
		return
	}

	response.Success(c, resp)
}

// LeaveVoice 离开语音频道
// POST /api/v1/channels/:cid/voice/leave
func (h *VoiceHandler) LeaveVoice(c *gin.Context) {
	userID := c.GetUint64("userID")

	channelIDStr := c.Param("cid")
	channelID, err := strconv.ParseUint(channelIDStr, 10, 64)
	if err != nil {
		response.Fail(c, errcode.ErrInvalidParam)
		return
	}

	if err = h.voiceService.LeaveVoice(c.Request.Context(), channelID, userID); err != nil {
		response.Fail(c, err)
		return
	}

	response.Success(c, nil)
}

// GetVoiceParticipants 获取语音频道参与者
// GET /api/v1/channels/:cid/voice/participants
func (h *VoiceHandler) GetVoiceParticipants(c *gin.Context) {
	userID := c.GetUint64("userID")

	channelIDStr := c.Param("cid")
	channelID, err := strconv.ParseUint(channelIDStr, 10, 64)
	if err != nil {
		response.Fail(c, errcode.ErrInvalidParam)
		return
	}

	resp, err := h.voiceService.GetVoiceParticipants(c.Request.Context(), channelID, userID)
	if err != nil {
		response.Fail(c, err)
		return
	}

	response.Success(c, resp)
}

// SetMute 设置静音
// POST /api/v1/channels/:cid/voice/mute
func (h *VoiceHandler) SetMute(c *gin.Context) {
	userID := c.GetUint64("userID")

	channelIDStr := c.Param("cid")
	channelID, err := strconv.ParseUint(channelIDStr, 10, 64)
	if err != nil {
		response.Fail(c, errcode.ErrInvalidParam)
		return
	}

	var req dto.SetMuteRequest
	if err = c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(c, nil, "参数校验失败")
		return
	}

	if err = h.voiceService.SetMute(c.Request.Context(), channelID, userID, req.IsMuted); err != nil {
		response.Fail(c, err)
		return
	}

	response.Success(c, nil)
}

// StartScreenShare 开始屏幕共享
// POST /api/v1/channels/:cid/screenshare/start
func (h *VoiceHandler) StartScreenShare(c *gin.Context) {
	userID := c.GetUint64("userID")

	channelIDStr := c.Param("cid")
	channelID, err := strconv.ParseUint(channelIDStr, 10, 64)
	if err != nil {
		response.Fail(c, errcode.ErrInvalidParam)
		return
	}

	resp, err := h.voiceService.StartScreenShare(c.Request.Context(), channelID, userID)
	if err != nil {
		response.Fail(c, err)
		return
	}

	response.Success(c, resp)
}

// StopScreenShare 停止屏幕共享
// POST /api/v1/channels/:cid/screenshare/stop
func (h *VoiceHandler) StopScreenShare(c *gin.Context) {
	userID := c.GetUint64("userID")

	channelIDStr := c.Param("cid")
	channelID, err := strconv.ParseUint(channelIDStr, 10, 64)
	if err != nil {
		response.Fail(c, errcode.ErrInvalidParam)
		return
	}

	if err = h.voiceService.StopScreenShare(c.Request.Context(), channelID, userID); err != nil {
		response.Fail(c, err)
		return
	}

	response.Success(c, nil)
}

// GetActiveScreenShare 获取当前屏幕共享
// GET /api/v1/channels/:cid/screenshare
func (h *VoiceHandler) GetActiveScreenShare(c *gin.Context) {
	channelIDStr := c.Param("cid")
	channelID, err := strconv.ParseUint(channelIDStr, 10, 64)
	if err != nil {
		response.Fail(c, errcode.ErrInvalidParam)
		return
	}

	resp, err := h.voiceService.GetActiveScreenShare(c.Request.Context(), channelID)
	if err != nil {
		response.Fail(c, err)
		return
	}

	response.Success(c, resp)
}
