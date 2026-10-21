package accounts_test

import (
	"context"
	"io"
	"log/slog"
	"path/filepath"
	"testing"
	"time"

	"github.com/kurtschwarz/homie/services/api/internal/accounts"
	"github.com/kurtschwarz/homie/services/api/internal/store"
)

func openService(t *testing.T) *accounts.Service {
	t.Helper()
	db, err := store.Open(filepath.Join(t.TempDir(), "accounts.db"), slog.New(slog.NewTextHandler(io.Discard, nil)))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	service, err := accounts.New(context.Background(), accounts.NewSQLiteRepository(db), slog.New(slog.NewTextHandler(io.Discard, nil)))
	if err != nil {
		t.Fatal(err)
	}
	return service
}

func TestSetupRequiresTheGeneratedCode(t *testing.T) {
	service := openService(t)
	code := service.SetupCode()
	if code == "" {
		t.Fatal("expected a setup code before any user exists")
	}

	if _, _, err := service.Setup(context.Background(), "1.2.3.4", "wrong-code", "kurt", "password123", "test-agent"); err != accounts.ErrInvalidSetupCode {
		t.Fatalf("wrong code: err=%v", err)
	}

	session, user, err := service.Setup(context.Background(), "1.2.3.4", code, "kurt", "password123", "test-agent")
	if err != nil {
		t.Fatalf("setup: %v", err)
	}
	if user.Username != "kurt" || user.Role != accounts.RoleAdmin || session.Token == "" || session.UserID != user.ID {
		t.Fatalf("setup result: user=%+v session=%+v", user, session)
	}

	// The code is spent, permanently, regardless of future attempts.
	if service.SetupCode() != "" {
		t.Fatal("setup code should be cleared once a user exists")
	}
	if _, _, err := service.Setup(context.Background(), "1.2.3.4", code, "someone-else", "password123", "test-agent"); err != accounts.ErrSetupComplete {
		t.Fatalf("setup after user exists: err=%v", err)
	}
}

func TestSetupValidatesUsernameAndPassword(t *testing.T) {
	service := openService(t)
	code := service.SetupCode()

	if _, _, err := service.Setup(context.Background(), "ip", code, "", "password123", "ua"); err != accounts.ErrUsernameRequired {
		t.Fatalf("blank username: err=%v", err)
	}
	if _, _, err := service.Setup(context.Background(), "ip", code, "kurt", "short", "ua"); err != accounts.ErrWeakPassword {
		t.Fatalf("short password: err=%v", err)
	}
}

func TestLoginAndCurrentUserRoundTrip(t *testing.T) {
	service := openService(t)
	code := service.SetupCode()
	_, user, err := service.Setup(context.Background(), "ip", code, "kurt", "password123", "ua")
	if err != nil {
		t.Fatal(err)
	}

	if _, _, err := service.Login(context.Background(), "ip", "kurt", "wrong-password", "ua"); err != accounts.ErrInvalidCredentials {
		t.Fatalf("wrong password: err=%v", err)
	}
	if _, _, err := service.Login(context.Background(), "ip", "someone-else", "password123", "ua"); err != accounts.ErrInvalidCredentials {
		t.Fatalf("unknown username: err=%v", err)
	}

	session, loggedIn, err := service.Login(context.Background(), "ip", "kurt", "password123", "ua")
	if err != nil || loggedIn.ID != user.ID {
		t.Fatalf("login: user=%+v err=%v", loggedIn, err)
	}

	resolved, resolvedSession, err := service.CurrentUser(context.Background(), session.Token)
	if err != nil || resolved.ID != user.ID || resolvedSession.ID != session.ID {
		t.Fatalf("current user: user=%+v session=%+v err=%v", resolved, resolvedSession, err)
	}

	if _, _, err := service.CurrentUser(context.Background(), "not-a-real-token"); err != accounts.ErrSessionNotFound {
		t.Fatalf("unknown token: err=%v", err)
	}
	if _, _, err := service.CurrentUser(context.Background(), ""); err != accounts.ErrSessionNotFound {
		t.Fatalf("empty token: err=%v", err)
	}
}

