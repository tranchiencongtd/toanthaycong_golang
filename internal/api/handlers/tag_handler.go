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

type TagHandler struct {
	db *sql.DB
}

func NewTagHandler(db *sql.DB) *TagHandler {
	return &TagHandler{db: db}
}

// GetTags godoc
// @Summary Lấy danh sách tags
// @Description Lấy danh sách tags với phân trang và tìm kiếm
// @Tags Tag
// @Accept json
// @Produce json
// @Param page query int false "Số trang" default(1)
// @Param limit query int false "Số item mỗi trang" default(10)
// @Param search query string false "Tìm kiếm theo tên tag"
// @Success 200 {object} dto.TagListResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/tags [get]
func (h *TagHandler) GetTags(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	search := c.Query("search")

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	offset := (page - 1) * limit

	// Build query với search
	query := `
		SELECT t.id, t.name, t.slug, t.description, t.color, t.created_at,
		       COALESCE(ct_count.course_count, 0) as course_count,
		       COUNT(*) OVER() as total_count
		FROM tags t
		LEFT JOIN (
			SELECT tag_id, COUNT(*) as course_count
			FROM course_tags
			GROUP BY tag_id
		) ct_count ON t.id = ct_count.tag_id
		WHERE 1=1`
	
	args := []interface{}{}
	argIndex := 1

	if search != "" {
		query += fmt.Sprintf(" AND (t.name ILIKE $%d OR t.slug ILIKE $%d)", argIndex, argIndex)
		searchTerm := "%" + search + "%"
		args = append(args, searchTerm)
		argIndex++
	}

	query += fmt.Sprintf(" ORDER BY t.name ASC LIMIT $%d OFFSET $%d", argIndex, argIndex+1)
	args = append(args, limit, offset)

	rows, err := h.db.Query(query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "Database error",
			Message: "Failed to fetch tags",
		})
		return
	}
	defer rows.Close()

	var tags []dto.TagDTO
	var totalCount int

	for rows.Next() {
		var tag dto.TagDTO
		var description sql.NullString
		var color sql.NullString
		var courseCount int

		err := rows.Scan(
			&tag.ID, &tag.Name, &tag.Slug, &description, &color,
			&tag.CreatedAt, &courseCount, &totalCount,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
				Error:   "Scan error",
				Message: "Failed to parse tag data",
			})
			return
		}

		if description.Valid {
			tag.Description = &description.String
		}
		if color.Valid {
			tag.Color = &color.String
		}
		tag.CourseCount = &courseCount

		tags = append(tags, tag)
	}

	// Tính toán pagination
	totalPages := (totalCount + limit - 1) / limit

	response := dto.TagListResponse{
		Data: tags,
		Pagination: dto.PaginationMeta{
			Page:       page,
			Limit:      limit,
			TotalItems: totalCount,
			TotalPages: totalPages,
		},
	}

	c.JSON(http.StatusOK, response)
}

// GetTag godoc
// @Summary Lấy thông tin tag theo ID
// @Description Lấy thông tin chi tiết tag theo ID
// @Tags Tag
// @Accept json
// @Produce json
// @Param id path string true "ID tag"
// @Success 200 {object} dto.TagDTO
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/tags/{id} [get]
func (h *TagHandler) GetTag(c *gin.Context) {
	id := c.Param("id")

	if _, err := uuid.Parse(id); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "Invalid ID",
			Message: "Invalid tag ID format",
		})
		return
	}

	query := `
		SELECT t.id, t.name, t.slug, t.description, t.color, t.created_at,
		       COALESCE(ct_count.course_count, 0) as course_count
		FROM tags t
		LEFT JOIN (
			SELECT tag_id, COUNT(*) as course_count
			FROM course_tags
			WHERE tag_id = $1
			GROUP BY tag_id
		) ct_count ON t.id = ct_count.tag_id
		WHERE t.id = $1`

	var tag dto.TagDTO
	var description sql.NullString
	var color sql.NullString
	var courseCount int

	err := h.db.QueryRow(query, id).Scan(
		&tag.ID, &tag.Name, &tag.Slug, &description, &color,
		&tag.CreatedAt, &courseCount,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, dto.ErrorResponse{
				Error:   "Not found",
				Message: "Tag not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "Database error",
			Message: "Failed to fetch tag",
		})
		return
	}

	if description.Valid {
		tag.Description = &description.String
	}
	if color.Valid {
		tag.Color = &color.String
	}
	tag.CourseCount = &courseCount

	c.JSON(http.StatusOK, tag)
}

