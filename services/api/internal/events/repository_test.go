package events_test

import (
	"context"
	"io"
	"log/slog"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/kurtschwarz/homie/services/api/internal/events"
	"github.com/kurtschwarz/homie/services/api/internal/store"
)

func openDB(t *testing.T) *sqlx.DB {
	t.Helper()
	db, err := store.Open(filepath.Join(t.TempDir(), "events.db"), slog.New(slog.NewTextHandler(io.Discard, nil)))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return db
}

func TestUpsertPersonRoundTrip(t *testing.T) {
	db := openDB(t)
	repo := events.NewSQLiteRepository(db)
	ctx := context.Background()
	id := uuid.MustParse("0192a780-0a00-7000-8000-000000000001")
	person := events.Person{ID: id, Name: "Lana", Emoji: "🦊", Swatch: "clay", Order: 1}
	if err := repo.UpsertPerson(ctx, person); err != nil {
		t.Fatal(err)
	}
	got, err := repo.ListPeople(ctx)
	if err != nil || len(got) != 1 {
		t.Fatalf("list: got=%+v err=%v", got, err)
	}
	if got[0] != person {
		t.Fatalf("round trip mismatch: got %+v, want %+v", got[0], person)
	}
}

func TestPeopleFamilyIsUnique(t *testing.T) {
	db := openDB(t)
	repo := events.NewSQLiteRepository(db)
	ctx := context.Background()
	a := uuid.MustParse("0192a780-0a00-7000-8000-000000000001")
	b := uuid.MustParse("0192a780-0a00-7000-8000-000000000002")
	if err := repo.UpsertPerson(ctx, events.Person{ID: a, Name: "A", Swatch: "clay"}); err != nil {
		t.Fatal(err)
	}
	if err := repo.SetFamily(ctx, a); err != nil {
		t.Fatal(err)
	}
	if _, err := db.ExecContext(ctx, `INSERT INTO people (id, name, emoji, swatch, sort_order, is_family) VALUES (?, 'B', '', 'pine', 0, 1)`, b[:]); err == nil {
		t.Fatal("expected a second is_family row to be rejected")
	}
}

func TestSetFamilyMovesTheFlag(t *testing.T) {
	db := openDB(t)
	repo := events.NewSQLiteRepository(db)
	ctx := context.Background()
	a := uuid.MustParse("0192a780-0a00-7000-8000-000000000001")
	b := uuid.MustParse("0192a780-0a00-7000-8000-000000000002")
	if err := repo.UpsertPerson(ctx, events.Person{ID: a, Name: "A", Swatch: "clay", IsFamily: true}); err != nil {
		t.Fatal(err)
	}
	if err := repo.UpsertPerson(ctx, events.Person{ID: b, Name: "B", Swatch: "pine"}); err != nil {
		t.Fatal(err)
	}
	if err := repo.SetFamily(ctx, b); err != nil {
		t.Fatal(err)
	}
	people, err := repo.ListPeople(ctx)
	if err != nil {
		t.Fatal(err)
	}
	familyCount := 0
	for _, p := range people {
		if p.IsFamily {
			familyCount++
			if p.ID != b {
				t.Fatalf("family flag on wrong person: %+v", p)
			}
		}
	}
	if familyCount != 1 {
		t.Fatalf("expected exactly one family person, got %d", familyCount)
	}
}

func TestDeletePersonCascadesTagMappings(t *testing.T) {
	db := openDB(t)
	repo := events.NewSQLiteRepository(db)
	ctx := context.Background()
	id := uuid.MustParse("0192a780-0a00-7000-8000-000000000001")
	if err := repo.UpsertPerson(ctx, events.Person{ID: id, Name: "Elina", Swatch: "clay"}); err != nil {
		t.Fatal(err)
	}
	if err := repo.UpsertTagMapping(ctx, events.TagMapping{Tag: "elina", Label: "Elina", Kind: events.TagKindPerson, PersonID: &id}); err != nil {
		t.Fatal(err)
	}
	if err := repo.DeletePerson(ctx, id); err != nil {
		t.Fatal(err)
	}
	mappings, err := repo.ListTagMappings(ctx)
	if err != nil || len(mappings) != 0 {
		t.Fatalf("expected mappings to cascade away, got %+v err=%v", mappings, err)
	}
}

