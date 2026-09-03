package handler

import (
	"context"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/yourorg/livemix/internal/dto"
	"github.com/yourorg/livemix/internal/middleware"
	"github.com/yourorg/livemix/internal/service"
	"github.com/yourorg/livemix/pkg/jwt"
	"github.com/yourorg/livemix/pkg/response"
)

// AuthHandler 认证处理器
type AuthHandler struct {
	userService *service.UserService
}

// NewAuthHandler 创建认证处理器实例
func NewAuthHandler(userService *service.UserService) *AuthHandler {
	return &AuthHandler{userService: userService}
}

// Register 用户注册
func (h *AuthHandler) Register(c *gin.Context) {
	var req dto.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(c, nil, "参数校验失败")
		return
	}

	user, err := h.userService.Register(c.Request.Context(), &req)
	if err != nil {
		response.Fail(c, err)
		return
	}

	response.Success(c, gin.H{
		"userId":   user.ID,
		"username": user.Username,
		"email":    user.Email,
	})
}

// Login 用户登录
func (h *AuthHandler) Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(c, nil, "参数校验失败")
		return
	}

	resp, err := h.userService.Login(c.Request.Context(), &req)
	if err != nil {
		response.Fail(c, err)
		return
	}

	response.Success(c, resp)
}

// RefreshToken 刷新Token
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var req dto.RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(c, nil, "参数校验失败")
		return
	}

	resp, err := h.userService.RefreshToken(c.Request.Context(), req.RefreshToken)
	if err != nil {
		response.Fail(c, err)
		return
	}

	response.Success(c, resp)
}

// Logout 用户退出登录
// 将当前用户的Access Token加入黑名单，使其失效
// 后续使用该Token的请求将被拒绝
func (h *AuthHandler) Logout(c *gin.Context) {
	// 从上下文获取Access Token
	accessToken := middleware.GetAccessToken(c)
	if accessToken == "" {
		response.Success(c, dto.LogoutResponse{Success: true})
		return
	}

	// 将Token加入黑名单
	ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
	defer cancel()

	if err := jwt.AddAccessTokenToBlacklist(ctx, accessToken); err != nil {
		// 即使加入黑名单失败，也返回成功
		// Token会在过期后自动失效
		// 日志记录在AddAccessTokenToBlacklist内部处理
	}

	response.Success(c, dto.LogoutResponse{Success: true})
}

// GetUserInfo 获取用户信息
func (h *AuthHandler) GetUserInfo(c *gin.Context) {
	userID := c.GetUint64("userID")

	resp, err := h.userService.GetUserInfo(c.Request.Context(), userID)
	if err != nil {
		response.Fail(c, err)
		return
	}

	response.Success(c, resp)
}

// UpdateProfile 更新用户资料
func (h *AuthHandler) UpdateProfile(c *gin.Context) {
	userID := c.GetUint64("userID")

	var req dto.UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(c, nil, "参数校验失败")
		return
	}

	if err := h.userService.UpdateProfile(c.Request.Context(), userID, &req); err != nil {
		response.Fail(c, err)
		return
	}

	response.Success(c, nil)
}

// ChangePassword 修改密码
// 修改密码成功后，将当前Token加入黑名单，需要重新登录
func (h *AuthHandler) ChangePassword(c *gin.Context) {
	userID := c.GetUint64("userID")

	var req dto.ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(c, nil, "参数校验失败")
		return
	}

	if err := h.userService.ChangePassword(c.Request.Context(), userID, &req); err != nil {
		response.Fail(c, err)
		return
	}

	// 修改密码成功后，将当前Token加入黑名单
	// 强制用户使用新密码重新登录
	accessToken := middleware.GetAccessToken(c)
	if accessToken != "" {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
		defer cancel()
		// 即使加入黑名单失败也不影响密码修改结果
		_ = jwt.AddAccessTokenToBlacklist(ctx, accessToken)
	}

	response.Success(c, gin.H{
		"success": true,
		"message": "密码修改成功，请重新登录",
	})
}

// SetCustomStatus 设置自定义状态
func (h *AuthHandler) SetCustomStatus(c *gin.Context) {
	userID := c.GetUint64("userID")

	var req dto.SetCustomStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(c, nil, "参数校验失败")
		return
	}

	if err := h.userService.SetCustomStatus(c.Request.Context(), userID, req.CustomStatus); err != nil {
		response.Fail(c, err)
		return
	}

	response.Success(c, nil)
}

// GetUserOnlineStatus 获取用户在线状态
func (h *AuthHandler) GetUserOnlineStatus(c *gin.Context) {
	userIDStr := c.Param("id")
	userID, err := strconv.ParseUint(userIDStr, 10, 64)
	if err != nil {
		response.Fail(c, nil)
		return
	}

	resp, err := h.userService.GetUserOnlineStatus(c.Request.Context(), userID)
	if err != nil {
		response.Fail(c, err)
		return
	}

	response.Success(c, resp)
}

// GetUserInfoByID 按 ID 查询用户公开信息
// GET /api/v1/users/:id/info  对应前端 roomService.getMemberInfo
// 复用 userService.GetUserInfo，返回 UserInfoResponse（不含密码等敏感字段）
func (h *AuthHandler) GetUserInfoByID(c *gin.Context) {
	userIDStr := c.Param("id")
	userID, err := strconv.ParseUint(userIDStr, 10, 64)
	if err != nil {
		response.Fail(c, nil)
		return
	}

	resp, err := h.userService.GetUserInfo(c.Request.Context(), userID)
	if err != nil {
		response.Fail(c, err)
		return
	}

	response.Success(c, resp)
}

// DeleteAccount 删除账户
// 删除账户后，将当前Token加入黑名单，使其立即失效
func (h *AuthHandler) DeleteAccount(c *gin.Context) {
	userID := c.GetUint64("userID")

	var req dto.DeleteAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(c, nil, "参数校验失败")
		return
	}

	if err := h.userService.DeleteAccount(c.Request.Context(), userID, req.Password); err != nil {
		response.Fail(c, err)
		return
	}

	// 删除账户后，将当前Token加入黑名单
	accessToken := middleware.GetAccessToken(c)
	if accessToken != "" {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
		defer cancel()
		_ = jwt.AddAccessTokenToBlacklist(ctx, accessToken)
	}

	response.Success(c, nil)
}