// CreateTag godoc
// @Summary Tạo tag mới
// @Description Tạo tag mới
// @Tags Tag
// @Accept json
// @Produce json
// @Param body body dto.CreateTagRequest true "Thông tin tag"
// @Success 201 {object} dto.TagDTO
// @Failure 400 {object} dto.ErrorResponse
// @Failure 409 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/tags [post]
func (h *TagHandler) CreateTag(c *gin.Context) {
	var req dto.CreateTagRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "Invalid input",
			Message: err.Error(),
		})
		return
	}

	// Kiểm tra name và slug đã tồn tại chưa
	var nameExists, slugExists bool
	h.db.QueryRow("SELECT EXISTS(SELECT 1 FROM tags WHERE name = $1)", req.Name).Scan(&nameExists)
	h.db.QueryRow("SELECT EXISTS(SELECT 1 FROM tags WHERE slug = $1)", req.Slug).Scan(&slugExists)

	if nameExists {
		c.JSON(http.StatusConflict, dto.ErrorResponse{
			Error:   "Conflict",
			Message: "Tag name already exists",
		})
		return
	}

	if slugExists {
		c.JSON(http.StatusConflict, dto.ErrorResponse{
			Error:   "Conflict",
			Message: "Tag slug already exists",
		})
		return
	}

	id := uuid.New().String()
	now := time.Now()

	var description sql.NullString
	if req.Description != nil {
		description = sql.NullString{String: *req.Description, Valid: true}
	}

	var color sql.NullString
	if req.Color != nil {
		color = sql.NullString{String: *req.Color, Valid: true}
	}

	query := `
		INSERT INTO tags (id, name, slug, description, color, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, name, slug, description, color, created_at`

	var tag dto.TagDTO

	err := h.db.QueryRow(query, id, req.Name, req.Slug, description, color, now).Scan(
		&tag.ID, &tag.Name, &tag.Slug, &description, &color, &tag.CreatedAt,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "Database error",
			Message: "Failed to create tag",
		})
		return
	}

	if description.Valid {
		tag.Description = &description.String
	}
	if color.Valid {
		tag.Color = &color.String
	}
	courseCount := 0
	tag.CourseCount = &courseCount

	c.JSON(http.StatusCreated, tag)
}