func TestTagMappingsRejectInconsistentKindAndPerson(t *testing.T) {
	db := openDB(t)
	ctx := context.Background()
	id := uuid.MustParse("0192a780-0a00-7000-8000-000000000001")
	if _, err := db.ExecContext(ctx, `INSERT INTO tag_mappings (tag, label, kind, person_id) VALUES ('x', 'X', 'person', NULL)`); err == nil {
		t.Fatal("expected kind=person with no person_id to be rejected")
	}
	if _, err := db.ExecContext(ctx, `INSERT INTO tag_mappings (tag, label, kind, person_id) VALUES ('x', 'X', 'include', ?)`, id[:]); err == nil {
		t.Fatal("expected kind=include with a person_id to be rejected")
	}
}

func TestReplaceWindow(t *testing.T) {
	db := openDB(t)
	repo := events.NewSQLiteRepository(db)
	ctx := context.Background()
	base := time.Unix(1_760_000_000, 0)
	windowFrom, windowTo := base, base.Add(48*time.Hour)

	outside := events.Event{
		ID: uuid.MustParse("0192a780-0a00-7000-8000-000000000099"), UID: "outside",
		Summary: "Outside", Start: base.Add(-240 * time.Hour), End: base.Add(-239 * time.Hour),
	}
	if err := repo.ReplaceWindow(ctx, windowFrom.Add(-time.Hour), windowFrom, []events.Event{outside}); err != nil {
		t.Fatal(err)
	}

	firstID := uuid.MustParse("0192a780-0a00-7000-8000-000000000001")
	first := events.Event{
		ID: firstID, UID: "abc", Summary: "First", Location: "Home",
		Start: base.Add(time.Hour), End: base.Add(2 * time.Hour), Categories: []string{"elina"},
	}
	if err := repo.ReplaceWindow(ctx, windowFrom, windowTo, []events.Event{first}); err != nil {
		t.Fatal(err)
	}
	got, err := repo.ListWindow(ctx, windowFrom, windowTo)
	if err != nil || len(got) != 1 || got[0].UID != "abc" {
		t.Fatalf("after first replace: got=%+v err=%v", got, err)
	}

	second := events.Event{
		ID: uuid.MustParse("0192a780-0a00-7000-8000-000000000002"), UID: "abc", Summary: "First renamed",
		Start: base.Add(time.Hour), End: base.Add(2 * time.Hour), Categories: []string{"elina"},
	}
	if err := repo.ReplaceWindow(ctx, windowFrom, windowTo, []events.Event{second}); err != nil {
		t.Fatal(err)
	}
	got, err = repo.ListWindow(ctx, windowFrom, windowTo)
	if err != nil || len(got) != 1 {
		t.Fatalf("after second replace: got=%+v err=%v", got, err)
	}
	if got[0].ID != firstID {
		t.Fatalf("id not preserved for unchanged (uid, recurrence_id): got %s, want %s", got[0].ID, firstID)
	}
	if got[0].Summary != "First renamed" {
		t.Fatalf("summary not updated: %+v", got[0])
	}

	if err := repo.ReplaceWindow(ctx, windowFrom, windowTo, nil); err != nil {
		t.Fatal(err)
	}
	got, err = repo.ListWindow(ctx, windowFrom, windowTo)
	if err != nil || len(got) != 0 {
		t.Fatalf("empty replace should empty the window: got=%+v err=%v", got, err)
	}

	stillThere, err := repo.ListWindow(ctx, windowFrom.Add(-240*time.Hour-time.Hour), windowFrom.Add(-238*time.Hour))
	if err != nil || len(stillThere) != 1 || stillThere[0].UID != "outside" {
		t.Fatalf("rows outside the window should survive: got=%+v err=%v", stillThere, err)
	}
}

