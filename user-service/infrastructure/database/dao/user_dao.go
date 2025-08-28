package dao

import (
	"context"
	"errors"
	"sync"

	"go-video/ddd/internal/resource"
	"go-video/ddd/user/infrastructure/database/po"
	"go-video/pkg/assert"
	"gorm.io/gorm"
)

var (
	userDaoOnce      sync.Once
	singletonUserDao *UserDao
)

// UserDao 用户数据访问对象
type UserDao struct {
	db *gorm.DB
}

// DefaultUserDao 获取用户DAO单例
func DefaultUserDao() *UserDao {
	assert.NotCircular()
	userDaoOnce.Do(func() {
		db := resource.DefaultMysqlResource()
		singletonUserDao = &UserDao{
			db: db.MainDB(),
		}
	})
	assert.NotNil(singletonUserDao)
	return singletonUserDao
}

// NewUserDao 创建用户DAO实例（支持依赖注入）
func NewUserDao() *UserDao {
	return &UserDao{
		db: resource.DefaultMysqlResource().MainDB(),
	}
}

// Create 创建用户
func (d *UserDao) Create(ctx context.Context, user *po.UserPO) error {
	return d.db.WithContext(ctx).Create(user).Error
}

// GetByID 根据内部ID获取用户（仅内部使用）
func (d *UserDao) GetByID(ctx context.Context, id uint64) (*po.UserPO, error) {
	var user po.UserPO
	err := d.db.WithContext(ctx).Where("id = ? AND is_deleted = 0", id).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

// GetByUUID 根据UUID获取用户
func (d *UserDao) GetByUUID(ctx context.Context, uuid string) (*po.UserPO, error) {
	var user po.UserPO
	err := d.db.WithContext(ctx).Where("uuid = ? AND is_deleted = 0", uuid).First(&user).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

// GetByUsername 根据用户名获取用户
func (d *UserDao) GetByUsername(ctx context.Context, username string) (*po.UserPO, error) {
	var user po.UserPO
	err := d.db.WithContext(ctx).Where("username = ? AND is_deleted = 0", username).First(&user).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

// GetByEmail 根据邮箱获取用户
func (d *UserDao) GetByEmail(ctx context.Context, email string) (*po.UserPO, error) {
	var user po.UserPO
	err := d.db.WithContext(ctx).Where("email = ? AND is_deleted = 0", email).First(&user).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

// ExistsByUsername 检查用户名是否存在
func (d *UserDao) ExistsByUsername(ctx context.Context, username string) (bool, error) {
	var count int64
	err := d.db.WithContext(ctx).Model(&po.UserPO{}).Where("username = ?", username).Count(&count).Error
	return count > 0, err
}

// ExistsByEmail 检查邮箱是否存在
func (d *UserDao) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	var count int64
	err := d.db.WithContext(ctx).Model(&po.UserPO{}).Where("email = ?", email).Count(&count).Error
	return count > 0, err
}

// GetByPage 分页获取用户
func (d *UserDao) GetByPage(ctx context.Context, offset, limit int) ([]*po.UserPO, error) {
	var users []*po.UserPO
	err := d.db.WithContext(ctx).Offset(offset).Limit(limit).Find(&users).Error
	return users, err
}

// Count 统计用户总数
func (d *UserDao) Count(ctx context.Context) (int64, error) {
	var count int64
	err := d.db.WithContext(ctx).Model(&po.UserPO{}).Count(&count).Error
	return count, err
}

// Update 更新用户
func (d *UserDao) Update(ctx context.Context, user *po.UserPO) error {
	return d.db.WithContext(ctx).Save(user).Error
}

// Delete 删除用户
func (d *UserDao) Delete(ctx context.Context, id uint64) error {
	return d.db.WithContext(ctx).Delete(&po.UserPO{}, id).Error
}

// GetActiveUsersByPage 分页获取激活用户
func (d *UserDao) GetActiveUsersByPage(ctx context.Context, offset, limit int) ([]*po.UserPO, error) {
	var users []*po.UserPO
	err := d.db.WithContext(ctx).Where("status = ?", 1).Offset(offset).Limit(limit).Find(&users).Error
	return users, err
}

// CountActiveUsers 统计激活用户数量
func (d *UserDao) CountActiveUsers(ctx context.Context) (int64, error) {
	var count int64
	err := d.db.WithContext(ctx).Model(&po.UserPO{}).Where("status = ?", 1).Count(&count).Error
	return count, err
}

// NewDatabaseTx 创建事务DAO
func (d *UserDao) NewDatabaseTx(tx *gorm.DB) *UserDao {
	return &UserDao{
		db: tx,
	}
}