// UpdateTag godoc
// @Summary Cập nhật tag
// @Description Cập nhật thông tin tag
// @Tags Tag
// @Accept json
// @Produce json
// @Param id path string true "ID tag"
// @Param body body dto.UpdateTagRequest true "Thông tin cập nhật"
// @Success 200 {object} dto.TagDTO
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/tags/{id} [put]
func (h *TagHandler) UpdateTag(c *gin.Context) {
	id := c.Param("id")

	if _, err := uuid.Parse(id); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "Invalid ID",
			Message: "Invalid tag ID format",
		})
		return
	}

	var req dto.UpdateTagRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "Invalid input",
			Message: err.Error(),
		})
		return
	}

	// Kiểm tra tag tồn tại
	var exists bool
	err := h.db.QueryRow("SELECT EXISTS(SELECT 1 FROM tags WHERE id = $1)", id).Scan(&exists)
	if err != nil || !exists {
		c.JSON(http.StatusNotFound, dto.ErrorResponse{
			Error:   "Not found",
			Message: "Tag not found",
		})
		return
	}

	// Build update query
	setParts := []string{}
	args := []interface{}{}
	argIndex := 1

	if req.Name != nil {
		// Kiểm tra name mới có trùng không
		var nameExists bool
		h.db.QueryRow("SELECT EXISTS(SELECT 1 FROM tags WHERE name = $1 AND id != $2)", *req.Name, id).Scan(&nameExists)
		if nameExists {
			c.JSON(http.StatusConflict, dto.ErrorResponse{
				Error:   "Conflict",
				Message: "Tag name already exists",
			})
			return
		}
		setParts = append(setParts, fmt.Sprintf("name = $%d", argIndex))
		args = append(args, *req.Name)
		argIndex++
	}

	if req.Slug != nil {
		// Kiểm tra slug mới có trùng không
		var slugExists bool
		h.db.QueryRow("SELECT EXISTS(SELECT 1 FROM tags WHERE slug = $1 AND id != $2)", *req.Slug, id).Scan(&slugExists)
		if slugExists {
			c.JSON(http.StatusConflict, dto.ErrorResponse{
				Error:   "Conflict",
				Message: "Tag slug already exists",
			})
			return
		}
		setParts = append(setParts, fmt.Sprintf("slug = $%d", argIndex))
		args = append(args, *req.Slug)
		argIndex++
	}

	if req.Description != nil {
		setParts = append(setParts, fmt.Sprintf("description = $%d", argIndex))
		args = append(args, *req.Description)
		argIndex++
	}

	if req.Color != nil {
		setParts = append(setParts, fmt.Sprintf("color = $%d", argIndex))
		args = append(args, *req.Color)
		argIndex++
	}

	if len(setParts) == 0 {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "No updates",
			Message: "No fields to update",
		})
		return
	}

	query := "UPDATE tags SET "
	for i, part := range setParts {
		if i > 0 {
			query += ", "
		}
		query += part
	}
	query += fmt.Sprintf(" WHERE id = $%d RETURNING id, name, slug, description, color, created_at", argIndex)
	args = append(args, id)

	var tag dto.TagDTO
	var description sql.NullString
	var color sql.NullString

	err = h.db.QueryRow(query, args...).Scan(
		&tag.ID, &tag.Name, &tag.Slug, &description, &color, &tag.CreatedAt,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "Database error",
			Message: "Failed to update tag",
		})
		return
	}

	if description.Valid {
		tag.Description = &description.String
	}
	if color.Valid {
		tag.Color = &color.String
	}

	// Lấy course count
	var courseCount int
	h.db.QueryRow("SELECT COUNT(*) FROM course_tags WHERE tag_id = $1", id).Scan(&courseCount)
	tag.CourseCount = &courseCount

	c.JSON(http.StatusOK, tag)
}

// DeleteTag godoc
// @Summary Xóa tag
// @Description Xóa tag theo ID
// @Tags Tag
// @Accept json
// @Produce json
// @Param id path string true "ID tag"
// @Success 204 "No Content"
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/tags/{id} [delete]
func (h *TagHandler) DeleteTag(c *gin.Context) {
	id := c.Param("id")

	if _, err := uuid.Parse(id); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "Invalid ID",
			Message: "Invalid tag ID format",
		})
		return
	}

	result, err := h.db.Exec("DELETE FROM tags WHERE id = $1", id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "Database error",
			Message: "Failed to delete tag",
		})
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, dto.ErrorResponse{
			Error:   "Not found",
			Message: "Tag not found",
		})
		return
	}

	c.Status(http.StatusNoContent)
}

// === COURSE TAGS ===

// GetCourseTags godoc
// @Summary Lấy danh sách tags của khóa học
// @Description Lấy danh sách tags của khóa học theo course ID
// @Tags Tag
// @Accept json
// @Produce json
// @Param course_id path string true "ID khóa học"
// @Success 200 {object} dto.CourseTagListResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/courses/{course_id}/tags [get]
func (h *TagHandler) GetCourseTags(c *gin.Context) {
	courseID := c.Param("course_id")

	if _, err := uuid.Parse(courseID); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "Invalid ID",
			Message: "Invalid course ID format",
		})
		return
	}

	query := `
		SELECT ct.course_id, ct.tag_id, t.name, t.slug, t.description, t.color, t.created_at
		FROM course_tags ct
		JOIN tags t ON ct.tag_id = t.id
		WHERE ct.course_id = $1
		ORDER BY t.name ASC`

	rows, err := h.db.Query(query, courseID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "Database error",
			Message: "Failed to fetch course tags",
		})
		return
	}
	defer rows.Close()

	var courseTags []dto.CourseTagDTO

	for rows.Next() {
		var courseTag dto.CourseTagDTO
		var tag dto.TagDTO
		var description sql.NullString
		var color sql.NullString

		err := rows.Scan(
			&courseTag.CourseID, &courseTag.TagID,
			&tag.Name, &tag.Slug, &description, &color, &tag.CreatedAt,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
				Error:   "Scan error",
				Message: "Failed to parse course tag data",
			})
			return
		}

		tag.ID = courseTag.TagID
		if description.Valid {
			tag.Description = &description.String
		}
		if color.Valid {
			tag.Color = &color.String
		}

		courseTag.Tag = &tag
		courseTags = append(courseTags, courseTag)
	}

	response := dto.CourseTagListResponse{
		Data: courseTags,
	}

	c.JSON(http.StatusOK, response)
}

