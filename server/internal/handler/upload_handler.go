package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/yourorg/livemix/internal/service"
	"github.com/yourorg/livemix/pkg/errcode"
	"github.com/yourorg/livemix/pkg/response"
)

// UploadHandler 文件上传处理器
type UploadHandler struct {
	uploadService *service.UploadService
}

// NewUploadHandler 创建文件上传处理器实例
func NewUploadHandler(uploadService *service.UploadService) *UploadHandler {
	return &UploadHandler{uploadService: uploadService}
}

// Upload 上传文件
// POST /api/v1/upload  Content-Type: multipart/form-data  字段: file
// 前端 uploadService.uploadFile 直接用 fetch 调用，手动注入 Bearer 头
func (h *UploadHandler) Upload(c *gin.Context) {
	userID := c.GetUint64("userID")

	fileHeader, err := c.FormFile("file")
	if err != nil {
		response.FailWithMessage(c, errcode.ErrInvalidParam, "缺少上传文件: file")
		return
	}

	// 构建可访问上传目录的基础URL
	// 使用请求 Host（含端口）拼接，保证返回的URL可被前端直接访问
	baseURL := "http://" + c.Request.Host + "/uploads"

	resp, err := h.uploadService.Upload(c.Request.Context(), userID, fileHeader, baseURL)
	if err != nil {
		response.Fail(c, err)
		return
	}

	response.Success(c, resp)
}
