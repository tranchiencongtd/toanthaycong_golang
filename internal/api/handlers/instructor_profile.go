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

type InstructorProfileHandler struct {
	db *sql.DB
}

func NewInstructorProfileHandler(db *sql.DB) *InstructorProfileHandler {
	return &InstructorProfileHandler{db: db}
}

// GET /api/instructor-profiles
func (h *InstructorProfileHandler) GetInstructorProfiles(c *gin.Context) {
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

	// Filter by approval status
	isApproved := c.Query("is_approved")
	var args []interface{}
	baseQuery := `
		SELECT id, user_id, title, expertise, experience_years, rating, total_students,
			   total_courses, total_reviews, website_url, linkedin_url, github_url,
			   is_approved, created_at, updated_at
		FROM instructor_profiles 
		WHERE 1=1`
	
	countQuery := "SELECT COUNT(*) FROM instructor_profiles WHERE 1=1"

	if isApproved != "" {
		baseQuery += " AND is_approved = $1"
		countQuery += " AND is_approved = $1"
		args = append(args, isApproved == "true")
	}

	// Get total count
	var total int64
	err := h.db.QueryRow(countQuery, args...).Scan(&total)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{
			Success: false,
			Message: "Failed to count instructor profiles",
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
			Message: "Failed to fetch instructor profiles",
			Error:   err.Error(),
		})
		return
	}
	defer rows.Close()

	var profiles []dto.InstructorProfileResponse
	for rows.Next() {
		var profile dto.InstructorProfileResponse
		err := rows.Scan(
			&profile.ID,
			&profile.UserID,
			&profile.Title,
			pq.Array(&profile.Expertise),
			&profile.ExperienceYears,
			&profile.Rating,
			&profile.TotalStudents,
			&profile.TotalCourses,
			&profile.TotalReviews,
			&profile.WebsiteURL,
			&profile.LinkedinURL,
			&profile.GithubURL,
			&profile.IsApproved,
			&profile.CreatedAt,
			&profile.UpdatedAt,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, dto.APIResponse{
				Success: false,
				Message: "Failed to scan instructor profile",
				Error:   err.Error(),
			})
			return
		}
		profiles = append(profiles, profile)
	}

	pagination := dto.NewPaginationResponse(total, query.Page, query.Limit)

	c.JSON(http.StatusOK, dto.APIResponse{
		Success: true,
		Message: "Instructor profiles retrieved successfully",
		Data: map[string]interface{}{
			"instructor_profiles": profiles,
			"pagination":          pagination,
		},
	})
}

// GET /api/instructor-profiles/:id
func (h *InstructorProfileHandler) GetInstructorProfile(c *gin.Context) {
	id := c.Param("id")
	
	if _, err := uuid.Parse(id); err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{
			Success: false,
			Message: "Invalid instructor profile ID format",
			Error:   err.Error(),
		})
		return
	}

	var profile dto.InstructorProfileResponse
	err := h.db.QueryRow(`
		SELECT id, user_id, title, expertise, experience_years, rating, total_students,
			   total_courses, total_reviews, website_url, linkedin_url, github_url,
			   is_approved, created_at, updated_at
		FROM instructor_profiles WHERE id = $1
	`, id).Scan(
		&profile.ID,
		&profile.UserID,
		&profile.Title,
		pq.Array(&profile.Expertise),
		&profile.ExperienceYears,
		&profile.Rating,
		&profile.TotalStudents,
		&profile.TotalCourses,
		&profile.TotalReviews,
		&profile.WebsiteURL,
		&profile.LinkedinURL,
		&profile.GithubURL,
		&profile.IsApproved,
		&profile.CreatedAt,
		&profile.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, dto.APIResponse{
				Success: false,
				Message: "Instructor profile not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, dto.APIResponse{
			Success: false,
			Message: "Failed to fetch instructor profile",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, dto.APIResponse{
		Success: true,
		Message: "Instructor profile retrieved successfully",
		Data:    profile,
	})
}

