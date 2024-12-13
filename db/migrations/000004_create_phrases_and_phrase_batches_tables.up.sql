CREATE TABLE IF NOT EXISTS phrase_batches (
    id BIGSERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE,
    CHECK (LENGTH(name) > 0)
);

CREATE TABLE IF NOT EXISTS phrases (
    id BIGSERIAL PRIMARY KEY,
    value TEXT NOT NULL,
    batch_id BIGINT REFERENCES phrase_batches (id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE,
    CHECK (LENGTH(value) > 0)
);

CREATE INDEX idx_phrase_value ON phrases (value)
WHERE deleted_at IS NULL;

CREATE INDEX idx_phrase_batch_id ON phrases (batch_id)
WHERE deleted_at IS NULL;
