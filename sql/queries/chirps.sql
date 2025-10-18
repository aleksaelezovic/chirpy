-- name: CreateChirp :one
INSERT INTO chirps (id, user_id, body)
VALUES (gen_random_uuid(), $1, $2)
RETURNING *;
