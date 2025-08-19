package errno

// code=0 请求成功
// code=4xx 客户端请求错误
// code=5xx 服务器端错误
// code=2xxxx 业务处理错误码

type Errno struct {
	Code    int
	Message string
}

var (
	OK = &Errno{Code: 200, Message: "Success"}

	ErrParameterInvalid = &Errno{Code: 400, Message: "Invalid parameter %s"}
	ErrUnauthorized     = &Errno{Code: 401, Message: "Unauthorized"}
	ErrNotFound         = &Errno{Code: 404, Message: "Not found"}

	ErrInternalServer = &Errno{Code: 500, Message: "Internal server error"}
	ErrDatabase       = &Errno{Code: 501, Message: "Database error"}
	ErrUnknown        = &Errno{Code: 510, Message: "Unknown error"}
)
