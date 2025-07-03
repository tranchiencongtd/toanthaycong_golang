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

type LectureProgressHandler struct {
	db *sql.DB
}

func NewLectureProgressHandler(db *sql.DB) *LectureProgressHandler {
	return &LectureProgressHandler{db: db}
}

// GetLectureProgresses godoc
// @Summary Lấy danh sách tiến độ bài giảng
// @Description Lấy danh sách tiến độ bài giảng với phân trang và lọc
// @Tags LectureProgress
// @Accept json
// @Produce json
// @Param page query int false "Số trang" default(1)
// @Param limit query int false "Số item mỗi trang" default(10)
// @Param user_id query string false "Lọc theo user ID"
// @Param lecture_id query string false "Lọc theo lecture ID"
// @Param completed query bool false "Lọc theo trạng thái hoàn thành"
// @Success 200 {object} dto.LectureProgressListResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/lecture-progress [get]
func (h *LectureProgressHandler) GetLectureProgresses(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	userID := c.Query("user_id")
	lectureID := c.Query("lecture_id")
	completed := c.Query("completed")

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	offset := (page - 1) * limit

	// Build query với filters
	query := `
		SELECT lp.id, lp.user_id, lp.lecture_id, lp.is_completed, 
		       lp.watch_time, lp.completed_at, lp.created_at, lp.updated_at,
		       COUNT(*) OVER() as total_count
		FROM lecture_progress lp
		WHERE 1=1`
	
	args := []interface{}{}
	argIndex := 1

	if userID != "" {
		query += fmt.Sprintf(" AND lp.user_id = $%d", argIndex)
		args = append(args, userID)
		argIndex++
	}

	if lectureID != "" {
		query += fmt.Sprintf(" AND lp.lecture_id = $%d", argIndex)
		args = append(args, lectureID)
		argIndex++
	}

	if completed != "" {
		if completed == "true" {
			query += fmt.Sprintf(" AND lp.is_completed = $%d", argIndex)
			args = append(args, true)
			argIndex++
		} else if completed == "false" {
			query += fmt.Sprintf(" AND lp.is_completed = $%d", argIndex)
			args = append(args, false)
			argIndex++
		}
	}

	query += fmt.Sprintf(" ORDER BY lp.updated_at DESC LIMIT $%d OFFSET $%d", argIndex, argIndex+1)
	args = append(args, limit, offset)

	rows, err := h.db.Query(query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "Database error",
			Message: "Failed to fetch lecture progress",
		})
		return
	}
	defer rows.Close()

	var progresses []dto.LectureProgressDTO
	var totalCount int

	for rows.Next() {
		var progress dto.LectureProgressDTO
		var completedAt sql.NullTime

		err := rows.Scan(
			&progress.ID, &progress.UserID, &progress.LectureID,
			&progress.IsCompleted, &progress.WatchTime, &completedAt,
			&progress.CreatedAt, &progress.UpdatedAt, &totalCount,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
				Error:   "Scan error",
				Message: "Failed to parse lecture progress data",
			})
			return
		}

		if completedAt.Valid {
			progress.CompletedAt = &completedAt.Time
		}

		progresses = append(progresses, progress)
	}

	// Tính toán pagination
	totalPages := (totalCount + limit - 1) / limit

	response := dto.LectureProgressListResponse{
		Data: progresses,
		Pagination: dto.PaginationMeta{
			Page:       page,
			Limit:      limit,
			TotalItems: totalCount,
			TotalPages: totalPages,
		},
	}

	c.JSON(http.StatusOK, response)
}

