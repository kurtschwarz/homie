package accounts

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/kurtschwarz/homie/services/api/internal/auth"
)

// Service is the gate behind the manager: one user, proven by a password
// today and perhaps a passkey later, and the sessions that remember a
// browser proved it.
type Service struct {
	repository Repository
	logger     *slog.Logger

	// setupCode is generated once at startup, only when no user exists yet,
	// and never persisted: it rotates on every restart, and once a user
	// exists it is cleared for good.
	setupCode string

	loginLimiter *auth.Limiter
	setupLimiter *auth.Limiter
}

// New creates the accounts service, generating a setup code if (and only
// if) no user has been created yet. Call SetupCode to log it once.
func New(ctx context.Context, repository Repository, logger *slog.Logger) (*Service, error) {
	service := &Service{
		repository:   repository,
		logger:       logger,
		loginLimiter: auth.NewLimiter(10, 15*time.Minute),
		setupLimiter: auth.NewLimiter(10, 15*time.Minute),
	}
	exists, err := repository.UsersExist(ctx)
	if err != nil {
		return nil, err
	}
	if !exists {
		code, err := auth.GenerateSetupCode()
		if err != nil {
			return nil, err
		}
		service.setupCode = code
	}
	return service, nil
}

// SetupCode returns the code to complete first-run setup, or "" once a
// user already exists.
func (s *Service) SetupCode() string { return s.setupCode }

// Setup creates the first user account as ADMIN, consuming the setup code.
// It fails permanently once a user exists, regardless of the code.
func (s *Service) Setup(ctx context.Context, clientIP, code, username, password, userAgent string) (Session, User, error) {
	if !s.setupLimiter.Allow("setup:" + clientIP) {
		return Session{}, User{}, ErrRateLimited
	}
	if s.setupCode == "" {
		return Session{}, User{}, ErrSetupComplete
	}
	if !auth.EqualSetupCode(s.setupCode, code) {
		return Session{}, User{}, ErrInvalidSetupCode
	}
	if username == "" {
		return Session{}, User{}, ErrUsernameRequired
	}
	if len(password) < 8 {
		return Session{}, User{}, ErrWeakPassword
	}

	id, err := uuid.NewV7()
	if err != nil {
		return Session{}, User{}, err
	}
	user := User{ID: id, Username: username, Role: RoleAdmin, CreatedAt: time.Now()}
	hash, err := auth.HashPassword(password)
	if err != nil {
		return Session{}, User{}, err
	}
	created, err := s.repository.CreateFirstUser(ctx, user, hash)
	if err != nil {
		return Session{}, User{}, err
	}
	if !created {
		// Another request won the race; the code is spent either way.
		s.setupCode = ""
		return Session{}, User{}, ErrSetupComplete
	}
	s.setupCode = ""

	session, err := s.createSession(ctx, user.ID, userAgent)
	if err != nil {
		return Session{}, User{}, err
	}
	return session, user, nil
}

// Login proves identity by password and starts a session.
func (s *Service) Login(ctx context.Context, clientIP, username, password, userAgent string) (Session, User, error) {
	if !s.loginLimiter.Allow("login:" + clientIP) {
		return Session{}, User{}, ErrRateLimited
	}
	user, hash, err := s.repository.GetUserByUsername(ctx, username)
	if err != nil {
		return Session{}, User{}, err
	}
	ok, err := auth.VerifyPassword(hash, password)
	if err != nil {
		return Session{}, User{}, err
	}
	if !ok {
		return Session{}, User{}, ErrInvalidCredentials
	}
	session, err := s.createSession(ctx, user.ID, userAgent)
	if err != nil {
		return Session{}, User{}, err
	}
	return session, user, nil
}

func (s *Service) createSession(ctx context.Context, userID uuid.UUID, userAgent string) (Session, error) {
	id, err := uuid.NewV7()
	if err != nil {
		return Session{}, err
	}
	var raw [32]byte
	if _, err := rand.Read(raw[:]); err != nil {
		return Session{}, err
	}
	token := hex.EncodeToString(raw[:])
	now := time.Now()
	session := Session{
		ID: id, UserID: userID, Token: token,
		CreatedAt: now, ExpiresAt: now.Add(SessionTTL), LastSeenAt: now,
		UserAgent: userAgent,
	}
	if err := s.repository.CreateSession(ctx, session, hashToken(token)); err != nil {
		return Session{}, err
	}
	return session, nil
}

// CurrentUser resolves a session token to the user it belongs to,
// extending the session's expiry. An empty, unknown or expired token
// returns ErrSessionNotFound.
func (s *Service) CurrentUser(ctx context.Context, token string) (User, Session, error) {
	if token == "" {
		return User{}, Session{}, ErrSessionNotFound
	}
	session, err := s.repository.GetSessionByTokenHash(ctx, hashToken(token))
	if err != nil {
		return User{}, Session{}, err
	}
	now := time.Now()
	if session.ExpiresAt.Before(now) {
		_ = s.repository.DeleteSession(ctx, session.ID)
		return User{}, Session{}, ErrSessionNotFound
	}
	user, err := s.repository.GetUser(ctx, session.UserID)
	if err != nil {
		return User{}, Session{}, err
	}
	expiresAt := now.Add(SessionTTL)
	if err := s.repository.TouchSession(ctx, session.ID, now, expiresAt); err != nil {
		s.logger.Warn("extend session", "error", err)
	} else {
		session.LastSeenAt, session.ExpiresAt = now, expiresAt
	}
	return user, session, nil
}

// Logout deletes the session behind token, if any. Logging out an unknown
// token is not an error: the browser ends up logged out either way.
func (s *Service) Logout(ctx context.Context, token string) error {
	if token == "" {
		return nil
	}
	session, err := s.repository.GetSessionByTokenHash(ctx, hashToken(token))
	if err == ErrSessionNotFound {
		return nil
	}
	if err != nil {
		return err
	}
	return s.repository.DeleteSession(ctx, session.ID)
}

// ListSessions returns userID's active sessions, most recently used first.
func (s *Service) ListSessions(ctx context.Context, userID uuid.UUID) ([]Session, error) {
	return s.repository.ListSessions(ctx, userID)
}

// RevokeSession deletes one of userID's own sessions. Scoping to userID
// stops one user from revoking a session that isn't theirs.
func (s *Service) RevokeSession(ctx context.Context, userID, sessionID uuid.UUID) error {
	sessions, err := s.repository.ListSessions(ctx, userID)
	if err != nil {
		return err
	}
	for _, session := range sessions {
		if session.ID == sessionID {
			return s.repository.DeleteSession(ctx, sessionID)
		}
	}
	return ErrSessionNotFound
}

func hashToken(token string) []byte {
	sum := sha256.Sum256([]byte(token))
	return sum[:]
}
