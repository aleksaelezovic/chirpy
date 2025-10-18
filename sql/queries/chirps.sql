-- name: CreateChirp :one
INSERT INTO chirps (id, user_id, body)
VALUES (gen_random_uuid(), $1, $2)
RETURNING *;

-- name: GetAllChirps :many
SELECT * FROM chirps ORDER BY created_at ASC;