// GetLectureProgress godoc
// @Summary Lấy thông tin tiến độ bài giảng theo ID
// @Description Lấy thông tin chi tiết tiến độ bài giảng theo ID
// @Tags LectureProgress
// @Accept json
// @Produce json
// @Param id path string true "ID tiến độ bài giảng"
// @Success 200 {object} dto.LectureProgressDTO
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/lecture-progress/{id} [get]
func (h *LectureProgressHandler) GetLectureProgress(c *gin.Context) {
	id := c.Param("id")

	if _, err := uuid.Parse(id); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "Invalid ID",
			Message: "Invalid lecture progress ID format",
		})
		return
	}

	query := `
		SELECT id, user_id, lecture_id, is_completed, watch_time, 
		       completed_at, created_at, updated_at
		FROM lecture_progress 
		WHERE id = $1`

	var progress dto.LectureProgressDTO
	var completedAt sql.NullTime

	err := h.db.QueryRow(query, id).Scan(
		&progress.ID, &progress.UserID, &progress.LectureID,
		&progress.IsCompleted, &progress.WatchTime, &completedAt,
		&progress.CreatedAt, &progress.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, dto.ErrorResponse{
				Error:   "Not found",
				Message: "Lecture progress not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "Database error",
			Message: "Failed to fetch lecture progress",
		})
		return
	}

	if completedAt.Valid {
		progress.CompletedAt = &completedAt.Time
	}

	c.JSON(http.StatusOK, progress)
}

// CreateLectureProgress godoc
// @Summary Tạo tiến độ bài giảng mới
// @Description Tạo tiến độ bài giảng mới
// @Tags LectureProgress
// @Accept json
// @Produce json
// @Param body body dto.CreateLectureProgressRequest true "Thông tin tiến độ bài giảng"
// @Success 201 {object} dto.LectureProgressDTO
// @Failure 400 {object} dto.ErrorResponse
// @Failure 409 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/lecture-progress [post]
func (h *LectureProgressHandler) CreateLectureProgress(c *gin.Context) {
	var req dto.CreateLectureProgressRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "Invalid input",
			Message: err.Error(),
		})
		return
	}

	// Kiểm tra user và lecture tồn tại
	var userExists, lectureExists bool
	h.db.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE id = $1)", req.UserID).Scan(&userExists)
	h.db.QueryRow("SELECT EXISTS(SELECT 1 FROM course_lectures WHERE id = $1)", req.LectureID).Scan(&lectureExists)

	if !userExists {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "Invalid user",
			Message: "User not found",
		})
		return
	}

	if !lectureExists {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "Invalid lecture",
			Message: "Lecture not found",
		})
		return
	}

	// Set default values
	isCompleted := false
	if req.IsCompleted != nil {
		isCompleted = *req.IsCompleted
	}

	watchTime := 0
	if req.WatchTime != nil {
		watchTime = *req.WatchTime
	}

	var completedAt sql.NullTime
	if isCompleted {
		completedAt = sql.NullTime{Time: time.Now(), Valid: true}
	}

	id := uuid.New().String()
	now := time.Now()

	query := `
		INSERT INTO lecture_progress (id, user_id, lecture_id, is_completed, watch_time, completed_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, user_id, lecture_id, is_completed, watch_time, completed_at, created_at, updated_at`

	var progress dto.LectureProgressDTO

	err := h.db.QueryRow(query, id, req.UserID, req.LectureID, isCompleted, watchTime, completedAt, now, now).Scan(
		&progress.ID, &progress.UserID, &progress.LectureID,
		&progress.IsCompleted, &progress.WatchTime, &completedAt,
		&progress.CreatedAt, &progress.UpdatedAt,
	)

	if err != nil {
		if err.Error() == `pq: duplicate key value violates unique constraint "lecture_progress_user_id_lecture_id_key"` {
			c.JSON(http.StatusConflict, dto.ErrorResponse{
				Error:   "Conflict",
				Message: "Lecture progress already exists for this user and lecture",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "Database error",
			Message: "Failed to create lecture progress",
		})
		return
	}

	if completedAt.Valid {
		progress.CompletedAt = &completedAt.Time
	}

	c.JSON(http.StatusCreated, progress)
}

