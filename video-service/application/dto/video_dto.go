package dto

type UploadVideoDto struct {
	VideoUUID string `json:"video_uuid"`
}

type VideoSyncVideoDto struct {
	VideoUUID string `json:"video_uuid"`
	TaskUUID  string `json:"task_uuid"`
}
