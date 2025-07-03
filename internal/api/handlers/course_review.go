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

type CourseReviewHandler struct {
	db *sql.DB
}

func NewCourseReviewHandler(db *sql.DB) *CourseReviewHandler {
	return &CourseReviewHandler{db: db}
}

// GetCourseReviews godoc
// @Summary Lấy danh sách đánh giá khóa học
// @Description Lấy danh sách đánh giá khóa học với phân trang và lọc
// @Tags CourseReview
// @Accept json
// @Produce json
// @Param page query int false "Số trang" default(1)
// @Param limit query int false "Số item mỗi trang" default(10)
// @Param course_id query string false "Lọc theo course ID"
// @Param user_id query string false "Lọc theo user ID"
// @Param rating query int false "Lọc theo rating"
// @Param approved query bool false "Lọc theo trạng thái approved"
// @Success 200 {object} dto.CourseReviewListResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/course-reviews [get]
func (h *CourseReviewHandler) GetCourseReviews(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	courseID := c.Query("course_id")
	userID := c.Query("user_id")
	rating := c.Query("rating")
	approved := c.Query("approved")

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	offset := (page - 1) * limit

	// Build query với filters
	query := `
		SELECT cr.id, cr.user_id, cr.course_id, cr.rating, 
		       cr.review_text, cr.is_approved, cr.created_at, cr.updated_at,
		       COUNT(*) OVER() as total_count
		FROM course_reviews cr
		WHERE 1=1`
	
	args := []interface{}{}
	argIndex := 1

	if courseID != "" {
		query += fmt.Sprintf(" AND cr.course_id = $%d", argIndex)
		args = append(args, courseID)
		argIndex++
	}

	if userID != "" {
		query += fmt.Sprintf(" AND cr.user_id = $%d", argIndex)
		args = append(args, userID)
		argIndex++
	}

	if rating != "" {
		if r, err := strconv.Atoi(rating); err == nil && r >= 1 && r <= 5 {
			query += fmt.Sprintf(" AND cr.rating = $%d", argIndex)
			args = append(args, r)
			argIndex++
		}
	}

	if approved != "" {
		if approved == "true" {
			query += fmt.Sprintf(" AND cr.is_approved = $%d", argIndex)
			args = append(args, true)
			argIndex++
		} else if approved == "false" {
			query += fmt.Sprintf(" AND cr.is_approved = $%d", argIndex)
			args = append(args, false)
			argIndex++
		}
	}

	query += fmt.Sprintf(" ORDER BY cr.created_at DESC LIMIT $%d OFFSET $%d", argIndex, argIndex+1)
	args = append(args, limit, offset)

	rows, err := h.db.Query(query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "Database error",
			Message: "Failed to fetch course reviews",
		})
		return
	}
	defer rows.Close()

	var reviews []dto.CourseReviewDTO
	var totalCount int

	for rows.Next() {
		var review dto.CourseReviewDTO
		var reviewText sql.NullString

		err := rows.Scan(
			&review.ID, &review.UserID, &review.CourseID,
			&review.Rating, &reviewText, &review.IsApproved,
			&review.CreatedAt, &review.UpdatedAt, &totalCount,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
				Error:   "Scan error",
				Message: "Failed to parse course review data",
			})
			return
		}

		if reviewText.Valid {
			review.ReviewText = &reviewText.String
		}

		reviews = append(reviews, review)
	}

	// Tính toán pagination
	totalPages := (totalCount + limit - 1) / limit

	response := dto.CourseReviewListResponse{
		Data: reviews,
		Pagination: dto.PaginationMeta{
			Page:       page,
			Limit:      limit,
			TotalItems: totalCount,
			TotalPages: totalPages,
		},
	}

	c.JSON(http.StatusOK, response)
}

// GetCourseReview godoc
// @Summary Lấy thông tin đánh giá khóa học theo ID
// @Description Lấy thông tin chi tiết đánh giá khóa học theo ID
// @Tags CourseReview
// @Accept json
// @Produce json
// @Param id path string true "ID đánh giá khóa học"
// @Success 200 {object} dto.CourseReviewDTO
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/course-reviews/{id} [get]
func (h *CourseReviewHandler) GetCourseReview(c *gin.Context) {
	id := c.Param("id")

	if _, err := uuid.Parse(id); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "Invalid ID",
			Message: "Invalid course review ID format",
		})
		return
	}

	query := `
		SELECT id, user_id, course_id, rating, review_text, 
		       is_approved, created_at, updated_at
		FROM course_reviews 
		WHERE id = $1`

	var review dto.CourseReviewDTO
	var reviewText sql.NullString

	err := h.db.QueryRow(query, id).Scan(
		&review.ID, &review.UserID, &review.CourseID,
		&review.Rating, &reviewText, &review.IsApproved,
		&review.CreatedAt, &review.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, dto.ErrorResponse{
				Error:   "Not found",
				Message: "Course review not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "Database error",
			Message: "Failed to fetch course review",
		})
		return
	}

	if reviewText.Valid {
		review.ReviewText = &reviewText.String
	}

	c.JSON(http.StatusOK, review)
}

