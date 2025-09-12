package restapi

import "github.com/gin-gonic/gin"

const (
	HeaderKeyRequestId = "x-request-id"
)

type PageQuery struct {
	PageSize int `form:"page_size,default=10"  binding:"omitempty,min=1,max=200" json:"page_size"`
	PageNum  int `form:"page_num,default=1"    binding:"omitempty,min=1" json:"page_num"`
}

func (q PageQuery) GetPageSize() int {
	return q.PageSize
}

func (q PageQuery) GetPageNum() int {
	return q.PageNum
}

func (p *PageQuery) Offset() int {
	return (p.PageNum - 1) * p.PageSize
}

func (p *PageQuery) Limit() int {
	return p.PageSize
}

type page interface {
	GetPageSize() int
	GetPageNum() int
}

func GetRequestId(c *gin.Context) string {
	v, ok := c.Get(HeaderKeyRequestId)
	if !ok {
		return ""
	}
	if requestId, ok := v.(string); ok {
		return requestId
	}
	return ""
}
