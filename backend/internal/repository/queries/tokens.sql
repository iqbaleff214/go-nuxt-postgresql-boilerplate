-- name: CreateToken :one
INSERT INTO tokens (user_id, token, type, expires_at)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetTokenByHash :one
SELECT * FROM tokens
WHERE token = $1 AND type = $2
LIMIT 1;

-- name: MarkTokenUsed :one
UPDATE tokens
SET used_at = NOW()
WHERE id = $1
RETURNING *;

-- name: DeleteTokensByUserAndType :exec
DELETE FROM tokens WHERE user_id = $1 AND type = $2;

-- name: DeleteExpiredTokens :exec
DELETE FROM tokens WHERE expires_at < NOW();

-- name: ListRecoveryTokensByUser :many
SELECT * FROM tokens
WHERE user_id = $1 AND type = 'totp_recovery' AND used_at IS NULL
ORDER BY created_at ASC;
