-- +goose Up
ALTER TABLE photos ADD COLUMN pushed_asset_id TEXT;
ALTER TABLE photos ADD COLUMN pushed_at INTEGER;

-- +goose Down
ALTER TABLE photos DROP COLUMN pushed_at;
ALTER TABLE photos DROP COLUMN pushed_asset_id;
