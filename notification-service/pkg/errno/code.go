package errno

// Errno 定义业务错误码。
type Errno struct {
	Code    int
	Message string
}

// Error 实现 error 接口。
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
)
