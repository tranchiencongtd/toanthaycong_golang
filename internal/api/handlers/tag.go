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

type TagHandler struct {
	db *sql.DB
}

func NewTagHandler(db *sql.DB) *TagHandler {
	return &TagHandler{db: db}
}

// GET /api/tags
func (h *TagHandler) GetTags(c *gin.Context) {
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

	// Search by name if provided
	search := c.Query("search")
	var args []interface{}
	baseQuery := `
		SELECT id, name, slug, description, color, created_at, updated_at
		FROM tags 
		WHERE 1=1`
	
	countQuery := "SELECT COUNT(*) FROM tags WHERE 1=1"

	if search != "" {
		baseQuery += " AND name ILIKE $1"
		countQuery += " AND name ILIKE $1"
		args = append(args, "%"+search+"%")
	}

	// Get total count
	var total int64
	err := h.db.QueryRow(countQuery, args...).Scan(&total)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{
			Success: false,
			Message: "Failed to count tags",
			Error:   err.Error(),
		})
		return
	}

	// Add pagination
	baseQuery += " ORDER BY name ASC LIMIT $" + strconv.Itoa(len(args)+1) + " OFFSET $" + strconv.Itoa(len(args)+2)
	args = append(args, query.Limit, query.GetOffset())

	rows, err := h.db.Query(baseQuery, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{
			Success: false,
			Message: "Failed to fetch tags",
			Error:   err.Error(),
		})
		return
	}
	defer rows.Close()

	var tags []dto.TagResponse
	for rows.Next() {
		var tag dto.TagResponse
		err := rows.Scan(
			&tag.ID,
			&tag.Name,
			&tag.Slug,
			&tag.Description,
			&tag.Color,
			&tag.CreatedAt,
			&tag.UpdatedAt,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, dto.APIResponse{
				Success: false,
				Message: "Failed to scan tag",
				Error:   err.Error(),
			})
			return
		}
		tags = append(tags, tag)
	}

	pagination := dto.NewPaginationResponse(total, query.Page, query.Limit)

	c.JSON(http.StatusOK, dto.APIResponse{
		Success: true,
		Message: "Tags retrieved successfully",
		Data: dto.TagListResponse{
			Tags:       tags,
			Pagination: pagination,
		},
	})
}

// GET /api/tags/:id
func (h *TagHandler) GetTag(c *gin.Context) {
	id := c.Param("id")
	
	if _, err := uuid.Parse(id); err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{
			Success: false,
			Message: "Invalid tag ID format",
			Error:   err.Error(),
		})
		return
	}

	var tag dto.TagResponse
	err := h.db.QueryRow(`
		SELECT id, name, slug, description, color, created_at, updated_at
		FROM tags WHERE id = $1
	`, id).Scan(
		&tag.ID,
		&tag.Name,
		&tag.Slug,
		&tag.Description,
		&tag.Color,
		&tag.CreatedAt,
		&tag.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, dto.APIResponse{
				Success: false,
				Message: "Tag not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, dto.APIResponse{
			Success: false,
			Message: "Failed to fetch tag",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, dto.APIResponse{
		Success: true,
		Message: "Tag retrieved successfully",
		Data:    tag,
	})
}

