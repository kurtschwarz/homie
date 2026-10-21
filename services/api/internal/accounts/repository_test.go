package accounts_test

import (
	"context"
	"io"
	"log/slog"
	"path/filepath"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/kurtschwarz/homie/services/api/internal/accounts"
	"github.com/kurtschwarz/homie/services/api/internal/store"
)

func openRepository(t *testing.T) *accounts.SQLiteRepository {
	t.Helper()
	db, err := store.Open(filepath.Join(t.TempDir(), "accounts.db"), slog.New(slog.NewTextHandler(io.Discard, nil)))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return accounts.NewSQLiteRepository(db)
}

func TestCreateFirstUserIsConditionalOnNoneExisting(t *testing.T) {
	repo := openRepository(t)
	ctx := context.Background()

	exists, err := repo.UsersExist(ctx)
	if err != nil || exists {
		t.Fatalf("exists before creation: %v err=%v", exists, err)
	}

	id, err := uuid.NewV7()
	if err != nil {
		t.Fatal(err)
	}
	created, err := repo.CreateFirstUser(ctx, accounts.User{ID: id, Username: "kurt", Role: accounts.RoleAdmin, CreatedAt: time.Now()}, "hash")
	if err != nil || !created {
		t.Fatalf("first create: created=%v err=%v", created, err)
	}

	exists, err = repo.UsersExist(ctx)
	if err != nil || !exists {
		t.Fatalf("exists after creation: %v err=%v", exists, err)
	}

	otherID, err := uuid.NewV7()
	if err != nil {
		t.Fatal(err)
	}
	created, err = repo.CreateFirstUser(ctx, accounts.User{ID: otherID, Username: "someone-else", Role: accounts.RoleAdmin, CreatedAt: time.Now()}, "hash")
	if err != nil || created {
		t.Fatalf("second create should lose the race: created=%v err=%v", created, err)
	}

	user, secret, err := repo.GetUserByUsername(ctx, "kurt")
	if err != nil || user.ID != id || user.Role != accounts.RoleAdmin || secret != "hash" {
		t.Fatalf("get by username: user=%+v secret=%q err=%v", user, secret, err)
	}
	if _, _, err := repo.GetUserByUsername(ctx, "someone-else"); err != accounts.ErrInvalidCredentials {
		t.Fatalf("losing user should not have been created: err=%v", err)
	}
}