// AddCourseTag godoc
// @Summary Thêm tag cho khóa học
// @Description Thêm tag cho khóa học
// @Tags Tag
// @Accept json
// @Produce json
// @Param body body dto.AddCourseTagRequest true "Thông tin course tag"
// @Success 201 {object} dto.CourseTagDTO
// @Failure 400 {object} dto.ErrorResponse
// @Failure 409 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/course-tags [post]
func (h *TagHandler) AddCourseTag(c *gin.Context) {
	var req dto.AddCourseTagRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "Invalid input",
			Message: err.Error(),
		})
		return
	}

	// Kiểm tra course và tag tồn tại
	var courseExists, tagExists bool
	h.db.QueryRow("SELECT EXISTS(SELECT 1 FROM courses WHERE id = $1)", req.CourseID).Scan(&courseExists)
	h.db.QueryRow("SELECT EXISTS(SELECT 1 FROM tags WHERE id = $1)", req.TagID).Scan(&tagExists)

	if !courseExists {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "Invalid course",
			Message: "Course not found",
		})
		return
	}

	if !tagExists {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "Invalid tag",
			Message: "Tag not found",
		})
		return
	}

	// Kiểm tra đã tồn tại chưa
	var exists bool
	h.db.QueryRow("SELECT EXISTS(SELECT 1 FROM course_tags WHERE course_id = $1 AND tag_id = $2)", req.CourseID, req.TagID).Scan(&exists)
	if exists {
		c.JSON(http.StatusConflict, dto.ErrorResponse{
			Error:   "Conflict",
			Message: "Tag already added to this course",
		})
		return
	}

	_, err := h.db.Exec("INSERT INTO course_tags (course_id, tag_id) VALUES ($1, $2)", req.CourseID, req.TagID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "Database error",
			Message: "Failed to add tag to course",
		})
		return
	}

	// Lấy thông tin tag để trả về
	query := `
		SELECT t.id, t.name, t.slug, t.description, t.color, t.created_at
		FROM tags t
		WHERE t.id = $1`

	var tag dto.TagDTO
	var description sql.NullString
	var color sql.NullString

	err = h.db.QueryRow(query, req.TagID).Scan(
		&tag.ID, &tag.Name, &tag.Slug, &description, &color, &tag.CreatedAt,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "Database error",
			Message: "Failed to fetch tag info",
		})
		return
	}

	if description.Valid {
		tag.Description = &description.String
	}
	if color.Valid {
		tag.Color = &color.String
	}

	courseTag := dto.CourseTagDTO{
		CourseID: req.CourseID,
		TagID:    req.TagID,
		Tag:      &tag,
	}

	c.JSON(http.StatusCreated, courseTag)
}

// RemoveCourseTag godoc
// @Summary Xóa tag khỏi khóa học
// @Description Xóa tag khỏi khóa học
// @Tags Tag
// @Accept json
// @Produce json
// @Param course_id query string true "ID khóa học"
// @Param tag_id query string true "ID tag"
// @Success 204 "No Content"
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/course-tags/remove [delete]
func (h *TagHandler) RemoveCourseTag(c *gin.Context) {
	courseID := c.Query("course_id")
	tagID := c.Query("tag_id")

	if courseID == "" || tagID == "" {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "Missing parameters",
			Message: "Both course_id and tag_id are required",
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

	if _, err := uuid.Parse(tagID); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "Invalid tag ID",
			Message: "Invalid tag ID format",
		})
		return
	}

	result, err := h.db.Exec("DELETE FROM course_tags WHERE course_id = $1 AND tag_id = $2", courseID, tagID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "Database error",
			Message: "Failed to remove tag from course",
		})
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, dto.ErrorResponse{
			Error:   "Not found",
			Message: "Course tag relationship not found",
		})
		return
	}

	c.Status(http.StatusNoContent)
}
