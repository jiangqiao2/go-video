package dto

import "upload-service/ddd/domain/entity"

type UploadVideoDto struct {
	UploadVideoUUID string `json:"upload_video_uuid"`
	ChunkSize       int    `json:"chunk_size"`
	TotalChunks     int    `json:"total_chunks"`
	UploadChunks    []int  `json:"upload_chunks"`
}

func NewUpadVideoDto(uploadVideoEntity *entity.UploadVideoEntity, uploadChunkEntitys []*entity.UploadChunkEntity) *UploadVideoDto {
	uploadChunks := make([]int, 0, len(uploadChunkEntitys))
	for _, v := range uploadChunkEntitys {
		uploadChunks = append(uploadChunks, v.ChunkIndex())
	}
	return &UploadVideoDto{
		UploadVideoUUID: uploadVideoEntity.UploadVideoUUID(),
		UploadChunks:    uploadChunks,
	}
}
