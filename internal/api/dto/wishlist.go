package dto

import "time"

// WishlistDTO - DTO cho danh sách yêu thích
type WishlistDTO struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	CourseID  string    `json:"course_id"`
	CreatedAt time.Time `json:"created_at"`
	
	// Thông tin liên quan
	Course *CourseDTO `json:"course,omitempty"`
}

// CreateWishlistRequest - Request thêm vào danh sách yêu thích
type CreateWishlistRequest struct {
	UserID   string `json:"user_id" binding:"required,uuid"`
	CourseID string `json:"course_id" binding:"required,uuid"`
}

// WishlistListResponse - Response danh sách yêu thích
type WishlistListResponse struct {
	Data       []WishlistDTO  `json:"data"`
	Pagination PaginationMeta `json:"pagination"`
}
