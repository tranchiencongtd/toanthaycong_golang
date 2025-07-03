package dto

import "time"

// CourseReviewDTO - DTO cho đánh giá khóa học
type CourseReviewDTO struct {
	ID         string    `json:"id"`
	UserID     string    `json:"user_id"`
	CourseID   string    `json:"course_id"`
	Rating     int       `json:"rating"`
	ReviewText *string   `json:"review_text,omitempty"`
	IsApproved bool      `json:"is_approved"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
	
	// Thông tin liên quan
	User   *UserDTO   `json:"user,omitempty"`
	Course *CourseDTO `json:"course,omitempty"`
}

// CreateCourseReviewRequest - Request tạo đánh giá khóa học
type CreateCourseReviewRequest struct {
	UserID     string  `json:"user_id" binding:"required,uuid"`
	CourseID   string  `json:"course_id" binding:"required,uuid"`
	Rating     int     `json:"rating" binding:"required,min=1,max=5"`
	ReviewText *string `json:"review_text,omitempty"`
}

// UpdateCourseReviewRequest - Request cập nhật đánh giá khóa học
type UpdateCourseReviewRequest struct {
	Rating     *int    `json:"rating,omitempty" binding:"omitempty,min=1,max=5"`
	ReviewText *string `json:"review_text,omitempty"`
	IsApproved *bool   `json:"is_approved,omitempty"`
}

// CourseReviewListResponse - Response danh sách đánh giá khóa học
type CourseReviewListResponse struct {
	Data       []CourseReviewDTO `json:"data"`
	Pagination PaginationMeta    `json:"pagination"`
}

// CourseReviewStatsDTO - Thống kê đánh giá khóa học
type CourseReviewStatsDTO struct {
	CourseID     string  `json:"course_id"`
	TotalReviews int     `json:"total_reviews"`
	AverageRating float64 `json:"average_rating"`
	RatingDistribution map[int]int `json:"rating_distribution"` // key: rating (1-5), value: count
}
