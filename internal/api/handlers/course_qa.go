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

type CourseQAHandler struct {
	db *sql.DB
}

func NewCourseQAHandler(db *sql.DB) *CourseQAHandler {
	return &CourseQAHandler{db: db}
}

// GetCourseQuestions godoc
// @Summary Lấy danh sách câu hỏi khóa học
// @Description Lấy danh sách câu hỏi khóa học với phân trang và lọc
// @Tags CourseQA
// @Accept json
// @Produce json
// @Param page query int false "Số trang" default(1)
// @Param limit query int false "Số item mỗi trang" default(10)
// @Param course_id query string false "Lọc theo course ID"
// @Param user_id query string false "Lọc theo user ID"
// @Param answered query bool false "Lọc theo trạng thái đã trả lời"
// @Success 200 {object} dto.CourseQuestionListResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/course-questions [get]
func (h *CourseQAHandler) GetCourseQuestions(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	courseID := c.Query("course_id")
	userID := c.Query("user_id")
	answered := c.Query("answered")

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	offset := (page - 1) * limit

	// Build query với filters
	query := `
		SELECT cq.id, cq.course_id, cq.lecture_id, cq.user_id, cq.title,
		       cq.question, cq.is_answered, cq.created_at, cq.updated_at,
		       COUNT(*) OVER() as total_count
		FROM course_questions cq
		WHERE 1=1`
	
	args := []interface{}{}
	argIndex := 1

	if courseID != "" {
		query += fmt.Sprintf(" AND cq.course_id = $%d", argIndex)
		args = append(args, courseID)
		argIndex++
	}

	if userID != "" {
		query += fmt.Sprintf(" AND cq.user_id = $%d", argIndex)
		args = append(args, userID)
		argIndex++
	}

	if answered != "" {
		if answered == "true" {
			query += fmt.Sprintf(" AND cq.is_answered = $%d", argIndex)
			args = append(args, true)
			argIndex++
		} else if answered == "false" {
			query += fmt.Sprintf(" AND cq.is_answered = $%d", argIndex)
			args = append(args, false)
			argIndex++
		}
	}

	query += fmt.Sprintf(" ORDER BY cq.created_at DESC LIMIT $%d OFFSET $%d", argIndex, argIndex+1)
	args = append(args, limit, offset)

	rows, err := h.db.Query(query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "Database error",
			Message: "Failed to fetch course questions",
		})
		return
	}
	defer rows.Close()

	var questions []dto.CourseQuestionDTO
	var totalCount int

	for rows.Next() {
		var question dto.CourseQuestionDTO
		var lectureID sql.NullString

		err := rows.Scan(
			&question.ID, &question.CourseID, &lectureID, &question.UserID,
			&question.Title, &question.Question, &question.IsAnswered,
			&question.CreatedAt, &question.UpdatedAt, &totalCount,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
				Error:   "Scan error",
				Message: "Failed to parse course question data",
			})
			return
		}

		if lectureID.Valid {
			question.LectureID = &lectureID.String
		}

		questions = append(questions, question)
	}

	// Tính toán pagination
	totalPages := (totalCount + limit - 1) / limit

	response := dto.CourseQuestionListResponse{
		Data: questions,
		Pagination: dto.PaginationMeta{
			Page:       page,
			Limit:      limit,
			TotalItems: totalCount,
			TotalPages: totalPages,
		},
	}

	c.JSON(http.StatusOK, response)
}

// GetCourseQuestion godoc
// @Summary Lấy thông tin câu hỏi khóa học theo ID
// @Description Lấy thông tin chi tiết câu hỏi khóa học theo ID
// @Tags CourseQA
// @Accept json
// @Produce json
// @Param id path string true "ID câu hỏi khóa học"
// @Success 200 {object} dto.CourseQuestionDTO
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/course-questions/{id} [get]
func (h *CourseQAHandler) GetCourseQuestion(c *gin.Context) {
	id := c.Param("id")

	if _, err := uuid.Parse(id); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "Invalid ID",
			Message: "Invalid course question ID format",
		})
		return
	}

	query := `
		SELECT id, course_id, lecture_id, user_id, title, question, is_answered, created_at, updated_at
		FROM course_questions 
		WHERE id = $1`

	var question dto.CourseQuestionDTO
	var lectureID sql.NullString

	err := h.db.QueryRow(query, id).Scan(
		&question.ID, &question.CourseID, &lectureID, &question.UserID,
		&question.Title, &question.Question, &question.IsAnswered,
		&question.CreatedAt, &question.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, dto.ErrorResponse{
				Error:   "Not found",
				Message: "Course question not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "Database error",
			Message: "Failed to fetch course question",
		})
		return
	}

	if lectureID.Valid {
		question.LectureID = &lectureID.String
	}

	c.JSON(http.StatusOK, question)
}