func TestListWindowIncludesEventStartingBeforeFrom(t *testing.T) {
	db := openDB(t)
	repo := events.NewSQLiteRepository(db)
	ctx := context.Background()
	base := time.Unix(1_760_000_000, 0)
	running := events.Event{
		ID: uuid.MustParse("0192a780-0a00-7000-8000-000000000001"), UID: "running",
		Summary: "Running now", Start: base.Add(-time.Hour), End: base.Add(time.Hour),
	}
	if err := repo.ReplaceWindow(ctx, base.Add(-time.Hour), base.Add(time.Hour), []events.Event{running}); err != nil {
		t.Fatal(err)
	}
	got, err := repo.ListWindow(ctx, base, base.Add(2*time.Hour))
	if err != nil || len(got) != 1 || got[0].UID != "running" {
		t.Fatalf("expected the still-running event to be included: got=%+v err=%v", got, err)
	}
}

func TestListWindowOrdersByStartAt(t *testing.T) {
	db := openDB(t)
	repo := events.NewSQLiteRepository(db)
	ctx := context.Background()
	base := time.Unix(1_760_000_000, 0)
	e1 := events.Event{ID: uuid.MustParse("0192a780-0a00-7000-8000-000000000001"), UID: "a", Summary: "A", Start: base.Add(2 * time.Hour), End: base.Add(3 * time.Hour)}
	e2 := events.Event{ID: uuid.MustParse("0192a780-0a00-7000-8000-000000000002"), UID: "b", Summary: "B", Start: base.Add(time.Hour), End: base.Add(2 * time.Hour)}
	if err := repo.ReplaceWindow(ctx, base, base.Add(4*time.Hour), []events.Event{e1, e2}); err != nil {
		t.Fatal(err)
	}
	got, err := repo.ListWindow(ctx, base, base.Add(4*time.Hour))
	if err != nil || len(got) != 2 || got[0].UID != "b" || got[1].UID != "a" {
		t.Fatalf("expected ordering by start_at: got=%+v err=%v", got, err)
	}
}

func TestUniqueUIDWithEmptyRecurrenceID(t *testing.T) {
	db := openDB(t)
	ctx := context.Background()
	uuid1 := uuid.MustParse("0192a780-0a00-7000-8000-000000000001")
	uuid2 := uuid.MustParse("0192a780-0a00-7000-8000-000000000002")
	id1 := uuid1[:]
	id2 := uuid2[:]
	if _, err := db.ExecContext(ctx, `INSERT INTO events (id, uid, recurrence_id, summary, start_at, end_at, all_day, synced_at) VALUES (?, 'dup', '', 'A', 0, 1, 0, 0)`, id1); err != nil {
		t.Fatal(err)
	}
	if _, err := db.ExecContext(ctx, `INSERT INTO events (id, uid, recurrence_id, summary, start_at, end_at, all_day, synced_at) VALUES (?, 'dup', '', 'B', 0, 1, 0, 0)`, id2); err == nil {
		t.Fatal("expected a duplicate (uid, '') to be rejected")
	}
}

func TestDistinctCategoriesDeduplicates(t *testing.T) {
	db := openDB(t)
	repo := events.NewSQLiteRepository(db)
	ctx := context.Background()
	base := time.Unix(1_760_000_000, 0)
	e1 := events.Event{ID: uuid.MustParse("0192a780-0a00-7000-8000-000000000001"), UID: "a", Summary: "A", Start: base, End: base.Add(time.Hour), Categories: []string{"elina", "yellow"}}
	e2 := events.Event{ID: uuid.MustParse("0192a780-0a00-7000-8000-000000000002"), UID: "b", Summary: "B", Start: base, End: base.Add(time.Hour), Categories: []string{"elina"}}
	if err := repo.ReplaceWindow(ctx, base, base.Add(2*time.Hour), []events.Event{e1, e2}); err != nil {
		t.Fatal(err)
	}
	got, err := repo.DistinctCategories(ctx)
	if err != nil {
		t.Fatal(err)
	}
	joined := strings.Join(got, ",")
	if len(got) != 2 || !strings.Contains(joined, "elina") || !strings.Contains(joined, "yellow") {
		t.Fatalf("expected deduplicated categories, got %+v", got)
	}
}
