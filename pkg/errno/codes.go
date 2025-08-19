package errno

type Errno struct {
	Code    int
	Message string
}

// 系统级错误码 (1000-1999)
const (
	// 成功
	CodeSuccess = 200

	// 系统错误
	CodeSystemError   = 1000
	CodeInternalError = 1001
	CodeDatabaseError = 1002
	CodeCacheError    = 1003
	CodeNetworkError  = 1004
	CodeConfigError   = 1005

	// 参数错误
	CodeInvalidParam  = 1100
	CodeMissingParam  = 1101
	CodeParamTooLong  = 1102
	CodeParamTooShort = 1103
	CodeInvalidFormat = 1104
)

// 业务级错误码 (2000-9999)
const (
	// 用户相关错误 (2000-2999)
	CodeUserNotFound       = 2000
	CodeUserAlreadyExists  = 2001
	CodeInvalidCredentials = 2002
	CodeUserDisabled       = 2003
	CodePasswordTooWeak    = 2004
	CodeUsernameInvalid    = 2005

	// 认证授权错误 (3000-3999)
	CodeUnauthorized     = 3000
	CodeForbidden        = 3001
	CodeTokenExpired     = 3002
	CodeTokenInvalid     = 3003
	CodePermissionDenied = 3004

	// 视频相关错误 (4000-4999)
	CodeVideoNotFound      = 4000
	CodeVideoUploadFailed  = 4001
	CodeVideoFormatInvalid = 4002
	CodeVideoTooLarge      = 4003
	CodeVideoProcessFailed = 4004

	// 内容相关错误 (5000-5999)
	CodeContentNotFound  = 5000
	CodeContentDeleted   = 5001
	CodeContentForbidden = 5002

	// 社交相关错误 (6000-6999)
	CodeFollowSelf      = 6000
	CodeAlreadyFollowed = 6001
	CodeNotFollowed     = 6002
	CodeCommentNotFound = 6003
	CodeLikeNotFound    = 6004
)

// 错误消息映射
var ErrorMessages = map[int]string{
	// 系统级错误
	CodeSuccess:       "操作成功",
	CodeSystemError:   "系统错误",
	CodeInternalError: "内部错误",
	CodeDatabaseError: "数据库错误",
	CodeCacheError:    "缓存错误",
	CodeNetworkError:  "网络错误",
	CodeConfigError:   "配置错误",

	// 参数错误
	CodeInvalidParam:  "参数无效",
	CodeMissingParam:  "缺少必要参数",
	CodeParamTooLong:  "参数过长",
	CodeParamTooShort: "参数过短",
	CodeInvalidFormat: "格式无效",

	// 用户相关错误
	CodeUserNotFound:       "用户不存在",
	CodeUserAlreadyExists:  "用户已存在",
	CodeInvalidCredentials: "用户名或密码错误",
	CodeUserDisabled:       "用户已被禁用",
	CodePasswordTooWeak:    "密码强度不够",
	CodeUsernameInvalid:    "用户名格式无效",

	// 认证授权错误
	CodeUnauthorized:     "未授权",
	CodeForbidden:        "禁止访问",
	CodeTokenExpired:     "令牌已过期",
	CodeTokenInvalid:     "令牌无效",
	CodePermissionDenied: "权限不足",

	// 视频相关错误
	CodeVideoNotFound:      "视频不存在",
	CodeVideoUploadFailed:  "视频上传失败",
	CodeVideoFormatInvalid: "视频格式不支持",
	CodeVideoTooLarge:      "视频文件过大",
	CodeVideoProcessFailed: "视频处理失败",

	// 内容相关错误
	CodeContentNotFound:  "内容不存在",
	CodeContentDeleted:   "内容已删除",
	CodeContentForbidden: "内容被禁止访问",

	// 社交相关错误
	CodeFollowSelf:      "不能关注自己",
	CodeAlreadyFollowed: "已经关注过了",
	CodeNotFollowed:     "未关注该用户",
	CodeCommentNotFound: "评论不存在",
	CodeLikeNotFound:    "点赞记录不存在",
}

// GetErrorMessage 获取错误消息
func GetErrorMessage(code int) string {
	if msg, ok := ErrorMessages[code]; ok {
		return msg
	}
	return "未知错误"
}

// 预定义常用错误
var (
	// 2xx 成功类
	ErrSuccess = NewBizError(&Errno{Code: CodeSuccess}, nil, GetErrorMessage(CodeSuccess))

	// 5xx 系统错误类
	ErrSystemError   = NewBizError(&Errno{Code: CodeSystemError}, nil, GetErrorMessage(CodeSystemError))
	ErrInternalError = NewBizError(&Errno{Code: CodeInternalError}, nil, GetErrorMessage(CodeInternalError))
	ErrDatabaseError = NewBizError(&Errno{Code: CodeDatabaseError}, nil, GetErrorMessage(CodeDatabaseError))

	// 4xx 客户端错误类
	ErrInvalidParam       = NewBizError(&Errno{Code: CodeInvalidParam}, nil, GetErrorMessage(CodeInvalidParam))
	ErrUserNotFound       = NewBizError(&Errno{Code: CodeUserNotFound}, nil, GetErrorMessage(CodeUserNotFound))
	ErrUserAlreadyExists  = NewBizError(&Errno{Code: CodeUserAlreadyExists}, nil, GetErrorMessage(CodeUserAlreadyExists))
	ErrInvalidCredentials = NewBizError(&Errno{Code: CodeInvalidCredentials}, nil, GetErrorMessage(CodeInvalidCredentials))
	ErrUnauthorized       = NewBizError(&Errno{Code: CodeUnauthorized}, nil, GetErrorMessage(CodeUnauthorized))
	ErrForbidden          = NewBizError(&Errno{Code: CodeForbidden}, nil, GetErrorMessage(CodeForbidden))
	ErrTokenExpired       = NewBizError(&Errno{Code: CodeTokenExpired}, nil, GetErrorMessage(CodeTokenExpired))
	ErrTokenInvalid       = NewBizError(&Errno{Code: CodeTokenInvalid}, nil, GetErrorMessage(CodeTokenInvalid))

	// 业务错误类
	ErrVideoNotFound = NewBizError(&Errno{Code: CodeVideoNotFound}, nil, GetErrorMessage(CodeVideoNotFound))
)
