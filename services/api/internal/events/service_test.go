package events_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/kurtschwarz/homie/services/api/internal/caldav"
	"github.com/kurtschwarz/homie/services/api/internal/events"
)

type stubFetcher struct {
	occurrences []caldav.Occurrence
	err         error
}

func (s stubFetcher) Report(context.Context, time.Time, time.Time) ([]caldav.Occurrence, error) {
	return s.occurrences, s.err
}

func seedEvent(t *testing.T, repo events.Repository, uid string, start time.Time, categories []string) {
	t.Helper()
	id, err := uuid.NewRandom()
	if err != nil {
		t.Fatal(err)
	}
	event := events.Event{ID: id, UID: uid, Summary: uid, Start: start, End: start.Add(time.Hour), Categories: categories}
	if err := repo.ReplaceWindow(context.Background(), start.Add(-time.Minute), start.Add(2*time.Hour), []events.Event{event}); err != nil {
		t.Fatal(err)
	}
}

func TestDisplayableExcludesUntaggedAndUnmapped(t *testing.T) {
	db := openDB(t)
	repo := events.NewSQLiteRepository(db)
	service := events.New(repo)
	ctx := context.Background()
	now := time.Now()

	seedEvent(t, repo, "untagged", now.Add(time.Hour), nil)
	seedEvent(t, repo, "unmapped", now.Add(2*time.Hour), []string{"yellow"})

	got, err := service.Displayable(ctx, 21)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 0 {
		t.Fatalf("expected no displayable events, got %+v", got)
	}
}

func TestDisplayablePersonMapping(t *testing.T) {
	db := openDB(t)
	repo := events.NewSQLiteRepository(db)
	service := events.New(repo)
	ctx := context.Background()
	now := time.Now()

	elina, err := uuid.NewRandom()
	if err != nil {
		t.Fatal(err)
	}
	lana, err := uuid.NewRandom()
	if err != nil {
		t.Fatal(err)
	}
	if err := repo.UpsertPerson(ctx, events.Person{ID: elina, Name: "Elina", Swatch: "clay", Order: 1}); err != nil {
		t.Fatal(err)
	}
	if err := repo.UpsertPerson(ctx, events.Person{ID: lana, Name: "Lana", Swatch: "pine", Order: 0}); err != nil {
		t.Fatal(err)
	}
	if err := repo.UpsertTagMapping(ctx, events.TagMapping{Tag: "elina", Label: "Elina", Kind: events.TagKindPerson, PersonID: &elina}); err != nil {
		t.Fatal(err)
	}
	if err := repo.UpsertTagMapping(ctx, events.TagMapping{Tag: "lana", Label: "Lana", Kind: events.TagKindPerson, PersonID: &lana}); err != nil {
		t.Fatal(err)
	}

	seedEvent(t, repo, "one-person", now.Add(time.Hour), []string{" Elina "})
	got, err := service.Displayable(ctx, 21)
	if err != nil || len(got) != 1 || len(got[0].People) != 1 || got[0].People[0].ID != elina {
		t.Fatalf("expected one person match: got=%+v err=%v", got, err)
	}

	seedEvent(t, repo, "two-people", now.Add(3*time.Hour), []string{"Elina", "Lana"})
	got, err = service.Displayable(ctx, 21)
	if err != nil || len(got) != 2 {
		t.Fatalf("expected two events: got=%+v err=%v", got, err)
	}
	var twoPeopleEvent *events.DisplayEvent
	for i := range got {
		if got[i].UID == "two-people" {
			twoPeopleEvent = &got[i]
		}
	}
	if twoPeopleEvent == nil || len(twoPeopleEvent.People) != 2 || twoPeopleEvent.People[0].ID != lana || twoPeopleEvent.People[1].ID != elina {
		t.Fatalf("expected two people ordered by sort_order: got=%+v", twoPeopleEvent)
	}
}

