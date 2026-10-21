package accounts

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

// Repository persists users, their credentials, and their sessions.
type Repository interface {
	// UsersExist reports whether any user row exists, without exposing it.
	UsersExist(ctx context.Context) (bool, error)
	// CreateFirstUser inserts a user and password credential together, but
	// only if no user row exists yet. It reports false, nil if another
	// request won the race.
	CreateFirstUser(ctx context.Context, user User, passwordHash string) (bool, error)
	GetUserByUsername(ctx context.Context, username string) (User, string, error)
	GetUser(ctx context.Context, id uuid.UUID) (User, error)
	CreateSession(ctx context.Context, session Session, tokenHash []byte) error
	GetSessionByTokenHash(ctx context.Context, tokenHash []byte) (Session, error)
	TouchSession(ctx context.Context, id uuid.UUID, lastSeenAt, expiresAt time.Time) error
	DeleteSession(ctx context.Context, id uuid.UUID) error
	ListSessions(ctx context.Context, userID uuid.UUID) ([]Session, error)
}

// SQLiteRepository implements Repository with SQLite.
type SQLiteRepository struct {
	db *sqlx.DB
}

// NewSQLiteRepository creates an accounts repository backed by SQLite.
func NewSQLiteRepository(db *sqlx.DB) *SQLiteRepository {
	return &SQLiteRepository{db: db}
}

var _ Repository = (*SQLiteRepository)(nil)

type userRow struct {
	ID        []byte `db:"id"`
	Username  string `db:"username"`
	Role      string `db:"role"`
	CreatedAt int64  `db:"created_at"`
}

func (r userRow) user() (User, error) {
	id, err := uuid.FromBytes(r.ID)
	if err != nil {
		return User{}, fmt.Errorf("decode user id: %w", err)
	}
	return User{ID: id, Username: r.Username, Role: Role(r.Role), CreatedAt: time.Unix(r.CreatedAt, 0)}, nil
}

// UsersExist reports whether any user row exists.
func (r *SQLiteRepository) UsersExist(ctx context.Context) (bool, error) {
	var exists bool
	if err := r.db.GetContext(ctx, &exists, `SELECT EXISTS(SELECT 1 FROM users)`); err != nil {
		return false, err
	}
	return exists, nil
}

// CreateFirstUser inserts a user and its password credential atomically, only
// when no user row exists yet. A single open connection already serializes
// writers, but the read-then-write of a plain SELECT-then-INSERT would
// still race, so the INSERT itself is conditional.
func (r *SQLiteRepository) CreateFirstUser(ctx context.Context, user User, passwordHash string) (bool, error) {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return false, err
	}
	defer func() { _ = tx.Rollback() }()

	result, err := tx.ExecContext(ctx, `
		INSERT INTO users (id, username, role, created_at)
		SELECT ?, ?, ?, ? WHERE NOT EXISTS (SELECT 1 FROM users)`,
		user.ID[:], user.Username, user.Role, user.CreatedAt.Unix())
	if err != nil {
		return false, err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return false, err
	}
	if affected == 0 {
		return false, nil
	}

	credentialID, err := uuid.NewV7()
	if err != nil {
		return false, err
	}
	if _, err := tx.ExecContext(ctx, `
		INSERT INTO user_credentials (id, user_id, kind, value, created_at)
		VALUES (?, ?, ?, ?, ?)`,
		credentialID[:], user.ID[:], CredentialKindPassword, passwordHash, time.Now().Unix()); err != nil {
		return false, err
	}

	if err := tx.Commit(); err != nil {
		return false, err
	}
	return true, nil
}

// GetUserByUsername returns the user and their password credential value.
func (r *SQLiteRepository) GetUserByUsername(ctx context.Context, username string) (User, string, error) {
	var row struct {
		userRow
		Value string `db:"value"`
	}
	err := r.db.GetContext(ctx, &row, `
		SELECT u.id AS id, u.username AS username, u.role AS role, u.created_at AS created_at, c.value AS value
		FROM users u
		JOIN user_credentials c ON c.user_id = u.id AND c.kind = ?
		WHERE u.username = ?`, CredentialKindPassword, username)
	if err == sql.ErrNoRows {
		return User{}, "", ErrInvalidCredentials
	}
	if err != nil {
		return User{}, "", err
	}
	user, err := row.user()
	if err != nil {
		return User{}, "", err
	}
	return user, row.Value, nil
}

