package handlers

import (
	"database/sql"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"internal/api/dto"
)

type CourseLectureHandler struct {
	db *sql.DB
}

func NewCourseLectureHandler(db *sql.DB) *CourseLectureHandler {
	return &CourseLectureHandler{db: db}
}

// GET /api/course-lectures
func (h *CourseLectureHandler) GetCourseLectures(c *gin.Context) {
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

	// Filter by section_id
	sectionID := c.Query("section_id")
	contentType := c.Query("content_type")
	var args []interface{}
	baseQuery := `
		SELECT id, section_id, title, description, content_type, video_url, video_duration,
			   article_content, file_url, sort_order, is_preview, is_downloadable, 
			   created_at, updated_at
		FROM course_lectures 
		WHERE 1=1`
	
	countQuery := "SELECT COUNT(*) FROM course_lectures WHERE 1=1"

	if sectionID != "" {
		if _, err := uuid.Parse(sectionID); err != nil {
			c.JSON(http.StatusBadRequest, dto.APIResponse{
				Success: false,
				Message: "Invalid section ID format",
				Error:   err.Error(),
			})
			return
		}
		baseQuery += " AND section_id = $" + strconv.Itoa(len(args)+1)
		countQuery += " AND section_id = $" + strconv.Itoa(len(args)+1)
		args = append(args, sectionID)
	}

	if contentType != "" {
		baseQuery += " AND content_type = $" + strconv.Itoa(len(args)+1)
		countQuery += " AND content_type = $" + strconv.Itoa(len(args)+1)
		args = append(args, contentType)
	}

	// Get total count
	var total int64
	err := h.db.QueryRow(countQuery, args...).Scan(&total)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{
			Success: false,
			Message: "Failed to count course lectures",
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
			Message: "Failed to fetch course lectures",
			Error:   err.Error(),
		})
		return
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
			c.JSON(http.StatusInternalServerError, dto.APIResponse{
				Success: false,
				Message: "Failed to scan course lecture",
				Error:   err.Error(),
			})
			return
		}
		lectures = append(lectures, lecture)
	}

	pagination := dto.NewPaginationResponse(total, query.Page, query.Limit)

	c.JSON(http.StatusOK, dto.APIResponse{
		Success: true,
		Message: "Course lectures retrieved successfully",
		Data: dto.CourseLectureListResponse{
			Lectures:   lectures,
			Pagination: pagination,
		},
	})
}

// GET /api/course-lectures/:id
func (h *CourseLectureHandler) GetCourseLecture(c *gin.Context) {
	id := c.Param("id")
	
	if _, err := uuid.Parse(id); err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{
			Success: false,
			Message: "Invalid course lecture ID format",
			Error:   err.Error(),
		})
		return
	}

	var lecture dto.CourseLectureResponse
	err := h.db.QueryRow(`
		SELECT id, section_id, title, description, content_type, video_url, video_duration,
			   article_content, file_url, sort_order, is_preview, is_downloadable, 
			   created_at, updated_at
		FROM course_lectures WHERE id = $1
	`, id).Scan(
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
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, dto.APIResponse{
				Success: false,
				Message: "Course lecture not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, dto.APIResponse{
			Success: false,
			Message: "Failed to fetch course lecture",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, dto.APIResponse{
		Success: true,
		Message: "Course lecture retrieved successfully",
		Data:    lecture,
	})
}

