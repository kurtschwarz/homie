// Package accounts manages the single user account that gates the
// manager: the user row, its password credential, and the sessions that
// prove a browser is logged in as them.
package accounts

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// SessionTTL is how long a session lasts since it was last used. It is
// extended on every use, so a home admin UI is not logged out by the clock
// while someone is actively using it.
const SessionTTL = 30 * 24 * time.Hour

// CredentialKindPassword is the only credential kind today. Passkeys would
// be another kind and another row, not a new column.
const CredentialKindPassword = "PASSWORD"

// Role identifies the permissions granted to a user.
type Role string

// RoleAdmin is the role assigned during first-run setup.
const RoleAdmin Role = "ADMIN"

// User is an account that can sign in to the manager.
type User struct {
	ID        uuid.UUID
	Username  string
	Role      Role
	CreatedAt time.Time
}

// Session is a logged-in browser's proof of identity. The raw Token is only
// ever available at creation; afterwards only its hash is stored.
type Session struct {
	ID         uuid.UUID
	UserID     uuid.UUID
	Token      string // set only when freshly created
	CreatedAt  time.Time
	ExpiresAt  time.Time
	LastSeenAt time.Time
	UserAgent  string
}

var (
	// ErrSetupComplete is returned by Setup once a user already exists: the
	// setup code is permanently spent at that point.
	ErrSetupComplete = errors.New("accounts: setup is already complete")
	// ErrInvalidSetupCode is returned for a wrong or already-used setup code.
	ErrInvalidSetupCode = errors.New("accounts: invalid setup code")
	// ErrInvalidCredentials is returned for a wrong username or password.
	ErrInvalidCredentials = errors.New("accounts: invalid username or password")
	// ErrRateLimited is returned when login or setup is attempted too often.
	ErrRateLimited = errors.New("accounts: too many attempts, try again later")
	// ErrSessionNotFound is returned when a session token doesn't resolve to
	// a live session, or a session id isn't one of the admin's own.
	ErrSessionNotFound = errors.New("accounts: session not found")
	// ErrWeakPassword is returned for a password too short to be worth hashing.
	ErrWeakPassword = errors.New("accounts: password must be at least 8 characters")
	// ErrUsernameRequired is returned for a blank username.
	ErrUsernameRequired = errors.New("accounts: username is required")
)
