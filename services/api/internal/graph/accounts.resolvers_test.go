package graph

import (
	"context"
	"io"
	"log/slog"
	"path/filepath"
	"testing"

	"github.com/kurtschwarz/homie/services/api/internal/accounts"
	"github.com/kurtschwarz/homie/services/api/internal/graph/model"
	"github.com/kurtschwarz/homie/services/api/internal/store"
)

// TestSessionWithoutAccountsReturnsNull keeps the public session query
// valid when account storage is unavailable.
func TestSessionWithoutAccountsReturnsNull(t *testing.T) {
	resolver := &queryResolver{Resolver: &Resolver{}}
	info, err := resolver.Session(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if info != nil {
		t.Fatalf("session=%+v, want nil when accounts are unconfigured", info)
	}
}

func TestHealthStatusReportsSetupRequiredUntilFirstUserExists(t *testing.T) {
	db, err := store.Open(filepath.Join(t.TempDir(), "accounts.db"), slog.New(slog.NewTextHandler(io.Discard, nil)))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })

	service, err := accounts.New(context.Background(), accounts.NewSQLiteRepository(db), slog.New(slog.NewTextHandler(io.Discard, nil)))
	if err != nil {
		t.Fatal(err)
	}
	resolver := &queryResolver{Resolver: &Resolver{Accounts: service}}

	health, err := resolver.Health(context.Background())
	if err != nil || health.Status != model.HealthStatusSetupRequired {
		t.Fatalf("health before setup: status=%s err=%v", health.Status, err)
	}
	if _, _, err := service.Setup(context.Background(), "127.0.0.1", service.SetupCode(), "kurt", "password123", "test"); err != nil {
		t.Fatal(err)
	}

	health, err = resolver.Health(context.Background())
	if err != nil || health.Status != model.HealthStatusOk {
		t.Fatalf("health after setup: status=%s err=%v", health.Status, err)
	}
}
