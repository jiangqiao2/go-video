package cqe

import "transcode-service/pkg/errno"

// TranscodeTaskCqe 转码任务CQE（别名）
type TranscodeTaskCqe = CreateTranscodeTaskReq

// CreateTranscodeTaskReq 创建转码任务请求
type CreateTranscodeTaskReq struct {
	UserUUID     string `json:"user_uuid" binding:"required"`     // 用户UUID
	VideoUUID    string `json:"video_uuid" binding:"required"`    // 视频UUID
	OriginalPath string `json:"original_path" binding:"required"` // 原始视频路径
	Resolution   string `json:"resolution" binding:"required"`    // 转码分辨率
	Bitrate      string `json:"bitrate" binding:"required"`       // 转码码率
}

func (req *CreateTranscodeTaskReq) Validate() error {
	if req.UserUUID == "" {
		return errno.ErrUserUUIDRequired
	}
	if req.VideoUUID == "" {
		return errno.ErrVideoUUIDRequired
	}
	if req.OriginalPath == "" {
		return errno.ErrOriginalPathRequired
	}
	if req.Resolution == "" {
		return errno.ErrResolutionRequired
	}
	if req.Bitrate == "" {
		return errno.ErrBitrateRequired
	}
	return nil
}

// QueryTranscodeTaskReq 查询转码任务请求
type QueryTranscodeTaskReq struct {
	TaskUUID string `uri:"task_uuid" binding:"required"`
	UserUUID string `header:"X-User-UUID" binding:"required"`
}

func (req *QueryTranscodeTaskReq) Validate() error {
	if req.TaskUUID == "" {
		return errno.ErrTaskUUIDRequired
	}
	if req.UserUUID == "" {
		return errno.ErrUserUUIDRequired
	}
	return nil
}

// UpdateTranscodeTaskStatusReq 更新转码任务状态请求
type UpdateTranscodeTaskStatusReq struct {
	TaskUUID string `uri:"task_uuid" binding:"required"`
	UserUUID string `header:"X-User-UUID" binding:"required"`
	Status   string `json:"status" binding:"required"`
}

func (req *UpdateTranscodeTaskStatusReq) Validate() error {
	if req.TaskUUID == "" {
		return errno.ErrTaskUUIDRequired
	}
	if req.UserUUID == "" {
		return errno.ErrUserUUIDRequired
	}
	if req.Status == "" {
		return errno.ErrStatusRequired
	}
	return nil
}

// ListTranscodeTasksReq 列表转码任务请求
type ListTranscodeTasksReq struct {
	UserUUID  string `header:"X-User-UUID" binding:"required"`
	VideoUUID string `form:"video_uuid"`
	Status    string `form:"status"`
	Page      int    `form:"page" binding:"min=1"`
	Size      int    `form:"size" binding:"min=1,max=100"`
}

func (req *ListTranscodeTasksReq) Validate() error {
	if req.UserUUID == "" {
		return errno.ErrUserUUIDRequired
	}
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.Size <= 0 || req.Size > 100 {
		req.Size = 10
	}
	return nil
}

// CancelTranscodeTaskReq 取消转码任务请求
type CancelTranscodeTaskReq struct {
	TaskUUID string `uri:"task_uuid" binding:"required"`
	UserUUID string `header:"X-User-UUID" binding:"required"`
}

func (req *CancelTranscodeTaskReq) Validate() error {
	if req.TaskUUID == "" {
		return errno.ErrTaskUUIDRequired
	}
	if req.UserUUID == "" {
		return errno.ErrUserUUIDRequired
	}
	return nil
}

// GetTranscodeProgressReq 获取转码进度请求
type GetTranscodeProgressReq struct {
	TaskUUID string `uri:"task_uuid" binding:"required"`
	UserUUID string `header:"X-User-UUID" binding:"required"`
}

func (req *GetTranscodeProgressReq) Validate() error {
	if req.TaskUUID == "" {
		return errno.ErrTaskUUIDRequired
	}
	if req.UserUUID == "" {
		return errno.ErrUserUUIDRequired
	}
	return nil
}
