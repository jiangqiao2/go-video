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
	Message  string `json:"message" example:"注册成功"`
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
	ExpiresIn    int64  `json:"expires_in" example:"7200"` // 秒
	Message      string `json:"message" example:"登录成功"`
}

// UserInfoResp 用户信息响应
type UserInfoResp struct {
	UserUUID string `json:"user_uuid" example:"550e8400-e29b-41d4-a716-446655440000"`
	Account  string `json:"account" example:"user123"`
	Message  string `json:"message" example:"查询成功"`
}
