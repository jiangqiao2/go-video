package cqe

// FollowReq 关注/取关请求
type FollowReq struct {
	UserUUID       string `json:"-" form:"-"`
	TargetUUID     string `json:"target_uuid" form:"target_uuid"`
	TargetUserUUID string `json:"target_user_uuid" form:"target_user_uuid"`
}

type FollowStatusReq struct {
	UserUUID       string `json:"-" form:"-"`
	TargetUUID     string `json:"target_uuid" form:"target_uuid"`
	TargetUserUUID string `json:"target_user_uuid" form:"target_user_uuid"`
}

type FollowListQuery struct {
	TargetUUID     string `form:"target_uuid"`
	TargetUserUUID string `form:"target_user_uuid"`
	// Cursor 为上一页最后一条记录的 "unixnano:id" 光标，空表示从最新开始
	Cursor string `form:"cursor"`
	Size   int    `form:"size"`
}

func (q *FollowListQuery) Normalize(defaultTarget string) {
	if q.TargetUUID == "" {
		q.TargetUUID = defaultTarget
	}
	if q.Size <= 0 || q.Size > 200 {
		q.Size = 20
	}
}
