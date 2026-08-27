-- name: CreateLens :one
INSERT INTO lenses (
    user_id,
    manufacturer,
    model,
    focal_length_mm,
    max_aperture
)
VALUES (
    sqlc.arg(user_id),
    sqlc.arg(manufacturer),
    sqlc.arg(model),
    sqlc.arg(focal_length_mm),
    sqlc.arg(max_aperture)
)
RETURNING
    id,
    manufacturer,
    model,
    focal_length_mm,
    max_aperture,
    created_at,
    user_id;

-- name: ListLenses :many
SELECT
    id,
    manufacturer,
    model,
    focal_length_mm,
    max_aperture,
    created_at,
    user_id
FROM lenses
WHERE user_id = sqlc.arg(user_id)
ORDER BY created_at DESC;

-- name: GetLensByID :one
SELECT
    id,
    manufacturer,
    model,
    focal_length_mm,
    max_aperture,
    created_at,
    user_id
FROM lenses
WHERE id = sqlc.arg(id)
  AND user_id = sqlc.arg(user_id);

-- name: UpdateLens :one
UPDATE lenses
SET
    manufacturer = sqlc.arg(manufacturer),
    model = sqlc.arg(model),
    focal_length_mm = sqlc.arg(focal_length_mm),
    max_aperture = sqlc.arg(max_aperture)
WHERE id = sqlc.arg(id)
  AND user_id = sqlc.arg(user_id)
RETURNING
    id,
    manufacturer,
    model,
    focal_length_mm,
    max_aperture,
    created_at,
    user_id;

-- name: DeleteLens :execrows
DELETE FROM lenses
WHERE id = sqlc.arg(id)
  AND user_id = sqlc.arg(user_id);
