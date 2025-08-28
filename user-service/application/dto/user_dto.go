package dto

import "time"

// CreateUserResponse 创建用户响应
type CreateUserResponse struct {
	UUID     string `json:"uuid" example:"550e8400-e29b-41d4-a716-446655440000"`
	Username string `json:"username" example:"john_doe"`
	Email    string `json:"email" example:"john@example.com"`
	Status   int    `json:"status" example:"2"`
}

// LoginResponse 登录响应
type LoginResponse struct {
	Token     string    `json:"token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	ExpiresAt time.Time `json:"expires_at" example:"2023-12-31T23:59:59Z"`
	User      *UserInfo `json:"user"`
}

// UserInfo 用户信息
type UserInfo struct {
	UUID     string `json:"uuid" example:"550e8400-e29b-41d4-a716-446655440000"`
	Username string `json:"username" example:"john_doe"`
	Email    string `json:"email" example:"john@example.com"`
	Nickname string `json:"nickname" example:"John"`
	Avatar   string `json:"avatar" example:"https://example.com/avatar.jpg"`
	Status   int    `json:"status" example:"1"`
}

// GetUserListResponse 获取用户列表响应
type GetUserListResponse struct {
	Users    []*UserInfo `json:"users"`
	Total    int64       `json:"total" example:"100"`
	Page     int         `json:"page" example:"1"`
	PageSize int         `json:"page_size" example:"10"`
}

// CommonResponse 通用响应
type CommonResponse struct {
	Message string `json:"message" example:"操作成功"`
}