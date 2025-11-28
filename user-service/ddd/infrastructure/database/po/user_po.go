package po

type UserPo struct {
	BaseModel
	UserUUID    string `gorm:"column:user_uuid" json:"user_uuid"`
	Account     string `gorm:"column:account" json:"account"`
	Password    string `gorm:"column:password" json:"password"`
	Nickname    string `gorm:"column:nickname;type:varchar(64)" json:"nickname"`
	AvatarUrl   string `gorm:"column:avatar_url;type:varchar(512)" json:"avatar_url"`
	Description string `gorm:"column:description;type:varchar(512)" json:"description"`
	CoverUrl    string `gorm:"column:cover_url;type:varchar(512)" json:"cover_url"`
}

func (UserPo) TableName() string {
	return "user"
}