// CreateCourseQuestion godoc
// @Summary Tạo câu hỏi khóa học mới
// @Description Tạo câu hỏi khóa học mới
// @Tags CourseQA
// @Accept json
// @Produce json
// @Param body body dto.CreateCourseQuestionRequest true "Thông tin câu hỏi khóa học"
// @Success 201 {object} dto.CourseQuestionDTO
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/course-questions [post]
func (h *CourseQAHandler) CreateCourseQuestion(c *gin.Context) {
	var req dto.CreateCourseQuestionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "Invalid input",
			Message: err.Error(),
		})
		return
	}

	// Kiểm tra user và course tồn tại
	var userExists, courseExists bool
	h.db.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE id = $1)", req.UserID).Scan(&userExists)
	h.db.QueryRow("SELECT EXISTS(SELECT 1 FROM courses WHERE id = $1)", req.CourseID).Scan(&courseExists)

	if !userExists {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "Invalid user",
			Message: "User not found",
		})
		return
	}

	if !courseExists {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "Invalid course",
			Message: "Course not found",
		})
		return
	}

	// Kiểm tra lecture nếu có
	if req.LectureID != nil {
		var lectureExists bool
		h.db.QueryRow("SELECT EXISTS(SELECT 1 FROM course_lectures WHERE id = $1)", *req.LectureID).Scan(&lectureExists)
		if !lectureExists {
			c.JSON(http.StatusBadRequest, dto.ErrorResponse{
				Error:   "Invalid lecture",
				Message: "Lecture not found",
			})
			return
		}
	}

	id := uuid.New().String()
	now := time.Now()

	var lectureID sql.NullString
	if req.LectureID != nil {
		lectureID = sql.NullString{String: *req.LectureID, Valid: true}
	}

	query := `
		INSERT INTO course_questions (id, course_id, lecture_id, user_id, title, question, is_answered, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, course_id, lecture_id, user_id, title, question, is_answered, created_at, updated_at`

	var question dto.CourseQuestionDTO

	err := h.db.QueryRow(query, id, req.CourseID, lectureID, req.UserID, req.Title, req.Question, false, now, now).Scan(
		&question.ID, &question.CourseID, &lectureID, &question.UserID,
		&question.Title, &question.Question, &question.IsAnswered,
		&question.CreatedAt, &question.UpdatedAt,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "Database error",
			Message: "Failed to create course question",
		})
		return
	}

	if lectureID.Valid {
		question.LectureID = &lectureID.String
	}

	c.JSON(http.StatusCreated, question)
}

