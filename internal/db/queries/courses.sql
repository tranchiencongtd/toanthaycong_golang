-- name: CreateCourse :one
INSERT INTO courses (
    title, slug, description, short_description, thumbnail_url, preview_video_url,
    instructor_id, category_id, price, discount_price, language, level, 
    requirements, what_you_learn, target_audience
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15
) RETURNING *;

-- name: GetCourse :one
SELECT * FROM courses WHERE id = $1 LIMIT 1;

-- name: GetCourseBySlug :one
SELECT * FROM courses WHERE slug = $1 LIMIT 1;

-- name: ListCourses :many
SELECT * FROM courses
WHERE status = 'published'
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

-- name: ListCoursesByInstructor :many
SELECT * FROM courses
WHERE instructor_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: ListCoursesByCategory :many
SELECT * FROM courses
WHERE category_id = $1 AND status = 'published'
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: SearchCourses :many
SELECT * FROM courses
WHERE 
    status = 'published' AND
    (title ILIKE '%' || $1 || '%' OR description ILIKE '%' || $1 || '%')
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: UpdateCourse :one
UPDATE courses 
SET 
    title = COALESCE($2, title),
    description = COALESCE($3, description),
    short_description = COALESCE($4, short_description),
    thumbnail_url = COALESCE($5, thumbnail_url),
    preview_video_url = COALESCE($6, preview_video_url),
    price = COALESCE($7, price),
    discount_price = COALESCE($8, discount_price),
    level = COALESCE($9, level),
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING *;

-- name: UpdateCourseStatus :one
UPDATE courses 
SET 
    status = $2,
    published_at = CASE WHEN $2 = 'published' THEN CURRENT_TIMESTAMP ELSE published_at END,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING *;

-- name: UpdateCourseStats :one
UPDATE courses 
SET 
    rating = $2,
    total_students = $3,
    total_reviews = $4,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING *;

-- name: DeleteCourse :exec
DELETE FROM courses WHERE id = $1;
