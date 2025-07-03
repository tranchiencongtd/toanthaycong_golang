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

type CourseAnnouncementHandler struct {
	db *sql.DB
}

func NewCourseAnnouncementHandler(db *sql.DB) *CourseAnnouncementHandler {
	return &CourseAnnouncementHandler{db: db}
}

// GetCourseAnnouncements godoc
// @Summary Lấy danh sách thông báo khóa học
// @Description Lấy danh sách thông báo khóa học với phân trang và lọc
// @Tags CourseAnnouncement
// @Accept json
// @Produce json
// @Param page query int false "Số trang" default(1)
// @Param limit query int false "Số item mỗi trang" default(10)
// @Param course_id query string false "Lọc theo course ID"
// @Param published query bool false "Lọc theo trạng thái published"
// @Success 200 {object} dto.CourseAnnouncementListResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/course-announcements [get]
func (h *CourseAnnouncementHandler) GetCourseAnnouncements(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	courseID := c.Query("course_id")
	published := c.Query("published")

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	offset := (page - 1) * limit

	// Build query với filters
	query := `
		SELECT ca.id, ca.course_id, ca.title, ca.content, ca.is_published,
		       ca.created_at, ca.updated_at,
		       COUNT(*) OVER() as total_count
		FROM course_announcements ca
		WHERE 1=1`
	
	args := []interface{}{}
	argIndex := 1

	if courseID != "" {
		query += fmt.Sprintf(" AND ca.course_id = $%d", argIndex)
		args = append(args, courseID)
		argIndex++
	}

	if published != "" {
		if published == "true" {
			query += fmt.Sprintf(" AND ca.is_published = $%d", argIndex)
			args = append(args, true)
			argIndex++
		} else if published == "false" {
			query += fmt.Sprintf(" AND ca.is_published = $%d", argIndex)
			args = append(args, false)
			argIndex++
		}
	}

	query += fmt.Sprintf(" ORDER BY ca.created_at DESC LIMIT $%d OFFSET $%d", argIndex, argIndex+1)
	args = append(args, limit, offset)

	rows, err := h.db.Query(query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "Database error",
			Message: "Failed to fetch course announcements",
		})
		return
	}
	defer rows.Close()

	var announcements []dto.CourseAnnouncementDTO
	var totalCount int

	for rows.Next() {
		var announcement dto.CourseAnnouncementDTO

		err := rows.Scan(
			&announcement.ID, &announcement.CourseID, &announcement.Title,
			&announcement.Content, &announcement.IsPublished,
			&announcement.CreatedAt, &announcement.UpdatedAt, &totalCount,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
				Error:   "Scan error",
				Message: "Failed to parse course announcement data",
			})
			return
		}

		announcements = append(announcements, announcement)
	}

	// Tính toán pagination
	totalPages := (totalCount + limit - 1) / limit

	response := dto.CourseAnnouncementListResponse{
		Data: announcements,
		Pagination: dto.PaginationMeta{
			Page:       page,
			Limit:      limit,
			TotalItems: totalCount,
			TotalPages: totalPages,
		},
	}

	c.JSON(http.StatusOK, response)
}

// GetCourseAnnouncement godoc
// @Summary Lấy thông tin thông báo khóa học theo ID
// @Description Lấy thông tin chi tiết thông báo khóa học theo ID
// @Tags CourseAnnouncement
// @Accept json
// @Produce json
// @Param id path string true "ID thông báo khóa học"
// @Success 200 {object} dto.CourseAnnouncementDTO
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/course-announcements/{id} [get]
func (h *CourseAnnouncementHandler) GetCourseAnnouncement(c *gin.Context) {
	id := c.Param("id")

	if _, err := uuid.Parse(id); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "Invalid ID",
			Message: "Invalid course announcement ID format",
		})
		return
	}

	query := `
		SELECT id, course_id, title, content, is_published, created_at, updated_at
		FROM course_announcements 
		WHERE id = $1`

	var announcement dto.CourseAnnouncementDTO

	err := h.db.QueryRow(query, id).Scan(
		&announcement.ID, &announcement.CourseID, &announcement.Title,
		&announcement.Content, &announcement.IsPublished,
		&announcement.CreatedAt, &announcement.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, dto.ErrorResponse{
				Error:   "Not found",
				Message: "Course announcement not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "Database error",
			Message: "Failed to fetch course announcement",
		})
		return
	}

	c.JSON(http.StatusOK, announcement)
}

