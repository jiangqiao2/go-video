package convertor

import (
	"transcode-service/ddd/domain/entity"
	"transcode-service/ddd/infrastructure/database/po"
)

// WorkerConvertor Worker转换器
type WorkerConvertor struct{}

// NewWorkerConvertor 创建Worker转换器
func NewWorkerConvertor() *WorkerConvertor {
	return &WorkerConvertor{}
}

// EntityToPO 实体转PO
func (c *WorkerConvertor) EntityToPO(entity *entity.WorkerEntity) (*po.WorkerPO, error) {
	if entity == nil {
		return nil, nil
	}
	
	// 转换系统信息
	systemInfoMap := make(po.JSONMap)
	for k, v := range entity.SystemInfo() {
		systemInfoMap[k] = v
	}
	
	// 转换元数据
	metadataMap := make(po.JSONMap)
	for k, v := range entity.Metadata() {
		metadataMap[k] = v
	}
	
	return &po.WorkerPO{
		WorkerID:        entity.WorkerID(),
		Name:            entity.Name(),
		Status:          entity.Status().String(),
		MaxTasks:        entity.MaxTasks(),
		CurrentTasks:    entity.CurrentTasks(),
		CPUUsage:        entity.CPUUsage(),
		MemoryUsage:     entity.MemoryUsage(),
		LastHeartbeatAt: entity.LastHeartbeatAt(),
		RegisteredAt:    entity.RegisteredAt(),
		UpdatedAt:       entity.UpdatedAt(),
		SystemInfo:      systemInfoMap,
		Metadata:        metadataMap,
	}, nil
}

// POToEntity PO转实体
func (c *WorkerConvertor) POToEntity(po *po.WorkerPO) (*entity.WorkerEntity, error) {
	if po == nil {
		return nil, nil
	}
	
	// 创建实体
	entity := entity.NewWorkerEntity(po.WorkerID, po.Name, po.MaxTasks)
	
	// 设置其他字段（需要通过反射或提供setter方法）
	// 这里简化处理，实际项目中可能需要更复杂的重建逻辑
	// 注意：需要设置status, currentTasks, cpuUsage等字段到entity中
	
	return entity, nil
}

// POListToEntityList PO列表转实体列表
func (c *WorkerConvertor) POListToEntityList(poList []*po.WorkerPO) ([]*entity.WorkerEntity, error) {
	if len(poList) == 0 {
		return nil, nil
	}
	
	entityList := make([]*entity.WorkerEntity, 0, len(poList))
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
func (c *WorkerConvertor) EntityListToPOList(entityList []*entity.WorkerEntity) ([]*po.WorkerPO, error) {
	if len(entityList) == 0 {
		return nil, nil
	}
	
	poList := make([]*po.WorkerPO, 0, len(entityList))
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