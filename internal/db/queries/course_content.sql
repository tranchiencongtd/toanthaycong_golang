-- name: CreateCourseSection :one
INSERT INTO course_sections (
    course_id, title, description, sort_order
) VALUES (
    $1, $2, $3, $4
) RETURNING *;

-- name: GetCourseSection :one
SELECT * FROM course_sections WHERE id = $1 LIMIT 1;

-- name: ListCourseSections :many
SELECT * FROM course_sections
WHERE course_id = $1
ORDER BY sort_order;

-- name: UpdateCourseSection :one
UPDATE course_sections 
SET 
    title = COALESCE($2, title),
    description = COALESCE($3, description),
    sort_order = COALESCE($4, sort_order),
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING *;

-- name: DeleteCourseSection :exec
DELETE FROM course_sections WHERE id = $1;

-- name: CreateCourseLecture :one
INSERT INTO course_lectures (
    section_id, title, description, content_type, video_url, video_duration,
    article_content, file_url, sort_order, is_preview, is_downloadable
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11
) RETURNING *;

-- name: GetCourseLecture :one
SELECT * FROM course_lectures WHERE id = $1 LIMIT 1;

-- name: ListSectionLectures :many
SELECT * FROM course_lectures
WHERE section_id = $1
ORDER BY sort_order;

-- name: ListCourseLectures :many
SELECT cl.*, cs.title as section_title
FROM course_lectures cl
JOIN course_sections cs ON cl.section_id = cs.id
WHERE cs.course_id = $1
ORDER BY cs.sort_order, cl.sort_order;

-- name: UpdateCourseLecture :one
UPDATE course_lectures 
SET 
    title = COALESCE($2, title),
    description = COALESCE($3, description),
    video_url = COALESCE($4, video_url),
    video_duration = COALESCE($5, video_duration),
    article_content = COALESCE($6, article_content),
    file_url = COALESCE($7, file_url),
    is_preview = COALESCE($8, is_preview),
    is_downloadable = COALESCE($9, is_downloadable),
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING *;

-- name: DeleteCourseLecture :exec
DELETE FROM course_lectures WHERE id = $1;
