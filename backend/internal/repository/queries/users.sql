-- name: CreateUser :one
INSERT INTO users (email, hashed_password, first_name, last_name, display_name, role)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: GetUserByID :one
SELECT * FROM users WHERE id = $1 AND deleted_at IS NULL LIMIT 1;

-- name: GetUserByEmail :one
SELECT * FROM users WHERE email = $1 AND deleted_at IS NULL LIMIT 1;

-- name: UpdateUserProfile :one
UPDATE users
SET first_name   = $2,
    last_name    = $3,
    display_name = $4,
    bio          = $5,
    updated_at   = NOW()
WHERE id = $1
RETURNING *;

-- name: UpdateUserEmail :one
UPDATE users
SET email      = $2,
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: UpdateUserPassword :one
UPDATE users
SET hashed_password = $2,
    updated_at      = NOW()
WHERE id = $1
RETURNING *;

-- name: UpdateUserStatus :one
UPDATE users
SET status     = $2,
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: UpdateUserRole :one
UPDATE users
SET role       = $2,
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: UpdateUserAvatarURL :one
UPDATE users
SET avatar_url = $2,
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: SetEmailVerified :one
UPDATE users
SET is_email_verified = TRUE,
    status            = 'active',
    updated_at        = NOW()
WHERE id = $1
RETURNING *;

-- name: Enable2FA :one
UPDATE users
SET totp_secret    = $2,
    is_2fa_enabled = $3,
    updated_at     = NOW()
WHERE id = $1
RETURNING *;

-- name: Disable2FA :one
UPDATE users
SET totp_secret    = NULL,
    is_2fa_enabled = FALSE,
    updated_at     = NOW()
WHERE id = $1
RETURNING *;

-- name: UpdateLastLogin :one
UPDATE users
SET last_login_at = NOW(),
    updated_at    = NOW()
WHERE id = $1
RETURNING *;

-- name: SoftDeleteUser :one
UPDATE users
SET deleted_at = NOW(),
    status     = 'inactive',
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: CancelSoftDelete :one
UPDATE users
SET deleted_at = NULL,
    status     = 'active',
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: HardDeleteUser :exec
DELETE FROM users WHERE id = $1;

-- name: ListUsers :many
SELECT * FROM users
WHERE deleted_at IS NULL
  AND ($1::user_role IS NULL OR role = $1)
  AND ($2::user_status IS NULL OR status = $2)
  AND ($3::boolean IS NULL OR is_email_verified = $3)
  AND ($4::text IS NULL OR email ILIKE '%' || $4 || '%' OR display_name ILIKE '%' || $4 || '%')
ORDER BY created_at DESC
LIMIT $5 OFFSET $6;

-- name: CountUsers :one
SELECT COUNT(*) FROM users
WHERE deleted_at IS NULL
  AND ($1::user_role IS NULL OR role = $1)
  AND ($2::user_status IS NULL OR status = $2)
  AND ($3::boolean IS NULL OR is_email_verified = $3)
  AND ($4::text IS NULL OR email ILIKE '%' || $4 || '%' OR display_name ILIKE '%' || $4 || '%');

-- name: ListUsersScheduledForHardDelete :many
SELECT * FROM users
WHERE deleted_at IS NOT NULL
  AND deleted_at < NOW() - INTERVAL '30 days';
