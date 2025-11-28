package po

import "time"

type FollowPo struct {
	BaseModel
	UserUUID   string    `gorm:"column:user_uuid"`
	TargetUUID string    `gorm:"column:target_uuid"`
	Status     string    `gorm:"column:status"`
	CreatedAt  time.Time `gorm:"column:created_at"`
	UpdatedAt  time.Time `gorm:"column:updated_at"`
}

func (FollowPo) TableName() string {
	return "user_follow"
}
