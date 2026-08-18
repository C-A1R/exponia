-- name: CreateCamera :one
INSERT INTO cameras (
    user_id,
    manufacturer,
    model
)
VALUES (
    sqlc.arg(user_id),
    sqlc.arg(manufacturer),
    sqlc.arg(model)
)
RETURNING id, manufacturer, model, created_at, user_id;

-- name: ListCameras :many
SELECT id, manufacturer, model, created_at, user_id
FROM cameras
WHERE user_id = sqlc.arg(user_id)
ORDER BY created_at DESC;

-- name: GetCameraByID :one
SELECT id, manufacturer, model, created_at, user_id
FROM cameras
WHERE id = sqlc.arg(id)
  AND user_id = sqlc.arg(user_id);

-- name: UpdateCamera :one
UPDATE cameras
SET
    manufacturer = sqlc.arg(manufacturer),
    model = sqlc.arg(model)
WHERE id = sqlc.arg(id)
  AND user_id = sqlc.arg(user_id)
RETURNING id, manufacturer, model, created_at, user_id;

-- name: DeleteCamera :execrows
DELETE FROM cameras
WHERE id = sqlc.arg(id)
  AND user_id = sqlc.arg(user_id);
