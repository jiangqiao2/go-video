package vo

// Page 分页值对象
type Page struct {
	page     int
	pageSize int
}

// NewPage 创建分页对象
func NewPage(page, pageSize int) *Page {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	if pageSize > 100 {
		pageSize = 100
	}
	return &Page{
		page:     page,
		pageSize: pageSize,
	}
}

// Page 获取页码
func (p *Page) Page() int {
	return p.page
}

// PageSize 获取页大小
func (p *Page) PageSize() int {
	return p.pageSize
}

// Offset 获取偏移量
func (p *Page) Offset() int {
	return (p.page - 1) * p.pageSize
}

// Limit 获取限制数量
func (p *Page) Limit() int {
	return p.pageSize
}