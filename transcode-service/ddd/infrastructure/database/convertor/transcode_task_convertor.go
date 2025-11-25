package convertor

import (
	"transcode-service/ddd/domain/entity"
	"transcode-service/ddd/domain/vo"
	"transcode-service/ddd/infrastructure/database/po"
)

type TranscodeTaskConvertor struct{}

func NewTranscodeTaskConvertor() *TranscodeTaskConvertor {
	return &TranscodeTaskConvertor{}
}

func (c *TranscodeTaskConvertor) ToEntity(job *po.TranscodeJob) *entity.TranscodeTaskEntity {
	if job == nil {
		return nil
	}
	var params vo.TranscodeParams
	if p, err := vo.NewTranscodeParams(job.Resolution, job.Bitrate); err == nil && p != nil {
		params = *p
	} else {
		params = vo.TranscodeParams{Resolution: job.Resolution, Bitrate: job.Bitrate}
	}
	status, err := vo.NewTaskStatusFromString(job.Status)
	if err != nil {
		status = vo.TaskStatusPending
	}
	return entity.NewTranscodeTaskEntityWithDetails(
		job.Id,
		job.JobUUID,
		job.UserUUID,
		job.VideoUUID,
		job.InputPath,
		job.OutputPath,
		status,
		job.Progress,
		job.Message,
		params,
		job.CreatedAt,
		job.UpdatedAt,
	)
}

func (c *TranscodeTaskConvertor) ToPO(entity *entity.TranscodeTaskEntity) *po.TranscodeJob {
	return &po.TranscodeJob{
		BaseModel:  po.BaseModel{Id: entity.ID(), CreatedAt: entity.CreatedAt(), UpdatedAt: entity.UpdatedAt()},
		JobUUID:    entity.TaskUUID(),
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
}

func (c *TranscodeTaskConvertor) ToEntities(pos []*po.TranscodeJob) []*entity.TranscodeTaskEntity {
	if pos == nil {
		return nil
	}
	entities := make([]*entity.TranscodeTaskEntity, 0, len(pos))
	for _, job := range pos {
		if job != nil {
			entities = append(entities, c.ToEntity(job))
		}
	}
	return entities
}

func (c *TranscodeTaskConvertor) ToPOs(entities []*entity.TranscodeTaskEntity) []*po.TranscodeJob {
	if entities == nil {
		return nil
	}
	pos := make([]*po.TranscodeJob, 0, len(entities))
	for _, e := range entities {
		if e != nil {
			pos = append(pos, c.ToPO(e))
		}
	}
	return pos
}
