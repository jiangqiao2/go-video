package po

import "time"

type UploadVideoPo struct {
	BaseModel
	UploadVideoUUID  string     `gorm:"column:upload_video_uuid" json:"-"`
	UserUUID         string     `gorm:"column:user_uuid" json:"-"`
	FileName         string     `gorm:"column:file_name" json:"-"`
	FileSize         int        `gorm:"column:file_size" json:"-"`
	FileHash         string     `gorm:"column:file_hash" json:"-"`
	TotalChunks      int        `gorm:"column:total_chunks" json:"-"`
	UploadedChunks   int        `gorm:"column:uploaded_chunks" json:"-"`
	ChunkStoragePath string     `gorm:"column:chunk_storage_path" json:"-"`
	Status           string     `gorm:"column:status" json:"-"`
	StoragePath      string     `gorm:"column:storage_path" json:"-"`
	CompletedTime    *time.Time `gorm:"column:completed_time" json:"-"`
}

func (UploadVideoPo) TableName() string {
	return "upload_video"
}
