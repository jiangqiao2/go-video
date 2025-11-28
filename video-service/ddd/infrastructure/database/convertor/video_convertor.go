package convertor

import (
	"video-service/ddd/domain/entity"
	"video-service/ddd/infrastructure/database/po"
)

func ToVideoPo(v *entity.Video) *po.Video {
	if v == nil {
		return nil
	}
	var uploadVideoPtr *string
	if v.UploadVideoUUID != "" {
		s := v.UploadVideoUUID
		uploadVideoPtr = &s
	}
	var transcodePtr *string
	if v.TranscodeTaskUUID != "" {
		s := v.TranscodeTaskUUID
		transcodePtr = &s
	}
	return &po.Video{
		ID:                v.ID,
		VideoUUID:         v.VideoUUID,
		UserUUID:          v.UserUUID,
		UploadVideoUUID:   uploadVideoPtr,
		Title:             v.Title,
		Description:       v.Description,
		CoverURL:          v.CoverURL,
		VideoURL:          v.VideoURL,
		DurationSec:       v.DurationSec,
		Resolution:        v.Resolution,
		SizeBytes:         v.SizeBytes,
		Status:            v.Status,
		TranscodeTaskUUID: transcodePtr,
		ErrorMessage:      v.ErrorMessage,
		Privacy:           v.Privacy,
		PublishedAt:       v.PublishedAt,
		CreatedAt:         v.CreatedAt,
		UpdatedAt:         v.UpdatedAt,
	}
}

func ToVideoEntity(p *po.Video) *entity.Video {
	if p == nil {
		return nil
	}
	var uploadVideo string
	if p.UploadVideoUUID != nil {
		uploadVideo = *p.UploadVideoUUID
	}
	var transcode string
	if p.TranscodeTaskUUID != nil {
		transcode = *p.TranscodeTaskUUID
	}
	return &entity.Video{
		ID:                p.ID,
		VideoUUID:         p.VideoUUID,
		UserUUID:          p.UserUUID,
		UploadVideoUUID:   uploadVideo,
		Title:             p.Title,
		Description:       p.Description,
		CoverURL:          p.CoverURL,
		VideoURL:          p.VideoURL,
		DurationSec:       p.DurationSec,
		Resolution:        p.Resolution,
		SizeBytes:         p.SizeBytes,
		Status:            p.Status,
		TranscodeTaskUUID: transcode,
		ErrorMessage:      p.ErrorMessage,
		Privacy:           p.Privacy,
		PublishedAt:       p.PublishedAt,
		CreatedAt:         p.CreatedAt,
		UpdatedAt:         p.UpdatedAt,
	}
}