// GET /api/instructor-profiles/user/:user_id
func (h *InstructorProfileHandler) GetInstructorProfileByUserID(c *gin.Context) {
	userID := c.Param("user_id")
	
	if _, err := uuid.Parse(userID); err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{
			Success: false,
			Message: "Invalid user ID format",
			Error:   err.Error(),
		})
		return
	}

	var profile dto.InstructorProfileResponse
	err := h.db.QueryRow(`
		SELECT id, user_id, title, expertise, experience_years, rating, total_students,
			   total_courses, total_reviews, website_url, linkedin_url, github_url,
			   is_approved, created_at, updated_at
		FROM instructor_profiles WHERE user_id = $1
	`, userID).Scan(
		&profile.ID,
		&profile.UserID,
		&profile.Title,
		pq.Array(&profile.Expertise),
		&profile.ExperienceYears,
		&profile.Rating,
		&profile.TotalStudents,
		&profile.TotalCourses,
		&profile.TotalReviews,
		&profile.WebsiteURL,
		&profile.LinkedinURL,
		&profile.GithubURL,
		&profile.IsApproved,
		&profile.CreatedAt,
		&profile.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, dto.APIResponse{
				Success: false,
				Message: "Instructor profile not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, dto.APIResponse{
			Success: false,
			Message: "Failed to fetch instructor profile",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, dto.APIResponse{
		Success: true,
		Message: "Instructor profile retrieved successfully",
		Data:    profile,
	})
}

// POST /api/instructor-profiles
func (h *InstructorProfileHandler) CreateInstructorProfile(c *gin.Context) {
	var req dto.CreateInstructorProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{
			Success: false,
			Message: "Invalid request body",
			Error:   err.Error(),
		})
		return
	}

	// Validate user_id
	if _, err := uuid.Parse(req.UserID); err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{
			Success: false,
			Message: "Invalid user ID format",
			Error:   err.Error(),
		})
		return
	}

	// Check if user exists and is an instructor
	var userRole string
	err := h.db.QueryRow("SELECT role FROM users WHERE id = $1", req.UserID).Scan(&userRole)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusBadRequest, dto.APIResponse{
				Success: false,
				Message: "User not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, dto.APIResponse{
			Success: false,
			Message: "Failed to verify user",
			Error:   err.Error(),
		})
		return
	}

	if userRole != "instructor" && userRole != "admin" {
		c.JSON(http.StatusBadRequest, dto.APIResponse{
			Success: false,
			Message: "User must be an instructor to create profile",
		})
		return
	}

	// Check if profile already exists
	var existingID string
	err = h.db.QueryRow("SELECT id FROM instructor_profiles WHERE user_id = $1", req.UserID).Scan(&existingID)
	if err != sql.ErrNoRows {
		c.JSON(http.StatusConflict, dto.APIResponse{
			Success: false,
			Message: "Instructor profile already exists for this user",
		})
		return
	}

	id := uuid.New().String()
	experienceYears := int32(0)
	if req.ExperienceYears != nil {
		experienceYears = *req.ExperienceYears
	}

	_, err = h.db.Exec(`
		INSERT INTO instructor_profiles (
			id, user_id, title, expertise, experience_years, website_url, 
			linkedin_url, github_url, created_at, updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	`, id, req.UserID, req.Title, pq.Array(req.Expertise), experienceYears,
		req.WebsiteURL, req.LinkedinURL, req.GithubURL)

	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{
			Success: false,
			Message: "Failed to create instructor profile",
			Error:   err.Error(),
		})
		return
	}

	// Fetch the created profile
	var profile dto.InstructorProfileResponse
	err = h.db.QueryRow(`
		SELECT id, user_id, title, expertise, experience_years, rating, total_students,
			   total_courses, total_reviews, website_url, linkedin_url, github_url,
			   is_approved, created_at, updated_at
		FROM instructor_profiles WHERE id = $1
	`, id).Scan(
		&profile.ID,
		&profile.UserID,
		&profile.Title,
		pq.Array(&profile.Expertise),
		&profile.ExperienceYears,
		&profile.Rating,
		&profile.TotalStudents,
		&profile.TotalCourses,
		&profile.TotalReviews,
		&profile.WebsiteURL,
		&profile.LinkedinURL,
		&profile.GithubURL,
		&profile.IsApproved,
		&profile.CreatedAt,
		&profile.UpdatedAt,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{
			Success: false,
			Message: "Failed to fetch created instructor profile",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, dto.APIResponse{
		Success: true,
		Message: "Instructor profile created successfully",
		Data:    profile,
	})
}

