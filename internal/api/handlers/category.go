package handlers

import (
	"database/sql"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"internal/api/dto"
)

type CategoryHandler struct {
	db *sql.DB
}

func NewCategoryHandler(db *sql.DB) *CategoryHandler {
	return &CategoryHandler{db: db}
}

// GET /api/categories
func (h *CategoryHandler) GetCategories(c *gin.Context) {
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

	// Build query with optional parent filter
	parentID := c.Query("parent_id")
	var sqlQuery string
	var args []interface{}
	
	baseQuery := `
		SELECT id, name, slug, description, icon_url, parent_id, sort_order, is_active, created_at, updated_at
		FROM categories 
		WHERE 1=1`
	
	countQuery := "SELECT COUNT(*) FROM categories WHERE 1=1"

	if parentID != "" {
		if parentID == "null" {
			baseQuery += " AND parent_id IS NULL"
			countQuery += " AND parent_id IS NULL"
		} else {
			baseQuery += " AND parent_id = $1"
			countQuery += " AND parent_id = $1"
			args = append(args, parentID)
		}
	}

	// Get total count
	var total int64
	err := h.db.QueryRow(countQuery, args...).Scan(&total)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{
			Success: false,
			Message: "Failed to count categories",
			Error:   err.Error(),
		})
		return
	}

	// Add pagination
	baseQuery += " ORDER BY sort_order ASC, name ASC LIMIT $" + strconv.Itoa(len(args)+1) + " OFFSET $" + strconv.Itoa(len(args)+2)
	args = append(args, query.Limit, query.GetOffset())

	rows, err := h.db.Query(baseQuery, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{
			Success: false,
			Message: "Failed to fetch categories",
			Error:   err.Error(),
		})
		return
	}
	defer rows.Close()

	var categories []dto.CategoryResponse
	for rows.Next() {
		var category dto.CategoryResponse
		err := rows.Scan(
			&category.ID,
			&category.Name,
			&category.Slug,
			&category.Description,
			&category.IconURL,
			&category.ParentID,
			&category.SortOrder,
			&category.IsActive,
			&category.CreatedAt,
			&category.UpdatedAt,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, dto.APIResponse{
				Success: false,
				Message: "Failed to scan category",
				Error:   err.Error(),
			})
			return
		}
		categories = append(categories, category)
	}

	pagination := dto.NewPaginationResponse(total, query.Page, query.Limit)

	c.JSON(http.StatusOK, dto.APIResponse{
		Success: true,
		Message: "Categories retrieved successfully",
		Data: dto.CategoryListResponse{
			Categories: categories,
			Pagination: pagination,
		},
	})
}

// GET /api/categories/:id
func (h *CategoryHandler) GetCategory(c *gin.Context) {
	id := c.Param("id")
	
	if _, err := uuid.Parse(id); err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{
			Success: false,
			Message: "Invalid category ID format",
			Error:   err.Error(),
		})
		return
	}

	var category dto.CategoryResponse
	err := h.db.QueryRow(`
		SELECT id, name, slug, description, icon_url, parent_id, sort_order, is_active, created_at, updated_at
		FROM categories WHERE id = $1
	`, id).Scan(
		&category.ID,
		&category.Name,
		&category.Slug,
		&category.Description,
		&category.IconURL,
		&category.ParentID,
		&category.SortOrder,
		&category.IsActive,
		&category.CreatedAt,
		&category.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, dto.APIResponse{
				Success: false,
				Message: "Category not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, dto.APIResponse{
			Success: false,
			Message: "Failed to fetch category",
			Error:   err.Error(),
		})
		return
	}

	// Get children categories
	children, err := h.getChildCategories(category.ID)
	if err != nil {
		// Log error but don't fail the request
		category.Children = []dto.CategoryResponse{}
	} else {
		category.Children = children
	}

	c.JSON(http.StatusOK, dto.APIResponse{
		Success: true,
		Message: "Category retrieved successfully",
		Data:    category,
	})
}

// POST /api/categories
func (h *CategoryHandler) CreateCategory(c *gin.Context) {
	var req dto.CreateCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{
			Success: false,
			Message: "Invalid request body",
			Error:   err.Error(),
		})
		return
	}

	// Validate parent_id if provided
	if req.ParentID != nil {
		if _, err := uuid.Parse(*req.ParentID); err != nil {
			c.JSON(http.StatusBadRequest, dto.APIResponse{
				Success: false,
				Message: "Invalid parent ID format",
				Error:   err.Error(),
			})
			return
		}
	}

	id := uuid.New().String()
	sortOrder := int32(0)
	if req.SortOrder != nil {
		sortOrder = *req.SortOrder
	}

	_, err := h.db.Exec(`
		INSERT INTO categories (id, name, slug, description, icon_url, parent_id, sort_order, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	`, id, req.Name, req.Slug, req.Description, req.IconURL, req.ParentID, sortOrder)

	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{
			Success: false,
			Message: "Failed to create category",
			Error:   err.Error(),
		})
		return
	}

	// Fetch the created category
	var category dto.CategoryResponse
	err = h.db.QueryRow(`
		SELECT id, name, slug, description, icon_url, parent_id, sort_order, is_active, created_at, updated_at
		FROM categories WHERE id = $1
	`, id).Scan(
		&category.ID,
		&category.Name,
		&category.Slug,
		&category.Description,
		&category.IconURL,
		&category.ParentID,
		&category.SortOrder,
		&category.IsActive,
		&category.CreatedAt,
		&category.UpdatedAt,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{
			Success: false,
			Message: "Failed to fetch created category",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, dto.APIResponse{
		Success: true,
		Message: "Category created successfully",
		Data:    category,
	})
}

