package dto

import (
	"time"
)

// Category DTOs
type CreateCategoryRequest struct {
	Name        string  `json:"name" binding:"required"`
	Slug        string  `json:"slug" binding:"required"`
	Description *string `json:"description"`
	IconURL     *string `json:"icon_url"`
	ParentID    *string `json:"parent_id"`
	SortOrder   *int32  `json:"sort_order"`
}

type UpdateCategoryRequest struct {
	Name        *string `json:"name"`
	Slug        *string `json:"slug"`
	Description *string `json:"description"`
	IconURL     *string `json:"icon_url"`
	ParentID    *string `json:"parent_id"`
	SortOrder   *int32  `json:"sort_order"`
	IsActive    *bool   `json:"is_active"`
}

type CategoryResponse struct {
	ID          string               `json:"id"`
	Name        string               `json:"name"`
	Slug        string               `json:"slug"`
	Description *string              `json:"description"`
	IconURL     *string              `json:"icon_url"`
	ParentID    *string              `json:"parent_id"`
	SortOrder   int32                `json:"sort_order"`
	IsActive    bool                 `json:"is_active"`
	CreatedAt   time.Time            `json:"created_at"`
	UpdatedAt   time.Time            `json:"updated_at"`
	Children    []CategoryResponse   `json:"children,omitempty"`
}

type CategoryListResponse struct {
	Categories []CategoryResponse `json:"categories"`
	Pagination PaginationResponse `json:"pagination"`
}
