-- name: CreateEnrollment :one
INSERT INTO enrollments (
    user_id, course_id
) VALUES (
    $1, $2
) RETURNING *;

-- name: GetEnrollment :one
SELECT * FROM enrollments 
WHERE user_id = $1 AND course_id = $2 
LIMIT 1;

-- name: ListUserEnrollments :many
SELECT e.*, c.title as course_title, c.thumbnail_url, c.instructor_id
FROM enrollments e
JOIN courses c ON e.course_id = c.id
WHERE e.user_id = $1
ORDER BY e.enrolled_at DESC
LIMIT $2 OFFSET $3;

-- name: ListCourseEnrollments :many
SELECT e.*, u.first_name, u.last_name, u.email
FROM enrollments e
JOIN users u ON e.user_id = u.id
WHERE e.course_id = $1
ORDER BY e.enrolled_at DESC
LIMIT $2 OFFSET $3;

-- name: UpdateEnrollmentProgress :one
UPDATE enrollments 
SET 
    progress_percentage = $3,
    last_accessed_at = CURRENT_TIMESTAMP,
    completed_at = CASE WHEN $3 >= 100 THEN CURRENT_TIMESTAMP ELSE completed_at END
WHERE user_id = $1 AND course_id = $2
RETURNING *;

-- name: IsUserEnrolled :one
SELECT EXISTS(
    SELECT 1 FROM enrollments 
    WHERE user_id = $1 AND course_id = $2
);

-- name: GetEnrollmentCount :one
SELECT COUNT(*) FROM enrollments WHERE course_id = $1;
