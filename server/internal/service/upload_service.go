package service

import (
	"context"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
	"github.com/yourorg/livemix/internal/dto"
	"github.com/yourorg/livemix/pkg/errcode"
	"github.com/yourorg/livemix/pkg/logger"
	"go.uber.org/zap"
)

const (
	// uploadDir 上传文件本地存储目录
	uploadDir = "./uploads"
	// maxUploadSize 最大上传文件大小 50MB
	// 与前端 uploadService.uploadAttachment 上限一致
	maxUploadSize = 50 * 1024 * 1024
)

// UploadService 文件上传服务
// 负责将上传文件保存到本地存储目录并返回可访问URL
type UploadService struct {
	storageDir string
}

// NewUploadService 创建文件上传服务实例
func NewUploadService() *UploadService {
	return &UploadService{storageDir: uploadDir}
}

// Upload 上传文件
// 将文件保存到本地存储目录，返回文件访问URL及元信息
// baseURL 为可访问上传目录的基础URL（如 http://localhost:8080/uploads）
func (s *UploadService) Upload(ctx context.Context, userID uint64, fileHeader *multipart.FileHeader, baseURL string) (*dto.UploadResponse, error) {
	// 校验文件大小
	if fileHeader.Size <= 0 {
		return nil, errcode.ErrBadRequest.WithMessage("文件为空")
	}
	if fileHeader.Size > maxUploadSize {
		return nil, errcode.ErrBadRequest.WithMessage("文件大小不能超过50MB")
	}

	// 确保存储目录存在
	if err := os.MkdirAll(s.storageDir, 0o755); err != nil {
		logger.Error("创建上传目录失败",
			zap.String("dir", s.storageDir),
			zap.Error(err),
		)
		return nil, errcode.ErrInternalServer.WithMessage("创建上传目录失败")
	}

	// 打开上传文件
	src, err := fileHeader.Open()
	if err != nil {
		return nil, errcode.ErrInternalServer.WithMessage("读取上传文件失败")
	}
	defer src.Close()

	// 生成唯一文件名，保留原始扩展名，避免文件名冲突和路径穿越
	ext := filepath.Ext(fileHeader.Filename)
	storedName := uuid.New().String() + ext
	storedPath := filepath.Join(s.storageDir, storedName)

	// 创建目标文件
	dst, err := os.Create(storedPath)
	if err != nil {
		logger.Error("创建目标文件失败",
			zap.String("path", storedPath),
			zap.Error(err),
		)
		return nil, errcode.ErrInternalServer.WithMessage("保存文件失败")
	}
	defer dst.Close()

	// 写入文件内容
	if _, err := dst.ReadFrom(src); err != nil {
		_ = os.Remove(storedPath)
		return nil, errcode.ErrInternalServer.WithMessage("写入文件失败")
	}

	// 推断文件分类
	fileType := detectFileType(fileHeader)

	// 构建可访问URL
	url := strings.TrimRight(baseURL, "/") + "/" + storedName

	logger.Info("文件上传成功",
		logger.WithUserID(userID),
		zap.String("filename", fileHeader.Filename),
		zap.Int64("size", fileHeader.Size),
		zap.String("type", fileType),
		zap.String("url", url),
	)

	return &dto.UploadResponse{
		URL:      url,
		Filename: fileHeader.Filename,
		Size:     fileHeader.Size,
		Type:     fileType,
	}, nil
}

// detectFileType 根据MIME类型推断文件分类
// 返回值: image, video, audio, file
func detectFileType(fileHeader *multipart.FileHeader) string {
	mimeType := fileHeader.Header.Get("Content-Type")
	switch {
	case strings.HasPrefix(mimeType, "image/"):
		return "image"
	case strings.HasPrefix(mimeType, "video/"):
		return "video"
	case strings.HasPrefix(mimeType, "audio/"):
		return "audio"
	default:
		return "file"
	}
}