func TestLoginIsRateLimited(t *testing.T) {
	service := openService(t)
	code := service.SetupCode()
	if _, _, err := service.Setup(context.Background(), "setup-ip", code, "kurt", "password123", "ua"); err != nil {
		t.Fatal(err)
	}
	var lastErr error
	for i := 0; i < 20; i++ {
		_, _, lastErr = service.Login(context.Background(), "rate-limit-ip", "kurt", "wrong-password", "ua")
		if lastErr == accounts.ErrRateLimited {
			break
		}
	}
	if lastErr != accounts.ErrRateLimited {
		t.Fatalf("expected rate limiting, last err=%v", lastErr)
	}
}

func TestLogoutDeletesTheSession(t *testing.T) {
	service := openService(t)
	code := service.SetupCode()
	session, _, err := service.Setup(context.Background(), "ip", code, "kurt", "password123", "ua")
	if err != nil {
		t.Fatal(err)
	}
	if err := service.Logout(context.Background(), session.Token); err != nil {
		t.Fatal(err)
	}
	if _, _, err := service.CurrentUser(context.Background(), session.Token); err != accounts.ErrSessionNotFound {
		t.Fatalf("session should be gone after logout: err=%v", err)
	}
	// Logging out an unknown token is not an error.
	if err := service.Logout(context.Background(), "unknown-token"); err != nil {
		t.Fatalf("logout of unknown token: %v", err)
	}
}

func TestListAndRevokeSessionsAreScopedToTheUser(t *testing.T) {
	service := openService(t)
	code := service.SetupCode()
	first, user, err := service.Setup(context.Background(), "ip", code, "kurt", "password123", "ua-one")
	if err != nil {
		t.Fatal(err)
	}
	second, _, err := service.Login(context.Background(), "ip", "kurt", "password123", "ua-two")
	if err != nil {
		t.Fatal(err)
	}

	sessions, err := service.ListSessions(context.Background(), user.ID)
	if err != nil || len(sessions) != 2 {
		t.Fatalf("list sessions: got=%+v err=%v", sessions, err)
	}

	if err := service.RevokeSession(context.Background(), user.ID, first.ID); err != nil {
		t.Fatal(err)
	}
	if _, _, err := service.CurrentUser(context.Background(), first.Token); err != accounts.ErrSessionNotFound {
		t.Fatalf("revoked session should be gone: err=%v", err)
	}
	if _, _, err := service.CurrentUser(context.Background(), second.Token); err != nil {
		t.Fatalf("other session should survive: err=%v", err)
	}

	if err := service.RevokeSession(context.Background(), user.ID, second.ID); err != nil {
		t.Fatal(err)
	}
	if err := service.RevokeSession(context.Background(), user.ID, second.ID); err != accounts.ErrSessionNotFound {
		t.Fatalf("revoking an already-revoked session: err=%v", err)
	}
}

func TestExpiredSessionIsRejectedAndDeleted(t *testing.T) {
	db, err := store.Open(filepath.Join(t.TempDir(), "accounts.db"), slog.New(slog.NewTextHandler(io.Discard, nil)))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	repository := accounts.NewSQLiteRepository(db)
	service, err := accounts.New(context.Background(), repository, slog.New(slog.NewTextHandler(io.Discard, nil)))
	if err != nil {
		t.Fatal(err)
	}
	session, user, err := service.Setup(context.Background(), "ip", service.SetupCode(), "kurt", "password123", "ua")
	if err != nil {
		t.Fatal(err)
	}
	if session.ExpiresAt.Before(time.Now().Add(29 * 24 * time.Hour)) {
		t.Fatalf("expected a ~30 day session, got expiry %v", session.ExpiresAt)
	}

	past := time.Now().Add(-time.Hour)
	if err := repository.TouchSession(context.Background(), session.ID, past, past); err != nil {
		t.Fatal(err)
	}
	if _, _, err := service.CurrentUser(context.Background(), session.Token); err != accounts.ErrSessionNotFound {
		t.Fatalf("expired session: err=%v", err)
	}
	sessions, err := service.ListSessions(context.Background(), user.ID)
	if err != nil || len(sessions) != 0 {
		t.Fatalf("expired session should have been deleted on use: sessions=%+v err=%v", sessions, err)
	}
}
