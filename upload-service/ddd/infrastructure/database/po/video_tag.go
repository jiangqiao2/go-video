package po

type VideoTag struct {
    BaseModel

    VideoUUID string `gorm:"column:video_uuid;type:varchar(36);not null;index:idx_video;uniqueIndex:uk_video_tag" json:"video_uuid"`
    TagUUID   string `gorm:"column:tag_uuid;type:varchar(36);not null;index:idx_tag;uniqueIndex:uk_video_tag" json:"tag_uuid"`
}

// 复合唯一索引：一个视频同一个标签只能绑定一次
func (VideoTag) TableName() string {
	return "video_tag"
}
