package dao

import (
	"context"
	"errors"
	"gorm.io/gorm"
	"user-service/ddd/infrastructure/database/po"
	"user-service/internal/resource"
)

type UserDao struct {
	db *gorm.DB
}

func NewUserDao() *UserDao {
	return &UserDao{
		db: resource.DefaultMysqlResource().MainDB(),
	}
}

func (d *UserDao) Create(ctx context.Context, userPo *po.UserPo) error {
	return d.db.WithContext(ctx).Create(userPo).Error
}

func (d *UserDao) QueryByAccount(ctx context.Context, account string) (*po.UserPo, error) {
	var user po.UserPo
	err := d.db.WithContext(ctx).Where("account = ?", account).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

func (d *UserDao) QueryByUUID(ctx context.Context, userUUID string) (*po.UserPo, error) {
	var user po.UserPo
	err := d.db.WithContext(ctx).Where("user_uuid = ?", userUUID).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

func (d *UserDao) QueryByID(ctx context.Context, id uint64) (*po.UserPo, error) {
	var user po.UserPo
	err := d.db.WithContext(ctx).Where("id = ?", id).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

func (d *UserDao) Update(ctx context.Context, userPo *po.UserPo) error {
	return d.db.WithContext(ctx).Save(userPo).Error
}

func (d *UserDao) DeleteByUUID(ctx context.Context, userUUID string) error {
	return d.db.WithContext(ctx).Where("user_uuid = ?", userUUID).Delete(&po.UserPo{}).Error
}

func (d *UserDao) ExistsByAccount(ctx context.Context, account string) (bool, error) {
	var count int64
	err := d.db.WithContext(ctx).Model(&po.UserPo{}).Where("account = ?", account).Count(&count).Error
	return count > 0, err
}

func (d *UserDao) ExistsByUUID(ctx context.Context, userUUID string) (bool, error) {
	var count int64
	err := d.db.WithContext(ctx).Model(&po.UserPo{}).Where("user_uuid = ?", userUUID).Count(&count).Error
	return count > 0, err
}