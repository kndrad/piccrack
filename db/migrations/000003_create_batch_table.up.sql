CREATE TABLE IF NOT EXISTS word_batches (
    id BIGSERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE
);

ALTER TABLE words
ADD COLUMN batch_id BIGINT,
ADD CONSTRAINT words_batch_fkey
FOREIGN KEY (batch_id) REFERENCES word_batches (id);

CREATE INDEX idx_words_batch_id ON words (batch_id) WHERE deleted_at IS NULL;
