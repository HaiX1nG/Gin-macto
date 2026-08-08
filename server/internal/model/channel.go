package model

import "time"

// 频道类型枚举
const (
	ChannelTypeText     int8 = 1 // 文字频道
	ChannelTypeVoice    int8 = 2 // 语音频道
	ChannelTypeCategory int8 = 3 // 分类
)

// Channel 频道表模型
// 服务器下的频道，支持文字、语音和分类三种类型
type Channel struct {
	ID        uint64    `gorm:"column:id;primaryKey;autoIncrement" json:"id"`                                                          // ID 频道唯一标识，自增主键
	ServerID  uint64    `gorm:"column:server_id;not null;index:idx_server_parent_position" json:"serverId"`                             // ServerID 所属服务器ID
	Name      string    `gorm:"column:name;type:varchar(100);not null" json:"name"`                                                    // Name 频道名称，最大100字符
	Type      int8      `gorm:"column:type;type:tinyint;not null" json:"type"`                                                          // Type 频道类型：1=文字，2=语音，3=分类
	Topic     string    `gorm:"column:topic;type:varchar(500)" json:"topic"`                                                            // Topic 频道主题，最大500字符
	ParentID  *uint64   `gorm:"column:parent_id" json:"parentId"`                                                                       // ParentID 父分类ID，顶层为 null
	Position  int       `gorm:"column:position;not null;default:0;index:idx_server_parent_position" json:"position"`                    // Position 排序权重
	Bitrate   int       `gorm:"column:bitrate;not null;default:64000" json:"bitrate"`                                                   // Bitrate 语音频道比特率，默认64000
	UserLimit int       `gorm:"column:user_limit;not null;default:0" json:"userLimit"`                                                  // UserLimit 语音频道人数上限，0=无限
	SlowMode  int       `gorm:"column:slow_mode;not null;default:0" json:"slowMode"`                                                    // SlowMode 文字频道慢速模式(秒)，0=关闭
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime" json:"createdAt"`                                                      // CreatedAt 创建时间
	UpdatedAt time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updatedAt"`                                                      // UpdatedAt 更新时间
}

// TableName 返回表名
func (Channel) TableName() string {
	return "channels"
}
