package vo

import "errors"

// Page 分页值对象
type Page struct {
	page     int
	pageSize int
	offset   int
}

// NewPage 创建分页对象
func NewPage(page, pageSize int) (*Page, error) {
	if page < 1 {
		return nil, errors.New("页码必须大于0")
	}

	if pageSize < 1 || pageSize > 100 {
		return nil, errors.New("每页大小必须在1-100之间")
	}

	offset := (page - 1) * pageSize

	return &Page{
		page:     page,
		pageSize: pageSize,
		offset:   offset,
	}, nil
}

// Page 获取页码
func (p *Page) Page() int {
	return p.page
}

// PageSize 获取每页大小
func (p *Page) PageSize() int {
	return p.pageSize
}

// Offset 获取偏移量
func (p *Page) Offset() int {
	return p.offset
}

// Limit 获取限制数量
func (p *Page) Limit() int {
	return p.pageSize
}