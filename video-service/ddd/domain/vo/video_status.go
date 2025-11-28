package vo

type VideoStatus struct {
	value string
}

var (
	VideoStatusProcessing = VideoStatus{"processing"}
	VideoStatusPublished  = VideoStatus{"published"}
	VideoStatusFailed     = VideoStatus{"failed"}
)

var videoStatusSet = []VideoStatus{
	VideoStatusProcessing,
	VideoStatusPublished,
	VideoStatusFailed,
}

func (s VideoStatus) Value() string { return s.value }

func (s VideoStatus) IsProcessing() bool { return s.value == VideoStatusProcessing.value }

func (s VideoStatus) IsPublished() bool { return s.value == VideoStatusPublished.value }

func (s VideoStatus) IsFailed() bool { return s.value == VideoStatusFailed.value }

func NewVideoStatus(value string) VideoStatus {
	switch value {
	case "processing":
		return VideoStatusProcessing
	case "published":
		return VideoStatusPublished
	case "failed":
		return VideoStatusFailed
	}
	return VideoStatusProcessing
}
