package vo

type UploadVideoStatus struct {
	value string
}

var (
	UploadVideoStatusInit      = UploadVideoStatus{"Init"}
	UploadVideoStatusSuccess   = UploadVideoStatus{"Success"}
	UploadVideoStatusFailed    = UploadVideoStatus{"Failed"}
	UploadVideoStatusUploading = UploadVideoStatus{"Uploading"}
	UploadVideoStatusMerging   = UploadVideoStatus{"Merging"}
)
var UploadVideoStatusArr = []UploadVideoStatus{
	UploadVideoStatusInit,
	UploadVideoStatusSuccess,
	UploadVideoStatusFailed,
	UploadVideoStatusUploading,
	UploadVideoStatusMerging,
}

func (u UploadVideoStatus) Value() string {
	return u.value
}

func NewUploadVideoStatus(value string) UploadVideoStatus {
	for _, v := range UploadVideoStatusArr {
		if v.value == value {
			return v
		}
	}
	return UploadVideoStatusInit
}

type GenerateStoragePathVO struct {
	userUUID        string
	uploadVideoUUID string
	fileName        string
}

// NewGenerateStoragePathVO 创建生成存储路径VO
func NewGenerateStoragePathVO(userUUID, uploadVideoUUID, fileName string) *GenerateStoragePathVO {
	return &GenerateStoragePathVO{
		userUUID:        userUUID,
		uploadVideoUUID: uploadVideoUUID,
		fileName:        fileName,
	}
}

// UserUUID 获取用户UUID
func (g *GenerateStoragePathVO) UserUUID() string {
	return g.userUUID
}

// UploadVideoUUID 获取上传视频UUID
func (g *GenerateStoragePathVO) UploadVideoUUID() string {
	return g.uploadVideoUUID
}

// FileName 获取文件名
func (g *GenerateStoragePathVO) FileName() string {
	return g.fileName
}
