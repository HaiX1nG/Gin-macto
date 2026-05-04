package service

import (
	"context"
	"errors"

	"github.com/yourorg/livemix/internal/dto"
	"github.com/yourorg/livemix/internal/model"
	"github.com/yourorg/livemix/internal/repository"
	"github.com/yourorg/livemix/pkg/errcode"
	"github.com/yourorg/livemix/pkg/jwt"
	"github.com/yourorg/livemix/pkg/util"
	"gorm.io/gorm"
)

// UserService 用户服务
type UserService struct {
	userRepo *repository.UserRepository
}

// NewUserService 创建用户服务实例
func NewUserService(userRepo *repository.UserRepository) *UserService {
	return &UserService{userRepo: userRepo}
}

// Register 用户注册
func (s *UserService) Register(ctx context.Context, req *dto.RegisterRequest) (*model.User, error) {
	// 检查用户名是否存在
	exists, err := s.userRepo.ExistsByUsername(ctx, req.Username)
	if err != nil {
		return nil, errcode.ErrDBError.WithMessage("检查用户名失败")
	}
	if exists {
		return nil, errcode.ErrUserAlreadyExists
	}

	// 检查邮箱是否存在
	exists, err = s.userRepo.ExistsByEmail(ctx, req.Email)
	if err != nil {
		return nil, errcode.ErrDBError.WithMessage("检查邮箱失败")
	}
	if exists {
		return nil, errcode.ErrEmailAlreadyUsed
	}

	// 密码加密
	passwordHash, err := util.HashPassword(req.Password)
	if err != nil {
		return nil, errcode.ErrInternalServer.WithMessage("密码加密失败")
	}

	// 创建用户
	user := &model.User{
		Username:     req.Username,
		PasswordHash: passwordHash,
		Email:        req.Email,
	}

	if err = s.userRepo.Create(ctx, user); err != nil {
		return nil, errcode.ErrDBError.WithMessage("创建用户失败")
	}

	return user, nil
}

// Login 用户登录
func (s *UserService) Login(ctx context.Context, req *dto.LoginRequest) (*dto.LoginResponse, error) {
	// 查询用户
	user, err := s.userRepo.FindByUsername(ctx, req.Username)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errcode.ErrUserNotFound
		}
		return nil, errcode.ErrDBError.WithMessage("查询用户失败")
	}

	// 验证密码
	if !util.CheckPassword(req.Password, user.PasswordHash) {
		return nil, errcode.ErrWrongPassword
	}

	// 生成Token
	tokenPair, err := jwt.GenerateToken(user.ID, user.Username)
	if err != nil {
		return nil, errcode.ErrInternalServer.WithMessage("生成Token失败")
	}

	return &dto.LoginResponse{
		UserID:       user.ID,
		Username:     user.Username,
		Email:        user.Email,
		AvatarURL:    user.AvatarURL,
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		ExpiresIn:    tokenPair.ExpiresIn,
	}, nil
}

// RefreshToken 刷新Token
func (s *UserService) RefreshToken(ctx context.Context, refreshToken string) (*dto.RefreshTokenResponse, error) {
	// 解析Refresh Token
	claims, err := jwt.ParseRefreshToken(refreshToken)
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, errcode.ErrTokenExpired
		}
		return nil, errcode.ErrTokenInvalid
	}

	// 验证用户是否存在
	user, err := s.userRepo.FindByID(ctx, claims.UserID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errcode.ErrUserNotFound
		}
		return nil, errcode.ErrDBError.WithMessage("查询用户失败")
	}

	// 生成新的Token对
	tokenPair, err := jwt.GenerateToken(user.ID, user.Username)
	if err != nil {
		return nil, errcode.ErrInternalServer.WithMessage("生成Token失败")
	}

	return &dto.RefreshTokenResponse{
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		ExpiresIn:    tokenPair.ExpiresIn,
	}, nil
}

// GetUserInfo 获取用户信息
func (s *UserService) GetUserInfo(ctx context.Context, userID uint64) (*dto.UserInfoResponse, error) {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errcode.ErrUserNotFound
		}
		return nil, errcode.ErrDBError.WithMessage("查询用户失败")
	}

	return &dto.UserInfoResponse{
		UserID:    user.ID,
		Username:  user.Username,
		Email:     user.Email,
		AvatarURL: user.AvatarURL,
		CreatedAt: user.CreatedAt.Format("2006-01-02 15:04:05"),
	}, nil
}

// UpdateProfile 更新用户资料
func (s *UserService) UpdateProfile(ctx context.Context, userID uint64, req *dto.UpdateProfileRequest) error {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errcode.ErrUserNotFound
		}
		return errcode.ErrDBError.WithMessage("查询用户失败")
	}

	// 更新用户名
	if req.Username != "" && req.Username != user.Username {
		exists, err := s.userRepo.ExistsByUsername(ctx, req.Username)
		if err != nil {
			return errcode.ErrDBError.WithMessage("检查用户名失败")
		}
		if exists {
			return errcode.ErrUserAlreadyExists
		}
		user.Username = req.Username
	}

	// 更新邮箱
	if req.Email != "" && req.Email != user.Email {
		exists, err := s.userRepo.ExistsByEmail(ctx, req.Email)
		if err != nil {
			return errcode.ErrDBError.WithMessage("检查邮箱失败")
		}
		if exists {
			return errcode.ErrEmailAlreadyUsed
		}
		user.Email = req.Email
	}

	// 更新头像
	if req.AvatarURL != "" {
		user.AvatarURL = req.AvatarURL
	}

	if err = s.userRepo.Update(ctx, user); err != nil {
		return errcode.ErrDBError.WithMessage("更新用户失败")
	}

	return nil
}

// ChangePassword 修改密码
func (s *UserService) ChangePassword(ctx context.Context, userID uint64, req *dto.ChangePasswordRequest) error {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errcode.ErrUserNotFound
		}
		return errcode.ErrDBError.WithMessage("查询用户失败")
	}

	// 验证旧密码
	if !util.CheckPassword(req.OldPassword, user.PasswordHash) {
		return errcode.ErrWrongPassword
	}

	// 新密码加密
	passwordHash, err := util.HashPassword(req.NewPassword)
	if err != nil {
		return errcode.ErrInternalServer.WithMessage("密码加密失败")
	}

	user.PasswordHash = passwordHash

	if err = s.userRepo.Update(ctx, user); err != nil {
		return errcode.ErrDBError.WithMessage("更新密码失败")
	}

	return nil
}
