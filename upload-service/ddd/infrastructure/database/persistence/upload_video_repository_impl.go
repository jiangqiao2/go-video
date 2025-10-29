package persistence

import (
	"context"
	"upload-service/ddd/domain/entity"
	"upload-service/ddd/domain/repo"
	"upload-service/ddd/domain/vo"
	"upload-service/ddd/infrastructure/database/convertor"
	"upload-service/ddd/infrastructure/database/dao"
)

type uploadVideoRepositoryImpl struct {
	uploadVideoDao *dao.UploadVideoDao
	uploadChunkDao *dao.UploadChunkDao
}

func NewUploadVideoRepository() repo.UploadVideoRepository {
	return &uploadVideoRepositoryImpl{
		uploadVideoDao: dao.NewUploadVideoDao(),
		uploadChunkDao: dao.NewUploadChunkDao(),
	}
}

func (u *uploadVideoRepositoryImpl) CreateUploadVideoAndChunks(ctx context.Context, uploadVideoEntity *entity.UploadVideoEntity,
	uploadChunkEntitys []*entity.UploadChunkEntity) error {
	uploadVideoPo := convertor.ToUploadVideoPo(uploadVideoEntity)
	uploadChunkPos := convertor.ToUploadChunkArrPo(uploadChunkEntitys)
	return u.uploadVideoDao.BatchCreate(ctx, uploadVideoPo, uploadChunkPos)
}

func (u *uploadVideoRepositoryImpl) QueryUploadVideoByName(ctx context.Context, userUUID, fileName string, fileSize int, fileHash string) (*entity.UploadVideoEntity, []*entity.UploadChunkEntity, error) {
	uploadVideoPo, err := u.uploadVideoDao.QueryByFileNameAndHash(ctx, userUUID, fileName, fileSize, fileHash)
	if err != nil {
		return nil, nil, err
	}
	if uploadVideoPo == nil {
		return nil, nil, nil
	}
	
	// 查询所有分片，而不仅仅是Initialized状态的分片
	uploadChunkPos, err := u.uploadChunkDao.QueryByUploadVideoUUID(ctx, uploadVideoPo.UploadVideoUUID)
	if err != nil {
		return nil, nil, err
	}
	return convertor.ToUploadVideoEntity(uploadVideoPo), convertor.ToUploadChunkEntityArr(uploadChunkPos), nil
}

func (u *uploadVideoRepositoryImpl) QueryUploadVideoByUUID(ctx context.Context, uploadVideoUUID string) (*entity.UploadVideoEntity, error) {
	uploadVideoPo, err := u.uploadVideoDao.QueryByUUID(ctx, uploadVideoUUID)
	if err != nil {
		return nil, err
	}
	return convertor.ToUploadVideoEntity(uploadVideoPo), nil
}

func (u *uploadVideoRepositoryImpl) QueryUploadVideoByChunkUUID(ctx context.Context, query *repo.UploadChunkCheckQuery) (*entity.UploadChunkEntity, error) {
	uploadChunkPo, err := u.uploadChunkDao.QueryUploadVideoByUUID(ctx, query)
	if err != nil {
		return nil, err
	}
	return convertor.ToUploadChunkEntity(uploadChunkPo), nil
}

func (u *uploadVideoRepositoryImpl) UpdateUploadChunkStatus(ctx context.Context, uploadChunkUUID string, uploadChunkStatus vo.UploadChunkStatus) error {
	return u.uploadChunkDao.UpdateStatusByUUID(ctx, uploadChunkUUID, uploadChunkStatus.Value())
}

func (u *uploadVideoRepositoryImpl) QueryByUserAndUUID(ctx context.Context, uploadVideoUUID, userUUID string) (*entity.UploadVideoEntity, error) {
	uploadVideoPo, err := u.uploadVideoDao.QueryByUserUUIDAndUUID(ctx, uploadVideoUUID, userUUID)
	if err != nil {
		return nil, err
	}
	return convertor.ToUploadVideoEntity(uploadVideoPo), nil
}

func (u *uploadVideoRepositoryImpl) CountChunkByUploadVideoUUID(ctx context.Context, uploadVideoUUID, status string) (int64, error) {
	count, err := u.uploadChunkDao.CountChunk(ctx, uploadVideoUUID, status)
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (u *uploadVideoRepositoryImpl) UpdateUploadVideoStatus(ctx context.Context, uploadVideoUUID string, status vo.UploadVideoStatus) error {
	return u.uploadVideoDao.UpdateStatusByUUID(ctx, uploadVideoUUID, status.Value())
}

func (u *uploadVideoRepositoryImpl) QueryByStoragePath(ctx context.Context, userUUID, chunkUUID string) (string, error) {
	res, err := u.uploadChunkDao.QueryStoragePathByUUID(ctx, userUUID, chunkUUID, vo.UploadChunkStatusCompleted.Value())
	if err != nil {
		return "", err
	}
	return res, nil
}
