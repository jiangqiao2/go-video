package vo

// VideoStatus represents the lifecycle status of a published video.
type VideoStatus struct {
	value string
}

var (
	VideoStatusDraft     = VideoStatus{value: "Draft"}
	VideoStatusPublished = VideoStatus{value: "Published"}
)

var videoStatusSet = []VideoStatus{
	VideoStatusDraft,
	VideoStatusPublished,
}

// NewVideoStatus constructs a VideoStatus from raw value, falling back to Draft when unknown.
func NewVideoStatus(value string) VideoStatus {
	for _, status := range videoStatusSet {
		if status.value == value {
			return status
		}
	}
	return VideoStatusDraft
}

// Value returns the underlying string value.
func (s VideoStatus) Value() string {
	return s.value
}

// IsDraft reports whether the status is Draft.
func (s VideoStatus) IsDraft() bool {
	return s.value == VideoStatusDraft.value
}

// IsPublished reports whether the status is Published.
func (s VideoStatus) IsPublished() bool {
	return s.value == VideoStatusPublished.value
}
