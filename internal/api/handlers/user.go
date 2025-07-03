package handlers

import (
	"database/sql"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
	"internal/api/dto"
)

type UserHandler struct {
	db *sql.DB
}

func NewUserHandler(db *sql.DB) *UserHandler {
	return &UserHandler{db: db}
}

// GET /api/users
func (h *UserHandler) GetUsers(c *gin.Context) {
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

	// Filter by role if provided
	role := c.Query("role")
	var sqlQuery string
	var args []interface{}
	
	baseQuery := `
		SELECT id, email, username, first_name, last_name, avatar_url, bio, role, is_verified, created_at, updated_at
		FROM users 
		WHERE 1=1`
	
	countQuery := "SELECT COUNT(*) FROM users WHERE 1=1"

	if role != "" {
		baseQuery += " AND role = $1"
		countQuery += " AND role = $1"
		args = append(args, role)
	}

	// Get total count
	var total int64
	err := h.db.QueryRow(countQuery, args...).Scan(&total)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{
			Success: false,
			Message: "Failed to count users",
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
			Message: "Failed to fetch users",
			Error:   err.Error(),
		})
		return
	}
	defer rows.Close()

	var users []dto.UserResponse
	for rows.Next() {
		var user dto.UserResponse
		err := rows.Scan(
			&user.ID,
			&user.Email,
			&user.Username,
			&user.FirstName,
			&user.LastName,
			&user.AvatarURL,
			&user.Bio,
			&user.Role,
			&user.IsVerified,
			&user.CreatedAt,
			&user.UpdatedAt,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, dto.APIResponse{
				Success: false,
				Message: "Failed to scan user",
				Error:   err.Error(),
			})
			return
		}
		users = append(users, user)
	}

	pagination := dto.NewPaginationResponse(total, query.Page, query.Limit)

	c.JSON(http.StatusOK, dto.APIResponse{
		Success: true,
		Message: "Users retrieved successfully",
		Data: dto.UserListResponse{
			Users:      users,
			Pagination: pagination,
		},
	})
}

// GET /api/users/:id
func (h *UserHandler) GetUser(c *gin.Context) {
	id := c.Param("id")
	
	if _, err := uuid.Parse(id); err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{
			Success: false,
			Message: "Invalid user ID format",
			Error:   err.Error(),
		})
		return
	}

	var user dto.UserResponse
	err := h.db.QueryRow(`
		SELECT id, email, username, first_name, last_name, avatar_url, bio, role, is_verified, created_at, updated_at
		FROM users WHERE id = $1
	`, id).Scan(
		&user.ID,
		&user.Email,
		&user.Username,
		&user.FirstName,
		&user.LastName,
		&user.AvatarURL,
		&user.Bio,
		&user.Role,
		&user.IsVerified,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, dto.APIResponse{
				Success: false,
				Message: "User not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, dto.APIResponse{
			Success: false,
			Message: "Failed to fetch user",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, dto.APIResponse{
		Success: true,
		Message: "User retrieved successfully",
		Data:    user,
	})
}

// POST /api/users
func (h *UserHandler) CreateUser(c *gin.Context) {
	var req dto.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{
			Success: false,
			Message: "Invalid request body",
			Error:   err.Error(),
		})
		return
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{
			Success: false,
			Message: "Failed to hash password",
			Error:   err.Error(),
		})
		return
	}

	// Set default role if not provided
	role := "student"
	if req.Role != nil {
		role = *req.Role
	}

	id := uuid.New().String()

	_, err = h.db.Exec(`
		INSERT INTO users (id, email, username, password_hash, first_name, last_name, avatar_url, bio, role, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	`, id, req.Email, req.Username, string(hashedPassword), req.FirstName, req.LastName, req.AvatarURL, req.Bio, role)

	if err != nil {
		// Check for unique constraint violations
		if pqErr, ok := err.(*pq.Error); ok {
			switch pqErr.Constraint {
			case "users_email_key":
				c.JSON(http.StatusConflict, dto.APIResponse{
					Success: false,
					Message: "Email already exists",
				})
				return
			case "users_username_key":
				c.JSON(http.StatusConflict, dto.APIResponse{
					Success: false,
					Message: "Username already exists",
				})
				return
			}
		}
		
		c.JSON(http.StatusInternalServerError, dto.APIResponse{
			Success: false,
			Message: "Failed to create user",
			Error:   err.Error(),
		})
		return
	}

	// Fetch the created user (without password)
	var user dto.UserResponse
	err = h.db.QueryRow(`
		SELECT id, email, username, first_name, last_name, avatar_url, bio, role, is_verified, created_at, updated_at
		FROM users WHERE id = $1
	`, id).Scan(
		&user.ID,
		&user.Email,
		&user.Username,
		&user.FirstName,
		&user.LastName,
		&user.AvatarURL,
		&user.Bio,
		&user.Role,
		&user.IsVerified,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{
			Success: false,
			Message: "Failed to fetch created user",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, dto.APIResponse{
		Success: true,
		Message: "User created successfully",
		Data:    user,
	})
}

