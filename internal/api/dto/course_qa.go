package dto

import "time"

// CourseQuestionDTO - DTO cho câu hỏi khóa học
type CourseQuestionDTO struct {
	ID         string     `json:"id"`
	CourseID   string     `json:"course_id"`
	LectureID  *string    `json:"lecture_id,omitempty"`
	UserID     string     `json:"user_id"`
	Title      string     `json:"title"`
	Question   string     `json:"question"`
	IsAnswered bool       `json:"is_answered"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
	
	// Thông tin liên quan
	User    *UserDTO             `json:"user,omitempty"`
	Course  *CourseDTO           `json:"course,omitempty"`
	Lecture *CourseLectureDTO    `json:"lecture,omitempty"`
	Answers []CourseAnswerDTO    `json:"answers,omitempty"`
}

// CreateCourseQuestionRequest - Request tạo câu hỏi khóa học
type CreateCourseQuestionRequest struct {
	CourseID  string  `json:"course_id" binding:"required,uuid"`
	LectureID *string `json:"lecture_id,omitempty" binding:"omitempty,uuid"`
	UserID    string  `json:"user_id" binding:"required,uuid"`
	Title     string  `json:"title" binding:"required,max=200"`
	Question  string  `json:"question" binding:"required"`
}

// UpdateCourseQuestionRequest - Request cập nhật câu hỏi khóa học
type UpdateCourseQuestionRequest struct {
	Title      *string `json:"title,omitempty" binding:"omitempty,max=200"`
	Question   *string `json:"question,omitempty"`
	IsAnswered *bool   `json:"is_answered,omitempty"`
}

// CourseQuestionListResponse - Response danh sách câu hỏi khóa học
type CourseQuestionListResponse struct {
	Data       []CourseQuestionDTO `json:"data"`
	Pagination PaginationMeta      `json:"pagination"`
}

// CourseAnswerDTO - DTO cho câu trả lời
type CourseAnswerDTO struct {
	ID                 string    `json:"id"`
	QuestionID         string    `json:"question_id"`
	UserID             string    `json:"user_id"`
	Answer             string    `json:"answer"`
	IsInstructorAnswer bool      `json:"is_instructor_answer"`
	Votes              int       `json:"votes"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
	
	// Thông tin liên quan
	User *UserDTO `json:"user,omitempty"`
}

// CreateCourseAnswerRequest - Request tạo câu trả lời
type CreateCourseAnswerRequest struct {
	QuestionID         string `json:"question_id" binding:"required,uuid"`
	UserID             string `json:"user_id" binding:"required,uuid"`
	Answer             string `json:"answer" binding:"required"`
	IsInstructorAnswer *bool  `json:"is_instructor_answer,omitempty"`
}

// UpdateCourseAnswerRequest - Request cập nhật câu trả lời
type UpdateCourseAnswerRequest struct {
	Answer             *string `json:"answer,omitempty"`
	IsInstructorAnswer *bool   `json:"is_instructor_answer,omitempty"`
	Votes              *int    `json:"votes,omitempty"`
}

// CourseAnswerListResponse - Response danh sách câu trả lời
type CourseAnswerListResponse struct {
	Data       []CourseAnswerDTO `json:"data"`
	Pagination PaginationMeta    `json:"pagination"`
}