// UpdateLectureProgress godoc
// @Summary Cập nhật tiến độ bài giảng
// @Description Cập nhật thông tin tiến độ bài giảng
// @Tags LectureProgress
// @Accept json
// @Produce json
// @Param id path string true "ID tiến độ bài giảng"
// @Param body body dto.UpdateLectureProgressRequest true "Thông tin cập nhật"
// @Success 200 {object} dto.LectureProgressDTO
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/lecture-progress/{id} [put]
func (h *LectureProgressHandler) UpdateLectureProgress(c *gin.Context) {
	id := c.Param("id")

	if _, err := uuid.Parse(id); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "Invalid ID",
			Message: "Invalid lecture progress ID format",
		})
		return
	}

	var req dto.UpdateLectureProgressRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "Invalid input",
			Message: err.Error(),
		})
		return
	}

	// Kiểm tra progress tồn tại
	var exists bool
	err := h.db.QueryRow("SELECT EXISTS(SELECT 1 FROM lecture_progress WHERE id = $1)", id).Scan(&exists)
	if err != nil || !exists {
		c.JSON(http.StatusNotFound, dto.ErrorResponse{
			Error:   "Not found",
			Message: "Lecture progress not found",
		})
		return
	}

	// Build update query
	setParts := []string{}
	args := []interface{}{}
	argIndex := 1

	if req.IsCompleted != nil {
		setParts = append(setParts, fmt.Sprintf("is_completed = $%d", argIndex))
		args = append(args, *req.IsCompleted)
		argIndex++

		// Update completed_at based on is_completed
		if *req.IsCompleted {
			setParts = append(setParts, fmt.Sprintf("completed_at = $%d", argIndex))
			args = append(args, time.Now())
			argIndex++
		} else {
			setParts = append(setParts, fmt.Sprintf("completed_at = $%d", argIndex))
			args = append(args, nil)
			argIndex++
		}
	}

	if req.WatchTime != nil {
		setParts = append(setParts, fmt.Sprintf("watch_time = $%d", argIndex))
		args = append(args, *req.WatchTime)
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
		UPDATE lecture_progress SET %s 
		WHERE id = $%d
		RETURNING id, user_id, lecture_id, is_completed, watch_time, completed_at, created_at, updated_at`,
		fmt.Sprintf("%s", setParts[0]),
		argIndex,
	)

	for i := 1; i < len(setParts); i++ {
		query = fmt.Sprintf(`
			UPDATE lecture_progress SET %s, %s 
			WHERE id = $%d
			RETURNING id, user_id, lecture_id, is_completed, watch_time, completed_at, created_at, updated_at`,
			setParts[0], setParts[i],
			argIndex,
		)
	}

	args = append(args, id)

	var progress dto.LectureProgressDTO
	var completedAt sql.NullTime

	err = h.db.QueryRow(query, args...).Scan(
		&progress.ID, &progress.UserID, &progress.LectureID,
		&progress.IsCompleted, &progress.WatchTime, &completedAt,
		&progress.CreatedAt, &progress.UpdatedAt,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "Database error",
			Message: "Failed to update lecture progress",
		})
		return
	}

	if completedAt.Valid {
		progress.CompletedAt = &completedAt.Time
	}

	c.JSON(http.StatusOK, progress)
}

// DeleteLectureProgress godoc
// @Summary Xóa tiến độ bài giảng
// @Description Xóa tiến độ bài giảng theo ID
// @Tags LectureProgress
// @Accept json
// @Produce json
// @Param id path string true "ID tiến độ bài giảng"
// @Success 204 "No Content"
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/lecture-progress/{id} [delete]
func (h *LectureProgressHandler) DeleteLectureProgress(c *gin.Context) {
	id := c.Param("id")

	if _, err := uuid.Parse(id); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "Invalid ID",
			Message: "Invalid lecture progress ID format",
		})
		return
	}

	result, err := h.db.Exec("DELETE FROM lecture_progress WHERE id = $1", id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "Database error",
			Message: "Failed to delete lecture progress",
		})
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, dto.ErrorResponse{
			Error:   "Not found",
			Message: "Lecture progress not found",
		})
		return
	}

	c.Status(http.StatusNoContent)
}
