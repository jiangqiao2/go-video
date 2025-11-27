package cqe

// FollowReq 关注/取关请求
type FollowReq struct {
	UserUUID   string `json:"-" form:"-"`
	TargetUUID string `json:"target_uuid" form:"target_uuid" binding:"required"`
}

type FollowStatusReq struct {
	UserUUID   string `json:"-" form:"-"`
	TargetUUID string `json:"target_uuid" form:"target_uuid"`
}

type FollowListQuery struct {
	TargetUUID string `form:"target_uuid"`
	Page       int    `form:"page"`
	Size       int    `form:"size"`
}

func (q *FollowListQuery) Normalize(defaultTarget string) {
	if q.TargetUUID == "" {
		q.TargetUUID = defaultTarget
	}
	if q.Page <= 0 {
		q.Page = 1
	}
	if q.Size <= 0 || q.Size > 200 {
		q.Size = 20
	}
}
