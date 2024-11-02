CREATE TABLE IF NOT EXISTS words (
    id BIGSERIAL PRIMARY KEY,
    value TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE,
    CONSTRAINT words_value_unique UNIQUE (value)
);
CREATE INDEX idx_words_value ON words(value)
WHERE deleted_at IS NULL;
