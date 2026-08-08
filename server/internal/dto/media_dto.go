package dto

// AddPlaylistItemRequest 添加播放项请求
type AddPlaylistItemRequest struct {
	Title    string `json:"title" binding:"required,min=1,max=200"`
	Artist   string `json:"artist" binding:"omitempty,max=200"`
	MusicURL string `json:"musicUrl" binding:"required,url,max=500"`
	Duration uint32 `json:"duration" binding:"omitempty"`
}

// PlaylistItemResponse 播放项响应
type PlaylistItemResponse struct {
	ID        uint64 `json:"id"`
	Title     string `json:"title"`
	Artist    string `json:"artist"`
	MusicURL  string `json:"musicUrl"`
	Duration  uint32 `json:"duration"`
	PlayOrder uint32 `json:"playOrder"`
	Status    int8   `json:"status"`
	AddedBy   uint64 `json:"addedBy"`
}

// ReorderPlaylistRequest 调整播放顺序请求
type ReorderPlaylistRequest struct {
	ItemIDs []uint64 `json:"itemIds" binding:"required,min=1"`
}
