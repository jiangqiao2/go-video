package dto

// UserDto 用户数据传输对象
type UserDto struct {
	UserUUID  string `json:"user_uuid"`
	Account   string `json:"account"`
	AvatarURL string `json:"avatar_url"`
}

// NewUserDto 创建用户DTO
func NewUserDto(userUUID, account, avatarURL string) *UserDto {
	return &UserDto{
		UserUUID:  userUUID,
		Account:   account,
		AvatarURL: avatarURL,
	}
}

// UserLoginDto 用户登录数据传输对象
type UserLoginDto struct {
	UserUUID     string `json:"user_uuid"`
	Account      string `json:"account"`
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
}

// NewUserLoginDto 创建用户登录DTO
func NewUserLoginDto(userUUID, account, accessToken, refreshToken string, expiresIn int64) *UserLoginDto {
	return &UserLoginDto{
		UserUUID:     userUUID,
		Account:      account,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    expiresIn,
	}
}