// CreateCourseReview godoc
// @Summary Tạo đánh giá khóa học mới
// @Description Tạo đánh giá khóa học mới
// @Tags CourseReview
// @Accept json
// @Produce json
// @Param body body dto.CreateCourseReviewRequest true "Thông tin đánh giá khóa học"
// @Success 201 {object} dto.CourseReviewDTO
// @Failure 400 {object} dto.ErrorResponse
// @Failure 409 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/course-reviews [post]
func (h *CourseReviewHandler) CreateCourseReview(c *gin.Context) {
	var req dto.CreateCourseReviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "Invalid input",
			Message: err.Error(),
		})
		return
	}

	// Kiểm tra user và course tồn tại
	var userExists, courseExists bool
	h.db.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE id = $1)", req.UserID).Scan(&userExists)
	h.db.QueryRow("SELECT EXISTS(SELECT 1 FROM courses WHERE id = $1)", req.CourseID).Scan(&courseExists)

	if !userExists {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "Invalid user",
			Message: "User not found",
		})
		return
	}

	if !courseExists {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "Invalid course",
			Message: "Course not found",
		})
		return
	}

	// Kiểm tra user đã đăng ký khóa học chưa
	var enrolled bool
	h.db.QueryRow("SELECT EXISTS(SELECT 1 FROM enrollments WHERE user_id = $1 AND course_id = $2)", req.UserID, req.CourseID).Scan(&enrolled)
	if !enrolled {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "Not enrolled",
			Message: "User must be enrolled in the course to review it",
		})
		return
	}

	id := uuid.New().String()
	now := time.Now()

	var reviewText sql.NullString
	if req.ReviewText != nil {
		reviewText = sql.NullString{String: *req.ReviewText, Valid: true}
	}

	query := `
		INSERT INTO course_reviews (id, user_id, course_id, rating, review_text, is_approved, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, user_id, course_id, rating, review_text, is_approved, created_at, updated_at`

	var review dto.CourseReviewDTO

	err := h.db.QueryRow(query, id, req.UserID, req.CourseID, req.Rating, reviewText, true, now, now).Scan(
		&review.ID, &review.UserID, &review.CourseID,
		&review.Rating, &reviewText, &review.IsApproved,
		&review.CreatedAt, &review.UpdatedAt,
	)

	if err != nil {
		if err.Error() == `pq: duplicate key value violates unique constraint "course_reviews_user_id_course_id_key"` {
			c.JSON(http.StatusConflict, dto.ErrorResponse{
				Error:   "Conflict",
				Message: "User has already reviewed this course",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "Database error",
			Message: "Failed to create course review",
		})
		return
	}

	if reviewText.Valid {
		review.ReviewText = &reviewText.String
	}

	// Cập nhật rating trung bình của khóa học
	h.updateCourseRating(req.CourseID)

	c.JSON(http.StatusCreated, review)
}

