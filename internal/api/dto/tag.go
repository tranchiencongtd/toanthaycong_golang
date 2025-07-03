package dto

import "time"

// TagDTO - DTO cho thẻ tag
type TagDTO struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Slug        string    `json:"slug"`
	Description *string   `json:"description,omitempty"`
	Color       *string   `json:"color,omitempty"` // màu hex
	CreatedAt   time.Time `json:"created_at"`
	
	// Thông tin liên quan
	CourseCount *int `json:"course_count,omitempty"` // số khóa học sử dụng tag này
}

// CreateTagRequest - Request tạo thẻ tag
type CreateTagRequest struct {
	Name        string  `json:"name" binding:"required,max=50"`
	Slug        string  `json:"slug" binding:"required,max=50"`
	Description *string `json:"description,omitempty"`
	Color       *string `json:"color,omitempty" binding:"omitempty,len=7"` // #RRGGBB format
}

// UpdateTagRequest - Request cập nhật thẻ tag
type UpdateTagRequest struct {
	Name        *string `json:"name,omitempty" binding:"omitempty,max=50"`
	Slug        *string `json:"slug,omitempty" binding:"omitempty,max=50"`
	Description *string `json:"description,omitempty"`
	Color       *string `json:"color,omitempty" binding:"omitempty,len=7"`
}

// TagListResponse - Response danh sách thẻ tag
type TagListResponse struct {
	Data       []TagDTO       `json:"data"`
	Pagination PaginationMeta `json:"pagination"`
}

// CourseTagDTO - DTO cho liên kết course-tag
type CourseTagDTO struct {
	CourseID string `json:"course_id"`
	TagID    string `json:"tag_id"`
	
	// Thông tin liên quan
	Course *CourseDTO `json:"course,omitempty"`
	Tag    *TagDTO    `json:"tag,omitempty"`
}

// AddCourseTagRequest - Request thêm tag cho khóa học
type AddCourseTagRequest struct {
	CourseID string `json:"course_id" binding:"required,uuid"`
	TagID    string `json:"tag_id" binding:"required,uuid"`
}

// CourseTagListResponse - Response danh sách tag của khóa học
type CourseTagListResponse struct {
	Data []CourseTagDTO `json:"data"`
}
