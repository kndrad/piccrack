CREATE TABLE IF NOT EXISTS sentence_batches (
    id BIGSERIAL PRIMARY KEY,
    name TEXT NOT NULL CHECK (name > 0),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE
);

CREATE TABLE IF NOT EXISTS sentences (
    id BIGSERIAL PRIMARY KEY,
    value TEXT NOT NULL CHECK (value > 0),
    batch_id BIGINT REFERENCES sentence_batches (id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX idx_sentences_value ON sentences (value)
WHERE deleted_at IS NULL;

CREATE INDEX idx_sentences_batch_id ON sentences (batch_id)
WHERE deleted_at IS NULL;
