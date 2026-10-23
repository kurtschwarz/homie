package calendars

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"path/filepath"
	"testing"

	"github.com/google/uuid"
	"github.com/kurtschwarz/homie/services/api/internal/store"
)

func TestSetCalendarVisiblePersistsAndReportsMissing(t *testing.T) {
	db, err := store.Open(filepath.Join(t.TempDir(), "calendars.db"), slog.New(slog.NewTextHandler(io.Discard, nil)))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	repo := NewSQLiteRepository(db)
	ctx := context.Background()
	accountID := uuid.New()
	account := Account{
		ID: accountID, Name: "Fastmail", Provider: ProviderCalDAV,
		CalDAV:          &CalDAVConnection{URL: "https://example.test/home/", Username: "me"},
		RefreshInterval: DefaultRefresh, WindowPast: DefaultWindowPast, WindowFuture: DefaultWindowFuture,
	}
	if err := repo.UpsertAccount(ctx, storedAccount{Account: account, Credentials: []byte("sealed")}); err != nil {
		t.Fatal(err)
	}
	calendar := Calendar{ID: uuid.New(), AccountID: accountID, ProviderRef: "/home/family/", Name: "Family"}
	if err := repo.UpsertDiscoveredCalendar(ctx, calendar); err != nil {
		t.Fatal(err)
	}
	discovered, err := repo.GetCalendar(ctx, calendar.ID)
	if err != nil || discovered.Visible {
		t.Fatalf("discovered calendars must start hidden: got=%+v err=%v", discovered, err)
	}
	for _, want := range []bool{true, false} {
		if err := repo.SetCalendarVisible(ctx, calendar.ID, want); err != nil {
			t.Fatal(err)
		}
		got, err := repo.GetCalendar(ctx, calendar.ID)
		if err != nil || got.Visible != want {
			t.Fatalf("visible=%v: got=%+v err=%v", want, got, err)
		}
	}
	if err := repo.SetCalendarVisible(ctx, uuid.New(), true); !errors.Is(err, ErrNotFound) {
		t.Fatalf("unknown calendar: got %v", err)
	}
}
