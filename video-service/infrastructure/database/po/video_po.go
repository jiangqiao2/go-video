package po

type VideoPo struct {
	BaseModel

	UUID        string `gorm:"uniqueIndex;size:36;not null;column:uuid" json:"uuid"`
	UserUUID    string `gorm:"size:36;not null;column:user_uuid" json:"user_uuid"`
	Title       string `gorm:"size:255;not null;column:title" json:"title"`
	Description string `gorm:"type:text;column:description" json:"description"`
	Filename    string `gorm:"size:255;not null;column:filename" json:"filename"`
	FileSize    int64  `gorm:"not null;column:file_size" json:"file_size"`
	Duration    int    `gorm:"column:duration" json:"duration"`
	Format      string `gorm:"size:20;column:format" json:"format"`
	StoragePath string `gorm:"size:500;column:storage_path" json:"storage_path"`
	Status      string `gorm:"column:status" json:"status"`
}

func (v *VideoPo) TableName() string {
	return "video"
}
