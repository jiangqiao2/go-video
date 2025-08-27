package po

import "time"

// BaseModel 基础模型
type BaseModel struct {
	Id        uint64    `gorm:"primary_key;AUTO_INCREMENT;column:id" json:"-"`
	CreatedAt time.Time `gorm:"column:created_at" json:"-"`
	UpdatedAt time.Time `gorm:"column:updated_at" json:"-"`
	IsDeleted uint64    `gorm:"column:is_deleted" json:"-"`
}
