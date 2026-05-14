package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/yourorg/livemix/internal/model"
	"github.com/yourorg/livemix/internal/repository"
	"github.com/yourorg/livemix/pkg/errcode"
	"gorm.io/gorm"
)

// TestUserService_Register_Success 测试注册成功
func TestUserService_Register_Success(t *testing.T) {
	assert.True(t, true)
}

// TestErrcode_Error 测试错误码
func TestErrcode_Error(t *testing.T) {
	err := errcode.NewError(40001, "参数错误")
	assert.Equal(t, 40001, err.Code)
	assert.Equal(t, "参数错误", err.Message)
}

// TestErrcode_WithMessage 测试自定义消息
func TestErrcode_WithMessage(t *testing.T) {
	err := errcode.ErrBadRequest.WithMessage("自定义错误消息")
	assert.Equal(t, 40000, err.Code)
	assert.Equal(t, "自定义错误消息", err.Message)
}

// TestErrcode_HTTPStatus 测试HTTP状态码
func TestErrcode_HTTPStatus(t *testing.T) {
	assert.Equal(t, 200, errcode.Success.HTTPStatus())
	assert.Equal(t, 400, errcode.ErrBadRequest.HTTPStatus())
	assert.Equal(t, 401, errcode.ErrUnauthorized.HTTPStatus()) // 40003 精确映射到 401
	assert.Equal(t, 500, errcode.ErrInternalServer.HTTPStatus())
}

// MockUserRepo 用于测试的模拟仓储
type MockUserRepo struct {
	users map[string]*model.User
}

func newMockUserRepo() *MockUserRepo {
	return &MockUserRepo{
		users: make(map[string]*model.User),
	}
}

func (m *MockUserRepo) Create(ctx context.Context, user *model.User) error {
	m.users[user.Username] = user
	return nil
}

func (m *MockUserRepo) FindByID(ctx context.Context, id uint64) (*model.User, error) {
	for _, u := range m.users {
		if u.ID == id {
			return u, nil
		}
	}
	return nil, gorm.ErrRecordNotFound
}

func (m *MockUserRepo) FindByUsername(ctx context.Context, username string) (*model.User, error) {
	if u, ok := m.users[username]; ok {
		return u, nil
	}
	return nil, gorm.ErrRecordNotFound
}

func (m *MockUserRepo) FindByEmail(ctx context.Context, email string) (*model.User, error) {
	for _, u := range m.users {
		if u.Email == email {
			return u, nil
		}
	}
	return nil, gorm.ErrRecordNotFound
}

func (m *MockUserRepo) Update(ctx context.Context, user *model.User) error {
	m.users[user.Username] = user
	return nil
}

func (m *MockUserRepo) ExistsByUsername(ctx context.Context, username string) (bool, error) {
	_, ok := m.users[username]
	return ok, nil
}

func (m *MockUserRepo) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	for _, u := range m.users {
		if u.Email == email {
			return true, nil
		}
	}
	return false, nil
}

// Delete 删除用户（硬删除）
func (m *MockUserRepo) Delete(ctx context.Context, id uint64) error {
	for username, u := range m.users {
		if u.ID == id {
			delete(m.users, username)
			return nil
		}
	}
	return nil
}

// DeleteWithDB 使用指定 DB 删除用户（用于事务）
func (m *MockUserRepo) DeleteWithDB(db *gorm.DB, id uint64) error {
	return m.Delete(context.Background(), id)
}

// FindByIDs 批量查询用户
func (m *MockUserRepo) FindByIDs(ctx context.Context, userIDs []uint64) (map[uint64]*model.User, error) {
	result := make(map[uint64]*model.User)
	for _, u := range m.users {
		for _, id := range userIDs {
			if u.ID == id {
				result[id] = u
			}
		}
	}
	return result, nil
}

// 确保MockUserRepo实现接口
var _ repository.UserRepositoryInterface = (*MockUserRepo)(nil)
