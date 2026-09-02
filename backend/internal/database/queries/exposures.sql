-- name: CreateExposure :one
INSERT INTO exposures (
    frame_id,
    exposure_index,
    camera_id,
    lens_id,
    aperture,
    shutter_speed_us,
    shot_at,
    note
)
SELECT
    f.id,
    COALESCE(
        (
            SELECT MAX(e.exposure_index) + 1
            FROM exposures e
            WHERE e.frame_id = f.id
        ),
        1
    ),
    fr.camera_id,
    sqlc.narg(lens_id),
    sqlc.narg(aperture),
    sqlc.narg(shutter_speed_us),
    sqlc.narg(shot_at),
    sqlc.narg(note)
FROM frames f
JOIN film_rolls fr
    ON fr.id = f.film_roll_id
WHERE f.id = sqlc.arg(frame_id)
  AND fr.user_id = sqlc.arg(user_id)
  AND fr.camera_id IS NOT NULL
  AND (
      sqlc.narg(lens_id)::BIGINT IS NULL
      OR EXISTS (
          SELECT 1
          FROM lenses l
          WHERE l.id = sqlc.narg(lens_id)
            AND l.user_id = sqlc.arg(user_id)
      )
  )
RETURNING
    id,
    frame_id,
    exposure_index,
    camera_id,
    lens_id,
    aperture,
    shutter_speed_us,
    shot_at,
    note,
    created_at;


-- name: GetExposureCreationState :one
SELECT
    fr.camera_id,
    (
        sqlc.narg(lens_id)::BIGINT IS NULL
        OR EXISTS (
            SELECT 1
            FROM lenses l
            WHERE l.id = sqlc.narg(lens_id)
              AND l.user_id = sqlc.arg(user_id)
        )
    ) AS lens_available
FROM frames f
JOIN film_rolls fr
    ON fr.id = f.film_roll_id
WHERE f.id = sqlc.arg(frame_id)
  AND fr.user_id = sqlc.arg(user_id);


-- name: GetExposureByID :one
SELECT
    e.id,
    e.frame_id,
    e.exposure_index,
    e.camera_id,
    e.lens_id,
    e.aperture,
    e.shutter_speed_us,
    e.shot_at,
    e.note,
    e.created_at
FROM exposures e
JOIN frames f
    ON f.id = e.frame_id
JOIN film_rolls fr
    ON fr.id = f.film_roll_id
WHERE e.id = sqlc.arg(id)
  AND fr.user_id = sqlc.arg(user_id);


-- name: ListExposures :many
SELECT
    e.id,
    e.frame_id,
    e.exposure_index,
    e.camera_id,
    e.lens_id,
    e.aperture,
    e.shutter_speed_us,
    e.shot_at,
    e.note,
    e.created_at
FROM exposures e
JOIN frames f
    ON f.id = e.frame_id
JOIN film_rolls fr
    ON fr.id = f.film_roll_id
WHERE e.frame_id = sqlc.arg(frame_id)
  AND fr.user_id = sqlc.arg(user_id)
ORDER BY e.exposure_index;


-- name: GetOwnedExposureFrameID :one
SELECT f.id
FROM frames f
JOIN film_rolls fr
    ON fr.id = f.film_roll_id
WHERE f.id = sqlc.arg(frame_id)
  AND fr.user_id = sqlc.arg(user_id);


-- name: UpdateExposure :one
UPDATE exposures e
SET
    camera_id = sqlc.arg(camera_id),
    lens_id = sqlc.narg(lens_id),
    aperture = sqlc.narg(aperture),
    shutter_speed_us = sqlc.narg(shutter_speed_us),
    shot_at = sqlc.narg(shot_at),
    note = sqlc.narg(note)
FROM frames f
JOIN film_rolls fr
    ON fr.id = f.film_roll_id
WHERE e.id = sqlc.arg(id)
  AND f.id = e.frame_id
  AND fr.user_id = sqlc.arg(user_id)
  AND EXISTS (
      SELECT 1
      FROM cameras c
      WHERE c.id = sqlc.arg(camera_id)
        AND c.user_id = sqlc.arg(user_id)
  )
  AND (
      sqlc.narg(lens_id)::BIGINT IS NULL
      OR EXISTS (
          SELECT 1
          FROM lenses l
          WHERE l.id = sqlc.narg(lens_id)
            AND l.user_id = sqlc.arg(user_id)
      )
  )
RETURNING
    e.id,
    e.frame_id,
    e.exposure_index,
    e.camera_id,
    e.lens_id,
    e.aperture,
    e.shutter_speed_us,
    e.shot_at,
    e.note,
    e.created_at;


-- name: GetExposureUpdateState :one
SELECT
    EXISTS (
        SELECT 1
        FROM cameras c
        WHERE c.id = sqlc.arg(camera_id)
          AND c.user_id = sqlc.arg(user_id)
    ) AS camera_available,
    (
        sqlc.narg(lens_id)::BIGINT IS NULL
        OR EXISTS (
            SELECT 1
            FROM lenses l
            WHERE l.id = sqlc.narg(lens_id)
              AND l.user_id = sqlc.arg(user_id)
        )
    ) AS lens_available
FROM exposures e
JOIN frames f
    ON f.id = e.frame_id
JOIN film_rolls fr
    ON fr.id = f.film_roll_id
WHERE e.id = sqlc.arg(id)
  AND fr.user_id = sqlc.arg(user_id);


-- name: DeleteExposure :execrows
DELETE FROM exposures e
USING frames f, film_rolls fr
WHERE e.id = sqlc.arg(id)
  AND f.id = e.frame_id
  AND fr.id = f.film_roll_id
  AND fr.user_id = sqlc.arg(user_id);
