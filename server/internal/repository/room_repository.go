package repository

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"time"

	"github.com/yourorg/livemix/internal/model"
	"gorm.io/gorm"
)

// RoomRepository 房间仓储
type RoomRepository struct {
	db *gorm.DB
}

// NewRoomRepository 创建房间仓储实例
func NewRoomRepository(db *gorm.DB) *RoomRepository {
	return &RoomRepository{db: db}
}

// Create 创建房间
func (r *RoomRepository) Create(ctx context.Context, room *model.Room) error {
	return r.db.WithContext(ctx).Create(room).Error
}

// CreateWithDB 使用指定 DB 创建房间（用于事务）
// 注意：传入的 db 应通过 TransactionContext.DB() 获取，该 DB 已包含 context
// 使用场景：在事务中创建房间，与创建参与者记录在同一事务中执行
func (r *RoomRepository) CreateWithDB(db *gorm.DB, room *model.Room) error {
	return db.Create(room).Error
}

// DeleteWithDB 使用指定 DB 删除房间（用于事务）
// 注意：传入的 db 应通过 TransactionContext.DB() 获取，该 DB 已包含 context
// 使用场景：在事务中删除房间，与删除参与者记录在同一事务中执行
func (r *RoomRepository) DeleteWithDB(db *gorm.DB, id uint64) error {
	return db.Delete(&model.Room{}, id).Error
}

// FindByID 根据ID查询房间
func (r *RoomRepository) FindByID(ctx context.Context, id uint64) (*model.Room, error) {
	var room model.Room
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&room).Error
	if err != nil {
		return nil, err
	}
	return &room, nil
}

// FindByInviteCode 根据邀请码查询房间
func (r *RoomRepository) FindByInviteCode(ctx context.Context, inviteCode string) (*model.Room, error) {
	var room model.Room
	err := r.db.WithContext(ctx).Where("invite_code = ?", inviteCode).First(&room).Error
	if err != nil {
		return nil, err
	}
	return &room, nil
}

// List 获取公开房间列表
func (r *RoomRepository) List(ctx context.Context, roomType int8, page, pageSize int) ([]model.Room, int64, error) {
	var rooms []model.Room
	var total int64

	query := r.db.WithContext(ctx).Model(&model.Room{}).Where("is_private = ?", false)
	if roomType > 0 {
		query = query.Where("room_type = ?", roomType)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&rooms).Error
	return rooms, total, err
}

// Update 更新房间
func (r *RoomRepository) Update(ctx context.Context, room *model.Room) error {
	return r.db.WithContext(ctx).Save(room).Error
}

// Delete 删除房间
func (r *RoomRepository) Delete(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Delete(&model.Room{}, id).Error
}

// GenerateInviteCode 生成邀请码
func (r *RoomRepository) GenerateInviteCode(ctx context.Context) (string, error) {
	for i := 0; i < 10; i++ {
		bytes := make([]byte, 4)
		if _, err := rand.Read(bytes); err != nil {
			return "", err
		}
		code := hex.EncodeToString(bytes)

		var count int64
		if err := r.db.WithContext(ctx).Model(&model.Room{}).Where("invite_code = ?", code).Count(&count).Error; err != nil {
			return "", err
		}
		if count == 0 {
			return code, nil
		}
	}
	return "", errors.New("生成邀请码失败")
}

// RoomParticipantRepository 房间参与者仓储
type RoomParticipantRepository struct {
	db *gorm.DB
}

// NewRoomParticipantRepository 创建房间参与者仓储实例
func NewRoomParticipantRepository(db *gorm.DB) *RoomParticipantRepository {
	return &RoomParticipantRepository{db: db}
}

// Create 创建参与者记录
func (r *RoomParticipantRepository) Create(ctx context.Context, participant *model.RoomParticipant) error {
	return r.db.WithContext(ctx).Create(participant).Error
}

// CreateWithDB 使用指定 DB 创建参与者记录（用于事务）
// 注意：传入的 db 应通过 TransactionContext.DB() 获取，该 DB 已包含 context
// 使用场景：在事务中创建参与者记录，与创建房间在同一事务中执行
func (r *RoomParticipantRepository) CreateWithDB(db *gorm.DB, participant *model.RoomParticipant) error {
	return db.Create(participant).Error
}

// LeaveWithDB 使用指定 DB 设置参与者离开（用于事务）
// 注意：传入的 db 应通过 TransactionContext.DB() 获取，该 DB 已包含 context
// 使用场景：在事务中设置参与者离开，与删除房间操作在同一事务中执行
func (r *RoomParticipantRepository) LeaveWithDB(db *gorm.DB, roomID, userID uint64) error {
	now := time.Now()
	return db.Model(&model.RoomParticipant{}).
		Where("room_id = ? AND user_id = ? AND is_active = ?", roomID, userID, true).
		Updates(map[string]any{
			"is_active": false,
			"left_at":   now,
		}).Error
}