func TestDisplayableIncludeOnlyAttributesToFamily(t *testing.T) {
	db := openDB(t)
	repo := events.NewSQLiteRepository(db)
	service := events.New(repo)
	ctx := context.Background()
	now := time.Now()

	family, err := uuid.NewRandom()
	if err != nil {
		t.Fatal(err)
	}
	if err := repo.UpsertPerson(ctx, events.Person{ID: family, Name: "Family", Swatch: "slate"}); err != nil {
		t.Fatal(err)
	}
	if err := repo.SetFamily(ctx, family); err != nil {
		t.Fatal(err)
	}
	if err := repo.UpsertTagMapping(ctx, events.TagMapping{Tag: "homie", Label: "Homie", Kind: events.TagKindInclude}); err != nil {
		t.Fatal(err)
	}
	seedEvent(t, repo, "homie-event", now.Add(time.Hour), []string{"Homie"})

	got, err := service.Displayable(ctx, 21)
	if err != nil || len(got) != 1 || len(got[0].People) != 1 || got[0].People[0].ID != family {
		t.Fatalf("expected the family person: got=%+v err=%v", got, err)
	}
}

func TestDisplayableIncludeWithoutFamilyYieldsEmptyPeople(t *testing.T) {
	db := openDB(t)
	repo := events.NewSQLiteRepository(db)
	service := events.New(repo)
	ctx := context.Background()
	now := time.Now()

	if err := repo.UpsertTagMapping(ctx, events.TagMapping{Tag: "homie", Label: "Homie", Kind: events.TagKindInclude}); err != nil {
		t.Fatal(err)
	}
	seedEvent(t, repo, "homie-event", now.Add(time.Hour), []string{"Homie"})

	got, err := service.Displayable(ctx, 21)
	if err != nil || len(got) != 1 {
		t.Fatalf("expected the event to still appear: got=%+v err=%v", got, err)
	}
	if len(got[0].People) != 0 {
		t.Fatalf("expected empty people with no family row configured: %+v", got[0])
	}
}

func TestDisplayablePersonAndIncludeKeepsOnlyPerson(t *testing.T) {
	db := openDB(t)
	repo := events.NewSQLiteRepository(db)
	service := events.New(repo)
	ctx := context.Background()
	now := time.Now()

	elina, err := uuid.NewRandom()
	if err != nil {
		t.Fatal(err)
	}
	family, err := uuid.NewRandom()
	if err != nil {
		t.Fatal(err)
	}
	if err := repo.UpsertPerson(ctx, events.Person{ID: elina, Name: "Elina", Swatch: "clay"}); err != nil {
		t.Fatal(err)
	}
	if err := repo.UpsertPerson(ctx, events.Person{ID: family, Name: "Family", Swatch: "slate"}); err != nil {
		t.Fatal(err)
	}
	if err := repo.SetFamily(ctx, family); err != nil {
		t.Fatal(err)
	}
	if err := repo.UpsertTagMapping(ctx, events.TagMapping{Tag: "elina", Label: "Elina", Kind: events.TagKindPerson, PersonID: &elina}); err != nil {
		t.Fatal(err)
	}
	if err := repo.UpsertTagMapping(ctx, events.TagMapping{Tag: "homie", Label: "Homie", Kind: events.TagKindInclude}); err != nil {
		t.Fatal(err)
	}
	seedEvent(t, repo, "both", now.Add(time.Hour), []string{"Elina", "Homie"})

	got, err := service.Displayable(ctx, 21)
	if err != nil || len(got) != 1 || len(got[0].People) != 1 || got[0].People[0].ID != elina {
		t.Fatalf("expected only the person, not the family: got=%+v err=%v", got, err)
	}
}

func TestDisplayableDaysBoundsTheWindow(t *testing.T) {
	db := openDB(t)
	repo := events.NewSQLiteRepository(db)
	service := events.New(repo)
	ctx := context.Background()
	now := time.Now()

	elina, err := uuid.NewRandom()
	if err != nil {
		t.Fatal(err)
	}
	if err := repo.UpsertPerson(ctx, events.Person{ID: elina, Name: "Elina", Swatch: "clay"}); err != nil {
		t.Fatal(err)
	}
	if err := repo.UpsertTagMapping(ctx, events.TagMapping{Tag: "elina", Label: "Elina", Kind: events.TagKindPerson, PersonID: &elina}); err != nil {
		t.Fatal(err)
	}
	seedEvent(t, repo, "far", now.AddDate(0, 0, 30), []string{"Elina"})

	got, err := service.Displayable(ctx, 21)
	if err != nil || len(got) != 0 {
		t.Fatalf("expected the far event to be excluded by a 21-day window: got=%+v err=%v", got, err)
	}
	got, err = service.Displayable(ctx, 31)
	if err != nil || len(got) != 1 {
		t.Fatalf("expected the far event with a wider window: got=%+v err=%v", got, err)
	}
}

