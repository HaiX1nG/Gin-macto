package model

import "time"

// User 用户表模型
// 存储用户基本信息，包括用户名、密码哈希、邮箱、头像、横幅和个人简介等
type User struct {
	ID           uint64    `gorm:"column:id;primaryKey;autoIncrement" json:"id"`                                      // ID 用户唯一标识，自增主键
	Username     string    `gorm:"column:username;type:varchar(50);not null;uniqueIndex:uk_username" json:"username"` // Username 用户名，唯一，最大50字符
	PasswordHash string    `gorm:"column:password_hash;type:varchar(255);not null" json:"-"`                          // PasswordHash 密码哈希值，不返回给前端
	Email        string    `gorm:"column:email;type:varchar(100);uniqueIndex:uk_email" json:"email"`                  // Email 用户邮箱，唯一，最大100字符
	AvatarURL    string    `gorm:"column:avatar_url;type:varchar(500)" json:"avatarUrl"`                              // AvatarURL 用户头像URL，最大500字符
	BannerURL    string    `gorm:"column:banner_url;type:varchar(500)" json:"bannerUrl"`                              // BannerURL 用户个人横幅URL，最大500字符
	Bio          string    `gorm:"column:bio;type:text" json:"bio"`                                                   // Bio 用户个人简介
	CreatedAt    time.Time `gorm:"column:created_at;autoCreateTime" json:"createdAt"`                                 // CreatedAt 创建时间，自动填充
	UpdatedAt    time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updatedAt"`                                 // UpdatedAt 更新时间，自动更新
}

// TableName 返回表名
func (User) TableName() string {
	return "users"
}

// UserStatus 用户在线状态表模型
// 记录用户的在线状态、自定义状态和最后活跃时间
type UserStatus struct {
	ID           uint64     `gorm:"column:id;primaryKey;autoIncrement" json:"id"`                          // ID 状态记录唯一标识，自增主键
	UserID       uint64     `gorm:"column:user_id;not null;uniqueIndex:uk_user" json:"userId"`             // UserID 用户ID，唯一
	IsOnline     bool       `gorm:"column:is_online;type:tinyint(1);not null;default:0" json:"isOnline"`   // IsOnline 是否在线
	CustomStatus string     `gorm:"column:custom_status;type:varchar(100);default:''" json:"customStatus"` // CustomStatus 用户自定义状态，最大100字符
	LastSeenAt   *time.Time `gorm:"column:last_seen_at" json:"lastSeenAt"`                                 // LastSeenAt 最后活跃时间
	CreatedAt    time.Time  `gorm:"column:created_at;autoCreateTime" json:"createdAt"`                     // CreatedAt 创建时间
	UpdatedAt    time.Time  `gorm:"column:updated_at;autoUpdateTime" json:"updatedAt"`                     // UpdatedAt 更新时间
}

// TableName 返回表名
func (UserStatus) TableName() string {
	return "user_status"
}
