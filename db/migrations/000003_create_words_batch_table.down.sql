ALTER TABLE words
DROP CONSTRAINT words_batch_fkey,
DROP COLUMN batch_id;

DROP INDEX IF EXISTS idx_words_batch_id;
DROP TABLE IF EXISTS word_batches;
