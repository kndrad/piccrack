-- name: CreateSentencesBatch :one
WITH batch AS (
    INSERT INTO sentence_batches (name)
    VALUES ($1)
    RETURNING id
)

INSERT INTO sentences (value, batch_id)
SELECT
    sentence_value,
    (SELECT id FROM batch)
FROM UNNEST($2::text []) AS sentence_value
RETURNING id, value, batch_id;
