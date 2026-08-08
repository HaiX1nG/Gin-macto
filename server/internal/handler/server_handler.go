package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/yourorg/livemix/internal/dto"
	"github.com/yourorg/livemix/internal/service"
	"github.com/yourorg/livemix/pkg/errcode"
	"github.com/yourorg/livemix/pkg/response"
)

// ServerHandler 服务器处理器
type ServerHandler struct {
	serverService *service.ServerService
}

// NewServerHandler 创建服务器处理器实例
func NewServerHandler(serverService *service.ServerService) *ServerHandler {
	return &ServerHandler{serverService: serverService}
}

// CreateServer 创建服务器
// POST /api/v1/servers
func (h *ServerHandler) CreateServer(c *gin.Context) {
	userID := c.GetUint64("userID")

	var req dto.CreateServerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(c, nil, "参数校验失败: "+err.Error())
		return
	}

	resp, err := h.serverService.CreateServer(c.Request.Context(), userID, &req)
	if err != nil {
		response.Fail(c, err)
		return
	}

	response.Success(c, resp)
}

// GetServerList 获取我的服务器列表
// GET /api/v1/servers
func (h *ServerHandler) GetServerList(c *gin.Context) {
	userID := c.GetUint64("userID")

	resp, err := h.serverService.GetServerList(c.Request.Context(), userID)
	if err != nil {
		response.Fail(c, err)
		return
	}

	response.Success(c, resp)
}

// GetServerDetail 获取服务器详情
// GET /api/v1/servers/:id
func (h *ServerHandler) GetServerDetail(c *gin.Context) {
	serverIDStr := c.Param("id")
	serverID, err := strconv.ParseUint(serverIDStr, 10, 64)
	if err != nil {
		response.Fail(c, errcode.ErrInvalidParam)
		return
	}

	resp, err := h.serverService.GetServerDetail(c.Request.Context(), serverID)
	if err != nil {
		response.Fail(c, err)
		return
	}

	response.Success(c, resp)
}

// UpdateServer 更新服务器
// PUT /api/v1/servers/:id
func (h *ServerHandler) UpdateServer(c *gin.Context) {
	userID := c.GetUint64("userID")

	serverIDStr := c.Param("id")
	serverID, err := strconv.ParseUint(serverIDStr, 10, 64)
	if err != nil {
		response.Fail(c, errcode.ErrInvalidParam)
		return
	}

	var req dto.UpdateServerRequest
	if err = c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(c, nil, "参数校验失败: "+err.Error())
		return
	}

	resp, err := h.serverService.UpdateServer(c.Request.Context(), userID, serverID, &req)
	if err != nil {
		response.Fail(c, err)
		return
	}

	response.Success(c, resp)
}

// DeleteServer 删除服务器
// DELETE /api/v1/servers/:id
func (h *ServerHandler) DeleteServer(c *gin.Context) {
	userID := c.GetUint64("userID")

	serverIDStr := c.Param("id")
	serverID, err := strconv.ParseUint(serverIDStr, 10, 64)
	if err != nil {
		response.Fail(c, errcode.ErrInvalidParam)
		return
	}

	if err = h.serverService.DeleteServer(c.Request.Context(), userID, serverID); err != nil {
		response.Fail(c, err)
		return
	}

	response.Success(c, nil)
}

// JoinServer 加入服务器
// POST /api/v1/servers/:id/join
func (h *ServerHandler) JoinServer(c *gin.Context) {
	userID := c.GetUint64("userID")

	serverIDStr := c.Param("id")
	serverID, err := strconv.ParseUint(serverIDStr, 10, 64)
	if err != nil {
		response.Fail(c, errcode.ErrInvalidParam)
		return
	}

	var req dto.JoinServerRequest
	if err = c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(c, nil, "参数校验失败: "+err.Error())
		return
	}

	if err = h.serverService.JoinServer(c.Request.Context(), userID, serverID, &req); err != nil {
		response.Fail(c, err)
		return
	}

	response.Success(c, nil)
}

// LeaveServer 离开服务器
// POST /api/v1/servers/:id/leave
func (h *ServerHandler) LeaveServer(c *gin.Context) {
	userID := c.GetUint64("userID")

	serverIDStr := c.Param("id")
	serverID, err := strconv.ParseUint(serverIDStr, 10, 64)
	if err != nil {
		response.Fail(c, errcode.ErrInvalidParam)
		return
	}

	if err = h.serverService.LeaveServer(c.Request.Context(), userID, serverID); err != nil {
		response.Fail(c, err)
		return
	}

	response.Success(c, nil)
}

// GetMembers 获取成员列表
// GET /api/v1/servers/:id/members
func (h *ServerHandler) GetMembers(c *gin.Context) {
	serverIDStr := c.Param("id")
	serverID, err := strconv.ParseUint(serverIDStr, 10, 64)
	if err != nil {
		response.Fail(c, errcode.ErrInvalidParam)
		return
	}

	resp, err := h.serverService.GetMembers(c.Request.Context(), serverID)
	if err != nil {
		response.Fail(c, err)
		return
	}

	response.Success(c, resp)
}

// GetMember 获取成员详情
// GET /api/v1/servers/:id/members/:uid
func (h *ServerHandler) GetMember(c *gin.Context) {
	serverIDStr := c.Param("id")
	serverID, err := strconv.ParseUint(serverIDStr, 10, 64)
	if err != nil {
		response.Fail(c, errcode.ErrInvalidParam)
		return
	}

	userIDStr := c.Param("uid")
	targetUserID, err := strconv.ParseUint(userIDStr, 10, 64)
	if err != nil {
		response.Fail(c, errcode.ErrInvalidParam)
		return
	}

	resp, err := h.serverService.GetMember(c.Request.Context(), serverID, targetUserID)
	if err != nil {
		response.Fail(c, err)
		return
	}

	response.Success(c, resp)
}

// UpdateMember 更新成员
// PUT /api/v1/servers/:id/members/:uid
func (h *ServerHandler) UpdateMember(c *gin.Context) {
	userID := c.GetUint64("userID")

	serverIDStr := c.Param("id")
	serverID, err := strconv.ParseUint(serverIDStr, 10, 64)
	if err != nil {
		response.Fail(c, errcode.ErrInvalidParam)
		return
	}

	userIDStr := c.Param("uid")
	targetUserID, err := strconv.ParseUint(userIDStr, 10, 64)
	if err != nil {
		response.Fail(c, errcode.ErrInvalidParam)
		return
	}

	var req dto.UpdateServerMemberRequest
	if err = c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(c, nil, "参数校验失败: "+err.Error())
		return
	}

	if err = h.serverService.UpdateMember(c.Request.Context(), userID, serverID, targetUserID, &req); err != nil {
		response.Fail(c, err)
		return
	}

	response.Success(c, nil)
}

// KickMember 踢出成员
// DELETE /api/v1/servers/:id/members/:uid
func (h *ServerHandler) KickMember(c *gin.Context) {
	userID := c.GetUint64("userID")

	serverIDStr := c.Param("id")
	serverID, err := strconv.ParseUint(serverIDStr, 10, 64)
	if err != nil {
		response.Fail(c, errcode.ErrInvalidParam)
		return
	}

	userIDStr := c.Param("uid")
	targetUserID, err := strconv.ParseUint(userIDStr, 10, 64)
	if err != nil {
		response.Fail(c, errcode.ErrInvalidParam)
		return
	}

	if err = h.serverService.KickMember(c.Request.Context(), userID, serverID, targetUserID); err != nil {
		response.Fail(c, err)
		return
	}

	response.Success(c, nil)
}
