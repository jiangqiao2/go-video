package persistence

import (
	"context"
	"transcode-service/ddd/domain/entity"
	"transcode-service/ddd/domain/repo"
	"transcode-service/ddd/infrastructure/database/convertor"
	"transcode-service/ddd/infrastructure/database/dao"
)

type transcodeRepositoryImpl struct {
	transTaskDao *dao.TranscodeTaskDAO
	convertor    *convertor.TranscodeTaskConvertor
}

func NewTranscodeRepository() repo.TranscodeTaskRepository {
	return &transcodeRepositoryImpl{
		transTaskDao: dao.NewTranscodeTaskDAO(),
		convertor:    convertor.NewTranscodeTaskConvertor(),
	}
}

func (t *transcodeRepositoryImpl) CreateTranscodeTask(ctx context.Context, task *entity.TranscodeTaskEntity) error {
	// 将实体转换为PO
	taskPo := t.convertor.ToPO(task)
	
	// 调用DAO创建任务
	return t.transTaskDao.Create(ctx, taskPo)
}
