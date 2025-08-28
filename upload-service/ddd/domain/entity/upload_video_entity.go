package entity

import (
	"github.com/google/uuid"
	"time"
	"upload-service/ddd/domain/vo"
)

type UploadVideoEntity struct {
	uploadVideoUUid string               //  视频文件UUID
	userUUID        string               // 用户UUID
	fileName        string               // 视频文件名字
	fileSize        int                  // 视频文件大小
	fileHash        string               // 视频文件Hash值
	totalChunks     int                  // 总分片数量
	uploadedChunks  int                  // 目前已经上传的分片数量
	status          vo.UploadVideoStatus // 目前文件的上传状态
	storagePath     string               // Minio文件的存储路径
	completedAt     *time.Time           // 完成时间 可空
}

func DefaultUploadVideoEntity(userUUID, fileName string,
	fileSize int,
	fileHash string,
	totalChunks, uploadedChunks int,
	status vo.UploadVideoStatus,
	storagePath string,
	completedAt *time.Time) *UploadVideoEntity {
	return &UploadVideoEntity{
		uploadVideoUUid: uuid.NewString(),
		userUUID:        userUUID,
		fileName:        fileName,
		fileSize:        fileSize,
		fileHash:        fileHash,
		totalChunks:     totalChunks,
		uploadedChunks:  uploadedChunks,
		status:          status,
		storagePath:     storagePath,
		completedAt:     completedAt,
	}
}
func NewUploadVideoEntity(uploadVideoUUID string,
	userUUID string,
	fileName string,
	fileSize int,
	fileHash string,
	totalChunks int,
	uploadedChunks int,
	status vo.UploadVideoStatus,
	storagePath string,
	completedAt *time.Time) *UploadVideoEntity {
	return &UploadVideoEntity{
		uploadVideoUUid: uploadVideoUUID,
		userUUID:        userUUID,
		fileName:        fileName,
		fileSize:        fileSize,
		fileHash:        fileHash,
		totalChunks:     totalChunks,
		uploadedChunks:  uploadedChunks,
		status:          status,
		storagePath:     storagePath,
		completedAt:     completedAt,
	}
}

func (e *UploadVideoEntity) FileName() string {
	return e.fileName
}
func (e *UploadVideoEntity) FileSize() int {
	return e.fileSize
}
func (e *UploadVideoEntity) FileHash() string {
	return e.fileHash
}
func (e *UploadVideoEntity) CompletedAt() *time.Time {
	return e.completedAt
}
func (e *UploadVideoEntity) UserUUID() string {
	return e.userUUID
}
func (e *UploadVideoEntity) UploadVideoStatus() vo.UploadVideoStatus {
	return e.status
}
func (e *UploadVideoEntity) UploadVideoUUID() string {
	return e.uploadVideoUUid
}

func (e *UploadVideoEntity) TotalChunks() int {
	return e.totalChunks
}

func (e *UploadVideoEntity) Status() vo.UploadVideoStatus {
	return e.status
}

func (e *UploadVideoEntity) StoragePath() string {
	return e.storagePath
}

func (e *UploadVideoEntity) SetStoragePath(storagePath string) *UploadVideoEntity {
	e.storagePath = storagePath
	return e
}

type UploadChunkEntity struct {
	chunkUUID       string
	uploadVideoUUID string
	chunkIndex      int
	chunkHash       string
	chunkSize       int
	storagePath     string
	completedAt     *time.Time
	status          vo.UploadChunkStatus
}

func DefaultUploadChunkEntity(uploadVideoUUID string,
	chunkIndex int,
	chunkHash string,
	chunkSize int,
	storagePath string,
	completedAt *time.Time,
	status vo.UploadChunkStatus) *UploadChunkEntity {
	return &UploadChunkEntity{
		chunkUUID:       uuid.NewString(),
		uploadVideoUUID: uploadVideoUUID,
		chunkIndex:      chunkIndex,
		chunkHash:       chunkHash,
		chunkSize:       chunkSize,
		storagePath:     storagePath,
		completedAt:     completedAt,
		status:          status,
	}
}

func NewUploadChunkEntity(chunkUUID string,
	uploadVideoUUID string,
	chunkIndex int,
	chunkHash string,
	chunkSize int,
	storagePath string,
	completedAt *time.Time,
	status vo.UploadChunkStatus) *UploadChunkEntity {
	return &UploadChunkEntity{
		chunkUUID:       chunkUUID,
		uploadVideoUUID: uploadVideoUUID,
		chunkIndex:      chunkIndex,
		chunkHash:       chunkHash,
		chunkSize:       chunkSize,
		storagePath:     storagePath,
		completedAt:     completedAt,
		status:          status,
	}
}
func (e *UploadChunkEntity) ChunkUUID() string {
	return e.chunkUUID
}

func (e *UploadChunkEntity) UploadVideoUUID() string {
	return e.uploadVideoUUID
}

func (e *UploadChunkEntity) ChunkIndex() int {
	return e.chunkIndex
}

func (e *UploadChunkEntity) ChunkHash() string {
	return e.chunkHash
}

func (e *UploadChunkEntity) ChunkSize() int {
	return e.chunkSize
}

func (e *UploadChunkEntity) StoragePath() string {
	return e.storagePath
}

func (e *UploadChunkEntity) CompletedAt() *time.Time {
	return e.completedAt
}

func (e *UploadChunkEntity) Status() vo.UploadChunkStatus {
	return e.status
}
