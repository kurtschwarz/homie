-- +goose Up
CREATE TABLE photos (
  id          BLOB    PRIMARY KEY,
  asset_id    TEXT    NOT NULL UNIQUE,
  version     TEXT    NOT NULL,
  file        TEXT    NOT NULL,
  width       INTEGER NOT NULL,
  height      INTEGER NOT NULL,
  taken_at    INTEGER NOT NULL,
  latitude    REAL,
  longitude   REAL,
  place       TEXT,
  imported_at INTEGER NOT NULL
) STRICT;

CREATE INDEX photos_taken_at ON photos (taken_at DESC, id DESC);

-- +goose Down
DROP INDEX photos_taken_at;
DROP TABLE photos;
