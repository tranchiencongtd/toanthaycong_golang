package handlers

import (
	"database/sql"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"internal/api/dto"
)

type CourseSectionHandler struct {
	db *sql.DB
}

func NewCourseSectionHandler(db *sql.DB) *CourseSectionHandler {
	return &CourseSectionHandler{db: db}
}

// GET /api/course-sections
func (h *CourseSectionHandler) GetCourseSections(c *gin.Context) {
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

	// Filter by course_id
	courseID := c.Query("course_id")
	var args []interface{}
	baseQuery := `
		SELECT id, course_id, title, description, sort_order, created_at, updated_at
		FROM course_sections 
		WHERE 1=1`
	
	countQuery := "SELECT COUNT(*) FROM course_sections WHERE 1=1"

	if courseID != "" {
		if _, err := uuid.Parse(courseID); err != nil {
			c.JSON(http.StatusBadRequest, dto.APIResponse{
				Success: false,
				Message: "Invalid course ID format",
				Error:   err.Error(),
			})
			return
		}
		baseQuery += " AND course_id = $1"
		countQuery += " AND course_id = $1"
		args = append(args, courseID)
	}

	// Get total count
	var total int64
	err := h.db.QueryRow(countQuery, args...).Scan(&total)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{
			Success: false,
			Message: "Failed to count course sections",
			Error:   err.Error(),
		})
		return
	}

	// Add pagination and ordering
	baseQuery += " ORDER BY sort_order ASC, created_at ASC LIMIT $" + strconv.Itoa(len(args)+1) + " OFFSET $" + strconv.Itoa(len(args)+2)
	args = append(args, query.Limit, query.GetOffset())

	rows, err := h.db.Query(baseQuery, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{
			Success: false,
			Message: "Failed to fetch course sections",
			Error:   err.Error(),
		})
		return
	}
	defer rows.Close()

	var sections []dto.CourseSectionResponse
	for rows.Next() {
		var section dto.CourseSectionResponse
		err := rows.Scan(
			&section.ID,
			&section.CourseID,
			&section.Title,
			&section.Description,
			&section.SortOrder,
			&section.CreatedAt,
			&section.UpdatedAt,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, dto.APIResponse{
				Success: false,
				Message: "Failed to scan course section",
				Error:   err.Error(),
			})
			return
		}

		// Get lectures for this section if requested
		includeLectures := c.Query("include_lectures")
		if includeLectures == "true" {
			lectures, err := h.getLecturesForSection(section.ID)
			if err == nil {
				section.Lectures = lectures
			}
		}

		sections = append(sections, section)
	}

	pagination := dto.NewPaginationResponse(total, query.Page, query.Limit)

	c.JSON(http.StatusOK, dto.APIResponse{
		Success: true,
		Message: "Course sections retrieved successfully",
		Data: dto.CourseSectionListResponse{
			Sections:   sections,
			Pagination: pagination,
		},
	})
}

// GET /api/course-sections/:id
func (h *CourseSectionHandler) GetCourseSection(c *gin.Context) {
	id := c.Param("id")
	
	if _, err := uuid.Parse(id); err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{
			Success: false,
			Message: "Invalid course section ID format",
			Error:   err.Error(),
		})
		return
	}

	var section dto.CourseSectionResponse
	err := h.db.QueryRow(`
		SELECT id, course_id, title, description, sort_order, created_at, updated_at
		FROM course_sections WHERE id = $1
	`, id).Scan(
		&section.ID,
		&section.CourseID,
		&section.Title,
		&section.Description,
		&section.SortOrder,
		&section.CreatedAt,
		&section.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, dto.APIResponse{
				Success: false,
				Message: "Course section not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, dto.APIResponse{
			Success: false,
			Message: "Failed to fetch course section",
			Error:   err.Error(),
		})
		return
	}

	// Get lectures for this section
	lectures, err := h.getLecturesForSection(section.ID)
	if err == nil {
		section.Lectures = lectures
	}

	c.JSON(http.StatusOK, dto.APIResponse{
		Success: true,
		Message: "Course section retrieved successfully",
		Data:    section,
	})
}

// POST /api/course-sections
func (h *CourseSectionHandler) CreateCourseSection(c *gin.Context) {
	var req dto.CreateCourseSectionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{
			Success: false,
			Message: "Invalid request body",
			Error:   err.Error(),
		})
		return
	}

	// Validate course_id
	if _, err := uuid.Parse(req.CourseID); err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{
			Success: false,
			Message: "Invalid course ID format",
			Error:   err.Error(),
		})
		return
	}

	// Verify course exists
	var courseExists bool
	err := h.db.QueryRow("SELECT EXISTS(SELECT 1 FROM courses WHERE id = $1)", req.CourseID).Scan(&courseExists)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{
			Success: false,
			Message: "Failed to verify course",
			Error:   err.Error(),
		})
		return
	}

	if !courseExists {
		c.JSON(http.StatusBadRequest, dto.APIResponse{
			Success: false,
			Message: "Course not found",
		})
		return
	}

	id := uuid.New().String()

	_, err = h.db.Exec(`
		INSERT INTO course_sections (id, course_id, title, description, sort_order, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	`, id, req.CourseID, req.Title, req.Description, req.SortOrder)

	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{
			Success: false,
			Message: "Failed to create course section",
			Error:   err.Error(),
		})
		return
	}

	// Fetch the created section
	var section dto.CourseSectionResponse
	err = h.db.QueryRow(`
		SELECT id, course_id, title, description, sort_order, created_at, updated_at
		FROM course_sections WHERE id = $1
	`, id).Scan(
		&section.ID,
		&section.CourseID,
		&section.Title,
		&section.Description,
		&section.SortOrder,
		&section.CreatedAt,
		&section.UpdatedAt,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{
			Success: false,
			Message: "Failed to fetch created course section",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, dto.APIResponse{
		Success: true,
		Message: "Course section created successfully",
		Data:    section,
	})
}

