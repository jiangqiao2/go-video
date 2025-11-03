package errno

// code=0 请求成功
// code=4xx 客户端请求错误
// code=5xx 服务器端错误
// code=2xxxx 业务处理错误码

type Errno struct {
	Code    int
	Message string
}

// Error 实现error接口
func (e *Errno) Error() string {
	return e.Message
}

var (
	OK = &Errno{Code: 200, Message: "Success"}

	ErrParameterInvalid = &Errno{Code: 400, Message: "Invalid parameter %s"}
	ErrInvalidParam     = &Errno{Code: 400, Message: "Invalid parameter"}
	ErrUnauthorized     = &Errno{Code: 401, Message: "Unauthorized"}
	ErrNotFound         = &Errno{Code: 404, Message: "Not found"}

	ErrInternalServer = &Errno{Code: 500, Message: "Internal server error"}
	ErrDatabase       = &Errno{Code: 501, Message: "Database error"}
	ErrUnknown        = &Errno{Code: 510, Message: "Unknown error"}

	// 业务错误码
	ErrMissingParam          = &Errno{Code: 20001, Message: "Missing required parameter"}
	ErrFileNameIllegal       = &Errno{Code: 20002, Message: "File name is illegal"}
	ErrFileSizeIllegal       = &Errno{Code: 20003, Message: "File size is illegal"}
	ErrUploadIllegal         = &Errno{Code: 20004, Message: "Upload files is illegal"}
	ErrMinIoBuckNameNotExist = &Errno{Code: 20006, Message: "Minio bucket name does not exist"}
	ErrUploadChunkLoding     = &Errno{Code: 20005, Message: "Upload chunks is loding"}
	ErrUploadError           = &Errno{Code: 20006, Message: "Upload error"}
	ErrChunkIncomplete       = &Errno{Code: 20007, Message: "Chunk is incomplete"}
	// 转码服务错误码
	ErrTranscodeTaskNotFound = &Errno{Code: 20008, Message: "Transcode task not found"}
	ErrInvalidTaskStatus     = &Errno{Code: 20009, Message: "Invalid task status"}
	ErrTranscodeTaskExists   = &Errno{Code: 20010, Message: "Transcode task already exists"}
	ErrWorkerNotAvailable    = &Errno{Code: 20011, Message: "No worker available"}
	ErrQueueFull             = &Errno{Code: 20012, Message: "Task queue is full"}
	ErrUserUUIDRequired      = &Errno{Code: 20013, Message: "User UUID is required"}
	ErrVideoUUIDRequired     = &Errno{Code: 20014, Message: "Video UUID is required"}
	ErrTaskUUIDRequired      = &Errno{Code: 20015, Message: "Task UUID is required"}
	ErrOriginalPathRequired  = &Errno{Code: 20016, Message: "Original path is required"}
	ErrResolutionRequired    = &Errno{Code: 20017, Message: "Resolution is required"}
	ErrBitrateRequired       = &Errno{Code: 20018, Message: "Bitrate is required"}
	ErrStatusRequired        = &Errno{Code: 20019, Message: "Status is required"}
	
	// HLS相关错误码
	ErrHLSResolutionsRequired = &Errno{Code: 20020, Message: "HLS resolutions are required when HLS is enabled"}
	ErrInvalidHLSResolution   = &Errno{Code: 20021, Message: "Invalid HLS resolution configuration"}
	ErrHLSBitrateRequired     = &Errno{Code: 20022, Message: "HLS bitrate is required"}
	ErrHLSGenerationFailed    = &Errno{Code: 20023, Message: "HLS slice generation failed"}
)
