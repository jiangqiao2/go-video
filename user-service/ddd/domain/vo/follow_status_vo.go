package vo

type FollowStatusVo struct {
	value string
}

var arrStatus []FollowStatusVo = []FollowStatusVo{
	{"follow"},
	{"unfollow"},
}

func CheckFollowStatus(value string) bool {
	for _, v := range arrStatus {
		if v.value == value {
			return true
		}
	}
	return false
}

func GetFollowStatus(value string) FollowStatusVo {
	for _, v := range arrStatus {
		if v.value == value {
			return v
		}
	}
	return arrStatus[0]
}

// Value 返回原始字符串值，便于上层根据状态做分支。
func (v FollowStatusVo) Value() string {
	return v.value
}
