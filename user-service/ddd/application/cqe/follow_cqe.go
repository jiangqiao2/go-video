package cqe

import (
	"strings"

	"user-service/ddd/domain/vo"
	"user-service/pkg/errno"
)

// FollowReq 关注/取关请求
type FollowReq struct {
	UserUUID       string `json:"-" form:"-"`
	TargetUUID     string `json:"target_uuid" form:"target_uuid"`
	TargetUserUUID string `json:"target_user_uuid" form:"target_user_uuid"`
}

// FollowToggleReq 关注/取关统一请求，通过 follow 字段控制是关注还是取消关注。
type FollowToggleReq struct {
	UserUUID       string `json:"-" form:"-"`
	TargetUUID     string `json:"target_uuid" form:"target_uuid"`
	TargetUserUUID string `json:"target_user_uuid" form:"target_user_uuid"`
	// Action 为操作类型：
	// - "follow"   表示关注
	// - "unfollow" 表示取消关注
	Action string `json:"action" form:"action"`
}

func (req *FollowToggleReq) Normalize() error {
	if req == nil || req.UserUUID == "" {
		return errno.ErrParameterInvalid
	}
	if req.TargetUUID == "" && req.TargetUserUUID != "" {
		req.TargetUUID = req.TargetUserUUID
	}
	if req.TargetUUID == "" {
		return errno.ErrParameterInvalid
	}
	if req.UserUUID == req.TargetUUID {
		return errno.ErrFollowSelf
	}
	req.Action = strings.ToLower(strings.TrimSpace(req.Action))
	if len(req.Action) <= 0 || !vo.CheckFollowStatus(req.Action) {
		return errno.ErrParameterInvalid
	}
	return nil
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
