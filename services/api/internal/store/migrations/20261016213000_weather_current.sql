-- +goose Up
CREATE TABLE weather_current (
  location       TEXT    PRIMARY KEY,
  condition_code TEXT    NOT NULL,
  temperature    REAL    NOT NULL,
  daylight       INTEGER NOT NULL CHECK (daylight IN (0, 1)),
  fetched_at     INTEGER NOT NULL,
  expires_at     INTEGER NOT NULL
) STRICT;

-- +goose Down
DROP TABLE weather_current;
