package entity

import "time"

// Video represents the domain entity for a video.
type Video struct {
	ID                uint64
	VideoUUID         string
	UserUUID          string
	UploadVideoUUID   string
	Title             string
	Description       string
	CoverURL          string
	VideoURL          string
	DurationSec       *int
	Resolution        string
	SizeBytes         *int64
	Status            string
	TranscodeTaskUUID string
	ErrorMessage      string
	Privacy           string
	PublishedAt       *time.Time
	CreatedAt         time.Time
	UpdatedAt         time.Time
	LikeCount         int64
	PlayCount         int64
	CommentCount      int64
}

func (v *Video) SetCounts(like, play, comment int64) {
	v.LikeCount = like
	v.PlayCount = play
	v.CommentCount = comment
}