// CreateCourseAnnouncement godoc
// @Summary Tạo thông báo khóa học mới
// @Description Tạo thông báo khóa học mới
// @Tags CourseAnnouncement
// @Accept json
// @Produce json
// @Param body body dto.CreateCourseAnnouncementRequest true "Thông tin thông báo khóa học"
// @Success 201 {object} dto.CourseAnnouncementDTO
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/course-announcements [post]
func (h *CourseAnnouncementHandler) CreateCourseAnnouncement(c *gin.Context) {
	var req dto.CreateCourseAnnouncementRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "Invalid input",
			Message: err.Error(),
		})
		return
	}

	// Kiểm tra course tồn tại
	var courseExists bool
	h.db.QueryRow("SELECT EXISTS(SELECT 1 FROM courses WHERE id = $1)", req.CourseID).Scan(&courseExists)
	if !courseExists {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "Invalid course",
			Message: "Course not found",
		})
		return
	}

	// Set default values
	isPublished := false
	if req.IsPublished != nil {
		isPublished = *req.IsPublished
	}

	id := uuid.New().String()
	now := time.Now()

	query := `
		INSERT INTO course_announcements (id, course_id, title, content, is_published, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, course_id, title, content, is_published, created_at, updated_at`

	var announcement dto.CourseAnnouncementDTO

	err := h.db.QueryRow(query, id, req.CourseID, req.Title, req.Content, isPublished, now, now).Scan(
		&announcement.ID, &announcement.CourseID, &announcement.Title,
		&announcement.Content, &announcement.IsPublished,
		&announcement.CreatedAt, &announcement.UpdatedAt,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "Database error",
			Message: "Failed to create course announcement",
		})
		return
	}

	c.JSON(http.StatusCreated, announcement)
}

// UpdateCourseAnnouncement godoc
// @Summary Cập nhật thông báo khóa học
// @Description Cập nhật thông tin thông báo khóa học
// @Tags CourseAnnouncement
// @Accept json
// @Produce json
// @Param id path string true "ID thông báo khóa học"
// @Param body body dto.UpdateCourseAnnouncementRequest true "Thông tin cập nhật"
// @Success 200 {object} dto.CourseAnnouncementDTO
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/course-announcements/{id} [put]
func (h *CourseAnnouncementHandler) UpdateCourseAnnouncement(c *gin.Context) {
	id := c.Param("id")

	if _, err := uuid.Parse(id); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "Invalid ID",
			Message: "Invalid course announcement ID format",
		})
		return
	}

	var req dto.UpdateCourseAnnouncementRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "Invalid input",
			Message: err.Error(),
		})
		return
	}

	// Kiểm tra announcement tồn tại
	var exists bool
	err := h.db.QueryRow("SELECT EXISTS(SELECT 1 FROM course_announcements WHERE id = $1)", id).Scan(&exists)
	if err != nil || !exists {
		c.JSON(http.StatusNotFound, dto.ErrorResponse{
			Error:   "Not found",
			Message: "Course announcement not found",
		})
		return
	}

	// Build update query
	setParts := []string{}
	args := []interface{}{}
	argIndex := 1

	if req.Title != nil {
		setParts = append(setParts, fmt.Sprintf("title = $%d", argIndex))
		args = append(args, *req.Title)
		argIndex++
	}

	if req.Content != nil {
		setParts = append(setParts, fmt.Sprintf("content = $%d", argIndex))
		args = append(args, *req.Content)
		argIndex++
	}

	if req.IsPublished != nil {
		setParts = append(setParts, fmt.Sprintf("is_published = $%d", argIndex))
		args = append(args, *req.IsPublished)
		argIndex++
	}

	if len(setParts) == 0 {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "No updates",
			Message: "No fields to update",
		})
		return
	}

	setParts = append(setParts, fmt.Sprintf("updated_at = $%d", argIndex))
	args = append(args, time.Now())
	argIndex++

	query := fmt.Sprintf(`
		UPDATE course_announcements SET %s 
		WHERE id = $%d
		RETURNING id, course_id, title, content, is_published, created_at, updated_at`,
		fmt.Sprintf("%s", setParts[0]),
		argIndex,
	)

	for i := 1; i < len(setParts); i++ {
		query = fmt.Sprintf(`
			UPDATE course_announcements SET %s, %s 
			WHERE id = $%d
			RETURNING id, course_id, title, content, is_published, created_at, updated_at`,
			setParts[0], setParts[i],
			argIndex,
		)
	}

	args = append(args, id)

	var announcement dto.CourseAnnouncementDTO

	err = h.db.QueryRow(query, args...).Scan(
		&announcement.ID, &announcement.CourseID, &announcement.Title,
		&announcement.Content, &announcement.IsPublished,
		&announcement.CreatedAt, &announcement.UpdatedAt,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "Database error",
			Message: "Failed to update course announcement",
		})
		return
	}

	c.JSON(http.StatusOK, announcement)
}

// DeleteCourseAnnouncement godoc
// @Summary Xóa thông báo khóa học
// @Description Xóa thông báo khóa học theo ID
// @Tags CourseAnnouncement
// @Accept json
// @Produce json
// @Param id path string true "ID thông báo khóa học"
// @Success 204 "No Content"
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/course-announcements/{id} [delete]
func (h *CourseAnnouncementHandler) DeleteCourseAnnouncement(c *gin.Context) {
	id := c.Param("id")

	if _, err := uuid.Parse(id); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "Invalid ID",
			Message: "Invalid course announcement ID format",
		})
		return
	}

	result, err := h.db.Exec("DELETE FROM course_announcements WHERE id = $1", id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "Database error",
			Message: "Failed to delete course announcement",
		})
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, dto.ErrorResponse{
			Error:   "Not found",
			Message: "Course announcement not found",
		})
		return
	}

	c.Status(http.StatusNoContent)
}
