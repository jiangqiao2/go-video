package dto

// 用户相关响应 DTO

type UserRegisterDto struct {
	UserUUID string `json:"user_uuid" example:"550e8400-e29b-41d4-a716-446655440000"`
	Account  string `json:"account" example:"user123"`
}

type UserLoginDto struct {
	UserUUID     string `json:"user_uuid" example:"550e8400-e29b-41d4-a716-446655440000"`
	Account      string `json:"account" example:"user123"`
	AccessToken  string `json:"access_token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	RefreshToken string `json:"refresh_token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	ExpiresIn    int64  `json:"expires_in" example:"7200"`
	AvatarURL    string `json:"avatar_url" example:"image/avatar/user-550e..."`
}

type TokenRefreshDto struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
}

type UserInfoDto struct {
	UserUUID  string `json:"user_uuid" example:"550e8400-e29b-41d4-a716-446655440000"`
	Nickname  string `json:"nickname,omitempty" example:"昵称"`
	AvatarUrl string `json:"avatar_url" example:"image/avatar/user-550e..."`
}
