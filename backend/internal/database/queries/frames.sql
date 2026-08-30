-- name: CreateFrame :one
INSERT INTO frames (
    film_roll_id,
    frame_index,
    frame_label,
    note
)
SELECT
    fr.id,
    COALESCE(
        (
            SELECT MAX(f.frame_index) + 1
            FROM frames f
            WHERE f.film_roll_id = fr.id
        ),
        1
    ),
    sqlc.narg(frame_label),
    sqlc.narg(note)
FROM film_rolls fr
WHERE fr.id = sqlc.arg(film_roll_id)
  AND fr.user_id = sqlc.arg(user_id)
RETURNING
    id,
    film_roll_id,
    frame_index,
    frame_label,
    note,
    created_at;


-- name: GetFrameByID :one
SELECT
    f.id,
    f.film_roll_id,
    f.frame_index,
    f.frame_label,
    f.note,
    f.created_at
FROM frames f
JOIN film_rolls fr
    ON fr.id = f.film_roll_id
WHERE f.id = sqlc.arg(id)
  AND fr.user_id = sqlc.arg(user_id);


-- name: ListFrames :many
SELECT
    f.id,
    f.film_roll_id,
    f.frame_index,
    f.frame_label,
    f.note,
    f.created_at
FROM frames f
JOIN film_rolls fr
    ON fr.id = f.film_roll_id
WHERE f.film_roll_id = sqlc.arg(film_roll_id)
  AND fr.user_id = sqlc.arg(user_id)
ORDER BY f.frame_index;


-- name: GetOwnedFilmRollID :one
SELECT id
FROM film_rolls
WHERE id = sqlc.arg(film_roll_id)
  AND user_id = sqlc.arg(user_id);


-- name: UpdateFrame :one
UPDATE frames f
SET
    frame_label = sqlc.narg(frame_label),
    note = sqlc.narg(note)
FROM film_rolls fr
WHERE f.id = sqlc.arg(id)
  AND fr.id = f.film_roll_id
  AND fr.user_id = sqlc.arg(user_id)
RETURNING
    f.id,
    f.film_roll_id,
    f.frame_index,
    f.frame_label,
    f.note,
    f.created_at;


-- name: DeleteFrame :execrows
DELETE FROM frames f
USING film_rolls fr
WHERE f.id = sqlc.arg(id)
  AND fr.id = f.film_roll_id
  AND fr.user_id = sqlc.arg(user_id);
