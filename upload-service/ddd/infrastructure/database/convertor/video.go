package convertor

import (
	"encoding/json"
	"time"

	"upload-service/ddd/domain/entity"
	"upload-service/ddd/domain/vo"
	"upload-service/ddd/infrastructure/database/po"
)

func ToVideoPo(video *entity.VideoEntity) *po.VideoPo {
	if video == nil {
		return nil
	}
	var publishedAt *time.Time
	if ts := video.PublishedAt(); ts != nil {
		t := *ts
		publishedAt = &t
	}
	return &po.VideoPo{
		VideoUUID:       video.VideoUUID(),
		UploadVideoUUID: video.UploadVideoUUID(),
		UserUUID:        video.UserUUID(),
		Title:           video.Title(),
		Description:     video.Description(),
		TagsJSON:        encodeTags(video.Tags()),
		CoverURL:        video.CoverURL(),
		Status:          video.Status().Value(),
		PublishedAt:     publishedAt,
		TranscodeTaskUUID: video.TranscodeTaskUUID(),
		VideoURL:          video.VideoURL(),
		ErrorMessage:      video.ErrorMessage(),
	}
}

func ToVideoEntity(video *po.VideoPo) *entity.VideoEntity {
	if video == nil {
		return nil
	}
	return entity.NewVideoEntity(
		video.VideoUUID,
		video.UploadVideoUUID,
		video.UserUUID,
		video.Title,
		video.Description,
		video.CoverURL,
		decodeTags(video.TagsJSON),
		vo.NewVideoStatus(video.Status),
		video.PublishedAt,
		video.TranscodeTaskUUID,
		video.VideoURL,
		video.ErrorMessage,
	)
}

func ToVideoEntities(videos []*po.VideoPo) []*entity.VideoEntity {
	if len(videos) == 0 {
		return nil
	}

	result := make([]*entity.VideoEntity, 0, len(videos))
	for _, v := range videos {
		if entity := ToVideoEntity(v); entity != nil {
			result = append(result, entity)
		}
	}
	return result
}

func encodeTags(tags []string) string {
	if len(tags) == 0 {
		return "[]"
	}
	b, err := json.Marshal(tags)
	if err != nil {
		return "[]"
	}
	return string(b)
}

func decodeTags(raw string) []string {
	if raw == "" {
		return nil
	}
	var tags []string
	if err := json.Unmarshal([]byte(raw), &tags); err != nil {
		return nil
	}
	if len(tags) == 0 {
		return nil
	}
	return tags
}
