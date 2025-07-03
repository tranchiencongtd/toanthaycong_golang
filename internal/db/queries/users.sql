-- name: CreateUser :one
INSERT INTO users (
    email, username, password_hash, first_name, last_name, avatar_url, bio, role
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8
) RETURNING *;

-- name: GetUser :one
SELECT * FROM users WHERE id = $1 LIMIT 1;

-- name: GetUserByEmail :one
SELECT * FROM users WHERE email = $1 LIMIT 1;

-- name: GetUserByUsername :one
SELECT * FROM users WHERE username = $1 LIMIT 1;

-- name: UpdateUser :one
UPDATE users 
SET 
    first_name = COALESCE($2, first_name),
    last_name = COALESCE($3, last_name),
    avatar_url = COALESCE($4, avatar_url),
    bio = COALESCE($5, bio),
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING *;

-- name: VerifyUser :one
UPDATE users 
SET 
    is_verified = true,
    verification_token = null,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING *;

-- name: SetPasswordResetToken :one
UPDATE users 
SET 
    reset_password_token = $2,
    reset_password_expires = $3,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING *;

-- name: ResetPassword :one
UPDATE users 
SET 
    password_hash = $2,
    reset_password_token = null,
    reset_password_expires = null,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING *;

-- name: ListUsers :many
SELECT * FROM users
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;