func TestSavePersonRejectsUnknownSwatch(t *testing.T) {
	db := openDB(t)
	repo := events.NewSQLiteRepository(db)
	service := events.New(repo)
	ctx := context.Background()
	if _, err := service.SavePerson(ctx, events.Person{Name: "A", Swatch: "scarlet"}); err != events.ErrUnknownSwatch {
		t.Fatalf("expected ErrUnknownSwatch, got %v", err)
	}
}

func TestSavePersonCreateAndUpdate(t *testing.T) {
	db := openDB(t)
	repo := events.NewSQLiteRepository(db)
	service := events.New(repo)
	ctx := context.Background()

	created, err := service.SavePerson(ctx, events.Person{Name: "A", Swatch: "clay"})
	if err != nil || created.ID == uuid.Nil {
		t.Fatalf("create: got=%+v err=%v", created, err)
	}
	created.Name = "A Renamed"
	updated, err := service.SavePerson(ctx, created)
	if err != nil || updated.ID != created.ID || updated.Name != "A Renamed" {
		t.Fatalf("update: got=%+v err=%v", updated, err)
	}
}

func TestSavePersonMovesFamilyFlag(t *testing.T) {
	db := openDB(t)
	repo := events.NewSQLiteRepository(db)
	service := events.New(repo)
	ctx := context.Background()

	first, err := service.SavePerson(ctx, events.Person{Name: "A", Swatch: "clay", IsFamily: true})
	if err != nil {
		t.Fatal(err)
	}
	second, err := service.SavePerson(ctx, events.Person{Name: "B", Swatch: "pine", IsFamily: true})
	if err != nil {
		t.Fatal(err)
	}
	people, err := repo.ListPeople(ctx)
	if err != nil {
		t.Fatal(err)
	}
	for _, p := range people {
		if p.ID == first.ID && p.IsFamily {
			t.Fatalf("expected the first person to lose the family flag: %+v", people)
		}
		if p.ID == second.ID && !p.IsFamily {
			t.Fatalf("expected the second person to hold the family flag: %+v", people)
		}
	}
}

func TestSaveTagMappingNormalisesAndPreservesLabel(t *testing.T) {
	db := openDB(t)
	repo := events.NewSQLiteRepository(db)
	service := events.New(repo)
	ctx := context.Background()
	elina, err := uuid.NewRandom()
	if err != nil {
		t.Fatal(err)
	}
	if err := repo.UpsertPerson(ctx, events.Person{ID: elina, Name: "Elina", Swatch: "clay"}); err != nil {
		t.Fatal(err)
	}
	mapping, err := service.SaveTagMapping(ctx, " Elina ", events.TagKindPerson, &elina)
	if err != nil {
		t.Fatal(err)
	}
	if mapping.Tag != "elina" || mapping.Label != "Elina" {
		t.Fatalf("expected normalised tag and preserved label: %+v", mapping)
	}
}

func TestSaveTagMappingPersonKindRequiresPersonID(t *testing.T) {
	db := openDB(t)
	repo := events.NewSQLiteRepository(db)
	service := events.New(repo)
	ctx := context.Background()
	if _, err := service.SaveTagMapping(ctx, "elina", events.TagKindPerson, nil); err != events.ErrPersonMappingNeedsPerson {
		t.Fatalf("expected ErrPersonMappingNeedsPerson, got %v", err)
	}
}

func TestRefreshStoresUntaggedAndTaggedEvents(t *testing.T) {
	repo, service := newEventsService(t)
	ctx := context.Background()
	now := time.Now()
	service.Client = stubFetcher{occurrences: []caldav.Occurrence{
		{UID: "tagged", Summary: "Tagged", Start: now, End: now.Add(time.Hour), Categories: []string{"elina"}},
		{UID: "untagged", Summary: "Untagged", Start: now, End: now.Add(time.Hour)},
	}}
	if err := service.Refresh(ctx); err != nil {
		t.Fatal(err)
	}
	got, err := repo.ListWindow(ctx, now.Add(-events.WindowPast), now.Add(events.WindowFuture))
	if err != nil || len(got) != 2 {
		t.Fatalf("expected both the tagged and untagged event to be stored: got=%+v err=%v", got, err)
	}
}

