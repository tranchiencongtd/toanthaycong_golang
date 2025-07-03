package handlers

import (
	"database/sql"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/lib/pq"
	"internal/api/dto"
)

type EnrollmentHandler struct {
	db *sql.DB
}

func NewEnrollmentHandler(db *sql.DB) *EnrollmentHandler {
	return &EnrollmentHandler{db: db}
}

// GET /api/enrollments
func (h *EnrollmentHandler) GetEnrollments(c *gin.Context) {
	var query dto.PaginationQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{
			Success: false,
			Message: "Invalid query parameters",
			Error:   err.Error(),
		})
		return
	}

	query.SetDefaults()

	// Filters
	userID := c.Query("user_id")
	courseID := c.Query("course_id")
	isCompleted := c.Query("is_completed")

	var args []interface{}
	baseQuery := `
		SELECT id, user_id, course_id, enrolled_at, completed_at, progress_percentage, 
			   last_accessed_at, certificate_url
		FROM enrollments 
		WHERE 1=1`
	
	countQuery := "SELECT COUNT(*) FROM enrollments WHERE 1=1"

	if userID != "" {
		baseQuery += " AND user_id = $" + strconv.Itoa(len(args)+1)
		countQuery += " AND user_id = $" + strconv.Itoa(len(args)+1)
		args = append(args, userID)
	}

	if courseID != "" {
		baseQuery += " AND course_id = $" + strconv.Itoa(len(args)+1)
		countQuery += " AND course_id = $" + strconv.Itoa(len(args)+1)
		args = append(args, courseID)
	}

	if isCompleted != "" {
		if isCompleted == "true" {
			baseQuery += " AND completed_at IS NOT NULL"
			countQuery += " AND completed_at IS NOT NULL"
		} else if isCompleted == "false" {
			baseQuery += " AND completed_at IS NULL"
			countQuery += " AND completed_at IS NULL"
		}
	}

	// Get total count
	var total int64
	err := h.db.QueryRow(countQuery, args...).Scan(&total)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{
			Success: false,
			Message: "Failed to count enrollments",
			Error:   err.Error(),
		})
		return
	}

	// Add pagination and ordering
	baseQuery += " ORDER BY enrolled_at DESC LIMIT $" + strconv.Itoa(len(args)+1) + " OFFSET $" + strconv.Itoa(len(args)+2)
	args = append(args, query.Limit, query.GetOffset())

	rows, err := h.db.Query(baseQuery, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{
			Success: false,
			Message: "Failed to fetch enrollments",
			Error:   err.Error(),
		})
		return
	}
	defer rows.Close()

	var enrollments []dto.EnrollmentResponse
	for rows.Next() {
		var enrollment dto.EnrollmentResponse
		err := rows.Scan(
			&enrollment.ID,
			&enrollment.UserID,
			&enrollment.CourseID,
			&enrollment.EnrolledAt,
			&enrollment.CompletedAt,
			&enrollment.ProgressPercentage,
			&enrollment.LastAccessedAt,
			&enrollment.CertificateURL,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, dto.APIResponse{
				Success: false,
				Message: "Failed to scan enrollment",
				Error:   err.Error(),
			})
			return
		}
		enrollments = append(enrollments, enrollment)
	}

	pagination := dto.NewPaginationResponse(total, query.Page, query.Limit)

	c.JSON(http.StatusOK, dto.APIResponse{
		Success: true,
		Message: "Enrollments retrieved successfully",
		Data: dto.EnrollmentListResponse{
			Enrollments: enrollments,
			Pagination:  pagination,
		},
	})
}

// GET /api/enrollments/:id
func (h *EnrollmentHandler) GetEnrollment(c *gin.Context) {
	id := c.Param("id")
	
	if _, err := uuid.Parse(id); err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{
			Success: false,
			Message: "Invalid enrollment ID format",
			Error:   err.Error(),
		})
		return
	}

	var enrollment dto.EnrollmentResponse
	err := h.db.QueryRow(`
		SELECT id, user_id, course_id, enrolled_at, completed_at, progress_percentage, 
			   last_accessed_at, certificate_url
		FROM enrollments WHERE id = $1
	`, id).Scan(
		&enrollment.ID,
		&enrollment.UserID,
		&enrollment.CourseID,
		&enrollment.EnrolledAt,
		&enrollment.CompletedAt,
		&enrollment.ProgressPercentage,
		&enrollment.LastAccessedAt,
		&enrollment.CertificateURL,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, dto.APIResponse{
				Success: false,
				Message: "Enrollment not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, dto.APIResponse{
			Success: false,
			Message: "Failed to fetch enrollment",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, dto.APIResponse{
		Success: true,
		Message: "Enrollment retrieved successfully",
		Data:    enrollment,
	})
}

