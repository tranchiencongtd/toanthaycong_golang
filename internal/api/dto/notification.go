package dto

import "time"

// NotificationDTO - DTO cho thông báo
type NotificationDTO struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	Title     string    `json:"title"`
	Message   string    `json:"message"`
	Type      string    `json:"type"` // 'course_update', 'new_announcement', 'question_answered', v.v.
	RelatedID *string   `json:"related_id,omitempty"`
	IsRead    bool      `json:"is_read"`
	CreatedAt time.Time `json:"created_at"`
}

// CreateNotificationRequest - Request tạo thông báo
type CreateNotificationRequest struct {
	UserID    string  `json:"user_id" binding:"required,uuid"`
	Title     string  `json:"title" binding:"required,max=200"`
	Message   string  `json:"message" binding:"required"`
	Type      string  `json:"type" binding:"required,max=50"`
	RelatedID *string `json:"related_id,omitempty" binding:"omitempty,uuid"`
}

// UpdateNotificationRequest - Request cập nhật thông báo
type UpdateNotificationRequest struct {
	IsRead *bool `json:"is_read,omitempty"`
}

// NotificationListResponse - Response danh sách thông báo
type NotificationListResponse struct {
	Data       []NotificationDTO `json:"data"`
	Pagination PaginationMeta    `json:"pagination"`
}

// MarkAllAsReadRequest - Request đánh dấu tất cả thông báo đã đọc
type MarkAllAsReadRequest struct {
	UserID string `json:"user_id" binding:"required,uuid"`
}

// NotificationStatsDTO - Thống kê thông báo
type NotificationStatsDTO struct {
	UserID           string `json:"user_id"`
	TotalCount       int    `json:"total_count"`
	UnreadCount      int    `json:"unread_count"`
	ReadCount        int    `json:"read_count"`
}