// UpdateCourseReview godoc
// @Summary Cập nhật đánh giá khóa học
// @Description Cập nhật thông tin đánh giá khóa học
// @Tags CourseReview
// @Accept json
// @Produce json
// @Param id path string true "ID đánh giá khóa học"
// @Param body body dto.UpdateCourseReviewRequest true "Thông tin cập nhật"
// @Success 200 {object} dto.CourseReviewDTO
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/course-reviews/{id} [put]
func (h *CourseReviewHandler) UpdateCourseReview(c *gin.Context) {
	id := c.Param("id")

	if _, err := uuid.Parse(id); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "Invalid ID",
			Message: "Invalid course review ID format",
		})
		return
	}

	var req dto.UpdateCourseReviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "Invalid input",
			Message: err.Error(),
		})
		return
	}

	// Lấy thông tin review hiện tại
	var courseID string
	var exists bool
	err := h.db.QueryRow("SELECT course_id FROM course_reviews WHERE id = $1", id).Scan(&courseID)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, dto.ErrorResponse{
				Error:   "Not found",
				Message: "Course review not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "Database error",
			Message: "Failed to fetch course review",
		})
		return
	}
	exists = true

	// Build update query
	setParts := []string{}
	args := []interface{}{}
	argIndex := 1

	if req.Rating != nil {
		setParts = append(setParts, fmt.Sprintf("rating = $%d", argIndex))
		args = append(args, *req.Rating)
		argIndex++
	}

	if req.ReviewText != nil {
		setParts = append(setParts, fmt.Sprintf("review_text = $%d", argIndex))
		args = append(args, *req.ReviewText)
		argIndex++
	}

	if req.IsApproved != nil {
		setParts = append(setParts, fmt.Sprintf("is_approved = $%d", argIndex))
		args = append(args, *req.IsApproved)
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
		UPDATE course_reviews SET %s 
		WHERE id = $%d
		RETURNING id, user_id, course_id, rating, review_text, is_approved, created_at, updated_at`,
		fmt.Sprintf("%s", setParts[0]),
		argIndex,
	)

	for i := 1; i < len(setParts); i++ {
		query = fmt.Sprintf(`
			UPDATE course_reviews SET %s, %s 
			WHERE id = $%d
			RETURNING id, user_id, course_id, rating, review_text, is_approved, created_at, updated_at`,
			setParts[0], setParts[i],
			argIndex,
		)
	}

	args = append(args, id)

	var review dto.CourseReviewDTO
	var reviewText sql.NullString

	err = h.db.QueryRow(query, args...).Scan(
		&review.ID, &review.UserID, &review.CourseID,
		&review.Rating, &reviewText, &review.IsApproved,
		&review.CreatedAt, &review.UpdatedAt,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "Database error",
			Message: "Failed to update course review",
		})
		return
	}

	if reviewText.Valid {
		review.ReviewText = &reviewText.String
	}

	// Cập nhật rating trung bình của khóa học nếu rating được thay đổi
	if req.Rating != nil {
		h.updateCourseRating(courseID)
	}

	c.JSON(http.StatusOK, review)
}

// DeleteCourseReview godoc
// @Summary Xóa đánh giá khóa học
// @Description Xóa đánh giá khóa học theo ID
// @Tags CourseReview
// @Accept json
// @Produce json
// @Param id path string true "ID đánh giá khóa học"
// @Success 204 "No Content"
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/course-reviews/{id} [delete]
func (h *CourseReviewHandler) DeleteCourseReview(c *gin.Context) {
	id := c.Param("id")

	if _, err := uuid.Parse(id); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "Invalid ID",
			Message: "Invalid course review ID format",
		})
		return
	}

	// Lấy course_id trước khi xóa để cập nhật rating
	var courseID string
	err := h.db.QueryRow("SELECT course_id FROM course_reviews WHERE id = $1", id).Scan(&courseID)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, dto.ErrorResponse{
				Error:   "Not found",
				Message: "Course review not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "Database error",
			Message: "Failed to fetch course review",
		})
		return
	}

	result, err := h.db.Exec("DELETE FROM course_reviews WHERE id = $1", id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "Database error",
			Message: "Failed to delete course review",
		})
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, dto.ErrorResponse{
			Error:   "Not found",
			Message: "Course review not found",
		})
		return
	}

	// Cập nhật rating trung bình của khóa học
	h.updateCourseRating(courseID)

	c.Status(http.StatusNoContent)
}

// GetCourseReviewStats godoc
// @Summary Lấy thống kê đánh giá khóa học
// @Description Lấy thống kê đánh giá khóa học theo course ID
// @Tags CourseReview
// @Accept json
// @Produce json
// @Param course_id path string true "ID khóa học"
// @Success 200 {object} dto.CourseReviewStatsDTO
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/courses/{course_id}/review-stats [get]
func (h *CourseReviewHandler) GetCourseReviewStats(c *gin.Context) {
	courseID := c.Param("course_id")

	if _, err := uuid.Parse(courseID); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "Invalid ID",
			Message: "Invalid course ID format",
		})
		return
	}

	// Lấy thống kê cơ bản
	var stats dto.CourseReviewStatsDTO
	var avgRating sql.NullFloat64

	query := `
		SELECT 
			COUNT(*) as total_reviews,
			AVG(rating) as average_rating
		FROM course_reviews 
		WHERE course_id = $1 AND is_approved = true`

	err := h.db.QueryRow(query, courseID).Scan(&stats.TotalReviews, &avgRating)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "Database error",
			Message: "Failed to fetch review stats",
		})
		return
	}

	stats.CourseID = courseID
	if avgRating.Valid {
		stats.AverageRating = avgRating.Float64
	}

	// Lấy phân phối rating
	distributionQuery := `
		SELECT rating, COUNT(*) as count
		FROM course_reviews 
		WHERE course_id = $1 AND is_approved = true
		GROUP BY rating
		ORDER BY rating`

	rows, err := h.db.Query(distributionQuery, courseID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "Database error",
			Message: "Failed to fetch rating distribution",
		})
		return
	}
	defer rows.Close()

	stats.RatingDistribution = make(map[int]int)
	// Initialize với 0 cho tất cả ratings
	for i := 1; i <= 5; i++ {
		stats.RatingDistribution[i] = 0
	}

	for rows.Next() {
		var rating, count int
		if err := rows.Scan(&rating, &count); err == nil {
			stats.RatingDistribution[rating] = count
		}
	}

	c.JSON(http.StatusOK, stats)
}

// Helper function để cập nhật rating trung bình của khóa học
func (h *CourseReviewHandler) updateCourseRating(courseID string) {
	query := `
		UPDATE courses 
		SET 
			rating = COALESCE((
				SELECT AVG(rating) 
				FROM course_reviews 
				WHERE course_id = $1 AND is_approved = true
			), 0),
			total_reviews = COALESCE((
				SELECT COUNT(*) 
				FROM course_reviews 
				WHERE course_id = $1 AND is_approved = true
			), 0),
			updated_at = $2
		WHERE id = $1`

	h.db.Exec(query, courseID, time.Now())
}
