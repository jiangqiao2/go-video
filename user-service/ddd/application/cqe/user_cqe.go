package cqe

// UserRegisterReq 用户注册请求
type UserRegisterReq struct {
	Account  string `json:"account" binding:"required,min=3,max=50" example:"user123"`
	Password string `json:"password" binding:"required,min=8,max=100" example:"Password123"`
}

// UserLoginReq 用户登录请求
type UserLoginReq struct {
	Account  string `json:"account" binding:"required" example:"user123"`
	Password string `json:"password" binding:"required" example:"Password123"`
}

type TokenRefreshReq struct {
	RefreshToken string `json:"refresh_token"`
}

// UserSaveReq 保存用户信息请求（字段可选，未提供的不更新）
type UserSaveReq struct {
	AvatarUrl string `json:"avatar_url,omitempty" example:"image/avatar/user-550e..."`
}
