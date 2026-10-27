-- +goose Up
CREATE TABLE display_modes (
  mode              TEXT    PRIMARY KEY CHECK (mode IN ('light', 'dark', 'night', 'off')),
  starts_at_minutes INTEGER NOT NULL CHECK (starts_at_minutes BETWEEN 0 AND 1439),
  brightness        INTEGER NOT NULL CHECK (brightness BETWEEN 0 AND 100)
) STRICT;

INSERT INTO display_modes (mode, starts_at_minutes, brightness) VALUES
  ('light', 390, 100),
  ('dark', 1140, 60),
  ('night', 1260, 20),
  ('off', 1380, 0);

CREATE TABLE display_settings (
  id                   INTEGER PRIMARY KEY CHECK (id = 1),
  power_saving_enabled INTEGER NOT NULL DEFAULT 0 CHECK (power_saving_enabled IN (0, 1)),
  wake_for_alerts      INTEGER NOT NULL DEFAULT 0 CHECK (wake_for_alerts IN (0, 1)),
  ddc_bus              TEXT    NOT NULL DEFAULT '' CHECK (ddc_bus = '' OR ddc_bus NOT GLOB '*[^0-9]*')
) STRICT;

INSERT INTO display_settings (id) VALUES (1);

-- +goose Down
DROP TABLE display_settings;
DROP TABLE display_modes;
