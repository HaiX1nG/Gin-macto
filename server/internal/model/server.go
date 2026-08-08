package model

import "time"

// Server 服务器表模型
// KOOK 风格的服务器，是频道的容器，用户通过加入服务器来访问频道
type Server struct {
	ID          uint64    `gorm:"column:id;primaryKey;autoIncrement" json:"id"`                          // ID 服务器唯一标识，自增主键
	Name        string    `gorm:"column:name;type:varchar(100);not null" json:"name"`                     // Name 服务器名称，最大100字符
	IconURL     string    `gorm:"column:icon_url;type:varchar(500)" json:"iconUrl"`                       // IconURL 服务器图标URL
	BannerURL   string    `gorm:"column:banner_url;type:varchar(500)" json:"bannerUrl"`                   // BannerURL 服务器横幅URL
	Description string    `gorm:"column:description;type:varchar(500)" json:"description"`                // Description 服务器描述，最大500字符
	OwnerID     uint64    `gorm:"column:owner_id;not null;index:idx_owner" json:"ownerId"`                // OwnerID 服务器创建者（拥有者）用户ID
	InviteCode  string    `gorm:"column:invite_code;type:varchar(20);not null;uniqueIndex:uk_invite_code" json:"inviteCode"` // InviteCode 邀请码，用于加入服务器，唯一
	IsPrivate   bool      `gorm:"column:is_private;type:tinyint(1);not null;default:1" json:"isPrivate"`  // IsPrivate 是否为私密服务器，默认私密
	MaxMembers  int       `gorm:"column:max_members;not null;default:500" json:"maxMembers"`              // MaxMembers 最大成员数量，默认500
	CreatedAt   time.Time `gorm:"column:created_at;autoCreateTime" json:"createdAt"`                       // CreatedAt 创建时间
	UpdatedAt   time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updatedAt"`                       // UpdatedAt 更新时间
}

// TableName 返回表名
func (Server) TableName() string {
	return "servers"
}

// ServerMember 服务器成员表模型
// 记录用户与服务器的多对多关系，以及服务器内昵称
type ServerMember struct {
	ID        uint64    `gorm:"column:id;primaryKey;autoIncrement" json:"id"`                                            // ID 成员记录唯一标识，自增主键
	ServerID  uint64    `gorm:"column:server_id;not null;uniqueIndex:uk_server_user;index:idx_server" json:"serverId"` // ServerID 服务器ID
	UserID    uint64    `gorm:"column:user_id;not null;uniqueIndex:uk_server_user;index:idx_user" json:"userId"`       // UserID 用户ID
	Nickname  string    `gorm:"column:nickname;type:varchar(50)" json:"nickname"`                                       // Nickname 服务器内昵称，最大50字符
	JoinedAt  time.Time `gorm:"column:joined_at;autoCreateTime" json:"joinedAt"`                                         // JoinedAt 加入时间
}

// TableName 返回表名
func (ServerMember) TableName() string {
	return "server_members"
}

// Role 角色表模型
// 服务器内的角色定义，包含权限位掩码、颜色和排序
type Role struct {
	ID            uint64    `gorm:"column:id;primaryKey;autoIncrement" json:"id"`                              // ID 角色唯一标识，自增主键
	ServerID      uint64    `gorm:"column:server_id;not null;index:idx_server_position" json:"serverId"`       // ServerID 所属服务器ID
	Name          string    `gorm:"column:name;type:varchar(50);not null" json:"name"`                         // Name 角色名称，最大50字符
	Color         string    `gorm:"column:color;type:varchar(7);not null;default:'#99aab5'" json:"color"`       // Color 角色颜色，hex 格式，默认灰色
	Position      int       `gorm:"column:position;not null;default:0;index:idx_server_position" json:"position"` // Position 排序权重，越大越优先
	Permissions   int64     `gorm:"column:permissions;not null;default:0" json:"permissions"`                   // Permissions 权限位掩码，见 permission.go
	IsMentionable bool      `gorm:"column:is_mentionable;type:tinyint(1);not null;default:1" json:"isMentionable"` // IsMentionable 是否可被提及
	CreatedAt     time.Time `gorm:"column:created_at;autoCreateTime" json:"createdAt"`                          // CreatedAt 创建时间
	UpdatedAt     time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updatedAt"`                          // UpdatedAt 更新时间
}

// TableName 返回表名
func (Role) TableName() string {
	return "roles"
}

// ServerMemberRole 成员-角色关联表模型
// 多对多关系：一个成员可以拥有多个角色，一个角色可以被多个成员拥有
type ServerMemberRole struct {
	ID       uint64 `gorm:"column:id;primaryKey;autoIncrement" json:"id"`                                        // ID 关联记录唯一标识，自增主键
	ServerID uint64 `gorm:"column:server_id;not null;index:idx_server" json:"serverId"`                          // ServerID 所属服务器ID
	MemberID uint64 `gorm:"column:member_id;not null;uniqueIndex:uk_member_role;index:idx_member" json:"memberId"` // MemberID server_members.id
	RoleID   uint64 `gorm:"column:role_id;not null;uniqueIndex:uk_member_role;index:idx_role" json:"roleId"`     // RoleID roles.id
}

// TableName 返回表名
func (ServerMemberRole) TableName() string {
	return "server_member_roles"
}
