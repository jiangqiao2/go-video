package dao

import (
    "context"
    "gorm.io/gorm"
    "upload-service/ddd/infrastructure/database/po"
    "upload-service/internal/resource"
)

type TagDao struct {
    db *gorm.DB
}

func NewTagDao() *TagDao {
    return &TagDao{db: resource.DefaultMysqlResource().MainDB()}
}

func (d *TagDao) CreateBatchCreate(ctx context.Context, tags []*po.Tag) error {
    return d.db.WithContext(ctx).Model(&po.Tag{}).Create(tags).Error
}

func (d *TagDao) ListAll(ctx context.Context) ([]*po.Tag, error) {
    var result []*po.Tag
    err := d.db.WithContext(ctx).Model(&po.Tag{}).Where("is_deleted = 0").Order("id ASC").Find(&result).Error
    if err != nil {
        return nil, err
    }
    return result, nil
}