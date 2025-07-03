package dto

import "time"

// Response wrapper
type APIResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

type PaginationQuery struct {
	Page  int `form:"page" binding:"omitempty,min=1"`
	Limit int `form:"limit" binding:"omitempty,min=1,max=100"`
}

func (p *PaginationQuery) SetDefaults() {
	if p.Page == 0 {
		p.Page = 1
	}
	if p.Limit == 0 {
		p.Limit = 10
	}
}

func (p *PaginationQuery) GetOffset() int {
	return (p.Page - 1) * p.Limit
}

type PaginationResponse struct {
	Total     int64 `json:"total"`
	Page      int   `json:"page"`
	Limit     int   `json:"limit"`
	TotalPage int   `json:"total_page"`
}

func NewPaginationResponse(total int64, page, limit int) PaginationResponse {
	totalPage := int((total + int64(limit) - 1) / int64(limit))
	return PaginationResponse{
		Total:     total,
		Page:      page,
		Limit:     limit,
		TotalPage: totalPage,
	}
}
