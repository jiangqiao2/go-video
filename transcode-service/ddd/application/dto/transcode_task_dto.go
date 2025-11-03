package dto

import (
	"strconv"
	"strings"
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
	HLSConfig    *HLSConfigDto      `json:"hls_config,omitempty"` // HLS配置
}

// HLSConfigDto HLS配置DTO
type HLSConfigDto struct {
	Enabled         bool                      `json:"enabled"`
	Resolutions     []HLSResolutionConfigDto  `json:"resolutions,omitempty"`
	SegmentDuration int                       `json:"segment_duration,omitempty"`
	ListSize        int                       `json:"list_size,omitempty"`
	Format          string                    `json:"format,omitempty"`
	Status          string                    `json:"status,omitempty"`
	Progress        int                       `json:"progress,omitempty"`
	OutputPath      string                    `json:"output_path,omitempty"`
	ErrorMessage    string                    `json:"error_message,omitempty"`
}

// HLSResolutionConfigDto HLS分辨率配置DTO
type HLSResolutionConfigDto struct {
	Width   int    `json:"width"`
	Height  int    `json:"height"`
	Bitrate string `json:"bitrate"`
}

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

	// 添加HLS配置
	if hlsConfig := entity.GetHLSConfig(); hlsConfig != nil {
		hlsDto := &HLSConfigDto{
			Enabled:         hlsConfig.IsEnabled(),
			SegmentDuration: hlsConfig.SegmentDuration,
			ListSize:        hlsConfig.ListSize,
			Format:          hlsConfig.Format,
			Status:          hlsConfig.Status.String(),
			Progress:        hlsConfig.Progress,
			OutputPath:      hlsConfig.OutputPath,
			ErrorMessage:    hlsConfig.ErrorMessage,
		}

		// 转换分辨率配置
		resolutions := hlsConfig.Resolutions
		if len(resolutions) > 0 {
			hlsDto.Resolutions = make([]HLSResolutionConfigDto, len(resolutions))
			for i, res := range resolutions {
				// 从分辨率字符串解析宽高
				heightStr := strings.TrimSuffix(res.Resolution, "p")
				height, _ := strconv.Atoi(heightStr)
				width := height * 16 / 9 // 假设16:9比例
				
				hlsDto.Resolutions[i] = HLSResolutionConfigDto{
					Width:   width,
					Height:  height,
					Bitrate: res.Bitrate,
				}
			}
		}

		dto.HLSConfig = hlsDto
	}

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