// PUT /api/users/:id
func (h *UserHandler) UpdateUser(c *gin.Context) {
	id := c.Param("id")
	
	if _, err := uuid.Parse(id); err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{
			Success: false,
			Message: "Invalid user ID format",
			Error:   err.Error(),
		})
		return
	}

	var req dto.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{
			Success: false,
			Message: "Invalid request body",
			Error:   err.Error(),
		})
		return
	}

	// Check if user exists
	var exists bool
	err := h.db.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE id = $1)", id).Scan(&exists)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{
			Success: false,
			Message: "Failed to check user existence",
			Error:   err.Error(),
		})
		return
	}

	if !exists {
		c.JSON(http.StatusNotFound, dto.APIResponse{
			Success: false,
			Message: "User not found",
		})
		return
	}

	// Build dynamic update query
	setParts := []string{"updated_at = CURRENT_TIMESTAMP"}
	args := []interface{}{}
	argIndex := 1

	if req.Email != nil {
		setParts = append(setParts, "email = $"+strconv.Itoa(argIndex))
		args = append(args, *req.Email)
		argIndex++
	}

	if req.Username != nil {
		setParts = append(setParts, "username = $"+strconv.Itoa(argIndex))
		args = append(args, *req.Username)
		argIndex++
	}

	if req.FirstName != nil {
		setParts = append(setParts, "first_name = $"+strconv.Itoa(argIndex))
		args = append(args, *req.FirstName)
		argIndex++
	}

	if req.LastName != nil {
		setParts = append(setParts, "last_name = $"+strconv.Itoa(argIndex))
		args = append(args, *req.LastName)
		argIndex++
	}

	if req.AvatarURL != nil {
		setParts = append(setParts, "avatar_url = $"+strconv.Itoa(argIndex))
		args = append(args, *req.AvatarURL)
		argIndex++
	}

	if req.Bio != nil {
		setParts = append(setParts, "bio = $"+strconv.Itoa(argIndex))
		args = append(args, *req.Bio)
		argIndex++
	}

	if req.Role != nil {
		setParts = append(setParts, "role = $"+strconv.Itoa(argIndex))
		args = append(args, *req.Role)
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

	query := "UPDATE users SET " + setParts[0]
	for i := 1; i < len(setParts); i++ {
		query += ", " + setParts[i]
	}
	query += " WHERE id = " + whereClause

	_, err = h.db.Exec(query, args...)
	if err != nil {
		// Check for unique constraint violations
		if pqErr, ok := err.(*pq.Error); ok {
			switch pqErr.Constraint {
			case "users_email_key":
				c.JSON(http.StatusConflict, dto.APIResponse{
					Success: false,
					Message: "Email already exists",
				})
				return
			case "users_username_key":
				c.JSON(http.StatusConflict, dto.APIResponse{
					Success: false,
					Message: "Username already exists",
				})
				return
			}
		}

		c.JSON(http.StatusInternalServerError, dto.APIResponse{
			Success: false,
			Message: "Failed to update user",
			Error:   err.Error(),
		})
		return
	}

	// Fetch updated user
	var user dto.UserResponse
	err = h.db.QueryRow(`
		SELECT id, email, username, first_name, last_name, avatar_url, bio, role, is_verified, created_at, updated_at
		FROM users WHERE id = $1
	`, id).Scan(
		&user.ID,
		&user.Email,
		&user.Username,
		&user.FirstName,
		&user.LastName,
		&user.AvatarURL,
		&user.Bio,
		&user.Role,
		&user.IsVerified,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{
			Success: false,
			Message: "Failed to fetch updated user",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, dto.APIResponse{
		Success: true,
		Message: "User updated successfully",
		Data:    user,
	})
}

// DELETE /api/users/:id
func (h *UserHandler) DeleteUser(c *gin.Context) {
	id := c.Param("id")
	
	if _, err := uuid.Parse(id); err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{
			Success: false,
			Message: "Invalid user ID format",
			Error:   err.Error(),
		})
		return
	}

	// Check if user has courses (for instructors)
	var courseCount int
	err := h.db.QueryRow("SELECT COUNT(*) FROM courses WHERE instructor_id = $1", id).Scan(&courseCount)
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
			Message: "Cannot delete user who has associated courses",
		})
		return
	}

	result, err := h.db.Exec("DELETE FROM users WHERE id = $1", id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{
			Success: false,
			Message: "Failed to delete user",
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
			Message: "User not found",
		})
		return
	}

	c.JSON(http.StatusOK, dto.APIResponse{
		Success: true,
		Message: "User deleted successfully",
	})
}
