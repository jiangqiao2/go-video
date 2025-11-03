package persistence

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"transcode-service/ddd/domain/entity"
	"transcode-service/ddd/domain/repo"
	"transcode-service/ddd/domain/vo"
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

func (t *transcodeRepositoryImpl) UpdateTranscodeTaskProgress(ctx context.Context, taskUUID string, progress int) error {
	return t.transTaskDao.UpdateTranscodeTaskProgress(ctx, taskUUID, progress)
}

func (t *transcodeRepositoryImpl) UpdateTranscodeTask(ctx context.Context, task *entity.TranscodeTaskEntity) error {
	taskPo := t.convertor.ToPO(task)
	return t.transTaskDao.UpdateTranscodeTask(ctx, taskPo)
}

func (t *transcodeRepositoryImpl) GetTranscodeTask(ctx context.Context, taskUUID string) (*entity.TranscodeTaskEntity, error) {
	taskPo, err := t.transTaskDao.FindByTaskUUID(ctx, taskUUID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
		return nil, err
	}
	return t.convertor.ToEntity(taskPo), nil
}

func (t *transcodeRepositoryImpl) UpdateTranscodeTaskStatus(ctx context.Context, taskUUID string, status vo.TaskStatus, message, outputPath string, progress int) error {
	return t.transTaskDao.UpdateTranscodeTaskStatus(ctx, taskUUID, status.String(), message, outputPath, progress)
}

func (t *transcodeRepositoryImpl) QueryTranscodeTasksByStatus(ctx context.Context, status vo.TaskStatus, limit int) ([]*entity.TranscodeTaskEntity, error) {
	taskPos, err := t.transTaskDao.QueryTranscodeTasksByStatus(ctx, status.String(), limit)
	if err != nil {
		return nil, err
	}
	return t.convertor.ToEntities(taskPos), nil
}

// HLS相关方法实现

func (t *transcodeRepositoryImpl) UpdateHLSProgress(ctx context.Context, taskUUID string, progress int) error {
	return t.transTaskDao.UpdateHLSProgress(ctx, taskUUID, progress)
}

func (t *transcodeRepositoryImpl) UpdateHLSStatus(ctx context.Context, taskUUID string, status string) error {
	return t.transTaskDao.UpdateHLSStatus(ctx, taskUUID, status)
}

func (t *transcodeRepositoryImpl) UpdateHLSOutputPath(ctx context.Context, taskUUID string, outputPath string) error {
	return t.transTaskDao.UpdateHLSOutputPath(ctx, taskUUID, outputPath)
}

func (t *transcodeRepositoryImpl) UpdateHLSError(ctx context.Context, taskUUID string, errorMessage string) error {
	return t.transTaskDao.UpdateHLSError(ctx, taskUUID, errorMessage)
}

func (t *transcodeRepositoryImpl) UpdateHLSCompleted(ctx context.Context, taskUUID string) error {
	return t.transTaskDao.UpdateHLSCompleted(ctx, taskUUID)
}

func (t *transcodeRepositoryImpl) UpdateHLSFailed(ctx context.Context, taskUUID string, errorMessage string) error {
	return t.transTaskDao.UpdateHLSFailed(ctx, taskUUID, errorMessage)
}

func (t *transcodeRepositoryImpl) QueryHLSEnabledTasks(ctx context.Context, status string, limit int) ([]*entity.TranscodeTaskEntity, error) {
	taskPos, err := t.transTaskDao.QueryHLSEnabledTasks(ctx, status, limit)
	if err != nil {
		return nil, err
	}
	return t.convertor.ToEntities(taskPos), nil
}