// POST /api/tags
func (h *TagHandler) CreateTag(c *gin.Context) {
	var req dto.CreateTagRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{
			Success: false,
			Message: "Invalid request body",
			Error:   err.Error(),
		})
		return
	}

	id := uuid.New().String()

	_, err := h.db.Exec(`
		INSERT INTO tags (id, name, slug, description, color, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	`, id, req.Name, req.Slug, req.Description, req.Color)

	if err != nil {
		// Check for unique constraint violations
		if pqErr, ok := err.(*pq.Error); ok {
			switch pqErr.Constraint {
			case "tags_name_key":
				c.JSON(http.StatusConflict, dto.APIResponse{
					Success: false,
					Message: "Tag name already exists",
				})
				return
			case "tags_slug_key":
				c.JSON(http.StatusConflict, dto.APIResponse{
					Success: false,
					Message: "Tag slug already exists",
				})
				return
			}
		}
		
		c.JSON(http.StatusInternalServerError, dto.APIResponse{
			Success: false,
			Message: "Failed to create tag",
			Error:   err.Error(),
		})
		return
	}

	// Fetch the created tag
	var tag dto.TagResponse
	err = h.db.QueryRow(`
		SELECT id, name, slug, description, color, created_at, updated_at
		FROM tags WHERE id = $1
	`, id).Scan(
		&tag.ID,
		&tag.Name,
		&tag.Slug,
		&tag.Description,
		&tag.Color,
		&tag.CreatedAt,
		&tag.UpdatedAt,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{
			Success: false,
			Message: "Failed to fetch created tag",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, dto.APIResponse{
		Success: true,
		Message: "Tag created successfully",
		Data:    tag,
	})
}

// PUT /api/tags/:id
func (h *TagHandler) UpdateTag(c *gin.Context) {
	id := c.Param("id")
	
	if _, err := uuid.Parse(id); err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{
			Success: false,
			Message: "Invalid tag ID format",
			Error:   err.Error(),
		})
		return
	}

	var req dto.UpdateTagRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{
			Success: false,
			Message: "Invalid request body",
			Error:   err.Error(),
		})
		return
	}

	// Check if tag exists
	var exists bool
	err := h.db.QueryRow("SELECT EXISTS(SELECT 1 FROM tags WHERE id = $1)", id).Scan(&exists)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{
			Success: false,
			Message: "Failed to check tag existence",
			Error:   err.Error(),
		})
		return
	}

	if !exists {
		c.JSON(http.StatusNotFound, dto.APIResponse{
			Success: false,
			Message: "Tag not found",
		})
		return
	}

	// Build dynamic update query
	setParts := []string{"updated_at = CURRENT_TIMESTAMP"}
	args := []interface{}{}
	argIndex := 1

	if req.Name != nil {
		setParts = append(setParts, "name = $"+strconv.Itoa(argIndex))
		args = append(args, *req.Name)
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

	if req.Color != nil {
		setParts = append(setParts, "color = $"+strconv.Itoa(argIndex))
		args = append(args, *req.Color)
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

	query := "UPDATE tags SET " + setParts[0]
	for i := 1; i < len(setParts); i++ {
		query += ", " + setParts[i]
	}
	query += " WHERE id = " + whereClause

	_, err = h.db.Exec(query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{
			Success: false,
			Message: "Failed to update tag",
			Error:   err.Error(),
		})
		return
	}

	// Fetch updated tag
	var tag dto.TagResponse
	err = h.db.QueryRow(`
		SELECT id, name, slug, description, color, created_at, updated_at
		FROM tags WHERE id = $1
	`, id).Scan(
		&tag.ID,
		&tag.Name,
		&tag.Slug,
		&tag.Description,
		&tag.Color,
		&tag.CreatedAt,
		&tag.UpdatedAt,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{
			Success: false,
			Message: "Failed to fetch updated tag",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, dto.APIResponse{
		Success: true,
		Message: "Tag updated successfully",
		Data:    tag,
	})
}

// DELETE /api/tags/:id
func (h *TagHandler) DeleteTag(c *gin.Context) {
	id := c.Param("id")
	
	if _, err := uuid.Parse(id); err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{
			Success: false,
			Message: "Invalid tag ID format",
			Error:   err.Error(),
		})
		return
	}

	result, err := h.db.Exec("DELETE FROM tags WHERE id = $1", id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{
			Success: false,
			Message: "Failed to delete tag",
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
			Message: "Tag not found",
		})
		return
	}

	c.JSON(http.StatusOK, dto.APIResponse{
		Success: true,
		Message: "Tag deleted successfully",
	})
}
