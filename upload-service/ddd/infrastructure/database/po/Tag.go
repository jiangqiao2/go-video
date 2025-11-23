package po

type Tag struct {
    BaseModel

    TagUUID       string `gorm:"column:tag_uuid;type:varchar(36);uniqueIndex:uk_tag_uuid" json:"tag_uuid"`
    Name        string `gorm:"column:name;type:varchar(64);not null;uniqueIndex:uk_tag_name;comment:'标签名称'" json:"name"`
    Code        string `gorm:"column:code;type:varchar(64);not null;uniqueIndex:uk_tag_code;comment:'标签编码(英文/拼音)'" json:"code"`
    Description string `gorm:"column:description;type:varchar(256);comment:'标签描述'" json:"description"`
}

func (Tag) TableName() string {
    return "tag"
}
