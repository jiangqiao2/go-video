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

	// HLS相关配置（可选）
	EnableHLS       bool                  `json:"enable_hls"`       // 是否启用HLS切片
	HLSResolutions  []HLSResolutionConfig `json:"hls_resolutions"`  // HLS分辨率配置
	SegmentDuration int                   `json:"segment_duration"` // HLS切片时长（秒）
	ListSize        int                   `json:"list_size"`        // HLS播放列表大小
	HLSFormat       string                `json:"hls_format"`       // HLS格式
}

// HLSResolutionConfig HLS分辨率配置
type HLSResolutionConfig struct {
	Width   int    `json:"width" binding:"required"`   // 宽度
	Height  int    `json:"height" binding:"required"`  // 高度
	Bitrate string `json:"bitrate" binding:"required"` // 码率
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

	// 验证HLS配置
	if req.EnableHLS {
		if len(req.HLSResolutions) == 0 {
			return errno.ErrHLSResolutionsRequired
		}

		for _, res := range req.HLSResolutions {
			if res.Width <= 0 || res.Height <= 0 {
				return errno.ErrInvalidHLSResolution
			}
			if res.Bitrate == "" {
				return errno.ErrHLSBitrateRequired
			}
		}

		if req.SegmentDuration <= 0 {
			req.SegmentDuration = 4 // 默认4秒
		}
		if req.ListSize <= 0 {
			req.ListSize = 0 // 默认0（保留所有切片）
		}
		if req.HLSFormat == "" {
			req.HLSFormat = "hls" // 默认格式
		}
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
