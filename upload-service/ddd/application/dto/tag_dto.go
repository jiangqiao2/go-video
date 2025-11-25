package dto

type TagDto struct {
	Id          uint64 `json:"id"`
	TagUUID     string `json:"tag_uuid"`
	Name        string `json:"name"`
	Code        string `json:"code"`
	Description string `json:"description"`
}

type TagListDto struct {
	List []*TagDto `json:"list"`
}

func NewTagListDto(tags []*TagDto) *TagListDto { return &TagListDto{List: tags} }
