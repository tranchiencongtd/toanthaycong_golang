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

type WishlistHandler struct {
	db *sql.DB
}

func NewWishlistHandler(db *sql.DB) *WishlistHandler {
	return &WishlistHandler{db: db}
}

// GetWishlists godoc
// @Summary Lấy danh sách wishlist
// @Description Lấy danh sách wishlist với phân trang và lọc
// @Tags Wishlist
// @Accept json
// @Produce json
// @Param page query int false "Số trang" default(1)
// @Param limit query int false "Số item mỗi trang" default(10)
// @Param user_id query string false "Lọc theo user ID"
// @Success 200 {object} dto.WishlistListResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/wishlists [get]
func (h *WishlistHandler) GetWishlists(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	userID := c.Query("user_id")

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	offset := (page - 1) * limit

	// Build query với filters
	query := `
		SELECT w.id, w.user_id, w.course_id, w.created_at,
		       COUNT(*) OVER() as total_count
		FROM wishlists w
		WHERE 1=1`
	
	args := []interface{}{}
	argIndex := 1

	if userID != "" {
		query += fmt.Sprintf(" AND w.user_id = $%d", argIndex)
		args = append(args, userID)
		argIndex++
	}

	query += fmt.Sprintf(" ORDER BY w.created_at DESC LIMIT $%d OFFSET $%d", argIndex, argIndex+1)
	args = append(args, limit, offset)

	rows, err := h.db.Query(query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "Database error",
			Message: "Failed to fetch wishlists",
		})
		return
	}
	defer rows.Close()

	var wishlists []dto.WishlistDTO
	var totalCount int

	for rows.Next() {
		var wishlist dto.WishlistDTO

		err := rows.Scan(
			&wishlist.ID, &wishlist.UserID, &wishlist.CourseID,
			&wishlist.CreatedAt, &totalCount,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
				Error:   "Scan error",
				Message: "Failed to parse wishlist data",
			})
			return
		}

		wishlists = append(wishlists, wishlist)
	}

	// Tính toán pagination
	totalPages := (totalCount + limit - 1) / limit

	response := dto.WishlistListResponse{
		Data: wishlists,
		Pagination: dto.PaginationMeta{
			Page:       page,
			Limit:      limit,
			TotalItems: totalCount,
			TotalPages: totalPages,
		},
	}

	c.JSON(http.StatusOK, response)
}

// GetWishlist godoc
// @Summary Lấy thông tin wishlist theo ID
// @Description Lấy thông tin chi tiết wishlist theo ID
// @Tags Wishlist
// @Accept json
// @Produce json
// @Param id path string true "ID wishlist"
// @Success 200 {object} dto.WishlistDTO
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/wishlists/{id} [get]
func (h *WishlistHandler) GetWishlist(c *gin.Context) {
	id := c.Param("id")

	if _, err := uuid.Parse(id); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "Invalid ID",
			Message: "Invalid wishlist ID format",
		})
		return
	}

	query := `
		SELECT id, user_id, course_id, created_at
		FROM wishlists 
		WHERE id = $1`

	var wishlist dto.WishlistDTO

	err := h.db.QueryRow(query, id).Scan(
		&wishlist.ID, &wishlist.UserID, &wishlist.CourseID,
		&wishlist.CreatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, dto.ErrorResponse{
				Error:   "Not found",
				Message: "Wishlist not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "Database error",
			Message: "Failed to fetch wishlist",
		})
		return
	}

	c.JSON(http.StatusOK, wishlist)
}

// CreateWishlist godoc
// @Summary Thêm khóa học vào wishlist
// @Description Thêm khóa học vào wishlist
// @Tags Wishlist
// @Accept json
// @Produce json
// @Param body body dto.CreateWishlistRequest true "Thông tin wishlist"
// @Success 201 {object} dto.WishlistDTO
// @Failure 400 {object} dto.ErrorResponse
// @Failure 409 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/wishlists [post]
func (h *WishlistHandler) CreateWishlist(c *gin.Context) {
	var req dto.CreateWishlistRequest
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
	if enrolled {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "Already enrolled",
			Message: "User is already enrolled in this course",
		})
		return
	}

	id := uuid.New().String()
	now := time.Now()

	query := `
		INSERT INTO wishlists (id, user_id, course_id, created_at)
		VALUES ($1, $2, $3, $4)
		RETURNING id, user_id, course_id, created_at`

	var wishlist dto.WishlistDTO

	err := h.db.QueryRow(query, id, req.UserID, req.CourseID, now).Scan(
		&wishlist.ID, &wishlist.UserID, &wishlist.CourseID,
		&wishlist.CreatedAt,
	)

	if err != nil {
		if err.Error() == `pq: duplicate key value violates unique constraint "wishlists_user_id_course_id_key"` {
			c.JSON(http.StatusConflict, dto.ErrorResponse{
				Error:   "Conflict",
				Message: "Course is already in user's wishlist",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "Database error",
			Message: "Failed to add course to wishlist",
		})
		return
	}

	c.JSON(http.StatusCreated, wishlist)
}

