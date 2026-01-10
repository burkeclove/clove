-- name: CreateUser :one
INSERT INTO users (email, password_hash)
VALUES ($1, $2)
RETURNING *;

-- name: GetUserById :one
SELECT * FROM users WHERE id = $1;

-- name: GetUserByEmailPassword :one
SELECT * FROM users WHERE email = $1 and password_hash = $2;