// UpdateCourseQuestion godoc
// @Summary Cập nhật câu hỏi khóa học
// @Description Cập nhật thông tin câu hỏi khóa học
// @Tags CourseQA
// @Accept json
// @Produce json
// @Param id path string true "ID câu hỏi khóa học"
// @Param body body dto.UpdateCourseQuestionRequest true "Thông tin cập nhật"
// @Success 200 {object} dto.CourseQuestionDTO
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/course-questions/{id} [put]
func (h *CourseQAHandler) UpdateCourseQuestion(c *gin.Context) {
	id := c.Param("id")

	if _, err := uuid.Parse(id); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "Invalid ID",
			Message: "Invalid course question ID format",
		})
		return
	}

	var req dto.UpdateCourseQuestionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "Invalid input",
			Message: err.Error(),
		})
		return
	}

	// Kiểm tra question tồn tại
	var exists bool
	err := h.db.QueryRow("SELECT EXISTS(SELECT 1 FROM course_questions WHERE id = $1)", id).Scan(&exists)
	if err != nil || !exists {
		c.JSON(http.StatusNotFound, dto.ErrorResponse{
			Error:   "Not found",
			Message: "Course question not found",
		})
		return
	}

	// Build update query
	setParts := []string{}
	args := []interface{}{}
	argIndex := 1

	if req.Title != nil {
		setParts = append(setParts, fmt.Sprintf("title = $%d", argIndex))
		args = append(args, *req.Title)
		argIndex++
	}

	if req.Question != nil {
		setParts = append(setParts, fmt.Sprintf("question = $%d", argIndex))
		args = append(args, *req.Question)
		argIndex++
	}

	if req.IsAnswered != nil {
		setParts = append(setParts, fmt.Sprintf("is_answered = $%d", argIndex))
		args = append(args, *req.IsAnswered)
		argIndex++
	}

	if len(setParts) == 0 {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "No updates",
			Message: "No fields to update",
		})
		return
	}

	setParts = append(setParts, fmt.Sprintf("updated_at = $%d", argIndex))
	args = append(args, time.Now())
	argIndex++

	query := fmt.Sprintf(`
		UPDATE course_questions SET %s 
		WHERE id = $%d
		RETURNING id, course_id, lecture_id, user_id, title, question, is_answered, created_at, updated_at`,
		fmt.Sprintf("%s", setParts[0]),
		argIndex,
	)

	for i := 1; i < len(setParts); i++ {
		query = fmt.Sprintf(`
			UPDATE course_questions SET %s, %s 
			WHERE id = $%d
			RETURNING id, course_id, lecture_id, user_id, title, question, is_answered, created_at, updated_at`,
			setParts[0], setParts[i],
			argIndex,
		)
	}

	args = append(args, id)

	var question dto.CourseQuestionDTO
	var lectureID sql.NullString

	err = h.db.QueryRow(query, args...).Scan(
		&question.ID, &question.CourseID, &lectureID, &question.UserID,
		&question.Title, &question.Question, &question.IsAnswered,
		&question.CreatedAt, &question.UpdatedAt,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "Database error",
			Message: "Failed to update course question",
		})
		return
	}

	if lectureID.Valid {
		question.LectureID = &lectureID.String
	}

	c.JSON(http.StatusOK, question)
}

// DeleteCourseQuestion godoc
// @Summary Xóa câu hỏi khóa học
// @Description Xóa câu hỏi khóa học theo ID
// @Tags CourseQA
// @Accept json
// @Produce json
// @Param id path string true "ID câu hỏi khóa học"
// @Success 204 "No Content"
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/course-questions/{id} [delete]
func (h *CourseQAHandler) DeleteCourseQuestion(c *gin.Context) {
	id := c.Param("id")

	if _, err := uuid.Parse(id); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "Invalid ID",
			Message: "Invalid course question ID format",
		})
		return
	}

	result, err := h.db.Exec("DELETE FROM course_questions WHERE id = $1", id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "Database error",
			Message: "Failed to delete course question",
		})
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, dto.ErrorResponse{
			Error:   "Not found",
			Message: "Course question not found",
		})
		return
	}

	c.Status(http.StatusNoContent)
}

// === COURSE ANSWERS ===

