-- +goose Up
ALTER TABLE photos ADD COLUMN focus TEXT CHECK (focus IS NULL OR json_valid(focus));

-- +goose Down
ALTER TABLE photos DROP COLUMN focus;
