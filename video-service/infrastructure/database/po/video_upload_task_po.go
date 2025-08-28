package po

import "time"

type VideoUploadTaskPo struct {
	BaseModel
	UUID        string     `json:"uuid"`         //  任务ID
	UserUUID    string     `json:"user_uuid"`    // 用户ID
	VideoUUID   string     `json:"video_uuid"`   // 视频ID
	Status      string     `json:"status"`       // 任务状态
	ErrorMsg    string     `json:"error_msg"`    // 任务失败情况
	CompletedAt *time.Time `json:"completed_at"` // 完成时间
	StoragePath string     `json:"storage_path"` //  Minio存储唯一对象名字
}

func (v *VideoUploadTaskPo) TableName() string {
	return "video_upload_task"
}
