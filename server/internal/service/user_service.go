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
	userRepo   *repository.UserRepository
	statusRepo *repository.UserStatusRepository
	txManager  *repository.GormTransactionManager
}

// NewUserService 创建用户服务实例
func NewUserService(userRepo *repository.UserRepository, statusRepo *repository.UserStatusRepository, txManager *repository.GormTransactionManager) *UserService {
	return &UserService{
		userRepo:   userRepo,
		statusRepo: statusRepo,
		txManager:  txManager,
	}
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

	// 获取用户状态
	status, err := s.statusRepo.FindByUserID(ctx, userID)
	isOnline := false
	customStatus := ""
	if err == nil {
		isOnline = status.IsOnline
		customStatus = status.CustomStatus
	}

	return &dto.UserInfoResponse{
		UserID:       user.ID,
		Username:     user.Username,
		Email:        user.Email,
		AvatarURL:    user.AvatarURL,
		IsOnline:     isOnline,
		CustomStatus: customStatus,
		CreatedAt:    user.CreatedAt.Format("2006-01-02 15:04:05"),
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

// SetOnline 设置用户在线
func (s *UserService) SetOnline(ctx context.Context, userID uint64) error {
	return s.statusRepo.SetOnline(ctx, userID)
}

// SetOffline 设置用户离线
func (s *UserService) SetOffline(ctx context.Context, userID uint64) error {
	return s.statusRepo.SetOffline(ctx, userID)
}

// SetCustomStatus 设置自定义状态
func (s *UserService) SetCustomStatus(ctx context.Context, userID uint64, customStatus string) error {
	return s.statusRepo.SetCustomStatus(ctx, userID, customStatus)
}

// GetUserOnlineStatus 获取用户在线状态
func (s *UserService) GetUserOnlineStatus(ctx context.Context, userID uint64) (*dto.UserOnlineStatusResponse, error) {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errcode.ErrUserNotFound
		}
		return nil, errcode.ErrDBError.WithMessage("查询用户失败")
	}

	status, err := s.statusRepo.FindByUserID(ctx, userID)
	isOnline := false
	customStatus := ""
	lastSeenAt := ""
	if err == nil {
		isOnline = status.IsOnline
		customStatus = status.CustomStatus
		if status.LastSeenAt != nil {
			lastSeenAt = status.LastSeenAt.Format("2006-01-02 15:04:05")
		}
	}

	return &dto.UserOnlineStatusResponse{
		UserID:       user.ID,
		Username:     user.Username,
		IsOnline:     isOnline,
		CustomStatus: customStatus,
		LastSeenAt:   lastSeenAt,
	}, nil
}

// GetUsersOnlineStatus 批量获取用户在线状态
func (s *UserService) GetUsersOnlineStatus(ctx context.Context, userIDs []uint64) ([]dto.UserOnlineStatusResponse, error) {
	statuses, err := s.statusRepo.FindByUserIDs(ctx, userIDs)
	if err != nil {
		return nil, errcode.ErrDBError.WithMessage("查询用户状态失败")
	}

	// 构建状态映射
	statusMap := make(map[uint64]*model.UserStatus)
	for i := range statuses {
		statusMap[statuses[i].UserID] = &statuses[i]
	}

	var responses []dto.UserOnlineStatusResponse
	for _, userID := range userIDs {
		user, err := s.userRepo.FindByID(ctx, userID)
		if err != nil {
			continue
		}

		isOnline := false
		customStatus := ""
		lastSeenAt := ""
		if status, ok := statusMap[userID]; ok {
			isOnline = status.IsOnline
			customStatus = status.CustomStatus
			if status.LastSeenAt != nil {
				lastSeenAt = status.LastSeenAt.Format("2006-01-02 15:04:05")
			}
		}

		responses = append(responses, dto.UserOnlineStatusResponse{
			UserID:       user.ID,
			Username:     user.Username,
			IsOnline:     isOnline,
			CustomStatus: customStatus,
			LastSeenAt:   lastSeenAt,
		})
	}

	return responses, nil
}

// DeleteAccount 删除账户
// 危险操作，需要验证用户密码
// 使用事务确保删除用户状态记录和删除用户记录的原子性
func (s *UserService) DeleteAccount(ctx context.Context, userID uint64, password string) error {
	// 查询用户
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errcode.ErrUserNotFound
		}
		return errcode.ErrDBError.WithMessage("查询用户失败")
	}

	// 验证密码
	if !util.CheckPassword(password, user.PasswordHash) {
		return errcode.ErrWrongPassword
	}

	// 使用事务删除用户状态记录和用户记录
	// 确保两个操作要么全部成功，要么全部回滚
	// 避免删除状态后用户删除失败导致的数据不一致问题
	err = s.txManager.Transactional(ctx, func(tx repository.TransactionContext) error {
		// 删除用户状态记录
		if deleteErr := s.statusRepo.DeleteByUserIDWithDB(tx.DB(), userID); deleteErr != nil {
			return deleteErr
		}

		// 删除用户记录
		if deleteErr := s.userRepo.DeleteWithDB(tx.DB(), userID); deleteErr != nil {
			return deleteErr
		}

		return nil
	})

	if err != nil {
		return errcode.ErrDBError.WithMessage("删除账户失败")
	}

	return nil
}