// POST /api/enrollments
func (h *EnrollmentHandler) CreateEnrollment(c *gin.Context) {
	var req dto.CreateEnrollmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{
			Success: false,
			Message: "Invalid request body",
			Error:   err.Error(),
		})
		return
	}

	// Validate UUIDs
	if _, err := uuid.Parse(req.UserID); err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{
			Success: false,
			Message: "Invalid user ID format",
			Error:   err.Error(),
		})
		return
	}

	if _, err := uuid.Parse(req.CourseID); err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{
			Success: false,
			Message: "Invalid course ID format",
			Error:   err.Error(),
		})
		return
	}

	// Check if user exists
	var userExists bool
	err := h.db.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE id = $1)", req.UserID).Scan(&userExists)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{
			Success: false,
			Message: "Failed to verify user",
			Error:   err.Error(),
		})
		return
	}

	if !userExists {
		c.JSON(http.StatusBadRequest, dto.APIResponse{
			Success: false,
			Message: "User not found",
		})
		return
	}

	// Check if course exists and is published
	var courseStatus string
	err = h.db.QueryRow("SELECT status FROM courses WHERE id = $1", req.CourseID).Scan(&courseStatus)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusBadRequest, dto.APIResponse{
				Success: false,
				Message: "Course not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, dto.APIResponse{
			Success: false,
			Message: "Failed to verify course",
			Error:   err.Error(),
		})
		return
	}

	if courseStatus != "published" {
		c.JSON(http.StatusBadRequest, dto.APIResponse{
			Success: false,
			Message: "Course is not available for enrollment",
		})
		return
	}

	// Check if already enrolled
	var existingID string
	err = h.db.QueryRow("SELECT id FROM enrollments WHERE user_id = $1 AND course_id = $2", req.UserID, req.CourseID).Scan(&existingID)
	if err != sql.ErrNoRows {
		c.JSON(http.StatusConflict, dto.APIResponse{
			Success: false,
			Message: "User already enrolled in this course",
		})
		return
	}

	id := uuid.New().String()

	_, err = h.db.Exec(`
		INSERT INTO enrollments (id, user_id, course_id, enrolled_at)
		VALUES ($1, $2, $3, CURRENT_TIMESTAMP)
	`, id, req.UserID, req.CourseID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{
			Success: false,
			Message: "Failed to create enrollment",
			Error:   err.Error(),
		})
		return
	}

	// Fetch the created enrollment
	var enrollment dto.EnrollmentResponse
	err = h.db.QueryRow(`
		SELECT id, user_id, course_id, enrolled_at, completed_at, progress_percentage, 
			   last_accessed_at, certificate_url
		FROM enrollments WHERE id = $1
	`, id).Scan(
		&enrollment.ID,
		&enrollment.UserID,
		&enrollment.CourseID,
		&enrollment.EnrolledAt,
		&enrollment.CompletedAt,
		&enrollment.ProgressPercentage,
		&enrollment.LastAccessedAt,
		&enrollment.CertificateURL,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{
			Success: false,
			Message: "Failed to fetch created enrollment",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, dto.APIResponse{
		Success: true,
		Message: "Enrollment created successfully",
		Data:    enrollment,
	})
}

