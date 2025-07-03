package dto

import "time"

// LectureProgressDTO - DTO cho tiến độ bài giảng
type LectureProgressDTO struct {
	ID          string    `json:"id"`
	UserID      string    `json:"user_id"`
	LectureID   string    `json:"lecture_id"`
	IsCompleted bool      `json:"is_completed"`
	WatchTime   int       `json:"watch_time"` // tính bằng giây
	CompletedAt *time.Time `json:"completed_at,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// CreateLectureProgressRequest - Request tạo tiến độ bài giảng
type CreateLectureProgressRequest struct {
	UserID      string `json:"user_id" binding:"required,uuid"`
	LectureID   string `json:"lecture_id" binding:"required,uuid"`
	IsCompleted *bool  `json:"is_completed,omitempty"`
	WatchTime   *int   `json:"watch_time,omitempty"`
}

// UpdateLectureProgressRequest - Request cập nhật tiến độ bài giảng
type UpdateLectureProgressRequest struct {
	IsCompleted *bool `json:"is_completed,omitempty"`
	WatchTime   *int  `json:"watch_time,omitempty"`
}

// LectureProgressListResponse - Response danh sách tiến độ bài giảng
type LectureProgressListResponse struct {
	Data       []LectureProgressDTO `json:"data"`
	Pagination PaginationMeta       `json:"pagination"`
}