// DeleteWishlist godoc
// @Summary Xóa khóa học khỏi wishlist
// @Description Xóa khóa học khỏi wishlist theo ID
// @Tags Wishlist
// @Accept json
// @Produce json
// @Param id path string true "ID wishlist"
// @Success 204 "No Content"
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/wishlists/{id} [delete]
func (h *WishlistHandler) DeleteWishlist(c *gin.Context) {
	id := c.Param("id")

	if _, err := uuid.Parse(id); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "Invalid ID",
			Message: "Invalid wishlist ID format",
		})
		return
	}

	result, err := h.db.Exec("DELETE FROM wishlists WHERE id = $1", id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "Database error",
			Message: "Failed to remove from wishlist",
		})
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, dto.ErrorResponse{
			Error:   "Not found",
			Message: "Wishlist item not found",
		})
		return
	}

	c.Status(http.StatusNoContent)
}

// RemoveFromWishlistByUserAndCourse godoc
// @Summary Xóa khóa học khỏi wishlist theo user và course
// @Description Xóa khóa học khỏi wishlist theo user ID và course ID
// @Tags Wishlist
// @Accept json
// @Produce json
// @Param user_id query string true "ID người dùng"
// @Param course_id query string true "ID khóa học"
// @Success 204 "No Content"
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/wishlists/remove [delete]
func (h *WishlistHandler) RemoveFromWishlistByUserAndCourse(c *gin.Context) {
	userID := c.Query("user_id")
	courseID := c.Query("course_id")

	if userID == "" || courseID == "" {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "Missing parameters",
			Message: "Both user_id and course_id are required",
		})
		return
	}

	if _, err := uuid.Parse(userID); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "Invalid user ID",
			Message: "Invalid user ID format",
		})
		return
	}

	if _, err := uuid.Parse(courseID); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "Invalid course ID",
			Message: "Invalid course ID format",
		})
		return
	}

	result, err := h.db.Exec("DELETE FROM wishlists WHERE user_id = $1 AND course_id = $2", userID, courseID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "Database error",
			Message: "Failed to remove from wishlist",
		})
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, dto.ErrorResponse{
			Error:   "Not found",
			Message: "Wishlist item not found",
		})
		return
	}

	c.Status(http.StatusNoContent)
}

// CheckWishlist godoc
// @Summary Kiểm tra khóa học có trong wishlist không
// @Description Kiểm tra khóa học có trong wishlist của user không
// @Tags Wishlist
// @Accept json
// @Produce json
// @Param user_id query string true "ID người dùng"
// @Param course_id query string true "ID khóa học"
// @Success 200 {object} map[string]bool "{"in_wishlist": true/false}"
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/wishlists/check [get]
func (h *WishlistHandler) CheckWishlist(c *gin.Context) {
	userID := c.Query("user_id")
	courseID := c.Query("course_id")

	if userID == "" || courseID == "" {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "Missing parameters",
			Message: "Both user_id and course_id are required",
		})
		return
	}

	if _, err := uuid.Parse(userID); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "Invalid user ID",
			Message: "Invalid user ID format",
		})
		return
	}

	if _, err := uuid.Parse(courseID); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "Invalid course ID",
			Message: "Invalid course ID format",
		})
		return
	}

	var exists bool
	err := h.db.QueryRow("SELECT EXISTS(SELECT 1 FROM wishlists WHERE user_id = $1 AND course_id = $2)", userID, courseID).Scan(&exists)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "Database error",
			Message: "Failed to check wishlist",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"in_wishlist": exists,
	})
}
