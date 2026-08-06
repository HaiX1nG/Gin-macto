package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/yourorg/livemix/internal/dto"
	"github.com/yourorg/livemix/internal/service"
	"github.com/yourorg/livemix/pkg/errcode"
	"github.com/yourorg/livemix/pkg/response"
)

// RoomHandler 房间处理器
type RoomHandler struct {
	roomService *service.RoomService
}

// NewRoomHandler 创建房间处理器实例
func NewRoomHandler(roomService *service.RoomService) *RoomHandler {
	return &RoomHandler{roomService: roomService}
}

// CreateRoom 创建房间
func (h *RoomHandler) CreateRoom(c *gin.Context) {
	userID := c.GetUint64("userID")

	var req dto.CreateRoomRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(c, nil, "参数校验失败: "+err.Error())
		return
	}

	resp, err := h.roomService.CreateRoom(c.Request.Context(), userID, &req)
	if err != nil {
		response.Fail(c, err)
		return
	}

	response.Success(c, resp)
}

// JoinRoom 加入房间
func (h *RoomHandler) JoinRoom(c *gin.Context) {
	userID := c.GetUint64("userID")

	roomIDStr := c.Param("id")
	roomID, err := strconv.ParseUint(roomIDStr, 10, 64)
	if err != nil {
		response.Fail(c, nil)
		return
	}

	var req dto.JoinRoomRequest
	c.ShouldBindJSON(&req)

	if err := h.roomService.JoinRoom(c.Request.Context(), userID, roomID, &req); err != nil {
		response.Fail(c, err)
		return
	}

	response.Success(c, nil)
}

// LeaveRoom 离开房间
func (h *RoomHandler) LeaveRoom(c *gin.Context) {
	userID := c.GetUint64("userID")

	roomIDStr := c.Param("id")
	roomID, err := strconv.ParseUint(roomIDStr, 10, 64)
	if err != nil {
		response.Fail(c, nil)
		return
	}

	if err := h.roomService.LeaveRoom(c.Request.Context(), userID, roomID); err != nil {
		response.Fail(c, err)
		return
	}

	response.Success(c, nil)
}

// GetRoomList 获取房间列表
func (h *RoomHandler) GetRoomList(c *gin.Context) {
	var req dto.RoomListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.FailWithMessage(c, nil, "参数校验失败: "+err.Error())
		return
	}

	resp, total, err := h.roomService.GetRoomList(c.Request.Context(), &req)
	if err != nil {
		response.Fail(c, err)
		return
	}

	response.Page(c, resp, total, req.Page, req.PageSize)
}

// GetRoomInfo 获取房间信息
func (h *RoomHandler) GetRoomInfo(c *gin.Context) {
	roomIDStr := c.Param("id")
	roomID, err := strconv.ParseUint(roomIDStr, 10, 64)
	if err != nil {
		response.Fail(c, nil)
		return
	}

	resp, err := h.roomService.GetRoomInfo(c.Request.Context(), roomID)
	if err != nil {
		response.Fail(c, err)
		return
	}

	response.Success(c, resp)
}

// GetRoomParticipants 获取房间参与者
func (h *RoomHandler) GetRoomParticipants(c *gin.Context) {
	roomIDStr := c.Param("id")
	roomID, err := strconv.ParseUint(roomIDStr, 10, 64)
	if err != nil {
		response.Fail(c, nil)
		return
	}

	resp, err := h.roomService.GetRoomParticipants(c.Request.Context(), roomID)
	if err != nil {
		response.Fail(c, err)
		return
	}

	response.Success(c, resp)
}

// GetOnlineCount 获取房间在线人数
func (h *RoomHandler) GetOnlineCount(c *gin.Context) {
	roomIDStr := c.Param("id")
	roomID, err := strconv.ParseUint(roomIDStr, 10, 64)
	if err != nil {
		response.Fail(c, nil)
		return
	}

	resp, err := h.roomService.GetOnlineCount(c.Request.Context(), roomID)
	if err != nil {
		response.Fail(c, err)
		return
	}

	response.Success(c, resp)
}

