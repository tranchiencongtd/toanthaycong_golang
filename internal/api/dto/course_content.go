package dto

import (
	"time"
)

// Course Section DTOs
type CreateCourseSectionRequest struct {
	CourseID    string  `json:"course_id" binding:"required"`
	Title       string  `json:"title" binding:"required"`
	Description *string `json:"description"`
	SortOrder   int32   `json:"sort_order" binding:"required"`
}

type UpdateCourseSectionRequest struct {
	Title       *string `json:"title"`
	Description *string `json:"description"`
	SortOrder   *int32  `json:"sort_order"`
}

type CourseSectionResponse struct {
	ID          string                  `json:"id"`
	CourseID    string                  `json:"course_id"`
	Title       string                  `json:"title"`
	Description *string                 `json:"description"`
	SortOrder   int32                   `json:"sort_order"`
	CreatedAt   time.Time               `json:"created_at"`
	UpdatedAt   time.Time               `json:"updated_at"`
	Lectures    []CourseLectureResponse `json:"lectures,omitempty"`
}

type CourseSectionListResponse struct {
	Sections   []CourseSectionResponse `json:"sections"`
	Pagination PaginationResponse      `json:"pagination"`
}

// Course Lecture DTOs
type CreateCourseLectureRequest struct {
	SectionID       string  `json:"section_id" binding:"required"`
	Title           string  `json:"title" binding:"required"`
	Description     *string `json:"description"`
	ContentType     string  `json:"content_type" binding:"required,oneof=video article quiz file"`
	VideoURL        *string `json:"video_url"`
	VideoDuration   *int32  `json:"video_duration"`
	ArticleContent  *string `json:"article_content"`
	FileURL         *string `json:"file_url"`
	SortOrder       int32   `json:"sort_order" binding:"required"`
	IsPreview       *bool   `json:"is_preview"`
	IsDownloadable  *bool   `json:"is_downloadable"`
}

type UpdateCourseLectureRequest struct {
	Title          *string `json:"title"`
	Description    *string `json:"description"`
	ContentType    *string `json:"content_type" binding:"omitempty,oneof=video article quiz file"`
	VideoURL       *string `json:"video_url"`
	VideoDuration  *int32  `json:"video_duration"`
	ArticleContent *string `json:"article_content"`
	FileURL        *string `json:"file_url"`
	SortOrder      *int32  `json:"sort_order"`
	IsPreview      *bool   `json:"is_preview"`
	IsDownloadable *bool   `json:"is_downloadable"`
}

type CourseLectureResponse struct {
	ID             string    `json:"id"`
	SectionID      string    `json:"section_id"`
	Title          string    `json:"title"`
	Description    *string   `json:"description"`
	ContentType    string    `json:"content_type"`
	VideoURL       *string   `json:"video_url"`
	VideoDuration  *int32    `json:"video_duration"`
	ArticleContent *string   `json:"article_content"`
	FileURL        *string   `json:"file_url"`
	SortOrder      int32     `json:"sort_order"`
	IsPreview      bool      `json:"is_preview"`
	IsDownloadable bool      `json:"is_downloadable"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type CourseLectureListResponse struct {
	Lectures   []CourseLectureResponse `json:"lectures"`
	Pagination PaginationResponse      `json:"pagination"`
}
