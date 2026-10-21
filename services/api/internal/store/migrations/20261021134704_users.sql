-- +goose Up
CREATE TABLE users (
  id         BLOB    PRIMARY KEY,
  username   TEXT    NOT NULL UNIQUE,
  role       TEXT    NOT NULL DEFAULT 'ADMIN' CHECK (role IN ('ADMIN')),
  created_at INTEGER NOT NULL
) STRICT;

CREATE TABLE user_credentials (
  id            BLOB    PRIMARY KEY,
  user_id       BLOB    NOT NULL REFERENCES users (id) ON DELETE CASCADE,
  kind          TEXT    NOT NULL CHECK (kind IN ('PASSWORD')),
  value         TEXT    NOT NULL,
  created_at    INTEGER NOT NULL,
  last_used_at  INTEGER
) STRICT;

CREATE INDEX user_credentials_user_id ON user_credentials (user_id);

CREATE TABLE user_sessions (
  id            BLOB    PRIMARY KEY,
  user_id       BLOB    NOT NULL REFERENCES users (id) ON DELETE CASCADE,
  token_hash    BLOB    NOT NULL UNIQUE,
  created_at    INTEGER NOT NULL,
  expires_at    INTEGER NOT NULL,
  last_seen_at  INTEGER NOT NULL,
  user_agent    TEXT    NOT NULL DEFAULT ''
) STRICT;

CREATE INDEX user_sessions_user_id ON user_sessions (user_id);

-- +goose Down
DROP INDEX user_sessions_user_id;
DROP TABLE user_sessions;
DROP INDEX user_credentials_user_id;
DROP TABLE user_credentials;
DROP TABLE users;
