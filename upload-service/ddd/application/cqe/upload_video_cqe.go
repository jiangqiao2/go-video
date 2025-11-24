package cqe

import "upload-service/pkg/errno"

type UploadVideoInitReq struct {
	FileName    string `json:"file_name"`    // 文件名字
	FileSize    int    `json:"file_size"`    // 文件大小
	TotalChunks int    `json:"total_chunks"` // 文件分片数量
	UserUUID    string `json:"user_uuid"`    //  用户名
	FileHash    string `json:"file_hash"`    // 文件Hash值
}

func (u *UploadVideoInitReq) Validate() error {
	if u.FileName == "" || len(u.FileName) > 256 {
		return errno.ErrFileNameIllegal
	}
	if u.FileSize <= 0 || u.FileSize > 1024*1024*1024*5 {
		return errno.ErrFileSizeIllegal
	}
	if u.TotalChunks <= 0 || u.TotalChunks > 512 {
		return errno.ErrFileSizeIllegal
	}
	return nil
}

type UploadChunkReq struct {
	ChunkUUID       string `json:"chunk_uuid"`        // 分片唯一标识
	UserUUID        string `json:"user_uuid"`         // 用户UUID
	UploadVideoUUID string `json:"upload_video_uuid"` // 上传视频唯一标识
	ChunkSize       int    `json:"chunk_size"`        // 分片大小
	ChunkIndex      int    `json:"chunk_index"`       // 分片索引
	ChunkData       []byte `json:"chunk_data"`        // 分片文件
	ChunkHash       string `json:"chunk_hash"`        // 分片Hash值
}

func (u *UploadChunkReq) Validate() error {
	// TODO 参数校验
	return nil
}

type MergeChunkReq struct {
	UploadVideoUUID string `json:"upload_video_uuid"`
	UserUUID        string `json:"user_uuid"`
}

func (u *MergeChunkReq) Validate() error {

	return nil
}

type UploadVideoStoragePathReq struct {
    UserUUID  string `form:"user_uuid"`
    ChunkUUID string `form:"chunk_uuid"`
}

type UploadVideoStatusReq struct {
    UploadVideoUUID string `form:"upload_video_uuid"`
    UserUUID        string `form:"-"`
}

func (u *UploadVideoStatusReq) Validate() error {
    if u.UploadVideoUUID == "" {
        return errno.ErrUploadIllegal
    }
    return nil
}

type PresignImageReq struct {
    FileName       string `json:"file_name"`
    ContentType    string `json:"content_type"`
    Category       string `json:"category"`
    ExpiresSeconds int    `json:"expires_seconds"`
    UserUUID       string `json:"user_uuid,omitempty"`
}

func (r *PresignImageReq) Normalize() {
    if r.Category == "" { r.Category = "avatar" }
    if r.ExpiresSeconds <= 0 || r.ExpiresSeconds > 604800 { r.ExpiresSeconds = 900 }
}

func (r *PresignImageReq) Validate() error {
    if r.FileName == "" || len(r.FileName) > 256 { return errno.ErrFileNameIllegal }
    return nil
}