// PUT /api/instructor-profiles/:id
func (h *InstructorProfileHandler) UpdateInstructorProfile(c *gin.Context) {
	id := c.Param("id")
	
	if _, err := uuid.Parse(id); err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{
			Success: false,
			Message: "Invalid instructor profile ID format",
			Error:   err.Error(),
		})
		return
	}

	var req dto.UpdateInstructorProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{
			Success: false,
			Message: "Invalid request body",
			Error:   err.Error(),
		})
		return
	}

	// Check if profile exists
	var exists bool
	err := h.db.QueryRow("SELECT EXISTS(SELECT 1 FROM instructor_profiles WHERE id = $1)", id).Scan(&exists)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{
			Success: false,
			Message: "Failed to check instructor profile existence",
			Error:   err.Error(),
		})
		return
	}

	if !exists {
		c.JSON(http.StatusNotFound, dto.APIResponse{
			Success: false,
			Message: "Instructor profile not found",
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

	if req.Expertise != nil {
		setParts = append(setParts, "expertise = $"+strconv.Itoa(argIndex))
		args = append(args, pq.Array(req.Expertise))
		argIndex++
	}

	if req.ExperienceYears != nil {
		setParts = append(setParts, "experience_years = $"+strconv.Itoa(argIndex))
		args = append(args, *req.ExperienceYears)
		argIndex++
	}

	if req.WebsiteURL != nil {
		setParts = append(setParts, "website_url = $"+strconv.Itoa(argIndex))
		args = append(args, *req.WebsiteURL)
		argIndex++
	}

	if req.LinkedinURL != nil {
		setParts = append(setParts, "linkedin_url = $"+strconv.Itoa(argIndex))
		args = append(args, *req.LinkedinURL)
		argIndex++
	}

	if req.GithubURL != nil {
		setParts = append(setParts, "github_url = $"+strconv.Itoa(argIndex))
		args = append(args, *req.GithubURL)
		argIndex++
	}

	if req.IsApproved != nil {
		setParts = append(setParts, "is_approved = $"+strconv.Itoa(argIndex))
		args = append(args, *req.IsApproved)
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

	query := "UPDATE instructor_profiles SET " + setParts[0]
	for i := 1; i < len(setParts); i++ {
		query += ", " + setParts[i]
	}
	query += " WHERE id = " + whereClause

	_, err = h.db.Exec(query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{
			Success: false,
			Message: "Failed to update instructor profile",
			Error:   err.Error(),
		})
		return
	}

	// Fetch updated profile
	var profile dto.InstructorProfileResponse
	err = h.db.QueryRow(`
		SELECT id, user_id, title, expertise, experience_years, rating, total_students,
			   total_courses, total_reviews, website_url, linkedin_url, github_url,
			   is_approved, created_at, updated_at
		FROM instructor_profiles WHERE id = $1
	`, id).Scan(
		&profile.ID,
		&profile.UserID,
		&profile.Title,
		pq.Array(&profile.Expertise),
		&profile.ExperienceYears,
		&profile.Rating,
		&profile.TotalStudents,
		&profile.TotalCourses,
		&profile.TotalReviews,
		&profile.WebsiteURL,
		&profile.LinkedinURL,
		&profile.GithubURL,
		&profile.IsApproved,
		&profile.CreatedAt,
		&profile.UpdatedAt,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{
			Success: false,
			Message: "Failed to fetch updated instructor profile",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, dto.APIResponse{
		Success: true,
		Message: "Instructor profile updated successfully",
		Data:    profile,
	})
}

// DELETE /api/instructor-profiles/:id
func (h *InstructorProfileHandler) DeleteInstructorProfile(c *gin.Context) {
	id := c.Param("id")
	
	if _, err := uuid.Parse(id); err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{
			Success: false,
			Message: "Invalid instructor profile ID format",
			Error:   err.Error(),
		})
		return
	}

	result, err := h.db.Exec("DELETE FROM instructor_profiles WHERE id = $1", id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{
			Success: false,
			Message: "Failed to delete instructor profile",
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
			Message: "Instructor profile not found",
		})
		return
	}

	c.JSON(http.StatusOK, dto.APIResponse{
		Success: true,
		Message: "Instructor profile deleted successfully",
	})
}
