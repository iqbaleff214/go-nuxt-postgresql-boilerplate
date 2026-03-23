-- name: CreateNotification :one
INSERT INTO notifications (user_id, type, title, body)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: ListNotificationsForUser :many
SELECT * FROM notifications
WHERE user_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: CountNotificationsForUser :one
SELECT COUNT(*) FROM notifications WHERE user_id = $1;

-- name: GetNotificationByID :one
SELECT * FROM notifications WHERE id = $1 LIMIT 1;

-- name: MarkNotificationRead :one
UPDATE notifications
SET read_at = NOW()
WHERE id = $1 AND user_id = $2 AND read_at IS NULL
RETURNING *;

-- name: MarkAllNotificationsRead :exec
UPDATE notifications
SET read_at = NOW()
WHERE user_id = $1 AND read_at IS NULL;

-- name: CountUnreadNotifications :one
SELECT COUNT(*) FROM notifications WHERE user_id = $1 AND read_at IS NULL;
