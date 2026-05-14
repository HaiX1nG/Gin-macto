package repository

import (
	"context"

	"gorm.io/gorm"
)

// GormTransactionManager GORM 事务管理器
// 实现 TransactionManager 接口，提供基于 GORM 的事务管理能力
type GormTransactionManager struct {
	db *gorm.DB
}

// NewTransactionManager 创建事务管理器实例
func NewTransactionManager(db *gorm.DB) *GormTransactionManager {
	return &GormTransactionManager{db: db}
}

// Begin 开始事务，返回事务上下文
// 使用示例：
//
//	tx := tm.Begin(ctx)
//	defer tx.Rollback() // defer 中调用 Rollback 是安全的，已提交时无操作
//	// 执行业务逻辑...
//	if err := tx.Commit(); err != nil {
//	    return err
//	}
func (m *GormTransactionManager) Begin(ctx context.Context) TransactionContext {
	// 使用 GORM 的 Begin 方法开启事务
	tx := m.db.WithContext(ctx).Begin()
	return &gormTransactionContext{
		tx:   tx,
		ctx:  ctx,
		done: false,
	}
}

// gormTransactionContext GORM 事务上下文
// 封装 GORM 事务的生命周期管理，确保事务安全提交或回滚
type gormTransactionContext struct {
	tx   *gorm.DB
	ctx  context.Context
	done bool // 标记事务是否已完成（提交或回滚）
}

// Commit 提交事务
// 如果事务已完成（提交或回滚），则返回 nil（幂等操作）
func (c *gormTransactionContext) Commit() error {
	if c.done {
		return nil
	}
	c.done = true
	return c.tx.Commit().Error
}

// Rollback 回滚事务
// 如果事务已完成（提交或回滚），则返回 nil（幂等操作）
// 此方法设计为可在 defer 中安全调用，即使事务已提交也不会报错
func (c *gormTransactionContext) Rollback() error {
	if c.done {
		return nil
	}
	c.done = true
	return c.tx.Rollback().Error
}

// Context 获取带有事务的 context.Context
// 用于传递给需要在事务中执行的 repository 方法
func (c *gormTransactionContext) Context() context.Context {
	return c.ctx
}

// DB 获取事务 DB 实例
// 用于在事务中执行数据库操作
// 注意：返回的 DB 已包含 context（在 Begin 时通过 WithContext 注入）
// 因此使用 DB() 返回的实例执行操作时，无需再次调用 WithContext
// 使用示例：
//
//	err := tm.Transactional(ctx, func(tx TransactionContext) error {
//	    // tx.DB() 已包含 context，直接使用即可
//	    if err := userRepo.DeleteWithDB(tx.DB(), userID); err != nil {
//	        return err
//	    }
//	    return nil
//	})
func (c *gormTransactionContext) DB() *gorm.DB {
	return c.tx
}

// Transactional 执行事务函数
// 辅助函数，简化事务处理流程
// 如果 fn 返回错误，事务将自动回滚；否则自动提交
//
// 使用示例：
//
//	err := tm.Transactional(ctx, func(tx TransactionContext) error {
//	    if err := userRepo.Create(tx.Context(), user); err != nil {
//	        return err
//	    }
//	    if err := statusRepo.Create(tx.Context(), status); err != nil {
//	        return err
//	    }
//	    return nil
//	})
func (m *GormTransactionManager) Transactional(ctx context.Context, fn func(tx TransactionContext) error) error {
	tx := m.Begin(ctx)
	defer tx.Rollback() // 安全回滚，已提交时无操作

	if err := fn(tx); err != nil {
		return err
	}

	return tx.Commit()
}
