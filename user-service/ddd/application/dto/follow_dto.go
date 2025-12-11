package dto

// 关注状态
type FollowStatusDto struct {
	Following bool `json:"following"`
}

// 简单关注用户
type FollowUser struct {
	UserUUID  string `json:"user_uuid"`
	CreatedAt string `json:"created_at"`
}

// 关注/粉丝列表
type FollowListDto struct {
	List       []FollowUser `json:"list"`
	Size       int          `json:"size"`
	Total      int64        `json:"total"`
	NextCursor string       `json:"next_cursor,omitempty"`
}
