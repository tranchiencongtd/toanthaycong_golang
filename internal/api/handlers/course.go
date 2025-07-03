package handlers

import (
	"database/sql"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/lib/pq"
	"internal/api/dto"
)

type CourseHandler struct {
	db *sql.DB
}

func NewCourseHandler(db *sql.DB) *CourseHandler {
	return &CourseHandler{db: db}
}

// GET /api/courses
func (h *CourseHandler) GetCourses(c *gin.Context) {
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
	categoryID := c.Query("category_id")
	instructorID := c.Query("instructor_id")
	level := c.Query("level")
	status := c.Query("status")

	var args []interface{}
	baseQuery := `
		SELECT id, title, slug, description, short_description, thumbnail_url, preview_video_url,
			   instructor_id, category_id, price, discount_price, language, level, duration_hours,
			   total_lectures, status, requirements, what_you_learn, target_audience,
			   rating, total_students, total_reviews, published_at, created_at, updated_at
		FROM courses 
		WHERE 1=1`
	
	countQuery := "SELECT COUNT(*) FROM courses WHERE 1=1"

	if categoryID != "" {
		baseQuery += " AND category_id = $" + strconv.Itoa(len(args)+1)
		countQuery += " AND category_id = $" + strconv.Itoa(len(args)+1)
		args = append(args, categoryID)
	}

	if instructorID != "" {
		baseQuery += " AND instructor_id = $" + strconv.Itoa(len(args)+1)
		countQuery += " AND instructor_id = $" + strconv.Itoa(len(args)+1)
		args = append(args, instructorID)
	}

	if level != "" {
		baseQuery += " AND level = $" + strconv.Itoa(len(args)+1)
		countQuery += " AND level = $" + strconv.Itoa(len(args)+1)
		args = append(args, level)
	}

	if status != "" {
		baseQuery += " AND status = $" + strconv.Itoa(len(args)+1)
		countQuery += " AND status = $" + strconv.Itoa(len(args)+1)
		args = append(args, status)
	}

	// Get total count
	var total int64
	err := h.db.QueryRow(countQuery, args...).Scan(&total)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{
			Success: false,
			Message: "Failed to count courses",
			Error:   err.Error(),
		})
		return
	}

	// Add pagination
	baseQuery += " ORDER BY created_at DESC LIMIT $" + strconv.Itoa(len(args)+1) + " OFFSET $" + strconv.Itoa(len(args)+2)
	args = append(args, query.Limit, query.GetOffset())

	rows, err := h.db.Query(baseQuery, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{
			Success: false,
			Message: "Failed to fetch courses",
			Error:   err.Error(),
		})
		return
	}
	defer rows.Close()

	var courses []dto.CourseResponse
	for rows.Next() {
		var course dto.CourseResponse
		err := rows.Scan(
			&course.ID,
			&course.Title,
			&course.Slug,
			&course.Description,
			&course.ShortDescription,
			&course.ThumbnailURL,
			&course.PreviewVideoURL,
			&course.InstructorID,
			&course.CategoryID,
			&course.Price,
			&course.DiscountPrice,
			&course.Language,
			&course.Level,
			&course.DurationHours,
			&course.TotalLectures,
			&course.Status,
			pq.Array(&course.Requirements),
			pq.Array(&course.WhatYouLearn),
			pq.Array(&course.TargetAudience),
			&course.Rating,
			&course.TotalStudents,
			&course.TotalReviews,
			&course.PublishedAt,
			&course.CreatedAt,
			&course.UpdatedAt,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, dto.APIResponse{
				Success: false,
				Message: "Failed to scan course",
				Error:   err.Error(),
			})
			return
		}
		courses = append(courses, course)
	}

	pagination := dto.NewPaginationResponse(total, query.Page, query.Limit)

	c.JSON(http.StatusOK, dto.APIResponse{
		Success: true,
		Message: "Courses retrieved successfully",
		Data: dto.CourseListResponse{
			Courses:    courses,
			Pagination: pagination,
		},
	})
}