// GetCourseAnswers godoc
// @Summary Lấy danh sách câu trả lời
// @Description Lấy danh sách câu trả lời cho câu hỏi với phân trang
// @Tags CourseQA
// @Accept json
// @Produce json
// @Param question_id path string true "ID câu hỏi"
// @Param page query int false "Số trang" default(1)
// @Param limit query int false "Số item mỗi trang" default(10)
// @Success 200 {object} dto.CourseAnswerListResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/course-questions/{question_id}/answers [get]
func (h *CourseQAHandler) GetCourseAnswers(c *gin.Context) {
	questionID := c.Param("question_id")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	if _, err := uuid.Parse(questionID); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "Invalid ID",
			Message: "Invalid question ID format",
		})
		return
	}

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	offset := (page - 1) * limit

	query := `
		SELECT ca.id, ca.question_id, ca.user_id, ca.answer, ca.is_instructor_answer,
		       ca.votes, ca.created_at, ca.updated_at,
		       COUNT(*) OVER() as total_count
		FROM course_answers ca
		WHERE ca.question_id = $1
		ORDER BY ca.is_instructor_answer DESC, ca.votes DESC, ca.created_at ASC
		LIMIT $2 OFFSET $3`

	rows, err := h.db.Query(query, questionID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "Database error",
			Message: "Failed to fetch course answers",
		})
		return
	}
	defer rows.Close()

	var answers []dto.CourseAnswerDTO
	var totalCount int

	for rows.Next() {
		var answer dto.CourseAnswerDTO

		err := rows.Scan(
			&answer.ID, &answer.QuestionID, &answer.UserID, &answer.Answer,
			&answer.IsInstructorAnswer, &answer.Votes,
			&answer.CreatedAt, &answer.UpdatedAt, &totalCount,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
				Error:   "Scan error",
				Message: "Failed to parse course answer data",
			})
			return
		}

		answers = append(answers, answer)
	}

	// Tính toán pagination
	totalPages := (totalCount + limit - 1) / limit

	response := dto.CourseAnswerListResponse{
		Data: answers,
		Pagination: dto.PaginationMeta{
			Page:       page,
			Limit:      limit,
			TotalItems: totalCount,
			TotalPages: totalPages,
		},
	}

	c.JSON(http.StatusOK, response)
}

// CreateCourseAnswer godoc
// @Summary Tạo câu trả lời mới
// @Description Tạo câu trả lời mới cho câu hỏi
// @Tags CourseQA
// @Accept json
// @Produce json
// @Param body body dto.CreateCourseAnswerRequest true "Thông tin câu trả lời"
// @Success 201 {object} dto.CourseAnswerDTO
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/course-answers [post]
func (h *CourseQAHandler) CreateCourseAnswer(c *gin.Context) {
	var req dto.CreateCourseAnswerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "Invalid input",
			Message: err.Error(),
		})
		return
	}

	// Kiểm tra user và question tồn tại
	var userExists, questionExists bool
	h.db.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE id = $1)", req.UserID).Scan(&userExists)
	h.db.QueryRow("SELECT EXISTS(SELECT 1 FROM course_questions WHERE id = $1)", req.QuestionID).Scan(&questionExists)

	if !userExists {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "Invalid user",
			Message: "User not found",
		})
		return
	}

	if !questionExists {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "Invalid question",
			Message: "Question not found",
		})
		return
	}

	// Set default values
	isInstructorAnswer := false
	if req.IsInstructorAnswer != nil {
		isInstructorAnswer = *req.IsInstructorAnswer
	}

	id := uuid.New().String()
	now := time.Now()

	query := `
		INSERT INTO course_answers (id, question_id, user_id, answer, is_instructor_answer, votes, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, question_id, user_id, answer, is_instructor_answer, votes, created_at, updated_at`

	var answer dto.CourseAnswerDTO

	err := h.db.QueryRow(query, id, req.QuestionID, req.UserID, req.Answer, isInstructorAnswer, 0, now, now).Scan(
		&answer.ID, &answer.QuestionID, &answer.UserID, &answer.Answer,
		&answer.IsInstructorAnswer, &answer.Votes,
		&answer.CreatedAt, &answer.UpdatedAt,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "Database error",
			Message: "Failed to create course answer",
		})
		return
	}

	// Cập nhật trạng thái answered cho question
	h.db.Exec("UPDATE course_questions SET is_answered = true, updated_at = $1 WHERE id = $2", now, req.QuestionID)

	c.JSON(http.StatusCreated, answer)
}

