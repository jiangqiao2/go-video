package dto

import (
	"time"
	"transcode-service/ddd/domain/entity"
)

// TranscodeTaskDto 转码任务数据传输对象
type TranscodeTaskDto struct {
	TaskUUID     string             `json:"task_uuid"`
	UserUUID     string             `json:"user_uuid"`
	VideoUUID    string             `json:"video_uuid"`
	OriginalPath string             `json:"original_path"`
	OutputPath   string             `json:"output_path"`
	Status       string             `json:"status"`
	Progress     float64            `json:"progress"`
	ErrorMessage string             `json:"error_message,omitempty"`
	CreatedAt    time.Time          `json:"created_at"`
	UpdatedAt    time.Time          `json:"updated_at"`
	Params       TranscodeParamsDto `json:"params"`
	// 已拆分，HLS配置不再包含在转码任务DTO中
}

// HLS 已拆分为独立作业模型，相关 DTO 将在 hls_job_dto.go 中定义

// TranscodeParamsDto 转码参数数据传输对象
type TranscodeParamsDto struct {
	Resolution string `json:"resolution"`
	Bitrate    string `json:"bitrate"`
}

// TranscodeTaskListDto 转码任务列表数据传输对象
type TranscodeTaskListDto struct {
	Tasks      []TranscodeTaskDto `json:"tasks"`
	Total      int64              `json:"total"`
	Page       int                `json:"page"`
	Size       int                `json:"size"`
	TotalPages int                `json:"total_pages"`
}

// TranscodeProgressDto 转码进度数据传输对象
type TranscodeProgressDto struct {
	TaskUUID     string  `json:"task_uuid"`
	Status       string  `json:"status"`
	Progress     float64 `json:"progress"`
	ErrorMessage string  `json:"error_message,omitempty"`
}

// UpdateTranscodeTaskStatusDTO 更新转码任务状态DTO
type UpdateTranscodeTaskStatusDTO struct {
	Status       string `json:"status" binding:"required"`
	ErrorMessage string `json:"error_message,omitempty"`
}

// TranscodeTaskDTO 转码任务DTO（别名）
type TranscodeTaskDTO = TranscodeTaskDto

// NewTranscodeTaskDto 从实体创建DTO
func NewTranscodeTaskDto(entity *entity.TranscodeTaskEntity) *TranscodeTaskDto {
	if entity == nil {
		return nil
	}

	dto := &TranscodeTaskDto{
		TaskUUID:     entity.TaskUUID(),
		UserUUID:     entity.UserUUID(),
		VideoUUID:    entity.VideoUUID(),
		OriginalPath: entity.OriginalPath(),
		OutputPath:   entity.OutputPath(),
		Status:       entity.Status().String(),
		Progress:     float64(entity.Progress()),
		ErrorMessage: entity.ErrorMessage(),
		CreatedAt:    entity.CreatedAt(),
		UpdatedAt:    entity.UpdatedAt(),
		Params: TranscodeParamsDto{
			Resolution: entity.GetParams().Resolution,
			Bitrate:    entity.GetParams().Bitrate,
		},
	}

	// 已拆分：不在转码任务DTO中携带HLS配置

	return dto
}

// NewTranscodeTaskListDto 创建任务列表DTO
func NewTranscodeTaskListDto(entities []*entity.TranscodeTaskEntity, total int64, page, size int) *TranscodeTaskListDto {
	tasks := make([]TranscodeTaskDto, 0, len(entities))
	for _, entity := range entities {
		if dto := NewTranscodeTaskDto(entity); dto != nil {
			tasks = append(tasks, *dto)
		}
	}

	totalPages := int(total) / size
	if int(total)%size > 0 {
		totalPages++
	}

	return &TranscodeTaskListDto{
		Tasks:      tasks,
		Total:      total,
		Page:       page,
		Size:       size,
		TotalPages: totalPages,
	}
}

// NewTranscodeProgressDto 创建进度DTO
func NewTranscodeProgressDto(entity *entity.TranscodeTaskEntity) *TranscodeProgressDto {
	if entity == nil {
		return nil
	}

	return &TranscodeProgressDto{
		TaskUUID:     entity.TaskUUID(),
		Status:       entity.Status().String(),
		Progress:     float64(entity.Progress()),
		ErrorMessage: entity.ErrorMessage(),
	}
}