// DeleteByRoomWithDB 使用指定 DB 删除房间内所有参与者记录（用于事务）
// 注意：传入的 db 应通过 TransactionContext.DB() 获取，该 DB 已包含 context
// 使用场景：在事务中删除房间内所有参与者，与删除房间操作在同一事务中执行
func (r *RoomParticipantRepository) DeleteByRoomWithDB(db *gorm.DB, roomID uint64) error {
	return db.Where("room_id = ?", roomID).Delete(&model.RoomParticipant{}).Error
}

// FindByRoomAndUser 查询用户在房间的参与记录
func (r *RoomParticipantRepository) FindByRoomAndUser(ctx context.Context, roomID, userID uint64) (*model.RoomParticipant, error) {
	var participant model.RoomParticipant
	err := r.db.WithContext(ctx).
		Where("room_id = ? AND user_id = ? AND is_active = ?", roomID, userID, true).
		First(&participant).Error
	if err != nil {
		return nil, err
	}
	return &participant, nil
}

// FindActiveByRoom 查询房间内的活跃参与者
func (r *RoomParticipantRepository) FindActiveByRoom(ctx context.Context, roomID uint64) ([]model.RoomParticipant, error) {
	var participants []model.RoomParticipant
	err := r.db.WithContext(ctx).
		Where("room_id = ? AND is_active = ?", roomID, true).
		Order("joined_at ASC").
		Find(&participants).Error
	return participants, err
}

// CountActiveByRoom 统计房间活跃参与者数量
func (r *RoomParticipantRepository) CountActiveByRoom(ctx context.Context, roomID uint64) (int, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&model.RoomParticipant{}).
		Where("room_id = ? AND is_active = ?", roomID, true).
		Count(&count).Error
	return int(count), err
}

// Update 更新参与者记录
func (r *RoomParticipantRepository) Update(ctx context.Context, participant *model.RoomParticipant) error {
	return r.db.WithContext(ctx).Save(participant).Error
}

// Leave 设置参与者离开
func (r *RoomParticipantRepository) Leave(ctx context.Context, roomID, userID uint64) error {
	now := time.Now()
	return r.db.WithContext(ctx).
		Model(&model.RoomParticipant{}).
		Where("room_id = ? AND user_id = ? AND is_active = ?", roomID, userID, true).
		Updates(map[string]any{
			"is_active": false,
			"left_at":   now,
		}).Error
}

// ExistsActive 检查用户是否在房间中
func (r *RoomParticipantRepository) ExistsActive(ctx context.Context, roomID, userID uint64) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&model.RoomParticipant{}).
		Where("room_id = ? AND user_id = ? AND is_active = ?", roomID, userID, true).
		Count(&count).Error
	return count > 0, err
}

// FindActiveByUser 查询用户当前活跃的房间参与记录
func (r *RoomParticipantRepository) FindActiveByUser(ctx context.Context, userID uint64) ([]model.RoomParticipant, error) {
	var participants []model.RoomParticipant
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND is_active = ?", userID, true).
		Order("joined_at DESC").
		Find(&participants).Error
	return participants, err
}

// FindActiveByUserWithRoom 查询用户当前活跃的房间参与记录（含房间信息）
func (r *RoomParticipantRepository) FindActiveByUserWithRoom(ctx context.Context, userID uint64) ([]struct {
	model.RoomParticipant
	Room model.Room
}, error) {
	var results []struct {
		model.RoomParticipant
		Room model.Room
	}
	err := r.db.WithContext(ctx).
		Table("room_participants").
		Select("room_participants.*, rooms.*").
		Joins("LEFT JOIN rooms ON room_participants.room_id = rooms.id").
		Where("room_participants.user_id = ? AND room_participants.is_active = ?", userID, true).
		Scan(&results).Error
	return results, err
}

// DeleteByRoom 删除房间内所有参与者记录
func (r *RoomParticipantRepository) DeleteByRoom(ctx context.Context, roomID uint64) error {
	return r.db.WithContext(ctx).
		Where("room_id = ?", roomID).
		Delete(&model.RoomParticipant{}).Error
}

// CountActiveByRoomsBatch 批量统计多个房间的活跃参与者数量（解决 N+1 查询问题）
// 返回以 roomID 为 key 的参与者数量 map
func (r *RoomParticipantRepository) CountActiveByRoomsBatch(ctx context.Context, roomIDs []uint64) (map[uint64]int, error) {
	if len(roomIDs) == 0 {
		return make(map[uint64]int), nil
	}

	type roomCount struct {
		RoomID uint64 `gorm:"column:room_id"`
		Count  int    `gorm:"column:count"`
	}

	var results []roomCount
	err := r.db.WithContext(ctx).Model(&model.RoomParticipant{}).
		Select("room_id, COUNT(*) as count").
		Where("room_id IN ? AND is_active = ?", roomIDs, true).
		Group("room_id").
		Scan(&results).Error
	if err != nil {
		return nil, err
	}

	// 转换为 map，并为没有活跃参与者的房间设置默认值 0
	countMap := make(map[uint64]int, len(roomIDs))
	for _, r := range results {
		countMap[r.RoomID] = r.Count
	}
	// 确保所有请求的房间都有记录
	for _, roomID := range roomIDs {
		if _, exists := countMap[roomID]; !exists {
			countMap[roomID] = 0
		}
	}
	return countMap, nil
}
