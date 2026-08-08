package model

import "time"

// PlaylistItem 音乐播放列表模型
// 记录语音频道中的音乐播放项，包括音乐信息、播放顺序和状态
type PlaylistItem struct {
	ID        uint64    `gorm:"column:id;primaryKey;autoIncrement" json:"id"`                                          // ID 播放项唯一标识，自增主键
	ChannelID uint64    `gorm:"column:channel_id;not null;index:idx_channel_order;index:idx_channel_status" json:"channelId"` // ChannelID 所属语音频道ID
	AddedBy   uint64    `gorm:"column:added_by;not null" json:"addedBy"`                                               // AddedBy 添加该音乐的用户ID
	Title     string    `gorm:"column:title;type:varchar(200);not null" json:"title"`                                  // Title 音乐标题，最大200字符
	Artist    string    `gorm:"column:artist;type:varchar(200)" json:"artist"`                                         // Artist 艺术家名称，最大200字符
	MusicURL  string    `gorm:"column:music_url;type:varchar(500);not null" json:"musicUrl"`                           // MusicURL 音乐文件URL，最大500字符
	Duration  uint32    `gorm:"column:duration" json:"duration"`                                                       // Duration 音乐时长（秒）
	PlayOrder uint32    `gorm:"column:play_order;not null;default:0;index:idx_channel_order" json:"playOrder"`         // PlayOrder 播放顺序
	Status    int8      `gorm:"column:status;type:tinyint;not null;default:0;index:idx_channel_status" json:"status"`  // Status 播放状态：0=等待，1=播放中，2=已播放
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime" json:"createdAt"`                                     // CreatedAt 创建时间
}

// TableName 返回表名
func (PlaylistItem) TableName() string {
	return "playlist_items"
}
