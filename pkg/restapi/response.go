package restapi

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go-video/pkg/errno"
)

// Response 统一响应结构
type Response struct {
	Code      int         `json:"code"`
	Message   string      `json:"message"`
	Data      interface{} `json:"data,omitempty"`
	Timestamp int64       `json:"timestamp"`
	RequestID string      `json:"request_id,omitempty"`
}

// PageResponse 分页响应结构
type PageResponse struct {
	Code      int         `json:"code"`
	Message   string      `json:"message"`
	Data      interface{} `json:"data,omitempty"`
	Total     int64       `json:"total"`
	Page      int         `json:"page"`
	PageSize  int         `json:"page_size"`
	Timestamp int64       `json:"timestamp"`
	RequestID string      `json:"request_id,omitempty"`
}

// Success 成功响应
func Success(c *gin.Context, data interface{}) {
	response := Response{
		Code:      errno.CodeSuccess,
		Message:   errno.GetErrorMessage(errno.CodeSuccess),
		Data:      data,
		Timestamp: time.Now().Unix(),
		RequestID: getRequestID(c),
	}
	c.JSON(http.StatusOK, response)
}

// SuccessWithMessage 带自定义消息的成功响应
func SuccessWithMessage(c *gin.Context, message string, data interface{}) {
	response := Response{
		Code:      errno.CodeSuccess,
		Message:   message,
		Data:      data,
		Timestamp: time.Now().Unix(),
		RequestID: getRequestID(c),
	}
	c.JSON(http.StatusOK, response)
}

// SuccessPage 分页成功响应
func SuccessPage(c *gin.Context, data interface{}, total int64, page, pageSize int) {
	response := PageResponse{
		Code:      errno.CodeSuccess,
		Message:   errno.GetErrorMessage(errno.CodeSuccess),
		Data:      data,
		Total:     total,
		Page:      page,
		PageSize:  pageSize,
		Timestamp: time.Now().Unix(),
		RequestID: getRequestID(c),
	}
	c.JSON(http.StatusOK, response)
}

// Error 错误响应
func Error(c *gin.Context, err error) {
	var bizErr *errno.BizError
	var code int
	var message string
	var httpStatus int

	if e, ok := err.(*errno.BizError); ok {
		bizErr = e
		code = e.Code
		message = e.Message
		httpStatus = getHTTPStatus(e.Code)
	} else {
		code = errno.CodeInternalError
		message = err.Error()
		httpStatus = http.StatusInternalServerError
	}

	response := Response{
		Code:      code,
		Message:   message,
		Timestamp: time.Now().Unix(),
		RequestID: getRequestID(c),
	}

	// 开发环境下返回错误堆栈
	if gin.Mode() == gin.DebugMode && bizErr != nil && bizErr.Stack != "" {
		response.Data = map[string]interface{}{
			"stack": bizErr.Stack,
		}
	}

	c.JSON(httpStatus, response)
}

// ErrorWithCode 指定错误码的错误响应
func ErrorWithCode(c *gin.Context, code int, message string) {
	response := Response{
		Code:      code,
		Message:   message,
		Timestamp: time.Now().Unix(),
		RequestID: getRequestID(c),
	}
	c.JSON(getHTTPStatus(code), response)
}

// getRequestID 获取请求ID
func getRequestID(c *gin.Context) string {
	if requestID := c.GetHeader("X-Request-ID"); requestID != "" {
		return requestID
	}
	if requestID := c.GetString("request_id"); requestID != "" {
		return requestID
	}
	return ""
}

// getHTTPStatus 根据业务错误码获取HTTP状态码
func getHTTPStatus(code int) int {
	switch {
	case code == errno.CodeSuccess:
		return http.StatusOK
	case code >= errno.CodeInvalidParam && code < errno.CodeUserNotFound:
		return http.StatusBadRequest
	case code >= errno.CodeUnauthorized && code < errno.CodeVideoNotFound:
		if code == errno.CodeUnauthorized {
			return http.StatusUnauthorized
		}
		return http.StatusForbidden
	case code >= errno.CodeUserNotFound && code < errno.CodeUnauthorized:
		return http.StatusNotFound
	case code >= errno.CodeVideoNotFound:
		return http.StatusNotFound
	default:
		return http.StatusInternalServerError
	}
}