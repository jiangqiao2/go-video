package cqe

// GetUserByIDQuery 根据ID获取用户查询
type GetUserByIDQuery struct {
	UserID uint64 `uri:"id" binding:"required" example:"1"`
}

// GetCurrentUserQuery 获取当前用户查询
type GetCurrentUserQuery struct {
	UserID uint64 `json:"-"` // 从JWT中获取，不需要绑定
}

// GetUserListQuery 获取用户列表查询
type GetUserListQuery struct {
	Page     int `form:"page" binding:"min=1" example:"1"`
	PageSize int `form:"page_size" binding:"min=1,max=100" example:"10"`
}

// ValidateTokenQuery 验证令牌查询
type ValidateTokenQuery struct {
	Token string `header:"Authorization" binding:"required"`
}
