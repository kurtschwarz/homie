-- +goose Up
ALTER TABLE photos ADD COLUMN kind TEXT NOT NULL DEFAULT 'PHOTO' CHECK (kind IN ('PHOTO', 'DRAWING'));

-- +goose Down
ALTER TABLE photos DROP COLUMN kind;
