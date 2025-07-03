package dto

import (
	"time"
)

// Enrollment DTOs
type CreateEnrollmentRequest struct {
	UserID   string `json:"user_id" binding:"required"`
	CourseID string `json:"course_id" binding:"required"`
}

type UpdateEnrollmentRequest struct {
	ProgressPercentage *float64   `json:"progress_percentage" binding:"omitempty,min=0,max=100"`
	CompletedAt        *time.Time `json:"completed_at"`
	CertificateURL     *string    `json:"certificate_url"`
}

type EnrollmentResponse struct {
	ID                 string     `json:"id"`
	UserID             string     `json:"user_id"`
	CourseID           string     `json:"course_id"`
	EnrolledAt         time.Time  `json:"enrolled_at"`
	CompletedAt        *time.Time `json:"completed_at"`
	ProgressPercentage float64    `json:"progress_percentage"`
	LastAccessedAt     *time.Time `json:"last_accessed_at"`
	CertificateURL     *string    `json:"certificate_url"`
}

type EnrollmentListResponse struct {
	Enrollments []EnrollmentResponse `json:"enrollments"`
	Pagination  PaginationResponse   `json:"pagination"`
}

// Lecture Progress DTOs
type CreateLectureProgressRequest struct {
	UserID    string `json:"user_id" binding:"required"`
	LectureID string `json:"lecture_id" binding:"required"`
	WatchTime *int32 `json:"watch_time" binding:"omitempty,min=0"`
}

type UpdateLectureProgressRequest struct {
	IsCompleted *bool  `json:"is_completed"`
	WatchTime   *int32 `json:"watch_time" binding:"omitempty,min=0"`
}

type LectureProgressResponse struct {
	ID          string     `json:"id"`
	UserID      string     `json:"user_id"`
	LectureID   string     `json:"lecture_id"`
	IsCompleted bool       `json:"is_completed"`
	WatchTime   int32      `json:"watch_time"`
	CompletedAt *time.Time `json:"completed_at"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

type LectureProgressListResponse struct {
	Progress   []LectureProgressResponse `json:"progress"`
	Pagination PaginationResponse        `json:"pagination"`
}

// Wishlist DTOs
type CreateWishlistRequest struct {
	UserID   string `json:"user_id" binding:"required"`
	CourseID string `json:"course_id" binding:"required"`
}

type WishlistResponse struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	CourseID  string    `json:"course_id"`
	CreatedAt time.Time `json:"created_at"`
}

type WishlistListResponse struct {
	Wishlists  []WishlistResponse `json:"wishlists"`
	Pagination PaginationResponse `json:"pagination"`
}
