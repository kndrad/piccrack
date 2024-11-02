-- name: AllWords :many
SELECT id,
    value,
    created_at
FROM words
WHERE deleted_at IS NULL
ORDER BY value ASC
LIMIT $1 OFFSET $2;
--
-- name: InsertWord :one
INSERT INTO words (value, created_at)
VALUES ($1, CURRENT_TIMESTAMP) ON CONFLICT (value) DO NOTHING
RETURNING id,
    value,
    created_at;
--
-- name: GetWordByValue :one
SELECT id,
    value,
    created_at
FROM words
WHERE value = $1
    AND deleted_at IS NULL
LIMIT 1;
--
-- name: SoftDeleteWord :exec
UPDATE words
SET deleted_at = CURRENT_TIMESTAMP
WHERE id = $1
    AND deleted_at IS NULL;
