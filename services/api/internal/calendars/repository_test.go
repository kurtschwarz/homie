package calendars

import (
	"context"
	"io"
	"log/slog"
	"path/filepath"
	"testing"

	"github.com/google/uuid"
	"github.com/kurtschwarz/homie/services/api/internal/store"
)

func TestDiscoveryPreservesCalendarSettingsAndAccountDeleteCascades(t *testing.T) {
	db, err := store.Open(filepath.Join(t.TempDir(), "calendars.db"), slog.New(slog.NewTextHandler(io.Discard, nil)))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	repo := NewSQLiteRepository(db)
	ctx := context.Background()
	accountID := uuid.New()
	account := Account{ID: accountID, Name: "Fastmail", Provider: ProviderCalDAV, CalDAV: &CalDAVConnection{URL: "https://example.test/home/", Username: "me"}}
	applyAccountDefaults(&account)
	if err := repo.UpsertAccount(ctx, storedAccount{Account: account, Credentials: []byte("sealed")}); err != nil {
		t.Fatal(err)
	}
	calendarID := uuid.New()
	first := Calendar{ID: calendarID, AccountID: accountID, ProviderRef: "/home/family/", Name: "Family", Visible: false}
	if err := repo.UpsertDiscoveredCalendar(ctx, first); err != nil {
		t.Fatal(err)
	}
	first.Visible = true
	if err := repo.UpsertCalendar(ctx, first); err != nil {
		t.Fatal(err)
	}
	rediscovered := Calendar{ID: uuid.New(), AccountID: accountID, ProviderRef: first.ProviderRef, Name: "Family renamed"}
	if err := repo.UpsertDiscoveredCalendar(ctx, rediscovered); err != nil {
		t.Fatal(err)
	}
	rows, err := repo.ListCalendars(ctx, &accountID)
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 1 || rows[0].ID != calendarID || !rows[0].Visible || rows[0].Name != "Family renamed" {
		t.Fatalf("discovery reset calendar settings: %+v", rows)
	}
	eventID := uuid.New()
	if _, err := db.ExecContext(ctx, `INSERT INTO events (id,calendar_id,uid,summary,start_at,end_at,all_day,synced_at) VALUES (?,?,'event','Event',0,1,0,0)`, eventID[:], calendarID[:]); err != nil {
		t.Fatal(err)
	}
	if _, err := db.ExecContext(ctx, `INSERT INTO tag_mappings (calendar_id,tag,label,kind) VALUES (?,'family','Family','include')`, calendarID[:]); err != nil {
		t.Fatal(err)
	}
	if err := repo.DeleteAccount(ctx, accountID); err != nil {
		t.Fatal(err)
	}
	rows, err = repo.ListCalendars(ctx, &accountID)
	if err != nil || len(rows) != 0 {
		t.Fatalf("account delete did not cascade: rows=%+v err=%v", rows, err)
	}
	for _, table := range []string{"events", "tag_mappings"} {
		var count int
		if err := db.GetContext(ctx, &count, `SELECT count(*) FROM `+table); err != nil || count != 0 {
			t.Fatalf("%s did not cascade: count=%d err=%v", table, count, err)
		}
	}
}
