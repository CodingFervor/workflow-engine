package pagination

import "strconv"

type Params struct {
	Page     int `form:"page"`
	PageSize int `form:"page_size"`
}

type Result struct {
	Items      interface{} `json:"items"`
	Total      int64       `json:"total"`
	Page       int         `json:"page"`
	PageSize   int         `json:"page_size"`
	TotalPages int         `json:"total_pages"`
}

func New(page, pageSize int) Params {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}
	return Params{Page: page, PageSize: pageSize}
}

func (p Params) Offset() int {
	return (p.Page - 1) * p.PageSize
}

func (p Params) Limit() int {
	return p.PageSize
}

func (p Params) Result(items interface{}, total int64) Result {
	tp := int(total) / p.PageSize
	if int(total)%p.PageSize != 0 {
		tp++
	}
	return Result{
		Items:      items,
		Total:      total,
		Page:       p.Page,
		PageSize:   p.PageSize,
		TotalPages: tp,
	}
}

func Parse(pageStr, pageSizeStr string) Params {
	page, _ := strconv.Atoi(pageStr)
	ps, _ := strconv.Atoi(pageSizeStr)
	return New(page, ps)
}
