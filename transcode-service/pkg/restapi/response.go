package restapi

import (
	"encoding/json"
	"transcode-service/pkg/encode"
	"transcode-service/pkg/errno"

	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type Response struct {
	Code        int         `json:"code"`
	Message     string      `json:"message"`
	Data        interface{} `json:"data"`
	DataVersion string      `json:"data_version"`
	RequestId   string      `json:"request_id"`
}

type PageInfo struct {
	TotalNum    int64 `json:"total_num"`
	CurrentPage int   `json:"current_page"`
	PageSize    int   `json:"page_size"`
}

type PageResult struct {
	PageInfo PageInfo    `json:"page_info"`
	Rows     interface{} `json:"rows"`
}

func Success(c *gin.Context, data interface{}) {
	sendResponse(c, http.StatusOK, data, nil)
}

func Failed(c *gin.Context, err error) {
	sendResponse(c, http.StatusInternalServerError, nil, err)
}

func FailedWithStatus(c *gin.Context, err error, httpStatus int) {
	sendResponse(c, httpStatus, nil, err)
}

func sendResponse(c *gin.Context, httpStatus int, data interface{}, err error) {
	bizErr := errno.AssertBizError(err)
	c.Set("x-bizError", bizErr)
	c.Set("x-httpStatus", httpStatus)
	c.Writer.Header().Add("x-biz-code", strconv.Itoa(bizErr.Code()))
	// 返回json格式数据
	c.JSON(httpStatus, generateResponseDataWithVersion(c, bizErr, data))
}
func SuccessWithPage(c *gin.Context, q page, list interface{}, total int64) {
	pageInfo := PageInfo{
		PageSize:    q.GetPageSize(),
		CurrentPage: q.GetPageNum(),
		TotalNum:    total,
	}
	pageResult := PageResult{
		PageInfo: pageInfo,
		Rows:     list,
	}
	Success(c, pageResult)
}
func generateResponseDataWithVersion(c *gin.Context, err errno.BizError, data interface{}) *Response {
	resp := &Response{
		RequestId: GetRequestId(c),
		Code:      err.Code(),
		Message:   err.Message(),
		Data:      data,
	}

	if data == nil {
		return resp
	}
	b, _ := json.Marshal(data)
	if len(b) > 0 {
		resp.DataVersion = encode.Crc32HashCode(b)
		c.Set("x-data-version", resp.DataVersion)
	}
	lastHash := c.Query("data_version")
	if lastHash != "" && lastHash == resp.DataVersion {
		resp.Data = nil
	}

	return resp
}