// UpdateCourseAnswer godoc
// @Summary Cập nhật câu trả lời
// @Description Cập nhật thông tin câu trả lời
// @Tags CourseQA
// @Accept json
// @Produce json
// @Param id path string true "ID câu trả lời"
// @Param body body dto.UpdateCourseAnswerRequest true "Thông tin cập nhật"
// @Success 200 {object} dto.CourseAnswerDTO
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/course-answers/{id} [put]
func (h *CourseQAHandler) UpdateCourseAnswer(c *gin.Context) {
	id := c.Param("id")

	if _, err := uuid.Parse(id); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "Invalid ID",
			Message: "Invalid course answer ID format",
		})
		return
	}

	var req dto.UpdateCourseAnswerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "Invalid input",
			Message: err.Error(),
		})
		return
	}

	// Kiểm tra answer tồn tại
	var exists bool
	err := h.db.QueryRow("SELECT EXISTS(SELECT 1 FROM course_answers WHERE id = $1)", id).Scan(&exists)
	if err != nil || !exists {
		c.JSON(http.StatusNotFound, dto.ErrorResponse{
			Error:   "Not found",
			Message: "Course answer not found",
		})
		return
	}

	// Build update query
	setParts := []string{}
	args := []interface{}{}
	argIndex := 1

	if req.Answer != nil {
		setParts = append(setParts, fmt.Sprintf("answer = $%d", argIndex))
		args = append(args, *req.Answer)
		argIndex++
	}

	if req.IsInstructorAnswer != nil {
		setParts = append(setParts, fmt.Sprintf("is_instructor_answer = $%d", argIndex))
		args = append(args, *req.IsInstructorAnswer)
		argIndex++
	}

	if req.Votes != nil {
		setParts = append(setParts, fmt.Sprintf("votes = $%d", argIndex))
		args = append(args, *req.Votes)
		argIndex++
	}

	if len(setParts) == 0 {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "No updates",
			Message: "No fields to update",
		})
		return
	}

	setParts = append(setParts, fmt.Sprintf("updated_at = $%d", argIndex))
	args = append(args, time.Now())
	argIndex++

	query := fmt.Sprintf(`
		UPDATE course_answers SET %s 
		WHERE id = $%d
		RETURNING id, question_id, user_id, answer, is_instructor_answer, votes, created_at, updated_at`,
		fmt.Sprintf("%s", setParts[0]),
		argIndex,
	)

	for i := 1; i < len(setParts); i++ {
		query = fmt.Sprintf(`
			UPDATE course_answers SET %s, %s 
			WHERE id = $%d
			RETURNING id, question_id, user_id, answer, is_instructor_answer, votes, created_at, updated_at`,
			setParts[0], setParts[i],
			argIndex,
		)
	}

	args = append(args, id)

	var answer dto.CourseAnswerDTO

	err = h.db.QueryRow(query, args...).Scan(
		&answer.ID, &answer.QuestionID, &answer.UserID, &answer.Answer,
		&answer.IsInstructorAnswer, &answer.Votes,
		&answer.CreatedAt, &answer.UpdatedAt,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "Database error",
			Message: "Failed to update course answer",
		})
		return
	}

	c.JSON(http.StatusOK, answer)
}

// DeleteCourseAnswer godoc
// @Summary Xóa câu trả lời
// @Description Xóa câu trả lời theo ID
// @Tags CourseQA
// @Accept json
// @Produce json
// @Param id path string true "ID câu trả lời"
// @Success 204 "No Content"
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/course-answers/{id} [delete]
func (h *CourseQAHandler) DeleteCourseAnswer(c *gin.Context) {
	id := c.Param("id")

	if _, err := uuid.Parse(id); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "Invalid ID",
			Message: "Invalid course answer ID format",
		})
		return
	}

	// Lấy question_id trước khi xóa để kiểm tra còn answer nào khác không
	var questionID string
	err := h.db.QueryRow("SELECT question_id FROM course_answers WHERE id = $1", id).Scan(&questionID)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, dto.ErrorResponse{
				Error:   "Not found",
				Message: "Course answer not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "Database error",
			Message: "Failed to fetch course answer",
		})
		return
	}

	result, err := h.db.Exec("DELETE FROM course_answers WHERE id = $1", id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "Database error",
			Message: "Failed to delete course answer",
		})
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, dto.ErrorResponse{
			Error:   "Not found",
			Message: "Course answer not found",
		})
		return
	}

	// Kiểm tra xem còn answer nào khác không, nếu không thì cập nhật is_answered = false
	var hasOtherAnswers bool
	h.db.QueryRow("SELECT EXISTS(SELECT 1 FROM course_answers WHERE question_id = $1)", questionID).Scan(&hasOtherAnswers)
	if !hasOtherAnswers {
		h.db.Exec("UPDATE course_questions SET is_answered = false, updated_at = $1 WHERE id = $2", time.Now(), questionID)
	}

	c.Status(http.StatusNoContent)
}
