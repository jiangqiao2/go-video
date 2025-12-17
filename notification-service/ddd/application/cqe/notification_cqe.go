package cqe

// ListNotificationsReq 列表查询请求。
type ListNotificationsReq struct {
	Page     int `form:"page"`
	PageSize int `form:"page_size"`
}

func (r *ListNotificationsReq) Normalize() {
	if r.Page <= 0 {
		r.Page = 1
	}
	if r.PageSize <= 0 || r.PageSize > 100 {
		r.PageSize = 20
	}
}

// MarkReadReq 标记已读请求。
type MarkReadReq struct {
	IDs []uint64 `json:"ids"`
}

func (r *MarkReadReq) Validate() bool {
	return len(r.IDs) > 0
}

// CreateNotificationReq 创建通知请求（内部接口使用）。
type CreateNotificationReq struct {
	UserUUID  string `json:"user_uuid"`
	Type      string `json:"type"`
	Title     string `json:"title"`
	Content   string `json:"content"`
	ExtraJSON string `json:"extra_json,omitempty"`
}

// Validate 校验必填字段是否完整。
func (r *CreateNotificationReq) Validate() bool {
	if r == nil {
		return false
	}
	return r.UserUUID != "" && r.Type != "" && r.Title != "" && r.Content != ""
}
