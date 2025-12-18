package entity

import "user-service/ddd/domain/vo"

type FollowEntity struct {
	userUUID       string
	targetUserUUID string
	status         vo.FollowStatusVo
}

func NewFollowEntity(userUUID string, targetUserUUID string, status string) *FollowEntity {
	return &FollowEntity{userUUID, targetUserUUID, vo.GetFollowStatus(status)}
}

func (f *FollowEntity) UserUUID() string {
	return f.userUUID
}

func (f *FollowEntity) TargetUserUUID() string {
	return f.targetUserUUID
}
func (f *FollowEntity) Status() vo.FollowStatusVo {
	return f.status
}
