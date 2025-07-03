-- name: CreateCourseReview :one
INSERT INTO course_reviews (
    user_id, course_id, rating, review_text
) VALUES (
    $1, $2, $3, $4
) RETURNING *;

-- name: GetCourseReview :one
SELECT * FROM course_reviews 
WHERE user_id = $1 AND course_id = $2 
LIMIT 1;

-- name: ListCourseReviews :many
SELECT cr.*, u.first_name, u.last_name, u.avatar_url
FROM course_reviews cr
JOIN users u ON cr.user_id = u.id
WHERE cr.course_id = $1 AND cr.is_approved = true
ORDER BY cr.created_at DESC
LIMIT $2 OFFSET $3;

-- name: UpdateCourseReview :one
UPDATE course_reviews 
SET 
    rating = COALESCE($3, rating),
    review_text = COALESCE($4, review_text),
    updated_at = CURRENT_TIMESTAMP
WHERE user_id = $1 AND course_id = $2
RETURNING *;

-- name: GetCourseRatingStats :one
SELECT 
    AVG(rating)::DECIMAL(3,2) as average_rating,
    COUNT(*) as total_reviews,
    COUNT(CASE WHEN rating = 5 THEN 1 END) as five_star,
    COUNT(CASE WHEN rating = 4 THEN 1 END) as four_star,
    COUNT(CASE WHEN rating = 3 THEN 1 END) as three_star,
    COUNT(CASE WHEN rating = 2 THEN 1 END) as two_star,
    COUNT(CASE WHEN rating = 1 THEN 1 END) as one_star
FROM course_reviews 
WHERE course_id = $1 AND is_approved = true;
