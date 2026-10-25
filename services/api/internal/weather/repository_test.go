package weather_test

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"path/filepath"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/kurtschwarz/homie/services/api/internal/store"
	"github.com/kurtschwarz/homie/services/api/internal/weather"
)

func openDB(t *testing.T) *sqlx.DB {
	t.Helper()
	db, err := store.Open(filepath.Join(t.TempDir(), "weather.db"), slog.New(slog.NewTextHandler(io.Discard, nil)))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return db
}

func TestSQLiteRepositoryCacheRoundTripUpsertAndCascade(t *testing.T) {
	db := openDB(t)
	repo := weather.NewSQLiteRepository(db)
	ctx := context.Background()
	location := weather.Location{ID: uuid.New(), Label: "Kitchener", Latitude: 43.45, Longitude: -80.49}
	if err := repo.UpsertLocation(ctx, location); err != nil {
		t.Fatal(err)
	}
	entry := weather.CacheEntry{ConditionCode: "PartlyCloudy", Temperature: 12.75, Daylight: true, FetchedAt: time.Unix(1_800_000_000, 0)}
	if err := repo.Save(ctx, location.ID, entry); err != nil {
		t.Fatal(err)
	}
	got, found, err := repo.Load(ctx, location.ID)
	if err != nil || !found || got != entry {
		t.Fatalf("round trip: got=%+v found=%v err=%v", got, found, err)
	}
	entry.ConditionCode, entry.Temperature = "Rain", 7.5
	if err := repo.Save(ctx, location.ID, entry); err != nil {
		t.Fatal(err)
	}
	if got, _, _ := repo.Load(ctx, location.ID); got != entry {
		t.Fatalf("upsert: got %+v want %+v", got, entry)
	}
	if err := repo.DeleteLocation(ctx, location.ID); err != nil {
		t.Fatal(err)
	}
	if _, found, err := repo.Load(ctx, location.ID); err != nil || found {
		t.Fatalf("cache should cascade with its location: found=%v err=%v", found, err)
	}
}

func TestSQLiteRepositoryProvidersAreUniqueAndSingleActive(t *testing.T) {
	repo := weather.NewSQLiteRepository(openDB(t))
	ctx := context.Background()
	if _, err := repo.GetActiveProvider(ctx); !errors.Is(err, weather.ErrNotFound) {
		t.Fatalf("no active provider: got %v", err)
	}
	stored := weather.NewStoredProviderForTest(uuid.New(), []byte("sealed"))
	if err := repo.UpsertProvider(ctx, stored); err != nil {
		t.Fatal(err)
	}
	if err := repo.UpsertProvider(ctx, weather.NewStoredProviderForTest(uuid.New(), []byte("other"))); !errors.Is(err, weather.ErrDuplicateProvider) {
		t.Fatalf("second provider of the same kind: got %v", err)
	}
	if err := repo.ActivateProvider(ctx, stored.ID); err != nil {
		t.Fatal(err)
	}
	active, err := repo.GetActiveProvider(ctx)
	if err != nil || active.ID != stored.ID || !active.Active {
		t.Fatalf("active provider: %+v err=%v", active, err)
	}
	if err := repo.ActivateProvider(ctx, uuid.New()); !errors.Is(err, weather.ErrNotFound) {
		t.Fatalf("activate unknown provider: got %v", err)
	}
	if active, _ := repo.GetActiveProvider(ctx); active.ID != stored.ID {
		t.Fatal("failed activation must not clear the active provider")
	}
}
