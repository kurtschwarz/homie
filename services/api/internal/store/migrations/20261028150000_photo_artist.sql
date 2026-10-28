-- +goose Up
ALTER TABLE photos ADD COLUMN artist_id BLOB REFERENCES people (id) ON DELETE SET NULL;

-- +goose Down
ALTER TABLE photos DROP COLUMN artist_id;
