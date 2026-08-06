-- name: CreateCamera :one
INSERT INTO cameras (
    manufacturer,
    model
)
VALUES ($1, $2)
RETURNING id, manufacturer, model, created_at;

-- name: ListCameras :many
SELECT id, manufacturer, model, created_at
FROM cameras
ORDER BY created_at DESC;

-- name: GetCameraByID :one
SELECT id, manufacturer, model, created_at
FROM cameras
WHERE id = $1;

-- name: UpdateCamera :one
UPDATE cameras
SET
    manufacturer = $2,
    model = $3
WHERE id = $1
RETURNING id, manufacturer, model, created_at;

-- name: DeleteCamera :execrows
DELETE FROM cameras
WHERE id = $1;