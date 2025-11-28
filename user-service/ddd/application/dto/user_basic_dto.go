package dto

// UserBasicInfoDto 用户基本信息（公开）
type UserBasicInfoDto struct {
	UserUUID    string `json:"user_uuid"`
	Account     string `json:"account"`
	Nickname    string `json:"nickname,omitempty"`
	AvatarUrl   string `json:"avatar_url,omitempty"`
	Description string `json:"description,omitempty"`
	CoverUrl    string `json:"cover_url,omitempty"`
	CreatedAt   string `json:"created_at,omitempty"`
}

// UserRelationStatDto 用户关系统计
type UserRelationStatDto struct {
	UserUUID       string `json:"user_uuid"`
	FollowerCount  int64  `json:"follower_count"`  // 粉丝数
	FollowingCount int64  `json:"following_count"` // 关注数
	IsFollowed     bool   `json:"is_followed"`     // 当前用户是否已关注此用户
}
