package entity

import (
	"time"

	"github.com/google/uuid"

	"upload-service/ddd/domain/vo"
)

type UploadVideoEntity struct {
	uploadVideoUUID  string
	userUUID         string
	fileName         string
	fileSize         int
	fileHash         string
	totalChunks      int
	uploadedChunks   int
	status           vo.UploadVideoStatus
	storagePath      string
	chunkStoragePath string
	completedAt      *time.Time
}

func DefaultUploadVideoEntity(
	userUUID string,
	fileName string,
	fileSize int,
	fileHash string,
	totalChunks int,
	uploadedChunks int,
	status vo.UploadVideoStatus,
	storagePath string,
	completedAt *time.Time,
) *UploadVideoEntity {
	return &UploadVideoEntity{
		uploadVideoUUID: uuid.NewString(),
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

func NewUploadVideoEntity(
	uploadVideoUUID string,
	userUUID string,
	fileName string,
	fileSize int,
	fileHash string,
	totalChunks int,
	uploadedChunks int,
	status vo.UploadVideoStatus,
	storagePath string,
	completedAt *time.Time,
	chunkStoragePath string,
) *UploadVideoEntity {
	return &UploadVideoEntity{
		uploadVideoUUID:  uploadVideoUUID,
		userUUID:         userUUID,
		fileName:         fileName,
		fileSize:         fileSize,
		fileHash:         fileHash,
		totalChunks:      totalChunks,
		uploadedChunks:   uploadedChunks,
		status:           status,
		storagePath:      storagePath,
		completedAt:      completedAt,
		chunkStoragePath: chunkStoragePath,
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
	return e.uploadVideoUUID
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

func (e *UploadVideoEntity) ChunkStoragePath() string {
	return e.chunkStoragePath
}

func (e *UploadVideoEntity) SetStoragePath(storagePath string) *UploadVideoEntity {
	e.storagePath = storagePath
	return e
}

func (e *UploadVideoEntity) SetChunkStoragePath(chunkStoragePath string) *UploadVideoEntity {
	e.chunkStoragePath = chunkStoragePath
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
	putURL          string
	presignExpiredAt *time.Time
}

func DefaultUploadChunkEntity(
	uploadVideoUUID string,
	chunkIndex int,
	chunkHash string,
	chunkSize int,
	storagePath string,
	completedAt *time.Time,
	status vo.UploadChunkStatus,
) *UploadChunkEntity {
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

func NewUploadChunkEntity(
	chunkUUID string,
	uploadVideoUUID string,
	chunkIndex int,
	chunkHash string,
	chunkSize int,
	storagePath string,
	completedAt *time.Time,
	status vo.UploadChunkStatus,
	putURL string,
	presignExpiredAt *time.Time,
) *UploadChunkEntity {
	return &UploadChunkEntity{
		chunkUUID:       chunkUUID,
		uploadVideoUUID: uploadVideoUUID,
		chunkIndex:      chunkIndex,
		chunkHash:       chunkHash,
		chunkSize:       chunkSize,
		storagePath:     storagePath,
		completedAt:     completedAt,
		status:          status,
		putURL:          putURL,
		presignExpiredAt: presignExpiredAt,
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

func (e *UploadChunkEntity) PutURL() string {
	return e.putURL
}

func (e *UploadChunkEntity) PresignExpiredAt() *time.Time {
	return e.presignExpiredAt
}

func (e *UploadChunkEntity) SetPresign(putURL string, expiredAt time.Time) {
	e.putURL = putURL
	e.presignExpiredAt = &expiredAt
}
