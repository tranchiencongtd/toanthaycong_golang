package handlers

import (
	"database/sql"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/toanthaycong_golang/internal/api/dto"
)

type NotificationHandler struct {
	db *sql.DB
}

func NewNotificationHandler(db *sql.DB) *NotificationHandler {
	return &NotificationHandler{db: db}
}

// GetNotifications godoc
// @Summary Lấy danh sách thông báo
// @Description Lấy danh sách thông báo với phân trang và lọc
// @Tags Notification
// @Accept json
// @Produce json
// @Param page query int false "Số trang" default(1)
// @Param limit query int false "Số item mỗi trang" default(10)
// @Param user_id query string false "Lọc theo user ID"
// @Param type query string false "Lọc theo loại thông báo"
// @Param read query bool false "Lọc theo trạng thái đã đọc"
// @Success 200 {object} dto.NotificationListResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/notifications [get]
func (h *NotificationHandler) GetNotifications(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	userID := c.Query("user_id")
	notificationType := c.Query("type")
	read := c.Query("read")

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	offset := (page - 1) * limit

	// Build query với filters
	query := `
		SELECT n.id, n.user_id, n.title, n.message, n.type, n.related_id, n.is_read, n.created_at,
		       COUNT(*) OVER() as total_count
		FROM notifications n
		WHERE 1=1`
	
	args := []interface{}{}
	argIndex := 1

	if userID != "" {
		query += fmt.Sprintf(" AND n.user_id = $%d", argIndex)
		args = append(args, userID)
		argIndex++
	}

	if notificationType != "" {
		query += fmt.Sprintf(" AND n.type = $%d", argIndex)
		args = append(args, notificationType)
		argIndex++
	}

	if read != "" {
		if read == "true" {
			query += fmt.Sprintf(" AND n.is_read = $%d", argIndex)
			args = append(args, true)
			argIndex++
		} else if read == "false" {
			query += fmt.Sprintf(" AND n.is_read = $%d", argIndex)
			args = append(args, false)
			argIndex++
		}
	}

	query += fmt.Sprintf(" ORDER BY n.created_at DESC LIMIT $%d OFFSET $%d", argIndex, argIndex+1)
	args = append(args, limit, offset)

	rows, err := h.db.Query(query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "Database error",
			Message: "Failed to fetch notifications",
		})
		return
	}
	defer rows.Close()

	var notifications []dto.NotificationDTO
	var totalCount int

	for rows.Next() {
		var notification dto.NotificationDTO
		var relatedID sql.NullString

		err := rows.Scan(
			&notification.ID, &notification.UserID, &notification.Title,
			&notification.Message, &notification.Type, &relatedID,
			&notification.IsRead, &notification.CreatedAt, &totalCount,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
				Error:   "Scan error",
				Message: "Failed to parse notification data",
			})
			return
		}

		if relatedID.Valid {
			notification.RelatedID = &relatedID.String
		}

		notifications = append(notifications, notification)
	}

	// Tính toán pagination
	totalPages := (totalCount + limit - 1) / limit

	response := dto.NotificationListResponse{
		Data: notifications,
		Pagination: dto.PaginationMeta{
			Page:       page,
			Limit:      limit,
			TotalItems: totalCount,
			TotalPages: totalPages,
		},
	}

	c.JSON(http.StatusOK, response)
}

