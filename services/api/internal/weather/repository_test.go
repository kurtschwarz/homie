package weather_test

import (
	"context"
	"io"
	"log/slog"
	"path/filepath"
	"testing"
	"time"

	"github.com/kurtschwarz/homie/services/api/internal/store"
	"github.com/kurtschwarz/homie/services/api/internal/weather"
)

func TestSQLiteRepositoryRoundTripAndUpsert(t *testing.T) {
	db, err := store.Open(filepath.Join(t.TempDir(), "weather.db"), slog.New(slog.NewTextHandler(io.Discard, nil)))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	repo := weather.NewSQLiteRepository(db)
	ctx := context.Background()
	location := "43.4422469,-80.4699426"
	entry := weather.CacheEntry{
		ConditionCode: "PartlyCloudy",
		Temperature:   12.75,
		Daylight:      true,
		FetchedAt:     time.Unix(1_800_000_000, 0),
	}
	if err := repo.Save(ctx, location, entry); err != nil {
		t.Fatalf("save first entry (named bind parameters): %v", err)
	}
	got, found, err := repo.Load(ctx, location)
	if err != nil || !found {
		t.Fatalf("load entry: found=%v err=%v", found, err)
	}
	if got != entry {
		t.Fatalf("round trip mismatch: got %+v, want %+v", got, entry)
	}

	entry.ConditionCode = "Rain"
	entry.Temperature = 7.5
	entry.Daylight = false
	if err := repo.Save(ctx, location, entry); err != nil {
		t.Fatalf("upsert entry: %v", err)
	}
	got, found, err = repo.Load(ctx, location)
	if err != nil || !found || got != entry {
		t.Fatalf("upsert result: got=%+v found=%v err=%v, want=%+v", got, found, err, entry)
	}

	if _, found, err := repo.Load(ctx, "missing"); err != nil || found {
		t.Fatalf("missing row: found=%v err=%v, want false, nil", found, err)
	}
}
