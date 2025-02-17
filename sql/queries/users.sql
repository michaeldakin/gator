-- name: CreateUser :one
INSERT INTO users (id, created_at, updated_at, name)
VALUES (
    $1, $2, $3, $4
)
RETURNING *;

-- name: GetUser :one
SELECT * FROM users
WHERE id = $1 LIMIT 1;

-- name: GetUserByName :one
SELECT * from users
WHERE name = $1 LIMIT 1;

-- name: GetAllUsers :many
SELECT * FROM users
ORDER BY created_at;

-- name: DropUser :one
DELETE FROM users
WHERE id = $1
RETURNING *;