// PUT /api/categories/:id
func (h *CategoryHandler) UpdateCategory(c *gin.Context) {
	id := c.Param("id")
	
	if _, err := uuid.Parse(id); err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{
			Success: false,
			Message: "Invalid category ID format",
			Error:   err.Error(),
		})
		return
	}

	var req dto.UpdateCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{
			Success: false,
			Message: "Invalid request body",
			Error:   err.Error(),
		})
		return
	}

	// Check if category exists
	var exists bool
	err := h.db.QueryRow("SELECT EXISTS(SELECT 1 FROM categories WHERE id = $1)", id).Scan(&exists)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{
			Success: false,
			Message: "Failed to check category existence",
			Error:   err.Error(),
		})
		return
	}

	if !exists {
		c.JSON(http.StatusNotFound, dto.APIResponse{
			Success: false,
			Message: "Category not found",
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

	if req.IconURL != nil {
		setParts = append(setParts, "icon_url = $"+strconv.Itoa(argIndex))
		args = append(args, *req.IconURL)
		argIndex++
	}

	if req.ParentID != nil {
		setParts = append(setParts, "parent_id = $"+strconv.Itoa(argIndex))
		args = append(args, *req.ParentID)
		argIndex++
	}

	if req.SortOrder != nil {
		setParts = append(setParts, "sort_order = $"+strconv.Itoa(argIndex))
		args = append(args, *req.SortOrder)
		argIndex++
	}

	if req.IsActive != nil {
		setParts = append(setParts, "is_active = $"+strconv.Itoa(argIndex))
		args = append(args, *req.IsActive)
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

	query := "UPDATE categories SET " + 
		setParts[0]
	for i := 1; i < len(setParts); i++ {
		query += ", " + setParts[i]
	}
	query += " WHERE id = " + whereClause

	_, err = h.db.Exec(query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{
			Success: false,
			Message: "Failed to update category",
			Error:   err.Error(),
		})
		return
	}

	// Fetch updated category
	var category dto.CategoryResponse
	err = h.db.QueryRow(`
		SELECT id, name, slug, description, icon_url, parent_id, sort_order, is_active, created_at, updated_at
		FROM categories WHERE id = $1
	`, id).Scan(
		&category.ID,
		&category.Name,
		&category.Slug,
		&category.Description,
		&category.IconURL,
		&category.ParentID,
		&category.SortOrder,
		&category.IsActive,
		&category.CreatedAt,
		&category.UpdatedAt,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{
			Success: false,
			Message: "Failed to fetch updated category",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, dto.APIResponse{
		Success: true,
		Message: "Category updated successfully",
		Data:    category,
	})
}

// DELETE /api/categories/:id
func (h *CategoryHandler) DeleteCategory(c *gin.Context) {
	id := c.Param("id")
	
	if _, err := uuid.Parse(id); err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{
			Success: false,
			Message: "Invalid category ID format",
			Error:   err.Error(),
		})
		return
	}

	// Check if category has children
	var childCount int
	err := h.db.QueryRow("SELECT COUNT(*) FROM categories WHERE parent_id = $1", id).Scan(&childCount)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{
			Success: false,
			Message: "Failed to check child categories",
			Error:   err.Error(),
		})
		return
	}

	if childCount > 0 {
		c.JSON(http.StatusBadRequest, dto.APIResponse{
			Success: false,
			Message: "Cannot delete category with child categories",
		})
		return
	}

	// Check if category is used by courses
	var courseCount int
	err = h.db.QueryRow("SELECT COUNT(*) FROM courses WHERE category_id = $1", id).Scan(&courseCount)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{
			Success: false,
			Message: "Failed to check associated courses",
			Error:   err.Error(),
		})
		return
	}

	if courseCount > 0 {
		c.JSON(http.StatusBadRequest, dto.APIResponse{
			Success: false,
			Message: "Cannot delete category that has associated courses",
		})
		return
	}

	result, err := h.db.Exec("DELETE FROM categories WHERE id = $1", id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{
			Success: false,
			Message: "Failed to delete category",
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
			Message: "Category not found",
		})
		return
	}

	c.JSON(http.StatusOK, dto.APIResponse{
		Success: true,
		Message: "Category deleted successfully",
	})
}

// Helper function to get child categories
func (h *CategoryHandler) getChildCategories(parentID string) ([]dto.CategoryResponse, error) {
	rows, err := h.db.Query(`
		SELECT id, name, slug, description, icon_url, parent_id, sort_order, is_active, created_at, updated_at
		FROM categories 
		WHERE parent_id = $1 
		ORDER BY sort_order ASC, name ASC
	`, parentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var children []dto.CategoryResponse
	for rows.Next() {
		var child dto.CategoryResponse
		err := rows.Scan(
			&child.ID,
			&child.Name,
			&child.Slug,
			&child.Description,
			&child.IconURL,
			&child.ParentID,
			&child.SortOrder,
			&child.IsActive,
			&child.CreatedAt,
			&child.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		children = append(children, child)
	}

	return children, nil
}
