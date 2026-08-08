package dto

// RegisterRequest 注册请求
type RegisterRequest struct {
	Username string `json:"username" binding:"required,min=3,max=50"`
	Password string `json:"password" binding:"required,min=6,max=50"`
	Email    string `json:"email" binding:"required,email"`
}

// LoginRequest 登录请求
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// LoginResponse 登录响应
type LoginResponse struct {
	UserID       uint64 `json:"userId"`
	Username     string `json:"username"`
	Email        string `json:"email"`
	AvatarURL    string `json:"avatarUrl"`
	BannerURL    string `json:"bannerUrl"`
	Bio          string `json:"bio"`
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
	ExpiresIn    int64  `json:"expiresIn"`
}

// LogoutResponse 退出登录响应
type LogoutResponse struct {
	Success bool `json:"success" example:"true"`
}

// RefreshTokenRequest 刷新Token请求
type RefreshTokenRequest struct {
	RefreshToken string `json:"refreshToken" binding:"required"`
}

// RefreshTokenResponse 刷新Token响应
type RefreshTokenResponse struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
	ExpiresIn    int64  `json:"expiresIn"`
}

// UpdateProfileRequest 更新用户资料请求
type UpdateProfileRequest struct {
	Username  string `json:"username" binding:"omitempty,min=3,max=50"`
	Email     string `json:"email" binding:"omitempty,email"`
	AvatarURL string `json:"avatarUrl" binding:"omitempty,url,max=500"`
	BannerURL string `json:"bannerUrl" binding:"omitempty,url,max=500"`
	Bio       string `json:"bio" binding:"omitempty,max=500"`
}

// ChangePasswordRequest 修改密码请求
type ChangePasswordRequest struct {
	OldPassword string `json:"oldPassword" binding:"required,min=6,max=50"`
	NewPassword string `json:"newPassword" binding:"required,min=6,max=50"`
}

// UserInfoResponse 用户信息响应
type UserInfoResponse struct {
	UserID       uint64 `json:"userId"`
	Username     string `json:"username"`
	Email        string `json:"email"`
	AvatarURL    string `json:"avatarUrl"`
	BannerURL    string `json:"bannerUrl"`
	Bio          string `json:"bio"`
	IsOnline     bool   `json:"isOnline"`
	CustomStatus string `json:"customStatus"`
	CreatedAt    string `json:"createdAt"`
}

// SetCustomStatusRequest 设置自定义状态请求
type SetCustomStatusRequest struct {
	CustomStatus string `json:"customStatus" binding:"omitempty,max=100"`
}

// UserOnlineStatusResponse 用户在线状态响应
type UserOnlineStatusResponse struct {
	UserID       uint64 `json:"userId"`
	Username     string `json:"username"`
	IsOnline     bool   `json:"isOnline"`
	CustomStatus string `json:"customStatus"`
	LastSeenAt   string `json:"lastSeenAt,omitempty"`
}

// DeleteAccountRequest 删除账户请求
type DeleteAccountRequest struct {
	Password string `json:"password" binding:"required,min=6,max=50"`
}
