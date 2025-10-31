package gateway

import "context"

// StorageGateway 存储网关
type StorageGateway interface {
	// UploadTranscodedFile 上传转码后的文件，返回可访问的对象路径
	UploadTranscodedFile(ctx context.Context, localPath, objectKey, contentType string) (string, error)
	
	// DownloadFile 从存储中下载文件到本地路径
	DownloadFile(ctx context.Context, objectKey, localPath string) error
}
