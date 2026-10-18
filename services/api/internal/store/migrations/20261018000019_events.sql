-- +goose Up
CREATE TABLE people (
  id         BLOB    PRIMARY KEY,
  name       TEXT    NOT NULL,
  emoji      TEXT    NOT NULL DEFAULT '',
  swatch     TEXT    NOT NULL,
  sort_order INTEGER NOT NULL,
  is_family  INTEGER NOT NULL DEFAULT 0 CHECK (is_family IN (0, 1))
) STRICT;

CREATE UNIQUE INDEX people_family ON people (is_family) WHERE is_family = 1;

CREATE TABLE tag_mappings (
  tag       TEXT NOT NULL PRIMARY KEY,
  label     TEXT NOT NULL,
  kind      TEXT NOT NULL CHECK (kind IN ('person', 'include')),
  person_id BLOB REFERENCES people (id) ON DELETE CASCADE,
  CHECK ((kind = 'person') = (person_id IS NOT NULL))
) STRICT;

CREATE TABLE events (
  id            BLOB    PRIMARY KEY,
  uid           TEXT    NOT NULL,
  recurrence_id TEXT    NOT NULL DEFAULT '',
  summary       TEXT    NOT NULL,
  location      TEXT    NOT NULL DEFAULT '',
  start_at      INTEGER NOT NULL,
  end_at        INTEGER NOT NULL,
  all_day       INTEGER NOT NULL CHECK (all_day IN (0, 1)),
  categories    TEXT    NOT NULL DEFAULT '',
  synced_at     INTEGER NOT NULL,
  UNIQUE (uid, recurrence_id)
) STRICT;

CREATE INDEX events_start_at ON events (start_at, id);

-- +goose Down
DROP INDEX events_start_at;
DROP TABLE events;
DROP TABLE tag_mappings;
DROP INDEX people_family;
DROP TABLE people;