// GetNotification godoc
// @Summary Lấy thông tin thông báo theo ID
// @Description Lấy thông tin chi tiết thông báo theo ID
// @Tags Notification
// @Accept json
// @Produce json
// @Param id path string true "ID thông báo"
// @Success 200 {object} dto.NotificationDTO
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/notifications/{id} [get]
func (h *NotificationHandler) GetNotification(c *gin.Context) {
	id := c.Param("id")

	if _, err := uuid.Parse(id); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "Invalid ID",
			Message: "Invalid notification ID format",
		})
		return
	}

	query := `
		SELECT id, user_id, title, message, type, related_id, is_read, created_at
		FROM notifications 
		WHERE id = $1`

	var notification dto.NotificationDTO
	var relatedID sql.NullString

	err := h.db.QueryRow(query, id).Scan(
		&notification.ID, &notification.UserID, &notification.Title,
		&notification.Message, &notification.Type, &relatedID,
		&notification.IsRead, &notification.CreatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, dto.ErrorResponse{
				Error:   "Not found",
				Message: "Notification not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "Database error",
			Message: "Failed to fetch notification",
		})
		return
	}

	if relatedID.Valid {
		notification.RelatedID = &relatedID.String
	}

	c.JSON(http.StatusOK, notification)
}

// CreateNotification godoc
// @Summary Tạo thông báo mới
// @Description Tạo thông báo mới
// @Tags Notification
// @Accept json
// @Produce json
// @Param body body dto.CreateNotificationRequest true "Thông tin thông báo"
// @Success 201 {object} dto.NotificationDTO
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/notifications [post]
func (h *NotificationHandler) CreateNotification(c *gin.Context) {
	var req dto.CreateNotificationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "Invalid input",
			Message: err.Error(),
		})
		return
	}

	// Kiểm tra user tồn tại
	var userExists bool
	h.db.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE id = $1)", req.UserID).Scan(&userExists)
	if !userExists {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "Invalid user",
			Message: "User not found",
		})
		return
	}

	id := uuid.New().String()
	now := time.Now()

	var relatedID sql.NullString
	if req.RelatedID != nil {
		relatedID = sql.NullString{String: *req.RelatedID, Valid: true}
	}

	query := `
		INSERT INTO notifications (id, user_id, title, message, type, related_id, is_read, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, user_id, title, message, type, related_id, is_read, created_at`

	var notification dto.NotificationDTO

	err := h.db.QueryRow(query, id, req.UserID, req.Title, req.Message, req.Type, relatedID, false, now).Scan(
		&notification.ID, &notification.UserID, &notification.Title,
		&notification.Message, &notification.Type, &relatedID,
		&notification.IsRead, &notification.CreatedAt,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "Database error",
			Message: "Failed to create notification",
		})
		return
	}

	if relatedID.Valid {
		notification.RelatedID = &relatedID.String
	}

	c.JSON(http.StatusCreated, notification)
}

// UpdateNotification godoc
// @Summary Cập nhật thông báo
// @Description Cập nhật trạng thái thông báo (chủ yếu để đánh dấu đã đọc)
// @Tags Notification
// @Accept json
// @Produce json
// @Param id path string true "ID thông báo"
// @Param body body dto.UpdateNotificationRequest true "Thông tin cập nhật"
// @Success 200 {object} dto.NotificationDTO
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/notifications/{id} [put]
func (h *NotificationHandler) UpdateNotification(c *gin.Context) {
	id := c.Param("id")

	if _, err := uuid.Parse(id); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "Invalid ID",
			Message: "Invalid notification ID format",
		})
		return
	}

	var req dto.UpdateNotificationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "Invalid input",
			Message: err.Error(),
		})
		return
	}

	// Kiểm tra notification tồn tại
	var exists bool
	err := h.db.QueryRow("SELECT EXISTS(SELECT 1 FROM notifications WHERE id = $1)", id).Scan(&exists)
	if err != nil || !exists {
		c.JSON(http.StatusNotFound, dto.ErrorResponse{
			Error:   "Not found",
			Message: "Notification not found",
		})
		return
	}

	if req.IsRead == nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "No updates",
			Message: "No fields to update",
		})
		return
	}

	query := `
		UPDATE notifications SET is_read = $1
		WHERE id = $2
		RETURNING id, user_id, title, message, type, related_id, is_read, created_at`

	var notification dto.NotificationDTO
	var relatedID sql.NullString

	err = h.db.QueryRow(query, *req.IsRead, id).Scan(
		&notification.ID, &notification.UserID, &notification.Title,
		&notification.Message, &notification.Type, &relatedID,
		&notification.IsRead, &notification.CreatedAt,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "Database error",
			Message: "Failed to update notification",
		})
		return
	}

	if relatedID.Valid {
		notification.RelatedID = &relatedID.String
	}

	c.JSON(http.StatusOK, notification)
}

