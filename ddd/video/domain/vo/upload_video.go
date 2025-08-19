package vo

type UploadVideoVO struct {
	fileName string
	file     []byte
	userUUID string
}

func NewUploadVideo(fileName string, fileBytes []byte, userUUID string) *UploadVideoVO {
	return &UploadVideoVO{
		fileName: fileName,
		file:     fileBytes,
		userUUID: userUUID,
	}
}
