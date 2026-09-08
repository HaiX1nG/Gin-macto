package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/yourorg/livemix/internal/dto"
	"github.com/yourorg/livemix/internal/service"
	"github.com/yourorg/livemix/pkg/response"
)

// NotificationHandler 通知处理器
// 负责处理通知相关的 REST API 请求
type NotificationHandler struct {
	notificationService *service.NotificationService
}

// NewNotificationHandler 创建通知处理器实例
func NewNotificationHandler(notificationService *service.NotificationService) *NotificationHandler {
	return &NotificationHandler{
		notificationService: notificationService,
	}
}

// GetNotifications 获取通知列表
// GET /api/v1/notifications
// 支持分页和类型过滤
func (h *NotificationHandler) GetNotifications(c *gin.Context) {
	userID := c.GetUint64("userID")

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	notificationType := c.Query("type")

	resp, err := h.notificationService.GetNotifications(c.Request.Context(), userID, page, pageSize, notificationType)
	if err != nil {
		response.Fail(c, err)
		return
	}

	response.Success(c, resp)
}

// MarkAsRead 标记通知为已读/未读
// PUT /api/v1/notifications/:id/read
func (h *NotificationHandler) MarkAsRead(c *gin.Context) {
	userID := c.GetUint64("userID")

	notificationIDStr := c.Param("id")
	notificationID, err := strconv.ParseUint(notificationIDStr, 10, 64)
	if err != nil {
		response.FailWithMessage(c, nil, "无效的通知ID")
		return
	}

	var req dto.MarkReadRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(c, nil, "参数校验失败: "+err.Error())
		return
	}

	if err = h.notificationService.MarkAsRead(c.Request.Context(), userID, notificationID, req.IsRead); err != nil {
		response.Fail(c, err)
		return
	}

	response.Success(c, nil)
}

// GetUnreadCount 获取未读通知数
// GET /api/v1/notifications/unread-count
func (h *NotificationHandler) GetUnreadCount(c *gin.Context) {
	userID := c.GetUint64("userID")

	resp, err := h.notificationService.GetUnreadCount(c.Request.Context(), userID)
	if err != nil {
		response.Fail(c, err)
		return
	}

	response.Success(c, resp)
}

// MarkAllAsRead 标记所有通知为已读
// PUT /api/v1/notifications/read-all
func (h *NotificationHandler) MarkAllAsRead(c *gin.Context) {
	userID := c.GetUint64("userID")

	if err := h.notificationService.MarkAllAsRead(c.Request.Context(), userID); err != nil {
		response.Fail(c, err)
		return
	}

	response.Success(c, nil)
}
