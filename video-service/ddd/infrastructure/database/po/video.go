package po

import "time"

type Video struct {
	ID                uint64  `gorm:"primaryKey"`
	VideoUUID         string  `gorm:"size:64;uniqueIndex"`
	UserUUID          string  `gorm:"size:64;index"`
	UploadVideoUUID   *string `gorm:"size:64"`
	Title             string  `gorm:"size:255"`
	Description       string  `gorm:"type:text"`
	CoverURL          string  `gorm:"size:512"`
	VideoURL          string  `gorm:"size:512"`
	DurationSec       *int
	Resolution        string `gorm:"size:64"`
	SizeBytes         *int64
	Status            string  `gorm:"size:32;index"`
	TranscodeTaskUUID *string `gorm:"size:64;uniqueIndex"`
	ErrorMessage      string  `gorm:"type:text"`
	Privacy           string  `gorm:"size:32"`
	PublishedAt       *time.Time
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

func (Video) TableName() string { return "videos" }
