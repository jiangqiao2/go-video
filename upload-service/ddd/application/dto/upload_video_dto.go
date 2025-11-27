package dto

import "upload-service/ddd/domain/entity"

type UploadVideoDto struct {
	UploadVideoUUID string           `json:"upload_video_uuid"`
	ChunkSize       int              `json:"chunk_size"`
	TotalChunks     int              `json:"total_chunks"`
	Status          string           `json:"status"` // 添加上传视频状态
	UploadChunks    []UploadChunkDto `json:"upload_chunks"`
}

type UploadImageDto struct {
	Bucket string `json:"bucket"`
	Key    string `json:"key"`
	URL    string `json:"url"`
}

func NewUpadVideoDto(uploadVideoEntity *entity.UploadVideoEntity, uploadChunkEntitys []*entity.UploadChunkEntity) *UploadVideoDto {
	uploadChunks := make([]UploadChunkDto, 0, len(uploadChunkEntitys))
	for _, v := range uploadChunkEntitys {
		uploadChunks = append(uploadChunks, UploadChunkDto{
			ChunkUUID:   v.ChunkUUID(),
			ChunkIndex:  v.ChunkIndex(),
			Status:      v.Status().Value(),
			StoragePath: v.StoragePath(),
		})
	}
	return &UploadVideoDto{
		UploadVideoUUID: uploadVideoEntity.UploadVideoUUID(),
		ChunkSize:       5242880, // 5MB chunk size (MinIO ComposeObject minimum requirement)
		TotalChunks:     uploadVideoEntity.TotalChunks(),
		Status:          uploadVideoEntity.Status().Value(), // 包含上传视频状态
		UploadChunks:    uploadChunks,
	}
}

type UploadVideoChunkDto struct {
	Status string `json:"status"`
}

type MergeChunkDto struct {
	UploadVideoUUID string `json:"upload_video_uuid"`
	Status          string `json:"status"`
}

type UploadVideoStoragePathDto struct {
	StoragePath string `json:"storage_path"`
}

type UploadChunkDto struct {
	ChunkUUID    string `json:"chunk_uuid"`
	ChunkIndex   int    `json:"chunk_index"`
	Status       string `json:"status"`
	StoragePath  string `json:"storage_path"`
	PutURL       string `json:"put_url,omitempty"`
	ExpiresInSec int    `json:"expires_seconds,omitempty"`
}

type UploadVideoStatusDto struct {
	UploadVideoUUID string `json:"upload_video_uuid"`
	Status          string `json:"status"`
}

type PresignImageDto struct {
	Bucket string `json:"bucket"`
	Key    string `json:"key"`
	PutURL string `json:"put_url"`
}

type PresignChunkDto struct {
	UploadVideoUUID string `json:"upload_video_uuid"`
	ChunkUUID       string `json:"chunk_uuid"`
	ChunkIndex      int    `json:"chunk_index"`
	Bucket          string `json:"bucket"`
	Key             string `json:"key"`
	PutURL          string `json:"put_url"`
	ExpiresSeconds  int    `json:"expires_seconds"`
}
