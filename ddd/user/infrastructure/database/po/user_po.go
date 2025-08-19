package po

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// UserPO 用户持久化对象
type UserPO struct {
	BaseModel
	UUID     string `gorm:"uniqueIndex;size:36;not null;column:uuid" json:"uuid"`
	Username string `gorm:"uniqueIndex;size:50;not null" json:"username"`
	Password string `gorm:"size:255;not null;column:password_hash" json:"-"`
	Email    string `gorm:"uniqueIndex;size:100;not null" json:"email"`
	Nickname string `gorm:"size:100" json:"nickname"`
	Avatar   string `gorm:"size:255" json:"avatar"`
	Status   int    `gorm:"default:2;not null" json:"status"` // 1:激活 2:未激活 3:禁用 4:删除
}

// BeforeCreate GORM钩子：创建前自动生成UUID
func (u *UserPO) BeforeCreate(tx *gorm.DB) error {
	if u.UUID == "" {
		u.UUID = uuid.New().String()
	}
	return nil
}

// TableName 指定表名
func (UserPO) TableName() string {
	return "users"
}