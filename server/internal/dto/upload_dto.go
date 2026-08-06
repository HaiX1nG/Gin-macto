package dto

// UploadResponse 上传文件响应
// 与前端 uploadService.UploadResponse 对齐
type UploadResponse struct {
	URL      string `json:"url"`      // URL 文件可访问地址
	Filename string `json:"filename"` // Filename 原始文件名
	Size     int64  `json:"size"`     // Size 文件大小（字节）
	Type     string `json:"type"`     // Type 文件分类：image, video, audio, file
}