// GET /api/courses/:id
func (h *CourseHandler) GetCourse(c *gin.Context) {
	id := c.Param("id")
	
	if _, err := uuid.Parse(id); err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{
			Success: false,
			Message: "Invalid course ID format",
			Error:   err.Error(),
		})
		return
	}

	var course dto.CourseResponse
	err := h.db.QueryRow(`
		SELECT id, title, slug, description, short_description, thumbnail_url, preview_video_url,
			   instructor_id, category_id, price, discount_price, language, level, duration_hours,
			   total_lectures, status, requirements, what_you_learn, target_audience,
			   rating, total_students, total_reviews, published_at, created_at, updated_at
		FROM courses WHERE id = $1
	`, id).Scan(
		&course.ID,
		&course.Title,
		&course.Slug,
		&course.Description,
		&course.ShortDescription,
		&course.ThumbnailURL,
		&course.PreviewVideoURL,
		&course.InstructorID,
		&course.CategoryID,
		&course.Price,
		&course.DiscountPrice,
		&course.Language,
		&course.Level,
		&course.DurationHours,
		&course.TotalLectures,
		&course.Status,
		pq.Array(&course.Requirements),
		pq.Array(&course.WhatYouLearn),
		pq.Array(&course.TargetAudience),
		&course.Rating,
		&course.TotalStudents,
		&course.TotalReviews,
		&course.PublishedAt,
		&course.CreatedAt,
		&course.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, dto.APIResponse{
				Success: false,
				Message: "Course not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, dto.APIResponse{
			Success: false,
			Message: "Failed to fetch course",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, dto.APIResponse{
		Success: true,
		Message: "Course retrieved successfully",
		Data:    course,
	})
}

// POST /api/courses
func (h *CourseHandler) CreateCourse(c *gin.Context) {
	var req dto.CreateCourseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{
			Success: false,
			Message: "Invalid request body",
			Error:   err.Error(),
		})
		return
	}

	// Validate instructor_id and category_id
	if _, err := uuid.Parse(req.InstructorID); err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{
			Success: false,
			Message: "Invalid instructor ID format",
			Error:   err.Error(),
		})
		return
	}

	if _, err := uuid.Parse(req.CategoryID); err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{
			Success: false,
			Message: "Invalid category ID format",
			Error:   err.Error(),
		})
		return
	}

	// Verify instructor exists and is an instructor
	var instructorRole string
	err := h.db.QueryRow("SELECT role FROM users WHERE id = $1", req.InstructorID).Scan(&instructorRole)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusBadRequest, dto.APIResponse{
				Success: false,
				Message: "Instructor not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, dto.APIResponse{
			Success: false,
			Message: "Failed to verify instructor",
			Error:   err.Error(),
		})
		return
	}

	if instructorRole != "instructor" && instructorRole != "admin" {
		c.JSON(http.StatusBadRequest, dto.APIResponse{
			Success: false,
			Message: "User is not an instructor",
		})
		return
	}

	// Verify category exists
	var categoryExists bool
	err = h.db.QueryRow("SELECT EXISTS(SELECT 1 FROM categories WHERE id = $1)", req.CategoryID).Scan(&categoryExists)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{
			Success: false,
			Message: "Failed to verify category",
			Error:   err.Error(),
		})
		return
	}

	if !categoryExists {
		c.JSON(http.StatusBadRequest, dto.APIResponse{
			Success: false,
			Message: "Category not found",
		})
		return
	}

	id := uuid.New().String()

	_, err = h.db.Exec(`
		INSERT INTO courses (
			id, title, slug, description, short_description, thumbnail_url, preview_video_url,
			instructor_id, category_id, price, discount_price, language, level,
			requirements, what_you_learn, target_audience, created_at, updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	`, id, req.Title, req.Slug, req.Description, req.ShortDescription, req.ThumbnailURL, req.PreviewVideoURL,
		req.InstructorID, req.CategoryID, req.Price, req.DiscountPrice, req.Language, req.Level,
		pq.Array(req.Requirements), pq.Array(req.WhatYouLearn), pq.Array(req.TargetAudience))

	if err != nil {
		// Check for unique constraint violations
		if pqErr, ok := err.(*pq.Error); ok {
			if pqErr.Constraint == "courses_slug_key" {
				c.JSON(http.StatusConflict, dto.APIResponse{
					Success: false,
					Message: "Course slug already exists",
				})
				return
			}
		}
		
		c.JSON(http.StatusInternalServerError, dto.APIResponse{
			Success: false,
			Message: "Failed to create course",
			Error:   err.Error(),
		})
		return
	}

	// Fetch the created course
	var course dto.CourseResponse
	err = h.db.QueryRow(`
		SELECT id, title, slug, description, short_description, thumbnail_url, preview_video_url,
			   instructor_id, category_id, price, discount_price, language, level, duration_hours,
			   total_lectures, status, requirements, what_you_learn, target_audience,
			   rating, total_students, total_reviews, published_at, created_at, updated_at
		FROM courses WHERE id = $1
	`, id).Scan(
		&course.ID,
		&course.Title,
		&course.Slug,
		&course.Description,
		&course.ShortDescription,
		&course.ThumbnailURL,
		&course.PreviewVideoURL,
		&course.InstructorID,
		&course.CategoryID,
		&course.Price,
		&course.DiscountPrice,
		&course.Language,
		&course.Level,
		&course.DurationHours,
		&course.TotalLectures,
		&course.Status,
		pq.Array(&course.Requirements),
		pq.Array(&course.WhatYouLearn),
		pq.Array(&course.TargetAudience),
		&course.Rating,
		&course.TotalStudents,
		&course.TotalReviews,
		&course.PublishedAt,
		&course.CreatedAt,
		&course.UpdatedAt,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{
			Success: false,
			Message: "Failed to fetch created course",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, dto.APIResponse{
		Success: true,
		Message: "Course created successfully",
		Data:    course,
	})
}

