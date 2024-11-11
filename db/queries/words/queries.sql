-- name: AllWords :many
SELECT id, value, created_at
FROM words
WHERE deleted_at IS NULL
ORDER BY value ASC
LIMIT $1 OFFSET $2;
-- name: InsertWord :one
INSERT INTO words (value, created_at)
VALUES ($1, CURRENT_TIMESTAMP)
RETURNING id, value, created_at;
-- name: GetWordsFrequencies :many
SELECT words.value, count(*) AS word_count
FROM words
WHERE deleted_at IS NULL
GROUP BY words.value
ORDER BY word_count ASC
LIMIT $1 OFFSET $2;
-- name: GetWordsRank :many
SELECT
    words.value,
    ROW_NUMBER() OVER (ORDER BY count(*) DESC) as rank
FROM words
WHERE deleted_at IS NULL
GROUP BY words.value
ORDER BY rank ASC
LIMIT $1 OFFSET $2;
