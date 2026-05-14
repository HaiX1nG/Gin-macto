package repository

import (
	"context"

	"github.com/yourorg/livemix/internal/model"
	"gorm.io/gorm"
)

// UserRepository 用户仓储
type UserRepository struct {
	db *gorm.DB
}

// NewUserRepository 创建用户仓储实例
func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

// Create 创建用户
func (r *UserRepository) Create(ctx context.Context, user *model.User) error {
	return r.db.WithContext(ctx).Create(user).Error
}

// FindByID 根据ID查询用户
func (r *UserRepository) FindByID(ctx context.Context, id uint64) (*model.User, error) {
	var user model.User
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// FindByUsername 根据用户名查询用户
func (r *UserRepository) FindByUsername(ctx context.Context, username string) (*model.User, error) {
	var user model.User
	err := r.db.WithContext(ctx).Where("username = ?", username).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// FindByEmail 根据邮箱查询用户
func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*model.User, error) {
	var user model.User
	err := r.db.WithContext(ctx).Where("email = ?", email).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// Update 更新用户
func (r *UserRepository) Update(ctx context.Context, user *model.User) error {
	return r.db.WithContext(ctx).Save(user).Error
}

// ExistsByUsername 检查用户名是否存在
func (r *UserRepository) ExistsByUsername(ctx context.Context, username string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.User{}).Where("username = ?", username).Count(&count).Error
	return count > 0, err
}

// ExistsByEmail 检查邮箱是否存在
func (r *UserRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.User{}).Where("email = ?", email).Count(&count).Error
	return count > 0, err
}

// Delete 删除用户（硬删除）
func (r *UserRepository) Delete(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Delete(&model.User{}, id).Error
}

// DeleteWithDB 使用指定 DB 删除用户（用于事务）
// 注意：传入的 db 应通过 TransactionContext.DB() 获取，该 DB 已包含 context
// 使用场景：在事务中删除用户，确保与其他操作在同一事务中执行
func (r *UserRepository) DeleteWithDB(db *gorm.DB, id uint64) error {
	return db.Delete(&model.User{}, id).Error
}

// FindByIDs 批量查询用户（解决 N+1 查询问题）
// 参数 userIDs 为用户ID列表，返回以 userID 为 key 的用户信息 map
func (r *UserRepository) FindByIDs(ctx context.Context, userIDs []uint64) (map[uint64]*model.User, error) {
	if len(userIDs) == 0 {
		return make(map[uint64]*model.User), nil
	}

	var users []model.User
	err := r.db.WithContext(ctx).Where("id IN ?", userIDs).Find(&users).Error
	if err != nil {
		return nil, err
	}

	// 转换为 map 以便快速查找
	userMap := make(map[uint64]*model.User, len(users))
	for i := range users {
		userMap[users[i].ID] = &users[i]
	}
	return userMap, nil
}

// UserStatusRepository 用户状态仓储
type UserStatusRepository struct {
	db *gorm.DB
}

// NewUserStatusRepository 创建用户状态仓储实例
func NewUserStatusRepository(db *gorm.DB) *UserStatusRepository {
	return &UserStatusRepository{db: db}
}

// Upsert 创建或更新用户状态
func (r *UserStatusRepository) Upsert(ctx context.Context, status *model.UserStatus) error {
	return r.db.WithContext(ctx).
		Exec(`INSERT INTO user_status (user_id, is_online, custom_status, last_seen_at, created_at, updated_at)
			VALUES (?, ?, ?, ?, NOW(), NOW())
			ON DUPLICATE KEY UPDATE is_online = ?, custom_status = ?, last_seen_at = ?, updated_at = NOW()`,
			status.UserID, status.IsOnline, status.CustomStatus, status.LastSeenAt,
			status.IsOnline, status.CustomStatus, status.LastSeenAt).Error
}

// FindByUserID 根据用户ID查询状态
func (r *UserStatusRepository) FindByUserID(ctx context.Context, userID uint64) (*model.UserStatus, error) {
	var status model.UserStatus
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).First(&status).Error
	if err != nil {
		return nil, err
	}
	return &status, nil
}

// FindByUserIDs 批量查询用户状态
func (r *UserStatusRepository) FindByUserIDs(ctx context.Context, userIDs []uint64) ([]model.UserStatus, error) {
	var statuses []model.UserStatus
	err := r.db.WithContext(ctx).Where("user_id IN ?", userIDs).Find(&statuses).Error
	return statuses, err
}

// SetOnline 设置用户在线
func (r *UserStatusRepository) SetOnline(ctx context.Context, userID uint64) error {
	return r.db.WithContext(ctx).
		Exec(`INSERT INTO user_status (user_id, is_online, updated_at, created_at)
			VALUES (?, 1, NOW(), NOW())
			ON DUPLICATE KEY UPDATE is_online = 1, updated_at = NOW()`, userID).Error
}

// SetOffline 设置用户离线
func (r *UserStatusRepository) SetOffline(ctx context.Context, userID uint64) error {
	return r.db.WithContext(ctx).
		Exec(`INSERT INTO user_status (user_id, is_online, last_seen_at, updated_at, created_at)
			VALUES (?, 0, NOW(), NOW(), NOW())
			ON DUPLICATE KEY UPDATE is_online = 0, last_seen_at = NOW(), updated_at = NOW()`, userID).Error
}

// SetCustomStatus 设置自定义状态
func (r *UserStatusRepository) SetCustomStatus(ctx context.Context, userID uint64, customStatus string) error {
	return r.db.WithContext(ctx).
		Exec(`INSERT INTO user_status (user_id, custom_status, updated_at, created_at)
			VALUES (?, ?, NOW(), NOW())
			ON DUPLICATE KEY UPDATE custom_status = ?, updated_at = NOW()`, userID, customStatus, customStatus).Error
}

// DeleteByUserID 根据用户ID删除状态记录
func (r *UserStatusRepository) DeleteByUserID(ctx context.Context, userID uint64) error {
	return r.db.WithContext(ctx).Where("user_id = ?", userID).Delete(&model.UserStatus{}).Error
}

// DeleteByUserIDWithDB 使用指定 DB 根据用户ID删除状态记录（用于事务）
// 注意：传入的 db 应通过 TransactionContext.DB() 获取，该 DB 已包含 context
// 使用场景：在事务中删除用户状态，与删除用户操作在同一事务中执行
func (r *UserStatusRepository) DeleteByUserIDWithDB(db *gorm.DB, userID uint64) error {
	return db.Where("user_id = ?", userID).Delete(&model.UserStatus{}).Error
}

// FindByUserIDs 批量查询用户状态（解决 N+1 查询问题）
// 参数 userIDs 为用户ID列表，返回以 userID 为 key 的用户状态 map
func (r *UserStatusRepository) FindByUserIDsMap(ctx context.Context, userIDs []uint64) (map[uint64]*model.UserStatus, error) {
	if len(userIDs) == 0 {
		return make(map[uint64]*model.UserStatus), nil
	}

	var statuses []model.UserStatus
	err := r.db.WithContext(ctx).Where("user_id IN ?", userIDs).Find(&statuses).Error
	if err != nil {
		return nil, err
	}

	// 转换为 map 以便快速查找
	statusMap := make(map[uint64]*model.UserStatus, len(statuses))
	for i := range statuses {
		statusMap[statuses[i].UserID] = &statuses[i]
	}
	return statusMap, nil
}
