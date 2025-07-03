package routes

import (
	"database/sql"

	"github.com/gin-gonic/gin"
	"internal/api/handlers"
	"internal/api/middleware"
)

func SetupRoutes(db *sql.DB) *gin.Engine {
	// Create Gin router
	r := gin.New()

	// Add middleware
	r.Use(middleware.StructuredLogger())
	r.Use(gin.Recovery())
	r.Use(middleware.CORS())
	r.Use(middleware.JSONMiddleware())

	// Initialize handlers
	categoryHandler := handlers.NewCategoryHandler(db)
	userHandler := handlers.NewUserHandler(db)
	courseHandler := handlers.NewCourseHandler(db)
	tagHandler := handlers.NewTagHandler(db)
	instructorProfileHandler := handlers.NewInstructorProfileHandler(db)
	courseSectionHandler := handlers.NewCourseSectionHandler(db)
	courseLectureHandler := handlers.NewCourseLectureHandler(db)
	enrollmentHandler := handlers.NewEnrollmentHandler(db)
	lectureProgressHandler := handlers.NewLectureProgressHandler(db)
	courseReviewHandler := handlers.NewCourseReviewHandler(db)
	wishlistHandler := handlers.NewWishlistHandler(db)
	couponHandler := handlers.NewCouponHandler(db)
	courseAnnouncementHandler := handlers.NewCourseAnnouncementHandler(db)
	courseQAHandler := handlers.NewCourseQAHandler(db)
	notificationHandler := handlers.NewNotificationHandler(db)

	// API routes
	api := r.Group("/api/v1")
	{
		// Health check
		api.GET("/health", func(c *gin.Context) {
			c.JSON(200, gin.H{
				"status":  "ok",
				"message": "API is running",
			})
		})

		// Categories routes
		categories := api.Group("/categories")
		{
			categories.GET("", categoryHandler.GetCategories)
			categories.GET("/:id", categoryHandler.GetCategory)
			categories.POST("", categoryHandler.CreateCategory)
			categories.PUT("/:id", categoryHandler.UpdateCategory)
			categories.DELETE("/:id", categoryHandler.DeleteCategory)
		}

		// Users routes
		users := api.Group("/users")
		{
			users.GET("", userHandler.GetUsers)
			users.GET("/:id", userHandler.GetUser)
			users.POST("", userHandler.CreateUser)
			users.PUT("/:id", userHandler.UpdateUser)
			users.DELETE("/:id", userHandler.DeleteUser)
			
			// User notification stats
			users.GET("/:user_id/notification-stats", notificationHandler.GetNotificationStats)
		}

		// Courses routes
		courses := api.Group("/courses")
		{
			courses.GET("", courseHandler.GetCourses)
			courses.GET("/:id", courseHandler.GetCourse)
			courses.POST("", courseHandler.CreateCourse)
			courses.PUT("/:id", courseHandler.UpdateCourse)
			courses.DELETE("/:id", courseHandler.DeleteCourse)
			
			// Course tags
			courses.GET("/:course_id/tags", tagHandler.GetCourseTags)
			
			// Course review stats
			courses.GET("/:course_id/review-stats", courseReviewHandler.GetCourseReviewStats)
		}

		// Tags routes
		tags := api.Group("/tags")
		{
			tags.GET("", tagHandler.GetTags)
			tags.GET("/:id", tagHandler.GetTag)
			tags.POST("", tagHandler.CreateTag)
			tags.PUT("/:id", tagHandler.UpdateTag)
			tags.DELETE("/:id", tagHandler.DeleteTag)
		}

		// Instructor Profiles routes
		instructorProfiles := api.Group("/instructor-profiles")
		{
			instructorProfiles.GET("", instructorProfileHandler.GetInstructorProfiles)
			instructorProfiles.GET("/:id", instructorProfileHandler.GetInstructorProfile)
			instructorProfiles.POST("", instructorProfileHandler.CreateInstructorProfile)
			instructorProfiles.PUT("/:id", instructorProfileHandler.UpdateInstructorProfile)
			instructorProfiles.DELETE("/:id", instructorProfileHandler.DeleteInstructorProfile)
		}

		// Course Sections routes
		courseSections := api.Group("/course-sections")
		{
			courseSections.GET("", courseSectionHandler.GetCourseSections)
			courseSections.GET("/:id", courseSectionHandler.GetCourseSection)
			courseSections.POST("", courseSectionHandler.CreateCourseSection)
			courseSections.PUT("/:id", courseSectionHandler.UpdateCourseSection)
			courseSections.DELETE("/:id", courseSectionHandler.DeleteCourseSection)
		}

		// Course Lectures routes
		courseLectures := api.Group("/course-lectures")
		{
			courseLectures.GET("", courseLectureHandler.GetCourseLectures)
			courseLectures.GET("/:id", courseLectureHandler.GetCourseLecture)
			courseLectures.POST("", courseLectureHandler.CreateCourseLecture)
			courseLectures.PUT("/:id", courseLectureHandler.UpdateCourseLecture)
			courseLectures.DELETE("/:id", courseLectureHandler.DeleteCourseLecture)
		}

		// Enrollments routes
		enrollments := api.Group("/enrollments")
		{
			enrollments.GET("", enrollmentHandler.GetEnrollments)
			enrollments.GET("/:id", enrollmentHandler.GetEnrollment)
			enrollments.POST("", enrollmentHandler.CreateEnrollment)
			enrollments.PUT("/:id", enrollmentHandler.UpdateEnrollment)
			enrollments.DELETE("/:id", enrollmentHandler.DeleteEnrollment)
		}

		// Lecture Progress routes
		lectureProgress := api.Group("/lecture-progress")
		{
			lectureProgress.GET("", lectureProgressHandler.GetLectureProgresses)
			lectureProgress.GET("/:id", lectureProgressHandler.GetLectureProgress)
			lectureProgress.POST("", lectureProgressHandler.CreateLectureProgress)
			lectureProgress.PUT("/:id", lectureProgressHandler.UpdateLectureProgress)
			lectureProgress.DELETE("/:id", lectureProgressHandler.DeleteLectureProgress)
		}

		// Course Reviews routes
		courseReviews := api.Group("/course-reviews")
		{
			courseReviews.GET("", courseReviewHandler.GetCourseReviews)
			courseReviews.GET("/:id", courseReviewHandler.GetCourseReview)
			courseReviews.POST("", courseReviewHandler.CreateCourseReview)
			courseReviews.PUT("/:id", courseReviewHandler.UpdateCourseReview)
			courseReviews.DELETE("/:id", courseReviewHandler.DeleteCourseReview)
		}

		// Wishlists routes
		wishlists := api.Group("/wishlists")
		{
			wishlists.GET("", wishlistHandler.GetWishlists)
			wishlists.GET("/:id", wishlistHandler.GetWishlist)
			wishlists.POST("", wishlistHandler.CreateWishlist)
			wishlists.DELETE("/:id", wishlistHandler.DeleteWishlist)
			wishlists.DELETE("/remove", wishlistHandler.RemoveFromWishlistByUserAndCourse)
			wishlists.GET("/check", wishlistHandler.CheckWishlist)
		}

		// Coupons routes
		coupons := api.Group("/coupons")
		{
			coupons.GET("", couponHandler.GetCoupons)
			coupons.GET("/:id", couponHandler.GetCoupon)
			coupons.POST("", couponHandler.CreateCoupon)
			coupons.PUT("/:id", couponHandler.UpdateCoupon)
			coupons.DELETE("/:id", couponHandler.DeleteCoupon)
			coupons.POST("/validate", couponHandler.ValidateCoupon)
		}

		// Course Announcements routes
		courseAnnouncements := api.Group("/course-announcements")
		{
			courseAnnouncements.GET("", courseAnnouncementHandler.GetCourseAnnouncements)
			courseAnnouncements.GET("/:id", courseAnnouncementHandler.GetCourseAnnouncement)
			courseAnnouncements.POST("", courseAnnouncementHandler.CreateCourseAnnouncement)
			courseAnnouncements.PUT("/:id", courseAnnouncementHandler.UpdateCourseAnnouncement)
			courseAnnouncements.DELETE("/:id", courseAnnouncementHandler.DeleteCourseAnnouncement)
		}

		// Course Q&A routes
		courseQuestions := api.Group("/course-questions")
		{
			courseQuestions.GET("", courseQAHandler.GetCourseQuestions)
			courseQuestions.GET("/:id", courseQAHandler.GetCourseQuestion)
			courseQuestions.POST("", courseQAHandler.CreateCourseQuestion)
			courseQuestions.PUT("/:id", courseQAHandler.UpdateCourseQuestion)
			courseQuestions.DELETE("/:id", courseQAHandler.DeleteCourseQuestion)
			courseQuestions.GET("/:question_id/answers", courseQAHandler.GetCourseAnswers)
		}

		// Course Answers routes
		courseAnswers := api.Group("/course-answers")
		{
			courseAnswers.POST("", courseQAHandler.CreateCourseAnswer)
			courseAnswers.PUT("/:id", courseQAHandler.UpdateCourseAnswer)
			courseAnswers.DELETE("/:id", courseQAHandler.DeleteCourseAnswer)
		}

		// Notifications routes
		notifications := api.Group("/notifications")
		{
			notifications.GET("", notificationHandler.GetNotifications)
			notifications.GET("/:id", notificationHandler.GetNotification)
			notifications.POST("", notificationHandler.CreateNotification)
			notifications.PUT("/:id", notificationHandler.UpdateNotification)
			notifications.DELETE("/:id", notificationHandler.DeleteNotification)
			notifications.PUT("/mark-all-read", notificationHandler.MarkAllAsRead)
		}

		// Course Tags routes
		courseTags := api.Group("/course-tags")
		{
			courseTags.POST("", tagHandler.AddCourseTag)
			courseTags.DELETE("/remove", tagHandler.RemoveCourseTag)
		}
	}

	return r
}
