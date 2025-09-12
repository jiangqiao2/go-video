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
)
