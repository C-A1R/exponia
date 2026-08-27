-- name: CreateFilmRoll :one
INSERT INTO film_rolls (
    user_id,
    film_stock_id,
    format_id,
    exposure_iso
)
SELECT
    sqlc.arg(user_id),
    fs.id,
    sqlc.arg(format_id),
    fs.iso
FROM film_stocks fs
WHERE fs.id = sqlc.arg(film_stock_id)
RETURNING id;


-- name: GetFilmRollByID :one
SELECT
    fr.id,

    fs.id AS film_stock_id,
    fs.manufacturer AS film_stock_manufacturer,
    fs.name AS film_stock_name,
    fs.iso AS film_stock_iso,

    ff.id AS format_id,
    ff.code AS format_code,
    ff.name AS format_name,

    c.id AS camera_id,
    c.manufacturer AS camera_manufacturer,
    c.model AS camera_model,

    fr.exposure_iso,
    fr.status,
    fr.created_at

FROM film_rolls fr
JOIN film_stocks fs
    ON fs.id = fr.film_stock_id
JOIN film_formats ff
    ON ff.id = fr.format_id
LEFT JOIN cameras c
    ON c.id = fr.camera_id
   AND c.user_id = fr.user_id
WHERE fr.id = sqlc.arg(id)
  AND fr.user_id = sqlc.arg(user_id);


-- name: ListFilmRolls :many
SELECT
    fr.id,

    fs.id AS film_stock_id,
    fs.manufacturer AS film_stock_manufacturer,
    fs.name AS film_stock_name,
    fs.iso AS film_stock_iso,

    ff.id AS format_id,
    ff.code AS format_code,
    ff.name AS format_name,

    c.id AS camera_id,
    c.manufacturer AS camera_manufacturer,
    c.model AS camera_model,

    fr.exposure_iso,
    fr.status,
    fr.created_at

FROM film_rolls fr
JOIN film_stocks fs
    ON fs.id = fr.film_stock_id
JOIN film_formats ff
    ON ff.id = fr.format_id
LEFT JOIN cameras c
    ON c.id = fr.camera_id
   AND c.user_id = fr.user_id
WHERE fr.user_id = sqlc.arg(user_id)
ORDER BY fr.created_at DESC;


-- name: UpdateFilmRollStatus :one
UPDATE film_rolls
SET status = sqlc.arg(status)
WHERE id = sqlc.arg(id)
  AND user_id = sqlc.arg(user_id)
RETURNING id;


-- name: UpdateFilmRollExposureISO :one
UPDATE film_rolls
SET exposure_iso = sqlc.arg(exposure_iso)
WHERE id = sqlc.arg(id)
  AND user_id = sqlc.arg(user_id)
RETURNING id;


-- name: UpdateFilmRollCamera :one
UPDATE film_rolls fr
SET camera_id = sqlc.narg(camera_id)
WHERE fr.id = sqlc.arg(id)
  AND fr.user_id = sqlc.arg(user_id)
  AND (
      sqlc.narg(camera_id)::BIGINT IS NULL
      OR EXISTS (
          SELECT 1
          FROM cameras c
          WHERE c.id = sqlc.narg(camera_id)
            AND c.user_id = sqlc.arg(user_id)
      )
  )
RETURNING fr.id;


-- name: DeleteFilmRoll :execrows
DELETE FROM film_rolls
WHERE id = sqlc.arg(id)
  AND user_id = sqlc.arg(user_id);
