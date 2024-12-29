-- name: CreatePhrasesBatch :one
WITH batch AS (
    INSERT INTO phrase_batches (name)
    VALUES ($1)
    RETURNING id
)

INSERT INTO phrases (value, batch_id)
SELECT
    phrase_value,
    (SELECT id FROM batch)
FROM UNNEST(sqlc.arg(phrases)::text []) AS phrase_value
RETURNING id, batch_id;
