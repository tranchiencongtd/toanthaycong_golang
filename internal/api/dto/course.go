package dto

import (
	"time"
)

// Course DTOs
type CreateCourseRequest struct {
	Title             string   `json:"title" binding:"required"`
	Slug              string   `json:"slug" binding:"required"`
	Description       *string  `json:"description"`
	ShortDescription  *string  `json:"short_description"`
	ThumbnailURL      *string  `json:"thumbnail_url"`
	PreviewVideoURL   *string  `json:"preview_video_url"`
	InstructorID      string   `json:"instructor_id" binding:"required"`
	CategoryID        string   `json:"category_id" binding:"required"`
	Price             float64  `json:"price" binding:"required,min=0"`
	DiscountPrice     *float64 `json:"discount_price" binding:"omitempty,min=0"`
	Language          string   `json:"language" binding:"required"`
	Level             string   `json:"level" binding:"required,oneof=beginner intermediate advanced"`
	Requirements      []string `json:"requirements"`
	WhatYouLearn      []string `json:"what_you_learn"`
	TargetAudience    []string `json:"target_audience"`
}

type UpdateCourseRequest struct {
	Title            *string  `json:"title"`
	Slug             *string  `json:"slug"`
	Description      *string  `json:"description"`
	ShortDescription *string  `json:"short_description"`
	ThumbnailURL     *string  `json:"thumbnail_url"`
	PreviewVideoURL  *string  `json:"preview_video_url"`
	CategoryID       *string  `json:"category_id"`
	Price            *float64 `json:"price" binding:"omitempty,min=0"`
	DiscountPrice    *float64 `json:"discount_price" binding:"omitempty,min=0"`
	Language         *string  `json:"language"`
	Level            *string  `json:"level" binding:"omitempty,oneof=beginner intermediate advanced"`
	Status           *string  `json:"status" binding:"omitempty,oneof=draft pending published archived"`
	Requirements     []string `json:"requirements"`
	WhatYouLearn     []string `json:"what_you_learn"`
	TargetAudience   []string `json:"target_audience"`
}

type CourseResponse struct {
	ID               string    `json:"id"`
	Title            string    `json:"title"`
	Slug             string    `json:"slug"`
	Description      *string   `json:"description"`
	ShortDescription *string   `json:"short_description"`
	ThumbnailURL     *string   `json:"thumbnail_url"`
	PreviewVideoURL  *string   `json:"preview_video_url"`
	InstructorID     string    `json:"instructor_id"`
	CategoryID       string    `json:"category_id"`
	Price            float64   `json:"price"`
	DiscountPrice    *float64  `json:"discount_price"`
	Language         string    `json:"language"`
	Level            string    `json:"level"`
	DurationHours    int32     `json:"duration_hours"`
	TotalLectures    int32     `json:"total_lectures"`
	Status           string    `json:"status"`
	Requirements     []string  `json:"requirements"`
	WhatYouLearn     []string  `json:"what_you_learn"`
	TargetAudience   []string  `json:"target_audience"`
	Rating           float64   `json:"rating"`
	TotalStudents    int32     `json:"total_students"`
	TotalReviews     int32     `json:"total_reviews"`
	PublishedAt      *time.Time `json:"published_at"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

type CourseListResponse struct {
	Courses    []CourseResponse   `json:"courses"`
	Pagination PaginationResponse `json:"pagination"`
}

// Tag DTOs
type CreateTagRequest struct {
	Name        string  `json:"name" binding:"required"`
	Slug        string  `json:"slug" binding:"required"`
	Description *string `json:"description"`
	Color       string  `json:"color" binding:"required"`
}

type UpdateTagRequest struct {
	Name        *string `json:"name"`
	Slug        *string `json:"slug"`
	Description *string `json:"description"`
	Color       *string `json:"color"`
}

type TagResponse struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Slug        string    `json:"slug"`
	Description *string   `json:"description"`
	Color       string    `json:"color"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type TagListResponse struct {
	Tags       []TagResponse      `json:"tags"`
	Pagination PaginationResponse `json:"pagination"`
}
