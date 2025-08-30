package vo

import "io"

type MinIoUploadChunkVo struct {
	storagePath string
	bucketName  string
	reader      io.Reader
	fileSize    int64
	contentType string
}

func NewMinIoUploadChunkVo(storagePath, bucketName string, reader io.Reader, fileSize int64, contentType string) *MinIoUploadChunkVo {
	return &MinIoUploadChunkVo{
		storagePath: storagePath,
		bucketName:  bucketName,
		reader:      reader,
		fileSize:    fileSize,
		contentType: contentType,
	}
}

func (m *MinIoUploadChunkVo) StoragePath() string {
	return m.storagePath
}
func (m *MinIoUploadChunkVo) FileSize() int64 {
	return m.fileSize
}
func (m *MinIoUploadChunkVo) ContentType() string {
	return m.contentType
}
func (m *MinIoUploadChunkVo) Reader() io.Reader {
	return m.reader
}
func (m *MinIoUploadChunkVo) BucketName() string {
	return m.bucketName
}
