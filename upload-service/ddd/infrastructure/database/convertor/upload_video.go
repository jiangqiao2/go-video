package convertor

import (
	"upload-service/ddd/domain/entity"
	"upload-service/ddd/domain/vo"
	"upload-service/ddd/infrastructure/database/po"
)

func ToUploadVideoPo(uploadVideoEntity *entity.UploadVideoEntity) *po.UploadVideoPo {
	return &po.UploadVideoPo{
		UploadVideoUUID:  uploadVideoEntity.UploadVideoUUID(),
		UserUUID:         uploadVideoEntity.UserUUID(),
		FileName:         uploadVideoEntity.FileName(),
		FileSize:         uploadVideoEntity.FileSize(),
		FileHash:         uploadVideoEntity.FileHash(),
		TotalChunks:      uploadVideoEntity.TotalChunks(),
		Status:           uploadVideoEntity.Status().Value(),
		StoragePath:      uploadVideoEntity.StoragePath(),
		CompletedTime:    uploadVideoEntity.CompletedAt(),
		ChunkStoragePath: uploadVideoEntity.ChunkStoragePath(),
	}
}

func ToUploadChunkPo(uploadChunkEntity *entity.UploadChunkEntity) *po.UploadChunkPo {
	return &po.UploadChunkPo{
		ChunkUUID:        uploadChunkEntity.ChunkUUID(),
		ChunkSize:        uploadChunkEntity.ChunkSize(),
		ChunkHash:        uploadChunkEntity.ChunkHash(),
		Status:           uploadChunkEntity.Status().Value(),
		StoragePath:      uploadChunkEntity.StoragePath(),
		CompletedTime:    uploadChunkEntity.CompletedAt(),
		UploadVideoUUID:  uploadChunkEntity.UploadVideoUUID(),
		ChunkIndex:       uploadChunkEntity.ChunkIndex(),
		PutURL:           uploadChunkEntity.PutURL(),
		PresignExpiredAt: uploadChunkEntity.PresignExpiredAt(),
	}
}

func ToUploadChunkArrPo(uploadChunkEntityArr []*entity.UploadChunkEntity) []*po.UploadChunkPo {
	uploadChunkPos := make([]*po.UploadChunkPo, 0, len(uploadChunkEntityArr))
	for _, uploadChunk := range uploadChunkEntityArr {
		uploadChunkPos = append(uploadChunkPos, ToUploadChunkPo(uploadChunk))
	}
	return uploadChunkPos
}

func ToUploadVideoEntity(uploadVideoPo *po.UploadVideoPo) *entity.UploadVideoEntity {
	if uploadVideoPo == nil {
		return nil
	}
	return entity.NewUploadVideoEntity(
		uploadVideoPo.UploadVideoUUID,
		uploadVideoPo.UserUUID,
		uploadVideoPo.FileName,
		uploadVideoPo.FileSize,
		uploadVideoPo.FileHash,
		uploadVideoPo.TotalChunks,
		uploadVideoPo.UploadedChunks,
		vo.NewUploadVideoStatus(uploadVideoPo.Status),
		uploadVideoPo.StoragePath,
		uploadVideoPo.CompletedTime,
		uploadVideoPo.ChunkStoragePath)
}

func ToUploadChunkEntity(uploadChunkPo *po.UploadChunkPo) *entity.UploadChunkEntity {
	if uploadChunkPo == nil {
		return nil
	}
	return entity.NewUploadChunkEntity(
		uploadChunkPo.ChunkUUID,
		uploadChunkPo.UploadVideoUUID,
		uploadChunkPo.ChunkIndex,
		uploadChunkPo.ChunkHash,
		uploadChunkPo.ChunkSize,
		uploadChunkPo.StoragePath,
		uploadChunkPo.CompletedTime,
		vo.NewUploadChunkStatus(uploadChunkPo.Status),
		uploadChunkPo.PutURL,
		uploadChunkPo.PresignExpiredAt,
	)
}

func ToUploadChunkEntityArr(uploadChunkPos []*po.UploadChunkPo) []*entity.UploadChunkEntity {
	if len(uploadChunkPos) == 0 {
		return nil
	}
	result := make([]*entity.UploadChunkEntity, 0, len(uploadChunkPos))
	for _, uploadChunkPo := range uploadChunkPos {
		result = append(result, ToUploadChunkEntity(uploadChunkPo))
	}
	return result
}
