package entity

import (
	"strings"
	"time"

	"github.com/google/uuid"

	"upload-service/ddd/domain/vo"
)

// VideoEntity aggregates published video metadata.
type VideoEntity struct {
	videoUUID       string
	uploadVideoUUID string
	userUUID        string
	title           string
	description     string
	tags            []string
	coverURL        string
	status          vo.VideoStatus
	publishedAt     *time.Time
	transcodeTaskUUID string
	videoURL          string
	errorMessage      string
}

// DefaultVideoEntity creates a new video entity intended for persistence.
func DefaultVideoEntity(userUUID, uploadVideoUUID, title, description, coverURL string, tags []string, status vo.VideoStatus) *VideoEntity {
	normalizedTags := normalizeTags(tags)
	entity := &VideoEntity{
		videoUUID:         uuid.NewString(),
		uploadVideoUUID:   uploadVideoUUID,
		userUUID:          userUUID,
		title:             title,
		description:       description,
		tags:              normalizedTags,
		coverURL:          coverURL,
		status:            status,
		transcodeTaskUUID: "",
		videoURL:          "",
		errorMessage:      "",
	}
	if status.IsPublished() {
		now := time.Now().UTC()
		entity.publishedAt = &now
	}
	return entity
}

// NewVideoEntity rebuilds a video entity from persisted state.
func NewVideoEntity(videoUUID, uploadVideoUUID, userUUID, title, description, coverURL string,
	tags []string, status vo.VideoStatus, publishedAt *time.Time, transcodeTaskUUID, videoURL, errorMessage string) *VideoEntity {
	return &VideoEntity{
		videoUUID:         videoUUID,
		uploadVideoUUID:   uploadVideoUUID,
		userUUID:          userUUID,
		title:             title,
		description:       description,
		tags:              cloneTags(tags),
		coverURL:          coverURL,
		status:            status,
		publishedAt:       publishedAt,
		transcodeTaskUUID: transcodeTaskUUID,
		videoURL:          videoURL,
		errorMessage:      errorMessage,
	}
}

// VideoUUID returns the unique video identifier.
func (v *VideoEntity) VideoUUID() string {
	return v.videoUUID
}

// UploadVideoUUID returns the associated upload task identifier.
func (v *VideoEntity) UploadVideoUUID() string {
	return v.uploadVideoUUID
}

// UserUUID returns the owner identifier.
func (v *VideoEntity) UserUUID() string {
	return v.userUUID
}

// Title returns the video title.
func (v *VideoEntity) Title() string {
	return v.title
}

// Description returns the video description.
func (v *VideoEntity) Description() string {
	return v.description
}

// Tags returns a defensive copy of tags.
func (v *VideoEntity) Tags() []string {
	return cloneTags(v.tags)
}

// CoverURL returns optional cover artwork URL.
func (v *VideoEntity) CoverURL() string {
	return v.coverURL
}

// Status returns publishing status.
func (v *VideoEntity) Status() vo.VideoStatus {
	return v.status
}

func (v *VideoEntity) SetStatus(status vo.VideoStatus) {
	v.status = status
}

// PublishedAt returns publish timestamp.
func (v *VideoEntity) PublishedAt() *time.Time {
	return v.publishedAt
}

// MarkPublished transitions entity to published state with timestamp.
func (v *VideoEntity) MarkPublished() {
	if !v.status.IsPublished() {
		v.status = vo.VideoStatusPublished
		now := time.Now().UTC()
		v.publishedAt = &now
	}
}

// SetPublishedAt manually sets the published timestamp.
func (v *VideoEntity) SetPublishedAt(ts *time.Time) {
	v.publishedAt = ts
}

// TranscodeTaskUUID returns associated transcode task identifier.
func (v *VideoEntity) TranscodeTaskUUID() string {
	return v.transcodeTaskUUID
}

// SetTranscodeTaskUUID updates the associated transcode task identifier.
func (v *VideoEntity) SetTranscodeTaskUUID(taskUUID string) {
	v.transcodeTaskUUID = taskUUID
}

// VideoURL returns the final playback URL if available.
func (v *VideoEntity) VideoURL() string {
	return v.videoURL
}

// SetVideoURL updates the playback URL.
func (v *VideoEntity) SetVideoURL(url string) {
	v.videoURL = url
}

// ErrorMessage returns the last error message.
func (v *VideoEntity) ErrorMessage() string {
	return v.errorMessage
}

// SetErrorMessage updates the error message.
func (v *VideoEntity) SetErrorMessage(msg string) {
	v.errorMessage = msg
}

func normalizeTags(tags []string) []string {
	if len(tags) == 0 {
		return nil
	}
	seen := make(map[string]struct{}, len(tags))
	result := make([]string, 0, len(tags))
	for _, tag := range tags {
		trimmed := strings.TrimSpace(tag)
		if trimmed == "" {
			continue
		}
		key := strings.ToLower(trimmed)
		if _, exists := seen[key]; exists {
			continue
		}
		seen[key] = struct{}{}
		result = append(result, trimmed)
	}
	if len(result) == 0 {
		return nil
	}
	return result
}

func cloneTags(tags []string) []string {
	if len(tags) == 0 {
		return nil
	}
	copied := make([]string, len(tags))
	copy(copied, tags)
	return copied
}
