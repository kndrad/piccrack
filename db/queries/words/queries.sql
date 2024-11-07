-- name: AllWords :many
SELECT id,
    value,
    created_at
FROM words
WHERE deleted_at IS NULL
ORDER BY value ASC
LIMIT $1 OFFSET $2;
-- name: InsertWord :one
INSERT INTO words (value, created_at)
VALUES ($1, CURRENT_TIMESTAMP) ON CONFLICT (value) DO NOTHING
RETURNING id,
    value,
    created_at;
