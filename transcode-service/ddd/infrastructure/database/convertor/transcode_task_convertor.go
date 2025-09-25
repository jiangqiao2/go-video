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
func (c *TranscodeTaskConvertor) ToEntity(po *po.TranscodeTask) (*entity.TranscodeTaskEntity, error) {
	if po == nil {
		return nil, nil
	}

	// 创建转码参数
	params, err := vo.NewTranscodeParams(po.Resolution, po.Bitrate)
	if err != nil {
		return nil, err
	}

	// 创建任务状态
	status, err := vo.NewTaskStatusFromString(po.Status)
	if err != nil {
		return nil, err
	}

	// 恢复转码任务实体
	task := entity.NewTranscodeTaskEntity(
		po.TaskUUID,
		po.VideoUUID,
		po.UserUUID,
		po.InputPath,
		po.OutputPath,
		*params,
		status,
		int(po.Progress),
		po.Message,
		po.CreatedAt,
		po.UpdatedAt,
		nil, // completedAt
	)

	return task, nil
}

// ToPO 将Entity转换为PO
func (c *TranscodeTaskConvertor) ToPO(entity *entity.TranscodeTaskEntity) *po.TranscodeTask {
	if entity == nil {
		return nil
	}

	return &po.TranscodeTask{
		BaseModel: po.BaseModel{
			CreatedAt: entity.CreatedAt(),
			UpdatedAt: entity.UpdatedAt(),
			IsDeleted: 0,
		},
		TaskUUID:   entity.TaskUUID(),
		UserUUID:   entity.UserUUID(),
		VideoUUID:  entity.VideoUUID(),
		InputPath:  entity.OriginalPath(),
		OutputPath: entity.OutputPath(),
		Resolution: entity.Params().Resolution,
		Bitrate:    entity.Params().Bitrate,
		Status:     entity.Status().String(),
		Progress:   float64(entity.Progress()),
		Message:    entity.ErrorMessage(),
	}
}

// ToEntities 批量将PO转换为Entity
func (c *TranscodeTaskConvertor) ToEntities(pos []*po.TranscodeTask) ([]*entity.TranscodeTaskEntity, error) {
	if len(pos) == 0 {
		return nil, nil
	}

	entities := make([]*entity.TranscodeTaskEntity, 0, len(pos))
	for _, p := range pos {
		entity, err := c.ToEntity(p)
		if err != nil {
			return nil, err
		}
		if entity != nil {
			entities = append(entities, entity)
		}
	}

	return entities, nil
}

// ToPOs 批量将Entity转换为PO
func (c *TranscodeTaskConvertor) ToPOs(entities []*entity.TranscodeTaskEntity) []*po.TranscodeTask {
	if len(entities) == 0 {
		return nil
	}

	pos := make([]*po.TranscodeTask, 0, len(entities))
	for _, e := range entities {
		po := c.ToPO(e)
		if po != nil {
			pos = append(pos, po)
		}
	}

	return pos
}
