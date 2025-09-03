package po

import "time"

type UploadVideoPo struct {
	BaseModel
	UploadVideoUUID  string     `gorm:"column:upload_video_uuid" json:"-"`  // 上传视频唯一UUID
	UserUUID         string     `gorm:"column:user_uuid" json:"-"`          // 上传用户的唯一UUID
	FileName         string     `gorm:"column:file_name" json:"-"`          // 原始文件名
	FileSize         int        `gorm:"column:file_size" json:"-"`          // 上传文件大小
	FileHash         string     `gorm:"column:file_hash" json:"-"`          // 文件内容Hash
	TotalChunks      int        `gorm:"column:total_chunks" json:"-"`       // 分片数量
	UploadedChunks   int        `gorm:"column:uploaded_chunks" json:"-"`    // 已经完成的分片数量
	ChunkStoragePath string     `gorm:"column:chunk_storage_path" json:"-"` // 分片路径
	Status           string     `gorm:"column:status" json:"-"`             // 状态
	StoragePath      string     `gorm:"column:storage_path" json:"-"`       // 合并后在Minio的路径
	CompletedTime    *time.Time `gorm:"column:completed_time" json:"-"`     // 完成时间
}

func (UploadVideoPo) TableName() string {
	return "upload_video"
}
