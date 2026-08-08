package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/yourorg/livemix/internal/dto"
	"github.com/yourorg/livemix/internal/service"
	"github.com/yourorg/livemix/pkg/errcode"
	"github.com/yourorg/livemix/pkg/response"
)

// RoleHandler 角色处理器
type RoleHandler struct {
	roleService *service.RoleService
}

// NewRoleHandler 创建角色处理器实例
func NewRoleHandler(roleService *service.RoleService) *RoleHandler {
	return &RoleHandler{roleService: roleService}
}

// CreateRole 创建角色
// POST /api/v1/servers/:id/roles
func (h *RoleHandler) CreateRole(c *gin.Context) {
	userID := c.GetUint64("userID")

	serverIDStr := c.Param("id")
	serverID, err := strconv.ParseUint(serverIDStr, 10, 64)
	if err != nil {
		response.Fail(c, errcode.ErrInvalidParam)
		return
	}

	var req dto.CreateRoleRequest
	if err = c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(c, nil, "参数校验失败: "+err.Error())
		return
	}

	resp, err := h.roleService.CreateRole(c.Request.Context(), serverID, userID, &req)
	if err != nil {
		response.Fail(c, err)
		return
	}

	response.Success(c, resp)
}

// GetRoles 获取角色列表
// GET /api/v1/servers/:id/roles
func (h *RoleHandler) GetRoles(c *gin.Context) {
	userID := c.GetUint64("userID")

	serverIDStr := c.Param("id")
	serverID, err := strconv.ParseUint(serverIDStr, 10, 64)
	if err != nil {
		response.Fail(c, errcode.ErrInvalidParam)
		return
	}

	resp, err := h.roleService.GetRoles(c.Request.Context(), serverID, userID)
	if err != nil {
		response.Fail(c, err)
		return
	}

	response.Success(c, resp)
}

// UpdateRole 更新角色
// PUT /api/v1/servers/:id/roles/:rid
func (h *RoleHandler) UpdateRole(c *gin.Context) {
	userID := c.GetUint64("userID")

	serverIDStr := c.Param("id")
	serverID, err := strconv.ParseUint(serverIDStr, 10, 64)
	if err != nil {
		response.Fail(c, errcode.ErrInvalidParam)
		return
	}

	roleIDStr := c.Param("rid")
	roleID, err := strconv.ParseUint(roleIDStr, 10, 64)
	if err != nil {
		response.Fail(c, errcode.ErrInvalidParam)
		return
	}

	var req dto.UpdateRoleRequest
	if err = c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(c, nil, "参数校验失败: "+err.Error())
		return
	}

	resp, err := h.roleService.UpdateRole(c.Request.Context(), serverID, roleID, userID, &req)
	if err != nil {
		response.Fail(c, err)
		return
	}

	response.Success(c, resp)
}

// DeleteRole 删除角色
// DELETE /api/v1/servers/:id/roles/:rid
func (h *RoleHandler) DeleteRole(c *gin.Context) {
	userID := c.GetUint64("userID")

	serverIDStr := c.Param("id")
	serverID, err := strconv.ParseUint(serverIDStr, 10, 64)
	if err != nil {
		response.Fail(c, errcode.ErrInvalidParam)
		return
	}

	roleIDStr := c.Param("rid")
	roleID, err := strconv.ParseUint(roleIDStr, 10, 64)
	if err != nil {
		response.Fail(c, errcode.ErrInvalidParam)
		return
	}

	if err = h.roleService.DeleteRole(c.Request.Context(), serverID, roleID, userID); err != nil {
		response.Fail(c, err)
		return
	}

	response.Success(c, nil)
}