// GetOnlineUsers 获取房间在线用户列表
func (h *RoomHandler) GetOnlineUsers(c *gin.Context) {
	roomIDStr := c.Param("id")
	roomID, err := strconv.ParseUint(roomIDStr, 10, 64)
	if err != nil {
		response.Fail(c, nil)
		return
	}

	resp, err := h.roomService.GetOnlineUsers(c.Request.Context(), roomID)
	if err != nil {
		response.Fail(c, err)
		return
	}

	response.Success(c, resp)
}

// GetUserStatus 获取用户状态
func (h *RoomHandler) GetUserStatus(c *gin.Context) {
	userIDStr := c.Param("id")
	userID, err := strconv.ParseUint(userIDStr, 10, 64)
	if err != nil {
		response.Fail(c, nil)
		return
	}

	resp, err := h.roomService.GetUserStatus(c.Request.Context(), userID)
	if err != nil {
		response.Fail(c, err)
		return
	}

	response.Success(c, resp)
}

// DeleteRoom 删除房间（仅房主可操作）
func (h *RoomHandler) DeleteRoom(c *gin.Context) {
	userID := c.GetUint64("userID")

	roomIDStr := c.Param("id")
	roomID, err := strconv.ParseUint(roomIDStr, 10, 64)
	if err != nil {
		response.Fail(c, nil)
		return
	}

	if err := h.roomService.DeleteRoom(c.Request.Context(), userID, roomID); err != nil {
		response.Fail(c, err)
		return
	}

	response.Success(c, nil)
}

// GetPublicRooms 获取公开房间列表
// GET /api/v1/rooms/public  对应前端 roomService.getPublicRooms
func (h *RoomHandler) GetPublicRooms(c *gin.Context) {
	var req dto.RoomListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.FailWithMessage(c, errcode.ErrInvalidParam, "参数校验失败: "+err.Error())
		return
	}

	resp, total, err := h.roomService.GetPublicRooms(c.Request.Context(), &req)
	if err != nil {
		response.Fail(c, err)
		return
	}

	response.Page(c, resp, total, req.Page, req.PageSize)
}

// KickMember 踢出房间成员
// POST /api/v1/rooms/:id/kick/:userId  对应前端 roomService.kickParticipant
func (h *RoomHandler) KickMember(c *gin.Context) {
	userID := c.GetUint64("userID")

	roomIDStr := c.Param("id")
	roomID, err := strconv.ParseUint(roomIDStr, 10, 64)
	if err != nil {
		response.Fail(c, errcode.ErrInvalidParam)
		return
	}

	targetUserIDStr := c.Param("userId")
	targetUserID, err := strconv.ParseUint(targetUserIDStr, 10, 64)
	if err != nil {
		response.Fail(c, errcode.ErrInvalidParam)
		return
	}

	if err = h.roomService.KickMember(c.Request.Context(), userID, roomID, targetUserID); err != nil {
		response.Fail(c, err)
		return
	}

	response.Success(c, nil)
}

// UpdateMemberRole 设置成员角色
// PUT /api/v1/rooms/:id/participants/:userId/role  对应前端 roomService.setParticipantRole
func (h *RoomHandler) UpdateMemberRole(c *gin.Context) {
	userID := c.GetUint64("userID")

	roomIDStr := c.Param("id")
	roomID, err := strconv.ParseUint(roomIDStr, 10, 64)
	if err != nil {
		response.Fail(c, errcode.ErrInvalidParam)
		return
	}

	targetUserIDStr := c.Param("userId")
	targetUserID, err := strconv.ParseUint(targetUserIDStr, 10, 64)
	if err != nil {
		response.Fail(c, errcode.ErrInvalidParam)
		return
	}

	var req dto.UpdateMemberRoleRequest
	if err = c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(c, errcode.ErrInvalidParam, "参数校验失败: "+err.Error())
		return
	}

	if err = h.roomService.UpdateMemberRole(c.Request.Context(), userID, roomID, targetUserID, &req); err != nil {
		response.Fail(c, err)
		return
	}

	response.Success(c, nil)
}
