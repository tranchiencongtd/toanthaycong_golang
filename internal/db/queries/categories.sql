-- name: CreateCategory :one
INSERT INTO categories (
    name, slug, description, icon_url, parent_id, sort_order
) VALUES (
    $1, $2, $3, $4, $5, $6
) RETURNING *;

-- name: GetCategory :one
SELECT * FROM categories WHERE id = $1 LIMIT 1;

-- name: GetCategoryBySlug :one
SELECT * FROM categories WHERE slug = $1 LIMIT 1;

-- name: ListCategories :many
SELECT * FROM categories
WHERE is_active = true
ORDER BY sort_order, name;

-- name: ListRootCategories :many
SELECT * FROM categories
WHERE parent_id IS NULL AND is_active = true
ORDER BY sort_order, name;

-- name: ListSubCategories :many
SELECT * FROM categories
WHERE parent_id = $1 AND is_active = true
ORDER BY sort_order, name;

-- name: UpdateCategory :one
UPDATE categories 
SET 
    name = COALESCE($2, name),
    description = COALESCE($3, description),
    icon_url = COALESCE($4, icon_url),
    sort_order = COALESCE($5, sort_order),
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING *;

-- name: DeleteCategory :exec
DELETE FROM categories WHERE id = $1;
