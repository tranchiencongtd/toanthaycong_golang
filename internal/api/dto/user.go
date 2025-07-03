package dto

import (
	"time"
)

// User DTOs
type CreateUserRequest struct {
	Email     string  `json:"email" binding:"required,email"`
	Username  string  `json:"username" binding:"required"`
	Password  string  `json:"password" binding:"required,min=6"`
	FirstName string  `json:"first_name" binding:"required"`
	LastName  string  `json:"last_name" binding:"required"`
	Role      *string `json:"role"`
	Bio       *string `json:"bio"`
	AvatarURL *string `json:"avatar_url"`
}

type UpdateUserRequest struct {
	Email     *string `json:"email" binding:"omitempty,email"`
	Username  *string `json:"username"`
	FirstName *string `json:"first_name"`
	LastName  *string `json:"last_name"`
	Role      *string `json:"role"`
	Bio       *string `json:"bio"`
	AvatarURL *string `json:"avatar_url"`
}

type UserResponse struct {
	ID         string    `json:"id"`
	Email      string    `json:"email"`
	Username   string    `json:"username"`
	FirstName  string    `json:"first_name"`
	LastName   string    `json:"last_name"`
	Role       string    `json:"role"`
	Bio        *string   `json:"bio"`
	AvatarURL  *string   `json:"avatar_url"`
	IsVerified bool      `json:"is_verified"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type UserListResponse struct {
	Users      []UserResponse     `json:"users"`
	Pagination PaginationResponse `json:"pagination"`
}

// Instructor Profile DTOs
type CreateInstructorProfileRequest struct {
	UserID          string   `json:"user_id" binding:"required"`
	Title           *string  `json:"title"`
	Expertise       []string `json:"expertise"`
	ExperienceYears *int32   `json:"experience_years"`
	WebsiteURL      *string  `json:"website_url"`
	LinkedinURL     *string  `json:"linkedin_url"`
	GithubURL       *string  `json:"github_url"`
}

type UpdateInstructorProfileRequest struct {
	Title           *string  `json:"title"`
	Expertise       []string `json:"expertise"`
	ExperienceYears *int32   `json:"experience_years"`
	WebsiteURL      *string  `json:"website_url"`
	LinkedinURL     *string  `json:"linkedin_url"`
	GithubURL       *string  `json:"github_url"`
	IsApproved      *bool    `json:"is_approved"`
}

type InstructorProfileResponse struct {
	ID              string    `json:"id"`
	UserID          string    `json:"user_id"`
	Title           *string   `json:"title"`
	Expertise       []string  `json:"expertise"`
	ExperienceYears int32     `json:"experience_years"`
	Rating          float64   `json:"rating"`
	TotalStudents   int32     `json:"total_students"`
	TotalCourses    int32     `json:"total_courses"`
	TotalReviews    int32     `json:"total_reviews"`
	WebsiteURL      *string   `json:"website_url"`
	LinkedinURL     *string   `json:"linkedin_url"`
	GithubURL       *string   `json:"github_url"`
	IsApproved      bool      `json:"is_approved"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}
