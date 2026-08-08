-- name: CreateLens :one
INSERT INTO lenses (
    manufacturer,
    model,
    focal_length_mm,
    max_aperture
)
VALUES ($1, $2, $3, $4)
RETURNING id, manufacturer, model, focal_length_mm, max_aperture, created_at;

-- name: ListLenses :many
SELECT id, manufacturer, model, focal_length_mm, max_aperture, created_at
FROM lenses
ORDER BY created_at DESC;

-- name: GetLensByID :one
SELECT id, manufacturer, model, focal_length_mm, max_aperture, created_at
FROM lenses
WHERE id = $1;

-- name: UpdateLens :one
UPDATE lenses
SET
    manufacturer = $2,
    model = $3,
    focal_length_mm = $4,
    max_aperture = $5
WHERE id = $1
RETURNING id, manufacturer, model, focal_length_mm, max_aperture, created_at;

-- name: DeleteLens :execrows
DELETE FROM lenses
WHERE id = $1;