func TestRefreshIsIdempotent(t *testing.T) {
	repo, service := newEventsService(t)
	ctx := context.Background()
	now := time.Now()
	service.Client = stubFetcher{occurrences: []caldav.Occurrence{
		{UID: "one", Summary: "One", Start: now, End: now.Add(time.Hour)},
	}}
	if err := service.Refresh(ctx); err != nil {
		t.Fatal(err)
	}
	if err := service.Refresh(ctx); err != nil {
		t.Fatal(err)
	}
	got, err := repo.ListWindow(ctx, now.Add(-events.WindowPast), now.Add(events.WindowFuture))
	if err != nil || len(got) != 1 {
		t.Fatalf("expected exactly one row after two refreshes: got=%+v err=%v", got, err)
	}
}

func TestRefreshDeletesEventsRemovedUpstream(t *testing.T) {
	repo, service := newEventsService(t)
	ctx := context.Background()
	now := time.Now()
	service.Client = stubFetcher{occurrences: []caldav.Occurrence{
		{UID: "gone-soon", Summary: "Gone soon", Start: now, End: now.Add(time.Hour)},
	}}
	if err := service.Refresh(ctx); err != nil {
		t.Fatal(err)
	}
	service.Client = stubFetcher{occurrences: nil}
	if err := service.Refresh(ctx); err != nil {
		t.Fatal(err)
	}
	got, err := repo.ListWindow(ctx, now.Add(-events.WindowPast), now.Add(events.WindowFuture))
	if err != nil || len(got) != 0 {
		t.Fatalf("expected the deleted event to disappear: got=%+v err=%v", got, err)
	}
}

func TestRefreshUpdatesCategoriesOnRetag(t *testing.T) {
	repo, service := newEventsService(t)
	ctx := context.Background()
	now := time.Now()
	service.Client = stubFetcher{occurrences: []caldav.Occurrence{
		{UID: "retag", Summary: "Retag", Start: now, End: now.Add(time.Hour), Categories: []string{"yellow"}},
	}}
	if err := service.Refresh(ctx); err != nil {
		t.Fatal(err)
	}
	service.Client = stubFetcher{occurrences: []caldav.Occurrence{
		{UID: "retag", Summary: "Retag", Start: now, End: now.Add(time.Hour), Categories: []string{"elina"}},
	}}
	if err := service.Refresh(ctx); err != nil {
		t.Fatal(err)
	}
	got, err := repo.ListWindow(ctx, now.Add(-events.WindowPast), now.Add(events.WindowFuture))
	if err != nil || len(got) != 1 || len(got[0].Categories) != 1 || got[0].Categories[0] != "elina" {
		t.Fatalf("expected the new categories after refresh: got=%+v err=%v", got, err)
	}
}

func TestRefreshLeavesRowsOutsideTheWindowUntouched(t *testing.T) {
	repo, service := newEventsService(t)
	ctx := context.Background()
	now := time.Now()
	seedEvent(t, repo, "far-future", now.AddDate(0, 0, 60), nil)
	service.Client = stubFetcher{occurrences: nil}
	if err := service.Refresh(ctx); err != nil {
		t.Fatal(err)
	}
	got, err := repo.ListWindow(ctx, now.AddDate(0, 0, 59), now.AddDate(0, 0, 61))
	if err != nil || len(got) != 1 || got[0].UID != "far-future" {
		t.Fatalf("expected the far-future row to survive: got=%+v err=%v", got, err)
	}
}

func TestRefreshIsANoOpWithoutAClient(t *testing.T) {
	_, service := newEventsService(t)
	if err := service.Refresh(context.Background()); err != nil {
		t.Fatalf("expected Refresh to be a no-op without a client, got %v", err)
	}
}

func newEventsService(t *testing.T) (events.Repository, *events.Service) {
	t.Helper()
	db := openDB(t)
	repo := events.NewSQLiteRepository(db)
	return repo, events.New(repo)
}