// PUT /api/courses/:id
func (h *CourseHandler) UpdateCourse(c *gin.Context) {
	id := c.Param("id")
	
	if _, err := uuid.Parse(id); err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{
			Success: false,
			Message: "Invalid course ID format",
			Error:   err.Error(),
		})
		return
	}

	var req dto.UpdateCourseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{
			Success: false,
			Message: "Invalid request body",
			Error:   err.Error(),
		})
		return
	}

	// Check if course exists
	var exists bool
	err := h.db.QueryRow("SELECT EXISTS(SELECT 1 FROM courses WHERE id = $1)", id).Scan(&exists)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{
			Success: false,
			Message: "Failed to check course existence",
			Error:   err.Error(),
		})
		return
	}

	if !exists {
		c.JSON(http.StatusNotFound, dto.APIResponse{
			Success: false,
			Message: "Course not found",
		})
		return
	}

	// Build dynamic update query
	setParts := []string{"updated_at = CURRENT_TIMESTAMP"}
	args := []interface{}{}
	argIndex := 1

	if req.Title != nil {
		setParts = append(setParts, "title = $"+strconv.Itoa(argIndex))
		args = append(args, *req.Title)
		argIndex++
	}

	if req.Slug != nil {
		setParts = append(setParts, "slug = $"+strconv.Itoa(argIndex))
		args = append(args, *req.Slug)
		argIndex++
	}

	if req.Description != nil {
		setParts = append(setParts, "description = $"+strconv.Itoa(argIndex))
		args = append(args, *req.Description)
		argIndex++
	}

	if req.ShortDescription != nil {
		setParts = append(setParts, "short_description = $"+strconv.Itoa(argIndex))
		args = append(args, *req.ShortDescription)
		argIndex++
	}

	if req.ThumbnailURL != nil {
		setParts = append(setParts, "thumbnail_url = $"+strconv.Itoa(argIndex))
		args = append(args, *req.ThumbnailURL)
		argIndex++
	}

	if req.PreviewVideoURL != nil {
		setParts = append(setParts, "preview_video_url = $"+strconv.Itoa(argIndex))
		args = append(args, *req.PreviewVideoURL)
		argIndex++
	}

	if req.CategoryID != nil {
		setParts = append(setParts, "category_id = $"+strconv.Itoa(argIndex))
		args = append(args, *req.CategoryID)
		argIndex++
	}

	if req.Price != nil {
		setParts = append(setParts, "price = $"+strconv.Itoa(argIndex))
		args = append(args, *req.Price)
		argIndex++
	}

	if req.DiscountPrice != nil {
		setParts = append(setParts, "discount_price = $"+strconv.Itoa(argIndex))
		args = append(args, *req.DiscountPrice)
		argIndex++
	}

	if req.Language != nil {
		setParts = append(setParts, "language = $"+strconv.Itoa(argIndex))
		args = append(args, *req.Language)
		argIndex++
	}

	if req.Level != nil {
		setParts = append(setParts, "level = $"+strconv.Itoa(argIndex))
		args = append(args, *req.Level)
		argIndex++
	}

	if req.Status != nil {
		setParts = append(setParts, "status = $"+strconv.Itoa(argIndex))
		args = append(args, *req.Status)
		argIndex++
		
		// If publishing, set published_at
		if *req.Status == "published" {
			setParts = append(setParts, "published_at = CURRENT_TIMESTAMP")
		}
	}

	if req.Requirements != nil {
		setParts = append(setParts, "requirements = $"+strconv.Itoa(argIndex))
		args = append(args, pq.Array(req.Requirements))
		argIndex++
	}

	if req.WhatYouLearn != nil {
		setParts = append(setParts, "what_you_learn = $"+strconv.Itoa(argIndex))
		args = append(args, pq.Array(req.WhatYouLearn))
		argIndex++
	}

	if req.TargetAudience != nil {
		setParts = append(setParts, "target_audience = $"+strconv.Itoa(argIndex))
		args = append(args, pq.Array(req.TargetAudience))
		argIndex++
	}

	if len(setParts) == 1 { // Only updated_at
		c.JSON(http.StatusBadRequest, dto.APIResponse{
			Success: false,
			Message: "No fields to update",
		})
		return
	}

	// Add WHERE clause
	args = append(args, id)
	whereClause := "$" + strconv.Itoa(argIndex)

	query := "UPDATE courses SET " + setParts[0]
	for i := 1; i < len(setParts); i++ {
		query += ", " + setParts[i]
	}
	query += " WHERE id = " + whereClause

	_, err = h.db.Exec(query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{
			Success: false,
			Message: "Failed to update course",
			Error:   err.Error(),
		})
		return
	}

	// Fetch updated course
	var course dto.CourseResponse
	err = h.db.QueryRow(`
		SELECT id, title, slug, description, short_description, thumbnail_url, preview_video_url,
			   instructor_id, category_id, price, discount_price, language, level, duration_hours,
			   total_lectures, status, requirements, what_you_learn, target_audience,
			   rating, total_students, total_reviews, published_at, created_at, updated_at
		FROM courses WHERE id = $1
	`, id).Scan(
		&course.ID,
		&course.Title,
		&course.Slug,
		&course.Description,
		&course.ShortDescription,
		&course.ThumbnailURL,
		&course.PreviewVideoURL,
		&course.InstructorID,
		&course.CategoryID,
		&course.Price,
		&course.DiscountPrice,
		&course.Language,
		&course.Level,
		&course.DurationHours,
		&course.TotalLectures,
		&course.Status,
		pq.Array(&course.Requirements),
		pq.Array(&course.WhatYouLearn),
		pq.Array(&course.TargetAudience),
		&course.Rating,
		&course.TotalStudents,
		&course.TotalReviews,
		&course.PublishedAt,
		&course.CreatedAt,
		&course.UpdatedAt,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{
			Success: false,
			Message: "Failed to fetch updated course",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, dto.APIResponse{
		Success: true,
		Message: "Course updated successfully",
		Data:    course,
	})
}

// DELETE /api/courses/:id
func (h *CourseHandler) DeleteCourse(c *gin.Context) {
	id := c.Param("id")
	
	if _, err := uuid.Parse(id); err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{
			Success: false,
			Message: "Invalid course ID format",
			Error:   err.Error(),
		})
		return
	}

	// Check if course has enrollments
	var enrollmentCount int
	err := h.db.QueryRow("SELECT COUNT(*) FROM enrollments WHERE course_id = $1", id).Scan(&enrollmentCount)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{
			Success: false,
			Message: "Failed to check course enrollments",
			Error:   err.Error(),
		})
		return
	}

	if enrollmentCount > 0 {
		c.JSON(http.StatusBadRequest, dto.APIResponse{
			Success: false,
			Message: "Cannot delete course that has enrollments",
		})
		return
	}

	result, err := h.db.Exec("DELETE FROM courses WHERE id = $1", id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{
			Success: false,
			Message: "Failed to delete course",
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
			Message: "Course not found",
		})
		return
	}

	c.JSON(http.StatusOK, dto.APIResponse{
		Success: true,
		Message: "Course deleted successfully",
	})
}
