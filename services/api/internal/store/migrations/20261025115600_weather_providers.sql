-- +goose Up
CREATE TABLE weather_providers (
  id                       BLOB    PRIMARY KEY,
  provider                 TEXT    NOT NULL CHECK (provider IN ('weatherkit')),
  config                   TEXT    NOT NULL DEFAULT '{}',
  credentials              BLOB    NOT NULL,
  active                   INTEGER NOT NULL DEFAULT 0 CHECK (active IN (0, 1)),
  refresh_interval_seconds INTEGER NOT NULL DEFAULT 1800 CHECK (refresh_interval_seconds >= 900),
  last_attempted_at        INTEGER,
  last_error               TEXT    NOT NULL DEFAULT ''
) STRICT;

CREATE UNIQUE INDEX weather_providers_provider ON weather_providers (provider);
CREATE UNIQUE INDEX weather_providers_active ON weather_providers (active) WHERE active = 1;

CREATE TABLE weather_locations (
  id         BLOB    PRIMARY KEY,
  label      TEXT    NOT NULL,
  latitude   REAL    NOT NULL CHECK (latitude BETWEEN -90 AND 90),
  longitude  REAL    NOT NULL CHECK (longitude BETWEEN -180 AND 180),
  sort_order INTEGER NOT NULL DEFAULT 0
) STRICT;

-- The cache is disposable and its old "lat,lon" key has no location row to point at.
DROP TABLE weather_current;
CREATE TABLE weather_current (
  location_id    BLOB    PRIMARY KEY REFERENCES weather_locations (id) ON DELETE CASCADE,
  condition_code TEXT    NOT NULL,
  temperature    REAL    NOT NULL,
  daylight       INTEGER NOT NULL CHECK (daylight IN (0, 1)),
  fetched_at     INTEGER NOT NULL
) STRICT;

-- +goose Down
DROP TABLE weather_current;
CREATE TABLE weather_current (
  location       TEXT    PRIMARY KEY,
  condition_code TEXT    NOT NULL,
  temperature    REAL    NOT NULL,
  daylight       INTEGER NOT NULL CHECK (daylight IN (0, 1)),
  fetched_at     INTEGER NOT NULL
) STRICT;
DROP TABLE weather_locations;
DROP INDEX weather_providers_active;
DROP INDEX weather_providers_provider;
DROP TABLE weather_providers;