// PUT /api/enrollments/:id
func (h *EnrollmentHandler) UpdateEnrollment(c *gin.Context) {
	id := c.Param("id")
	
	if _, err := uuid.Parse(id); err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{
			Success: false,
			Message: "Invalid enrollment ID format",
			Error:   err.Error(),
		})
		return
	}

	var req dto.UpdateEnrollmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{
			Success: false,
			Message: "Invalid request body",
			Error:   err.Error(),
		})
		return
	}

	// Check if enrollment exists
	var exists bool
	err := h.db.QueryRow("SELECT EXISTS(SELECT 1 FROM enrollments WHERE id = $1)", id).Scan(&exists)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{
			Success: false,
			Message: "Failed to check enrollment existence",
			Error:   err.Error(),
		})
		return
	}

	if !exists {
		c.JSON(http.StatusNotFound, dto.APIResponse{
			Success: false,
			Message: "Enrollment not found",
		})
		return
	}

	// Build dynamic update query
	setParts := []string{}
	args := []interface{}{}
	argIndex := 1

	if req.ProgressPercentage != nil {
		setParts = append(setParts, "progress_percentage = $"+strconv.Itoa(argIndex))
		args = append(args, *req.ProgressPercentage)
		argIndex++
	}

	if req.CompletedAt != nil {
		setParts = append(setParts, "completed_at = $"+strconv.Itoa(argIndex))
		args = append(args, *req.CompletedAt)
		argIndex++
	}

	if req.CertificateURL != nil {
		setParts = append(setParts, "certificate_url = $"+strconv.Itoa(argIndex))
		args = append(args, *req.CertificateURL)
		argIndex++
	}

	// Always update last_accessed_at
	setParts = append(setParts, "last_accessed_at = CURRENT_TIMESTAMP")

	if len(setParts) == 1 && setParts[0] == "last_accessed_at = CURRENT_TIMESTAMP" {
		// Only updating access time, which is fine
	}

	// Add WHERE clause
	args = append(args, id)
	whereClause := "$" + strconv.Itoa(argIndex)

	query := "UPDATE enrollments SET " + setParts[0]
	for i := 1; i < len(setParts); i++ {
		query += ", " + setParts[i]
	}
	query += " WHERE id = " + whereClause

	_, err = h.db.Exec(query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{
			Success: false,
			Message: "Failed to update enrollment",
			Error:   err.Error(),
		})
		return
	}

	// Fetch updated enrollment
	var enrollment dto.EnrollmentResponse
	err = h.db.QueryRow(`
		SELECT id, user_id, course_id, enrolled_at, completed_at, progress_percentage, 
			   last_accessed_at, certificate_url
		FROM enrollments WHERE id = $1
	`, id).Scan(
		&enrollment.ID,
		&enrollment.UserID,
		&enrollment.CourseID,
		&enrollment.EnrolledAt,
		&enrollment.CompletedAt,
		&enrollment.ProgressPercentage,
		&enrollment.LastAccessedAt,
		&enrollment.CertificateURL,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{
			Success: false,
			Message: "Failed to fetch updated enrollment",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, dto.APIResponse{
		Success: true,
		Message: "Enrollment updated successfully",
		Data:    enrollment,
	})
}

// DELETE /api/enrollments/:id (Unenroll)
func (h *EnrollmentHandler) DeleteEnrollment(c *gin.Context) {
	id := c.Param("id")
	
	if _, err := uuid.Parse(id); err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{
			Success: false,
			Message: "Invalid enrollment ID format",
			Error:   err.Error(),
		})
		return
	}

	result, err := h.db.Exec("DELETE FROM enrollments WHERE id = $1", id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{
			Success: false,
			Message: "Failed to delete enrollment",
			Error:   err.Error(),
		})
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{
			Success: false,
			Message: "Failed to check delete result",
			Error:   err.Error(),
		})
		return
	}

	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, dto.APIResponse{
			Success: false,
			Message: "Enrollment not found",
		})
		return
	}

	c.JSON(http.StatusOK, dto.APIResponse{
		Success: true,
		Message: "Enrollment deleted successfully",
	})
}

// PUT /api/enrollments/:id/access
func (h *EnrollmentHandler) UpdateLastAccess(c *gin.Context) {
	id := c.Param("id")
	
	if _, err := uuid.Parse(id); err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{
			Success: false,
			Message: "Invalid enrollment ID format",
			Error:   err.Error(),
		})
		return
	}

	result, err := h.db.Exec(`
		UPDATE enrollments 
		SET last_accessed_at = CURRENT_TIMESTAMP 
		WHERE id = $1
	`, id)

	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{
			Success: false,
			Message: "Failed to update last access",
			Error:   err.Error(),
		})
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{
			Success: false,
			Message: "Failed to check update result",
			Error:   err.Error(),
		})
		return
	}

	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, dto.APIResponse{
			Success: false,
			Message: "Enrollment not found",
		})
		return
	}

	c.JSON(http.StatusOK, dto.APIResponse{
		Success: true,
		Message: "Last access updated successfully",
	})
}