// GetUser returns the user with id.
func (r *SQLiteRepository) GetUser(ctx context.Context, id uuid.UUID) (User, error) {
	var row userRow
	if err := r.db.GetContext(ctx, &row, `SELECT id, username, role, created_at FROM users WHERE id = ?`, id[:]); err != nil {
		return User{}, err
	}
	return row.user()
}

type sessionRow struct {
	ID         []byte `db:"id"`
	UserID     []byte `db:"user_id"`
	CreatedAt  int64  `db:"created_at"`
	ExpiresAt  int64  `db:"expires_at"`
	LastSeenAt int64  `db:"last_seen_at"`
	UserAgent  string `db:"user_agent"`
}

func (r sessionRow) session() (Session, error) {
	id, err := uuid.FromBytes(r.ID)
	if err != nil {
		return Session{}, fmt.Errorf("decode session id: %w", err)
	}
	userID, err := uuid.FromBytes(r.UserID)
	if err != nil {
		return Session{}, fmt.Errorf("decode session user id: %w", err)
	}
	return Session{
		ID: id, UserID: userID,
		CreatedAt: time.Unix(r.CreatedAt, 0), ExpiresAt: time.Unix(r.ExpiresAt, 0), LastSeenAt: time.Unix(r.LastSeenAt, 0),
		UserAgent: r.UserAgent,
	}, nil
}

// CreateSession inserts session, storing tokenHash rather than the raw
// token: a database leak must not be a session leak.
func (r *SQLiteRepository) CreateSession(ctx context.Context, session Session, tokenHash []byte) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO user_sessions (id, user_id, token_hash, created_at, expires_at, last_seen_at, user_agent)
		VALUES (?, ?, ?, ?, ?, ?, ?)`,
		session.ID[:], session.UserID[:], tokenHash, session.CreatedAt.Unix(), session.ExpiresAt.Unix(), session.LastSeenAt.Unix(), session.UserAgent)
	return err
}

// GetSessionByTokenHash returns the session matching tokenHash.
func (r *SQLiteRepository) GetSessionByTokenHash(ctx context.Context, tokenHash []byte) (Session, error) {
	var row sessionRow
	err := r.db.GetContext(ctx, &row, `
		SELECT id, user_id, created_at, expires_at, last_seen_at, user_agent
		FROM user_sessions WHERE token_hash = ?`, tokenHash)
	if err == sql.ErrNoRows {
		return Session{}, ErrSessionNotFound
	}
	if err != nil {
		return Session{}, err
	}
	return row.session()
}

// TouchSession extends a session's expiry on use.
func (r *SQLiteRepository) TouchSession(ctx context.Context, id uuid.UUID, lastSeenAt, expiresAt time.Time) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE user_sessions SET last_seen_at = ?, expires_at = ? WHERE id = ?`,
		lastSeenAt.Unix(), expiresAt.Unix(), id[:])
	return err
}

// DeleteSession removes a session, logging that browser out immediately.
func (r *SQLiteRepository) DeleteSession(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM user_sessions WHERE id = ?`, id[:])
	return err
}

// ListSessions returns a user's sessions, most recently used first.
func (r *SQLiteRepository) ListSessions(ctx context.Context, userID uuid.UUID) ([]Session, error) {
	var rows []sessionRow
	if err := r.db.SelectContext(ctx, &rows, `
		SELECT id, user_id, created_at, expires_at, last_seen_at, user_agent
		FROM user_sessions WHERE user_id = ? ORDER BY last_seen_at DESC`, userID[:]); err != nil {
		return nil, err
	}
	sessions := make([]Session, 0, len(rows))
	for _, row := range rows {
		session, err := row.session()
		if err != nil {
			return nil, err
		}
		sessions = append(sessions, session)
	}
	return sessions, nil
}
