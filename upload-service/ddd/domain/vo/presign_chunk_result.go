package vo

type PresignChunkResult struct {
	UploadVideoUUID string
	ChunkUUID       string
	ChunkIndex      int
	Bucket          string
	Key             string
	PutURL          string
	ExpiresSeconds  int
}

type CompleteChunkResult struct {
	Status string
}
