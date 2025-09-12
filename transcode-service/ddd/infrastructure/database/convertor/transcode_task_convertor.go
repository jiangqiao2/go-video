package convertor

import (
	"encoding/json"
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

// EntityToPO 实体转PO
func (c *TranscodeTaskConvertor) EntityToPO(entity *entity.TranscodeTaskEntity) (*po.TranscodeTaskPO, error) {
	if entity == nil {
		return nil, nil
	}

	// 转换配置
	configMap := make(po.JSONMap)
	if entity.Config() != nil {
		configBytes, err := json.Marshal(entity.Config())
		if err != nil {
			return nil, err
		}
		err = json.Unmarshal(configBytes, &configMap)
		if err != nil {
			return nil, err
		}
	}

	// 转换元数据
	metadataMap := make(po.JSONMap)
	for k, v := range entity.Metadata() {
		metadataMap[k] = v
	}

	// 转换时间
	var estimatedTime, actualTime *int64
	if entity.EstimatedTime() != nil {
		t := int64(*entity.EstimatedTime())
		estimatedTime = &t
	}
	if entity.ActualTime() != nil {
		t := int64(*entity.ActualTime())
		actualTime = &t
	}

	return &po.TranscodeTaskPO{
		TaskID:          entity.TaskID(),
		UserID:          entity.UserID(),
		SourceVideoPath: entity.SourceVideoPath(),
		OutputPath:      entity.OutputPath(),
		Config:          configMap,
		Status:          entity.Status().String(),
		WorkerID:        entity.WorkerID(),
		Priority:        entity.Priority(),
		RetryCount:      entity.RetryCount(),
		MaxRetryCount:   entity.MaxRetryCount(),
		ErrorMessage:    entity.ErrorMessage(),
		Progress:        entity.Progress(),
		CreatedAt:       entity.CreatedAt(),
		UpdatedAt:       entity.UpdatedAt(),
		StartedAt:       entity.StartedAt(),
		CompletedAt:     entity.CompletedAt(),
		EstimatedTime:   estimatedTime,
		ActualTime:      actualTime,
		Metadata:        metadataMap,
	}, nil
}

// POToEntity PO转实体
func (c *TranscodeTaskConvertor) POToEntity(po *po.TranscodeTaskPO) (*entity.TranscodeTaskEntity, error) {
	if po == nil {
		return nil, nil
	}

	// 转换配置
	var config *vo.TranscodeConfig
	if len(po.Config) > 0 {
		configBytes, err := json.Marshal(po.Config)
		if err != nil {
			return nil, err
		}
		config = &vo.TranscodeConfig{}
		err = json.Unmarshal(configBytes, config)
		if err != nil {
			return nil, err
		}
	}

	// 转换元数据
	metadata := make(map[string]interface{})
	for k, v := range po.Metadata {
		metadata[k] = v
	}

	// 创建实体（使用反射或构造函数重建）
	entity := entity.NewTranscodeTaskEntity(
		po.UserID,
		po.SourceVideoPath,
		po.OutputPath,
		config,
		po.Priority,
		po.MaxRetryCount,
	)

	// 设置其他字段（需要通过反射或提供setter方法）
	// 这里简化处理，实际项目中可能需要更复杂的重建逻辑
	// 注意：estimatedTime 和 actualTime 在实际项目中需要通过setter设置到entity中

	return entity, nil
}

// POListToEntityList PO列表转实体列表
func (c *TranscodeTaskConvertor) POListToEntityList(poList []*po.TranscodeTaskPO) ([]*entity.TranscodeTaskEntity, error) {
	if len(poList) == 0 {
		return nil, nil
	}

	entityList := make([]*entity.TranscodeTaskEntity, 0, len(poList))
	for _, p := range poList {
		entity, err := c.POToEntity(p)
		if err != nil {
			return nil, err
		}
		if entity != nil {
			entityList = append(entityList, entity)
		}
	}

	return entityList, nil
}

// EntityListToPOList 实体列表转PO列表
func (c *TranscodeTaskConvertor) EntityListToPOList(entityList []*entity.TranscodeTaskEntity) ([]*po.TranscodeTaskPO, error) {
	if len(entityList) == 0 {
		return nil, nil
	}

	poList := make([]*po.TranscodeTaskPO, 0, len(entityList))
	for _, e := range entityList {
		po, err := c.EntityToPO(e)
		if err != nil {
			return nil, err
		}
		if po != nil {
			poList = append(poList, po)
		}
	}

	return poList, nil
}
