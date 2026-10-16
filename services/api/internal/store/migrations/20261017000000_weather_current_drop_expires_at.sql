-- +goose Up
ALTER TABLE weather_current DROP COLUMN expires_at;

-- +goose Down
ALTER TABLE weather_current ADD COLUMN expires_at INTEGER NOT NULL DEFAULT 0;
