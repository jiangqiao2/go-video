package convertor

import (
	"transcode-service/ddd/domain/entity"
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

	return entity.NewTranscodeTaskEntity(
		po.TaskUUID, po.UserUUID, po.VideoUUID, po.InputPath, po.OutputPath)
}

// ToPO 将Entity转换为PO
func (c *TranscodeTaskConvertor) ToPO(entity *entity.TranscodeTaskEntity) *po.TranscodeTask {
	return &po.TranscodeTask{
		TaskUUID:   entity.TaskUUID(),
		UserUUID:   entity.UserUUID(),
		VideoUUID:  entity.VideoUUID(),
		InputPath:  entity.InputPath(),
		OutputPath: entity.OutputPath(),
		Status:     entity.Status().String(),
	}
}

// ToEntities 批量将PO转换为Entity
func (c *TranscodeTaskConvertor) ToEntities(pos []*po.TranscodeTask) []*entity.TranscodeTaskEntity {

}

// ToPOs 批量将Entity转换为PO
func (c *TranscodeTaskConvertor) ToPOs(entities []*entity.TranscodeTaskEntity) []*po.TranscodeTask {

}
