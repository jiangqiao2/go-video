package po

import "time"

type UploadChunkPo struct {
	BaseModel
	ChunkUUID       string     `gorm:"column:chunk_uuid" json:"chunk_uuid"`               // 分片唯一UUID
	UploadVideoUUID string     `gorm:"column:upload_video_uuid" json:"upload_video_uuid"` // 关联 上传视频的uuid
	ChunkIndex      int        `gorm:"column:chunk_index" json:"chunk_index"`             // 分片索引，从0开始
	ChunkHash       string     `gorm:"column:chunk_hash" json:"chunk_hash"`               // 分片Hash
	ChunkSize       int        `gorm:"column:chunk_size" json:"chunk_size"`               // 分片大小
	StoragePath     string     `gorm:"column:storage_path" json:"-"`                      // 合并后在Minio的路径
	Status          string     `gorm:"column:status" json:"-"`                            //状态
	CompletedTime   *time.Time `gorm:"column:completed_time" json:"-"`                    // 完成时间
}

func (UploadChunkPo) TableName() string {
	return "upload_chunk"
}
