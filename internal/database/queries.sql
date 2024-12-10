-- name: ListWords :many
SELECT
    id,
    value,
    created_at
FROM words
WHERE deleted_at IS NULL
ORDER BY value ASC
LIMIT $1 OFFSET $2;

-- name: CreateWord :one
INSERT INTO words (value, created_at)
VALUES ($1, CURRENT_TIMESTAMP)
RETURNING id, value, created_at;

-- name: ListWordFrequencies :many
SELECT
    words.value,
    COUNT(*) AS total
FROM words
WHERE words.deleted_at IS NULL
GROUP BY words.value
ORDER BY total ASC
LIMIT $1 OFFSET $2;

-- name: ListWordRankings :many
SELECT
    words.value,
    ROW_NUMBER() OVER (ORDER BY COUNT(*) DESC) AS ranking
FROM words
WHERE words.deleted_at IS NULL
GROUP BY words.value
ORDER BY ranking ASC
LIMIT $1 OFFSET $2;

-- name: ListWordBatches :many
SELECT
    id,
    name,
    created_at
FROM word_batches
WHERE deleted_at IS NULL
ORDER BY created_at ASC
LIMIT $1 OFFSET $2;

-- name: CreateWordsBatch :one
WITH new_batch AS (
    INSERT INTO word_batches (name)
    VALUES ($1)
    RETURNING id
)

INSERT INTO words (value, batch_id)
SELECT
    word_value,
    (SELECT id FROM new_batch)
FROM UNNEST($2::text []) AS word_value
RETURNING id, value, batch_id;

-- name: ListWordsByBatchName :many
SELECT
    wb.name AS batch_name,
    w.value AS word_value
FROM word_batches AS wb
INNER JOIN words AS w ON wb.id = w.id
WHERE wb.name = $1 AND wb.deleted_at IS NULL
ORDER BY wb.created_at DESC;