// MarkAllAsRead godoc
// @Summary Đánh dấu tất cả thông báo đã đọc
// @Description Đánh dấu tất cả thông báo của user đã đọc
// @Tags Notification
// @Accept json
// @Produce json
// @Param body body dto.MarkAllAsReadRequest true "User ID"
// @Success 200 {object} map[string]interface{} "{"updated_count": 5}"
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/notifications/mark-all-read [put]
func (h *NotificationHandler) MarkAllAsRead(c *gin.Context) {
	var req dto.MarkAllAsReadRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "Invalid input",
			Message: err.Error(),
		})
		return
	}

	// Kiểm tra user tồn tại
	var userExists bool
	h.db.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE id = $1)", req.UserID).Scan(&userExists)
	if !userExists {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "Invalid user",
			Message: "User not found",
		})
		return
	}

	result, err := h.db.Exec("UPDATE notifications SET is_read = true WHERE user_id = $1 AND is_read = false", req.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "Database error",
			Message: "Failed to mark notifications as read",
		})
		return
	}

	rowsAffected, _ := result.RowsAffected()

	c.JSON(http.StatusOK, gin.H{
		"updated_count": rowsAffected,
	})
}

// DeleteNotification godoc
// @Summary Xóa thông báo
// @Description Xóa thông báo theo ID
// @Tags Notification
// @Accept json
// @Produce json
// @Param id path string true "ID thông báo"
// @Success 204 "No Content"
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/notifications/{id} [delete]
func (h *NotificationHandler) DeleteNotification(c *gin.Context) {
	id := c.Param("id")

	if _, err := uuid.Parse(id); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "Invalid ID",
			Message: "Invalid notification ID format",
		})
		return
	}

	result, err := h.db.Exec("DELETE FROM notifications WHERE id = $1", id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "Database error",
			Message: "Failed to delete notification",
		})
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, dto.ErrorResponse{
			Error:   "Not found",
			Message: "Notification not found",
		})
		return
	}

	c.Status(http.StatusNoContent)
}

// GetNotificationStats godoc
// @Summary Lấy thống kê thông báo
// @Description Lấy thống kê thông báo của user
// @Tags Notification
// @Accept json
// @Produce json
// @Param user_id path string true "ID người dùng"
// @Success 200 {object} dto.NotificationStatsDTO
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/users/{user_id}/notification-stats [get]
func (h *NotificationHandler) GetNotificationStats(c *gin.Context) {
	userID := c.Param("user_id")

	if _, err := uuid.Parse(userID); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "Invalid ID",
			Message: "Invalid user ID format",
		})
		return
	}

	// Kiểm tra user tồn tại
	var userExists bool
	h.db.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE id = $1)", userID).Scan(&userExists)
	if !userExists {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "Invalid user",
			Message: "User not found",
		})
		return
	}

	var stats dto.NotificationStatsDTO
	stats.UserID = userID

	query := `
		SELECT 
			COUNT(*) as total_count,
			COUNT(CASE WHEN is_read = false THEN 1 END) as unread_count,
			COUNT(CASE WHEN is_read = true THEN 1 END) as read_count
		FROM notifications 
		WHERE user_id = $1`

	err := h.db.QueryRow(query, userID).Scan(&stats.TotalCount, &stats.UnreadCount, &stats.ReadCount)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "Database error",
			Message: "Failed to fetch notification stats",
		})
		return
	}

	c.JSON(http.StatusOK, stats)
}
