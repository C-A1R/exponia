-- name: CreateUser :one
INSERT INTO users (
    email,
    display_name,
    auth_issuer,
    auth_subject
)
VALUES (
    sqlc.arg(email),
    sqlc.arg(display_name),
    sqlc.arg(auth_issuer),
    sqlc.arg(auth_subject)
)
RETURNING
    id,
    email,
    display_name,
    auth_issuer,
    auth_subject,
    created_at;

-- name: GetUserByID :one
SELECT
    id,
    email,
    display_name,
    auth_issuer,
    auth_subject,
    created_at
FROM users
WHERE id = sqlc.arg(id);

-- name: GetUserByAuthIdentity :one
SELECT
    id,
    email,
    display_name,
    auth_issuer,
    auth_subject,
    created_at
FROM users
WHERE auth_issuer = sqlc.arg(auth_issuer)
  AND auth_subject = sqlc.arg(auth_subject);
