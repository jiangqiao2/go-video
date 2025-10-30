package vo

// VideoStatus represents the lifecycle status of a published video.
type VideoStatus struct {
	value string
}

var (
	VideoStatusDraft      = VideoStatus{value: "Draft"}
	VideoStatusProcessing = VideoStatus{value: "Processing"}
	VideoStatusPublished  = VideoStatus{value: "Published"}
	VideoStatusFailed     = VideoStatus{value: "Failed"}
)

var videoStatusSet = []VideoStatus{
	VideoStatusDraft,
	VideoStatusProcessing,
	VideoStatusPublished,
	VideoStatusFailed,
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

// IsProcessing reports whether the status indicates ongoing processing.
func (s VideoStatus) IsProcessing() bool {
	return s.value == VideoStatusProcessing.value
}

// IsPublished reports whether the status is Published.
func (s VideoStatus) IsPublished() bool {
	return s.value == VideoStatusPublished.value
}

// IsFailed reports whether the status is Failed.
func (s VideoStatus) IsFailed() bool {
	return s.value == VideoStatusFailed.value
}