// POST /api/course-lectures
func (h *CourseLectureHandler) CreateCourseLecture(c *gin.Context) {
	var req dto.CreateCourseLectureRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{
			Success: false,
			Message: "Invalid request body",
			Error:   err.Error(),
		})
		return
	}

	// Validate section_id
	if _, err := uuid.Parse(req.SectionID); err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{
			Success: false,
			Message: "Invalid section ID format",
			Error:   err.Error(),
		})
		return
	}

	// Verify section exists
	var sectionExists bool
	err := h.db.QueryRow("SELECT EXISTS(SELECT 1 FROM course_sections WHERE id = $1)", req.SectionID).Scan(&sectionExists)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{
			Success: false,
			Message: "Failed to verify section",
			Error:   err.Error(),
		})
		return
	}

	if !sectionExists {
		c.JSON(http.StatusBadRequest, dto.APIResponse{
			Success: false,
			Message: "Section not found",
		})
		return
	}

	id := uuid.New().String()

	// Set default values
	isPreview := false
	isDownloadable := false
	if req.IsPreview != nil {
		isPreview = *req.IsPreview
	}
	if req.IsDownloadable != nil {
		isDownloadable = *req.IsDownloadable
	}

	_, err = h.db.Exec(`
		INSERT INTO course_lectures (
			id, section_id, title, description, content_type, video_url, video_duration,
			article_content, file_url, sort_order, is_preview, is_downloadable, 
			created_at, updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	`, id, req.SectionID, req.Title, req.Description, req.ContentType, req.VideoURL, req.VideoDuration,
		req.ArticleContent, req.FileURL, req.SortOrder, isPreview, isDownloadable)

	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{
			Success: false,
			Message: "Failed to create course lecture",
			Error:   err.Error(),
		})
		return
	}

	// Fetch the created lecture
	var lecture dto.CourseLectureResponse
	err = h.db.QueryRow(`
		SELECT id, section_id, title, description, content_type, video_url, video_duration,
			   article_content, file_url, sort_order, is_preview, is_downloadable, 
			   created_at, updated_at
		FROM course_lectures WHERE id = $1
	`, id).Scan(
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
		c.JSON(http.StatusInternalServerError, dto.APIResponse{
			Success: false,
			Message: "Failed to fetch created course lecture",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, dto.APIResponse{
		Success: true,
		Message: "Course lecture created successfully",
		Data:    lecture,
	})
}

// PUT /api/course-lectures/:id
func (h *CourseLectureHandler) UpdateCourseLecture(c *gin.Context) {
	id := c.Param("id")
	
	if _, err := uuid.Parse(id); err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{
			Success: false,
			Message: "Invalid course lecture ID format",
			Error:   err.Error(),
		})
		return
	}

	var req dto.UpdateCourseLectureRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{
			Success: false,
			Message: "Invalid request body",
			Error:   err.Error(),
		})
		return
	}

	// Check if lecture exists
	var exists bool
	err := h.db.QueryRow("SELECT EXISTS(SELECT 1 FROM course_lectures WHERE id = $1)", id).Scan(&exists)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{
			Success: false,
			Message: "Failed to check course lecture existence",
			Error:   err.Error(),
		})
		return
	}

	if !exists {
		c.JSON(http.StatusNotFound, dto.APIResponse{
			Success: false,
			Message: "Course lecture not found",
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

	if req.ContentType != nil {
		setParts = append(setParts, "content_type = $"+strconv.Itoa(argIndex))
		args = append(args, *req.ContentType)
		argIndex++
	}

	if req.VideoURL != nil {
		setParts = append(setParts, "video_url = $"+strconv.Itoa(argIndex))
		args = append(args, *req.VideoURL)
		argIndex++
	}

	if req.VideoDuration != nil {
		setParts = append(setParts, "video_duration = $"+strconv.Itoa(argIndex))
		args = append(args, *req.VideoDuration)
		argIndex++
	}

	if req.ArticleContent != nil {
		setParts = append(setParts, "article_content = $"+strconv.Itoa(argIndex))
		args = append(args, *req.ArticleContent)
		argIndex++
	}

	if req.FileURL != nil {
		setParts = append(setParts, "file_url = $"+strconv.Itoa(argIndex))
		args = append(args, *req.FileURL)
		argIndex++
	}

	if req.SortOrder != nil {
		setParts = append(setParts, "sort_order = $"+strconv.Itoa(argIndex))
		args = append(args, *req.SortOrder)
		argIndex++
	}

	if req.IsPreview != nil {
		setParts = append(setParts, "is_preview = $"+strconv.Itoa(argIndex))
		args = append(args, *req.IsPreview)
		argIndex++
	}

	if req.IsDownloadable != nil {
		setParts = append(setParts, "is_downloadable = $"+strconv.Itoa(argIndex))
		args = append(args, *req.IsDownloadable)
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

	query := "UPDATE course_lectures SET " + setParts[0]
	for i := 1; i < len(setParts); i++ {
		query += ", " + setParts[i]
	}
	query += " WHERE id = " + whereClause

	_, err = h.db.Exec(query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{
			Success: false,
			Message: "Failed to update course lecture",
			Error:   err.Error(),
		})
		return
	}

	// Fetch updated lecture
	var lecture dto.CourseLectureResponse
	err = h.db.QueryRow(`
		SELECT id, section_id, title, description, content_type, video_url, video_duration,
			   article_content, file_url, sort_order, is_preview, is_downloadable, 
			   created_at, updated_at
		FROM course_lectures WHERE id = $1
	`, id).Scan(
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
		c.JSON(http.StatusInternalServerError, dto.APIResponse{
			Success: false,
			Message: "Failed to fetch updated course lecture",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, dto.APIResponse{
		Success: true,
		Message: "Course lecture updated successfully",
		Data:    lecture,
	})
}

// DELETE /api/course-lectures/:id
func (h *CourseLectureHandler) DeleteCourseLecture(c *gin.Context) {
	id := c.Param("id")
	
	if _, err := uuid.Parse(id); err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{
			Success: false,
			Message: "Invalid course lecture ID format",
			Error:   err.Error(),
		})
		return
	}

	result, err := h.db.Exec("DELETE FROM course_lectures WHERE id = $1", id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{
			Success: false,
			Message: "Failed to delete course lecture",
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
			Message: "Course lecture not found",
		})
		return
	}

	c.JSON(http.StatusOK, dto.APIResponse{
		Success: true,
		Message: "Course lecture deleted successfully",
	})
}
