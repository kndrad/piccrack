ALTER TABLE IF EXISTS phrase_batches
ADD CONSTRAINT phrase_batches_name_unique UNIQUE (
    name
);
