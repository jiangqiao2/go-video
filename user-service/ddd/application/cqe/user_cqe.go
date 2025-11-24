package cqe

// UserRegisterReq 用户注册请求
type UserRegisterReq struct {
	Account  string `json:"account" binding:"required,min=3,max=50" example:"user123"`
	Password string `json:"password" binding:"required,min=8,max=100" example:"Password123"`
}

// UserRegisterResp 用户注册响应
type UserRegisterResp struct {
	UserUUID string `json:"user_uuid" example:"550e8400-e29b-41d4-a716-446655440000"`
	Account  string `json:"account" example:"user123"`
}

// UserLoginReq 用户登录请求
type UserLoginReq struct {
	Account  string `json:"account" binding:"required" example:"user123"`
	Password string `json:"password" binding:"required" example:"Password123"`
}

// UserLoginResp 用户登录响应
type UserLoginResp struct {
	UserUUID     string `json:"user_uuid" example:"550e8400-e29b-41d4-a716-446655440000"`
	Account      string `json:"account" example:"user123"`
	AccessToken  string `json:"access_token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	RefreshToken string `json:"refresh_token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	ExpiresIn    int64  `json:"expires_in" example:"7200"`
	AvatarURL    string `json:"avatar_url" example:"image/avatar/user-550e..."`
}

type TokenRefreshReq struct {
	RefreshToken string `json:"refresh_token"`
}

type TokenRefreshResp struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
}

// UserInfoResp 用户信息响应
type UserInfoResp struct {
	UserUUID  string `json:"user_uuid" example:"550e8400-e29b-41d4-a716-446655440000"`
	Account   string `json:"account" example:"user123"`
	AvatarUrl string `json:"avatar_url" example:"image/avatar/user-550e..."`
}

// UserSaveReq 保存用户信息请求（字段可选，未提供的不更新）
type UserSaveReq struct {
	AvatarUrl *string `json:"avatar_url,omitempty" example:"image/avatar/user-550e..."`
}
