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
