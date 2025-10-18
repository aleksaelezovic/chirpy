-- name: CreateUser :one
INSERT INTO users (id, email, hashed_password)
VALUES (gen_random_uuid(), $1, $2)
RETURNING *;

-- name: DeleteAllUsers :exec
DELETE FROM users;

-- name: GetUserByEmail :one
SELECT * FROM users WHERE email = $1;
