package cqe

// CreateUserCommand 创建用户命令
type CreateUserCommand struct {
	Username string `json:"username" binding:"required,min=3,max=20" example:"john_doe"`
	Password string `json:"password" binding:"required,min=8,max=128" example:"password123"`
	Email    string `json:"email" binding:"required,email" example:"john@example.com"`
}

// LoginCommand 登录命令
type LoginCommand struct {
	Username string `json:"username" binding:"required" example:"john_doe"`
	Password string `json:"password" binding:"required" example:"password123"`
}

// UpdateProfileCommand 更新用户资料命令
type UpdateProfileCommand struct {
	UserID   uint64 `json:"-"` // 从JWT中获取，不需要绑定
	Nickname string `json:"nickname" binding:"max=50" example:"John Doe"`
	Avatar   string `json:"avatar" binding:"url" example:"https://example.com/avatar.jpg"`
}

// ChangePasswordCommand 修改密码命令
type ChangePasswordCommand struct {
	UserID      uint64 `json:"-"` // 从JWT中获取，不需要绑定
	OldPassword string `json:"old_password" binding:"required" example:"oldpassword123"`
	NewPassword string `json:"new_password" binding:"required,min=8,max=128" example:"newpassword123"`
}

// ActivateUserCommand 激活用户命令
type ActivateUserCommand struct {
	UserID uint64 `json:"user_id" binding:"required" example:"1"`
}

// DisableUserCommand 禁用用户命令
type DisableUserCommand struct {
	UserID uint64 `json:"user_id" binding:"required" example:"1"`
}

// DeleteUserCommand 删除用户命令
type DeleteUserCommand struct {
	UserID uint64 `json:"user_id" binding:"required" example:"1"`
}