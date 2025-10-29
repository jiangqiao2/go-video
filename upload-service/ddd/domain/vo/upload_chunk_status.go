package vo

type UploadChunkStatus struct {
	value string
}

var (
	UploadChunkStatusInitialized = UploadChunkStatus{value: "Initialized"} // 初始化
	UploadChunkStatusUploading   = UploadChunkStatus{value: "Uploading"}   // 上传中
	UploadChunkStatusFailed      = UploadChunkStatus{value: "Failed"}      // 上传失败，需前端重试
	UploadChunkStatusCompleted   = UploadChunkStatus{value: "Completed"}   // 上传完成
)

var UploadChunkStatusArr = []UploadChunkStatus{
	UploadChunkStatusInitialized,
	UploadChunkStatusUploading,
	UploadChunkStatusFailed,
	UploadChunkStatusCompleted,
}

func NewUploadChunkStatus(value string) UploadChunkStatus {
	for _, v := range UploadChunkStatusArr {
		if v.value == value {
			return v
		}
	}
	return UploadChunkStatusInitialized
}

func (u UploadChunkStatus) Value() string {
	return u.value
}

func (u UploadChunkStatus) IsInitialized() bool {
	return u.value == UploadChunkStatusInitialized.Value()
}

func (u UploadChunkStatus) IsUploading() bool {
	return u.value == UploadChunkStatusUploading.Value()
}

func (u UploadChunkStatus) IsFailed() bool {
	return u.value == UploadChunkStatusFailed.Value()
}

func (u UploadChunkStatus) IsCompleted() bool {
	return u.value == UploadChunkStatusCompleted.Value()
}
