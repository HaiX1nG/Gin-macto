package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/yourorg/livemix/internal/dto"
	"github.com/yourorg/livemix/internal/service"
	"github.com/yourorg/livemix/pkg/errcode"
	"github.com/yourorg/livemix/pkg/response"
)

// ScreenShareHandler 屏幕共享处理器
type ScreenShareHandler struct {
	screenShareService *service.ScreenShareService
}

// NewScreenShareHandler 创建屏幕共享处理器实例
func NewScreenShareHandler(screenShareService *service.ScreenShareService) *ScreenShareHandler {
	return &ScreenShareHandler{screenShareService: screenShareService}
}

// StartScreenShare 开始屏幕共享
func (h *ScreenShareHandler) StartScreenShare(c *gin.Context) {
	roomIDStr := c.Param("id")
	roomID, err := strconv.ParseUint(roomIDStr, 10, 64)
	if err != nil {
		response.Fail(c, errcode.ErrInvalidParam)
		return
	}

	userID := c.GetUint64("userID")

	resp, err := h.screenShareService.StartScreenShare(c.Request.Context(), roomID, userID)
	if err != nil {
		response.Fail(c, err)
		return
	}

	response.Success(c, resp)
}

// StopScreenShare 停止屏幕共享
func (h *ScreenShareHandler) StopScreenShare(c *gin.Context) {
	roomIDStr := c.Param("id")
	roomID, err := strconv.ParseUint(roomIDStr, 10, 64)
	if err != nil {
		response.Fail(c, errcode.ErrInvalidParam)
		return
	}

	userID := c.GetUint64("userID")

	if err = h.screenShareService.StopScreenShare(c.Request.Context(), roomID, userID); err != nil {
		response.Fail(c, err)
		return
	}

	response.Success(c, nil)
}

// GetActiveScreenShare 获取当前屏幕共享
func (h *ScreenShareHandler) GetActiveScreenShare(c *gin.Context) {
	roomIDStr := c.Param("id")
	roomID, err := strconv.ParseUint(roomIDStr, 10, 64)
	if err != nil {
		response.Fail(c, errcode.ErrInvalidParam)
		return
	}

	resp, err := h.screenShareService.GetActiveScreenShare(c.Request.Context(), roomID)
	if err != nil {
		response.Fail(c, err)
		return
	}

	response.Success(c, resp)
}

// VoiceHandler 语音处理器
type VoiceHandler struct {
	voiceService *service.VoiceService
}

// NewVoiceHandler 创建语音处理器实例
func NewVoiceHandler(voiceService *service.VoiceService) *VoiceHandler {
	return &VoiceHandler{voiceService: voiceService}
}

// JoinVoice 加入语音
func (h *VoiceHandler) JoinVoice(c *gin.Context) {
	roomIDStr := c.Param("id")
	roomID, err := strconv.ParseUint(roomIDStr, 10, 64)
	if err != nil {
		response.Fail(c, errcode.ErrInvalidParam)
		return
	}

	userID := c.GetUint64("userID")

	resp, err := h.voiceService.JoinVoice(c.Request.Context(), roomID, userID)
	if err != nil {
		response.Fail(c, err)
		return
	}

	response.Success(c, resp)
}

// LeaveVoice 离开语音
func (h *VoiceHandler) LeaveVoice(c *gin.Context) {
	roomIDStr := c.Param("id")
	roomID, err := strconv.ParseUint(roomIDStr, 10, 64)
	if err != nil {
		response.Fail(c, errcode.ErrInvalidParam)
		return
	}

	userID := c.GetUint64("userID")

	if err = h.voiceService.LeaveVoice(c.Request.Context(), roomID, userID); err != nil {
		response.Fail(c, err)
		return
	}

	response.Success(c, nil)
}

// GetVoiceParticipants 获取语音参与者
func (h *VoiceHandler) GetVoiceParticipants(c *gin.Context) {
	roomIDStr := c.Param("id")
	roomID, err := strconv.ParseUint(roomIDStr, 10, 64)
	if err != nil {
		response.Fail(c, errcode.ErrInvalidParam)
		return
	}

	resp, err := h.voiceService.GetVoiceParticipants(c.Request.Context(), roomID)
	if err != nil {
		response.Fail(c, err)
		return
	}

	response.Success(c, resp)
}

// SetMuteRequest 静音请求
type SetMuteRequest struct {
	Muted bool `json:"muted"`
}

// SetMute 设置静音状态
func (h *VoiceHandler) SetMute(c *gin.Context) {
	roomIDStr := c.Param("id")
	roomID, err := strconv.ParseUint(roomIDStr, 10, 64)
	if err != nil {
		response.Fail(c, errcode.ErrInvalidParam)
		return
	}

	userID := c.GetUint64("userID")

	var req SetMuteRequest
	if err = c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(c, errcode.ErrInvalidParam, "参数校验失败")
		return
	}

	if err = h.voiceService.SetMute(c.Request.Context(), roomID, userID, req.Muted); err != nil {
		response.Fail(c, err)
		return
	}

	response.Success(c, nil)
}

// SendSignal 发送 WebRTC 信令（REST 备用通道）
// POST /api/v1/rooms/:id/webrtc/signal  对应前端 webrtcService.sendSignal
// 实时信令主要通过 WebSocket 传输，此 REST 端点为备用通道
func (h *VoiceHandler) SendSignal(c *gin.Context) {
	roomIDStr := c.Param("id")
	roomID, err := strconv.ParseUint(roomIDStr, 10, 64)
	if err != nil {
		response.Fail(c, errcode.ErrInvalidParam)
		return
	}

	userID := c.GetUint64("userID")

	var req dto.WebRTCSignalRequest
	if err = c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(c, errcode.ErrInvalidParam, "参数校验失败: "+err.Error())
		return
	}

	if err = h.voiceService.SendSignal(c.Request.Context(), roomID, userID, &req); err != nil {
		response.Fail(c, err)
		return
	}

	response.Success(c, nil)
}
