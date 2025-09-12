package po

type UserPo struct {
	BaseModel
	UserUUID string `gorm:"column:user_uuid" json:"user_uuid"`
	Account  string `gorm:"column:account" json:"account"`
	Password string `gorm:"column:password" json:"password"`
}

func (UserPo) TableName() string {
	return "user"
}
