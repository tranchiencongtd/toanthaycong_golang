package dto

import "time"

// CourseAnnouncementDTO - DTO cho thông báo khóa học
type CourseAnnouncementDTO struct {
	ID          string    `json:"id"`
	CourseID    string    `json:"course_id"`
	Title       string    `json:"title"`
	Content     string    `json:"content"`
	IsPublished bool      `json:"is_published"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	
	// Thông tin liên quan
	Course *CourseDTO `json:"course,omitempty"`
}

// CreateCourseAnnouncementRequest - Request tạo thông báo khóa học
type CreateCourseAnnouncementRequest struct {
	CourseID    string `json:"course_id" binding:"required,uuid"`
	Title       string `json:"title" binding:"required,max=200"`
	Content     string `json:"content" binding:"required"`
	IsPublished *bool  `json:"is_published,omitempty"`
}

// UpdateCourseAnnouncementRequest - Request cập nhật thông báo khóa học
type UpdateCourseAnnouncementRequest struct {
	Title       *string `json:"title,omitempty" binding:"omitempty,max=200"`
	Content     *string `json:"content,omitempty"`
	IsPublished *bool   `json:"is_published,omitempty"`
}

// CourseAnnouncementListResponse - Response danh sách thông báo khóa học
type CourseAnnouncementListResponse struct {
	Data       []CourseAnnouncementDTO `json:"data"`
	Pagination PaginationMeta          `json:"pagination"`
}
