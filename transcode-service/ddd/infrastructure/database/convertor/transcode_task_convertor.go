package convertor

import (
	"transcode-service/ddd/domain/entity"
	"transcode-service/ddd/domain/vo"
	"transcode-service/ddd/infrastructure/database/po"
)

// TranscodeTaskConvertor 转码任务转换器
type TranscodeTaskConvertor struct{}

// NewTranscodeTaskConvertor 创建转码任务转换器
func NewTranscodeTaskConvertor() *TranscodeTaskConvertor {
	return &TranscodeTaskConvertor{}
}

// ToEntity 将PO转换为Entity
func (c *TranscodeTaskConvertor) ToEntity(po *po.TranscodeTask) *entity.TranscodeTaskEntity {
	// 创建基础实体
	var params vo.TranscodeParams
	if p, err := vo.NewTranscodeParams(po.Resolution, po.Bitrate); err == nil && p != nil {
		params = *p
	} else {
		params = vo.TranscodeParams{
			Resolution: po.Resolution,
			Bitrate:    po.Bitrate,
		}
	}

	status, err := vo.NewTaskStatusFromString(po.Status)
	if err != nil {
		status = vo.TaskStatusPending
	}

	// 构建HLS配置
	var hlsConfig *vo.HLSConfig
	if po.HLSEnabled {
		hlsConfig = &vo.HLSConfig{
			EnableHLS:       true,
			SegmentDuration: po.HLSSegmentDuration,
			ListSize:        po.HLSListSize,
			Format:          po.HLSFormat,
		}
		
		if po.HLSStatus != nil {
			switch *po.HLSStatus {
			case "pending":
				hlsConfig.SetStatus(vo.HLSStatusPending)
			case "processing":
				hlsConfig.SetStatus(vo.HLSStatusProcessing)
			case "completed":
				hlsConfig.SetStatus(vo.HLSStatusCompleted)
			case "failed":
				hlsConfig.SetStatus(vo.HLSStatusFailed)
			}
		}
		
		hlsConfig.SetProgress(po.HLSProgress)
		
		if po.HLSOutputPath != nil {
			hlsConfig.SetOutputPath(*po.HLSOutputPath)
		}
		
		if po.HLSErrorMessage != nil {
			hlsConfig.SetErrorMessage(*po.HLSErrorMessage)
		}
		
		// 注意：HLSConfig 不包含时间戳字段，时间戳由数据库层管理
	} else {
		hlsConfig = vo.DefaultHLSConfig()
	}

	taskEntity := entity.NewTranscodeTaskEntityWithDetails(
		po.Id, // 传递数据库ID
		po.TaskUUID,
		po.UserUUID,
		po.VideoUUID,
		po.InputPath,
		po.OutputPath,
		status,
		int(po.Progress),
		po.Message,
		params,
		hlsConfig,
		po.CreatedAt,
		po.UpdatedAt,
	)

	return taskEntity
}

// ToPO 将Entity转换为PO
func (c *TranscodeTaskConvertor) ToPO(entity *entity.TranscodeTaskEntity) *po.TranscodeTask {
	taskPO := &po.TranscodeTask{
		BaseModel: po.BaseModel{
			Id:        entity.ID(),
			CreatedAt: entity.CreatedAt(),
			UpdatedAt: entity.UpdatedAt(),
		},
		TaskUUID:   entity.TaskUUID(),
		UserUUID:   entity.UserUUID(),
		VideoUUID:  entity.VideoUUID(),
		InputPath:  entity.InputPath(),
		OutputPath: entity.OutputPath(),
		Resolution: entity.GetParams().Resolution,
		Bitrate:    entity.GetParams().Bitrate,
		Status:     entity.Status().String(),
		Message:    entity.ErrorMessage(),
		Progress:   entity.Progress(),
	}

	// 处理HLS字段
	hlsConfig := entity.GetHLSConfig()
	if hlsConfig != nil {
		taskPO.HLSEnabled = hlsConfig.IsEnabled()
		taskPO.HLSSegmentDuration = hlsConfig.SegmentDuration
		taskPO.HLSListSize = hlsConfig.ListSize
		taskPO.HLSFormat = hlsConfig.Format
		taskPO.HLSProgress = hlsConfig.Progress
		
		if hlsConfig.IsEnabled() {
			status := hlsConfig.Status.String()
			taskPO.HLSStatus = &status
			
			if outputPath := hlsConfig.OutputPath; outputPath != "" {
				taskPO.HLSOutputPath = &outputPath
			}
			
			if errorMsg := hlsConfig.ErrorMessage; errorMsg != "" {
				taskPO.HLSErrorMessage = &errorMsg
			}
			
			// 注意：HLSConfig 不包含时间戳字段，时间戳由数据库层管理
		}
	}

	return taskPO
}

// ToEntities 批量将PO转换为Entity
func (c *TranscodeTaskConvertor) ToEntities(pos []*po.TranscodeTask) []*entity.TranscodeTaskEntity {
	if pos == nil {
		return nil
	}

	entities := make([]*entity.TranscodeTaskEntity, 0, len(pos))
	for _, po := range pos {
		if po != nil {
			entity := c.ToEntity(po)
			entities = append(entities, entity)
		}
	}

	return entities
}

// ToPOs 批量将Entity转换为PO
func (c *TranscodeTaskConvertor) ToPOs(entities []*entity.TranscodeTaskEntity) []*po.TranscodeTask {
	if entities == nil {
		return nil
	}

	pos := make([]*po.TranscodeTask, 0, len(entities))
	for _, entity := range entities {
		if entity != nil {
			po := c.ToPO(entity)
			pos = append(pos, po)
		}
	}

	return pos
}
