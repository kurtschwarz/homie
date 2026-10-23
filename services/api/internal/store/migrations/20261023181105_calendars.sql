-- +goose Up
CREATE TABLE calendar_accounts (
  id                       BLOB    PRIMARY KEY,
  name                     TEXT    NOT NULL,
  provider                 TEXT    NOT NULL CHECK (provider IN ('caldav')),
  config                   TEXT    NOT NULL DEFAULT '{}',
  credentials              BLOB    NOT NULL,
  refresh_interval_seconds INTEGER NOT NULL DEFAULT 900 CHECK (refresh_interval_seconds >= 300),
  window_past_seconds      INTEGER NOT NULL DEFAULT 86400 CHECK (window_past_seconds >= 0),
  window_future_seconds    INTEGER NOT NULL DEFAULT 1814400 CHECK (window_future_seconds > 0),
  last_attempted_at        INTEGER
) STRICT;

CREATE TABLE calendars (
  id                       BLOB    PRIMARY KEY,
  account_id               BLOB    NOT NULL REFERENCES calendar_accounts (id) ON DELETE CASCADE,
  provider_ref             TEXT    NOT NULL,
  name                     TEXT    NOT NULL,
  visible                  INTEGER NOT NULL DEFAULT 0 CHECK (visible IN (0, 1)),
  read_only                INTEGER NOT NULL DEFAULT 0 CHECK (read_only IN (0, 1)),
  sort_order               INTEGER NOT NULL DEFAULT 0,
  last_synced_at           INTEGER,
  last_error               TEXT    NOT NULL DEFAULT '',
  UNIQUE (account_id, provider_ref)
) STRICT;

DROP TABLE tag_mappings;
DROP INDEX events_start_at;
DROP TABLE events;

CREATE TABLE tag_mappings (
  calendar_id BLOB NOT NULL REFERENCES calendars (id) ON DELETE CASCADE,
  tag         TEXT NOT NULL,
  label       TEXT NOT NULL,
  kind        TEXT NOT NULL CHECK (kind IN ('person', 'include')),
  person_id   BLOB REFERENCES people (id) ON DELETE CASCADE,
  PRIMARY KEY (calendar_id, tag),
  CHECK ((kind = 'person') = (person_id IS NOT NULL))
) STRICT;

CREATE TABLE events (
  id            BLOB    PRIMARY KEY,
  calendar_id   BLOB    NOT NULL REFERENCES calendars (id) ON DELETE CASCADE,
  uid           TEXT    NOT NULL,
  recurrence_id TEXT    NOT NULL DEFAULT '',
  summary       TEXT    NOT NULL,
  location      TEXT    NOT NULL DEFAULT '',
  start_at      INTEGER NOT NULL,
  end_at        INTEGER NOT NULL,
  all_day       INTEGER NOT NULL CHECK (all_day IN (0, 1)),
  categories    TEXT    NOT NULL DEFAULT '',
  synced_at     INTEGER NOT NULL,
  UNIQUE (calendar_id, uid, recurrence_id)
) STRICT;

CREATE INDEX events_start_at ON events (start_at, id);
CREATE INDEX calendars_stale ON calendar_accounts (last_attempted_at);

-- +goose Down
-- This rollback is intentionally lossy: the old schema cannot represent multiple calendars.
DROP INDEX calendars_stale;
DROP INDEX events_start_at;
DROP TABLE events;
DROP TABLE tag_mappings;

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
DROP TABLE calendars;
DROP TABLE calendar_accounts;