// PUT /api/course-sections/:id
func (h *CourseSectionHandler) UpdateCourseSection(c *gin.Context) {
	id := c.Param("id")
	
	if _, err := uuid.Parse(id); err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{
			Success: false,
			Message: "Invalid course section ID format",
			Error:   err.Error(),
		})
		return
	}

	var req dto.UpdateCourseSectionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{
			Success: false,
			Message: "Invalid request body",
			Error:   err.Error(),
		})
		return
	}

	// Check if section exists
	var exists bool
	err := h.db.QueryRow("SELECT EXISTS(SELECT 1 FROM course_sections WHERE id = $1)", id).Scan(&exists)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{
			Success: false,
			Message: "Failed to check course section existence",
			Error:   err.Error(),
		})
		return
	}

	if !exists {
		c.JSON(http.StatusNotFound, dto.APIResponse{
			Success: false,
			Message: "Course section not found",
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

	if req.Description != nil {
		setParts = append(setParts, "description = $"+strconv.Itoa(argIndex))
		args = append(args, *req.Description)
		argIndex++
	}

	if req.SortOrder != nil {
		setParts = append(setParts, "sort_order = $"+strconv.Itoa(argIndex))
		args = append(args, *req.SortOrder)
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

	query := "UPDATE course_sections SET " + setParts[0]
	for i := 1; i < len(setParts); i++ {
		query += ", " + setParts[i]
	}
	query += " WHERE id = " + whereClause

	_, err = h.db.Exec(query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{
			Success: false,
			Message: "Failed to update course section",
			Error:   err.Error(),
		})
		return
	}

	// Fetch updated section
	var section dto.CourseSectionResponse
	err = h.db.QueryRow(`
		SELECT id, course_id, title, description, sort_order, created_at, updated_at
		FROM course_sections WHERE id = $1
	`, id).Scan(
		&section.ID,
		&section.CourseID,
		&section.Title,
		&section.Description,
		&section.SortOrder,
		&section.CreatedAt,
		&section.UpdatedAt,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{
			Success: false,
			Message: "Failed to fetch updated course section",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, dto.APIResponse{
		Success: true,
		Message: "Course section updated successfully",
		Data:    section,
	})
}

// DELETE /api/course-sections/:id
func (h *CourseSectionHandler) DeleteCourseSection(c *gin.Context) {
	id := c.Param("id")
	
	if _, err := uuid.Parse(id); err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{
			Success: false,
			Message: "Invalid course section ID format",
			Error:   err.Error(),
		})
		return
	}

	// Check if section has lectures
	var lectureCount int
	err := h.db.QueryRow("SELECT COUNT(*) FROM course_lectures WHERE section_id = $1", id).Scan(&lectureCount)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{
			Success: false,
			Message: "Failed to check section lectures",
			Error:   err.Error(),
		})
		return
	}

	if lectureCount > 0 {
		c.JSON(http.StatusBadRequest, dto.APIResponse{
			Success: false,
			Message: "Cannot delete section that contains lectures",
		})
		return
	}

	result, err := h.db.Exec("DELETE FROM course_sections WHERE id = $1", id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{
			Success: false,
			Message: "Failed to delete course section",
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
			Message: "Course section not found",
		})
		return
	}

	c.JSON(http.StatusOK, dto.APIResponse{
		Success: true,
		Message: "Course section deleted successfully",
	})
}

// Helper function to get lectures for a section
func (h *CourseSectionHandler) getLecturesForSection(sectionID string) ([]dto.CourseLectureResponse, error) {
	rows, err := h.db.Query(`
		SELECT id, section_id, title, description, content_type, video_url, video_duration,
			   article_content, file_url, sort_order, is_preview, is_downloadable, 
			   created_at, updated_at
		FROM course_lectures 
		WHERE section_id = $1 
		ORDER BY sort_order ASC
	`, sectionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var lectures []dto.CourseLectureResponse
	for rows.Next() {
		var lecture dto.CourseLectureResponse
		err := rows.Scan(
			&lecture.ID,
			&lecture.SectionID,
			&lecture.Title,
			&lecture.Description,
			&lecture.ContentType,
			&lecture.VideoURL,
			&lecture.VideoDuration,
			&lecture.ArticleContent,
			&lecture.FileURL,
			&lecture.SortOrder,
			&lecture.IsPreview,
			&lecture.IsDownloadable,
			&lecture.CreatedAt,
			&lecture.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		lectures = append(lectures, lecture)
	}

	return lectures, nil
}
