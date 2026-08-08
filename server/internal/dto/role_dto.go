package dto

// CreateRoleRequest 创建角色请求
type CreateRoleRequest struct {
	Name          string `json:"name" binding:"required,min=1,max=50"`
	Color         string `json:"color" binding:"omitempty,len=7"`
	Permissions   int64  `json:"permissions"`
	IsMentionable *bool  `json:"isMentionable"`
}

// UpdateRoleRequest 更新角色请求
type UpdateRoleRequest struct {
	Name          string `json:"name" binding:"omitempty,min=1,max=50"`
	Color         string `json:"color" binding:"omitempty,len=7"`
	Permissions   *int64 `json:"permissions"`
	Position      *int   `json:"position"`
	IsMentionable *bool  `json:"isMentionable"`
}

// RoleResponse 角色响应
type RoleResponse struct {
	ID            uint64 `json:"id"`
	ServerID      uint64 `json:"serverId"`
	Name          string `json:"name"`
	Color         string `json:"color"`
	Position      int    `json:"position"`
	Permissions   int64  `json:"permissions"`
	IsMentionable bool   `json:"isMentionable"`
	CreatedAt     string `json:"createdAt"`
}
