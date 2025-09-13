package entity

// UserEntity 用户实体
type UserEntity struct {
	userUUID string
	account  string
	password string
}

// DefaultUserEntity 创建默认用户实体
func DefaultUserEntity(userUUID, account, password string) *UserEntity {
	return &UserEntity{
		userUUID: userUUID,
		account:  account,
		password: password,
	}
}

// NewUserEntity 创建用户实体
func NewUserEntity(userUUID, account, password string) *UserEntity {
	return &UserEntity{
		userUUID: userUUID,
		account:  account,
		password: password,
	}
}

// GetUserUUID 获取用户UUID
func (u *UserEntity) GetUserUUID() string {
	return u.userUUID
}

// GetAccount 获取账号
func (u *UserEntity) GetAccount() string {
	return u.account
}

// GetPassword 获取密码
func (u *UserEntity) GetPassword() string {
	return u.password
}

// SetPassword 设置密码
func (u *UserEntity) SetPassword(password string) {
	u.password = password
}