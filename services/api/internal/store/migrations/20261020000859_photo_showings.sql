-- +goose Up
CREATE TABLE photo_showings (
  id         BLOB    PRIMARY KEY,
  photo_id   BLOB    NOT NULL REFERENCES photos (id) ON DELETE CASCADE,
  started_at INTEGER NOT NULL,
  ends_at    INTEGER NOT NULL,
  ended_at   INTEGER,
  source     TEXT    NOT NULL CHECK (source IN ('SCHEDULE', 'TAP'))
) STRICT;

CREATE INDEX photo_showings_started_at ON photo_showings (started_at DESC, id DESC);
CREATE INDEX photo_showings_photo ON photo_showings (photo_id, started_at DESC);

-- +goose Down
DROP INDEX photo_showings_photo;
DROP INDEX photo_showings_started_at;
DROP TABLE photo_showings;
