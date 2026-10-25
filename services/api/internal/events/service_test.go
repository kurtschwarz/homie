package events_test

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/kurtschwarz/homie/services/api/internal/calendars"
	"github.com/kurtschwarz/homie/services/api/internal/calendars/caldav"
	"github.com/kurtschwarz/homie/services/api/internal/events"
)

type stubFetcher struct {
	occurrences []caldav.Occurrence
	err         error
}

func (s stubFetcher) Report(context.Context, time.Time, time.Time) ([]caldav.Occurrence, error) {
	return s.occurrences, s.err
}

// stubWriter is a Fetcher+Writer whose Put appends to the occurrences the
// next Report (and therefore the next Refresh) will return, standing in for
// a round trip through a real CalDAV server.
type stubWriter struct {
	occurrences *[]caldav.Occurrence
	putErr      error
}

type stubCalendarClient struct {
	mu          sync.Mutex
	occurrences []calendars.Occurrence
	err         error
	reports     int
}

func (s *stubCalendarClient) Report(context.Context, time.Time, time.Time) ([]calendars.Occurrence, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.reports++
	return s.occurrences, s.err
}
func (s *stubCalendarClient) Put(context.Context, calendars.Occurrence) error { return nil }
func (s *stubCalendarClient) reportCount() int                                { s.mu.Lock(); defer s.mu.Unlock(); return s.reports }

type stubSourceProvider struct {
	mu                  sync.Mutex
	sources             []calendars.Source
	byID                map[uuid.UUID]calendars.Source
	successes, failures map[uuid.UUID]int
}

func (s *stubSourceProvider) Sources(context.Context) ([]calendars.Source, error) {
	return s.sources, nil
}
func (s *stubSourceProvider) Source(_ context.Context, id uuid.UUID) (calendars.Source, error) {
	source, ok := s.byID[id]
	if !ok {
		return calendars.Source{}, calendars.ErrNotFound
	}
	return source, nil
}
func (s *stubSourceProvider) RecordSuccess(_ context.Context, id uuid.UUID, _ time.Time) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.successes[id]++
	return nil
}
func (s *stubSourceProvider) RecordFailure(_ context.Context, id uuid.UUID, _ string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.failures[id]++
	return nil
}
func (s *stubSourceProvider) RecordAttempt(_ context.Context, id uuid.UUID, at time.Time) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for index := range s.sources {
		if s.sources[index].Account.ID == id {
			s.sources[index].Account.LastAttemptedAt = &at
		}
	}
	for calendarID, source := range s.byID {
		if source.Account.ID == id {
			source.Account.LastAttemptedAt = &at
			s.byID[calendarID] = source
		}
	}
	return nil
}

func (s stubWriter) Report(context.Context, time.Time, time.Time) ([]caldav.Occurrence, error) {
	return *s.occurrences, nil
}

func (s stubWriter) Put(_ context.Context, occ caldav.Occurrence) error {
	if s.putErr != nil {
		return s.putErr
	}
	*s.occurrences = append(*s.occurrences, occ)
	return nil
}

func associate(t *testing.T, repo events.Repository, personID uuid.UUID, scope events.AssociationScope, labels ...string) {
	t.Helper()
	association := events.Association{CalendarID: repositoryTestCalendarID, PersonID: personID, Scope: scope, Terms: events.ParseTerms(labels)}
	if err := repo.InsertAssociation(context.Background(), association); err != nil {
		t.Fatal(err)
	}
}

func seedPerson(t *testing.T, repo events.Repository, name string, order int) uuid.UUID {
	t.Helper()
	id := uuid.New()
	if err := repo.UpsertPerson(context.Background(), events.Person{ID: id, Name: name, Swatch: "clay", Order: order}); err != nil {
		t.Fatal(err)
	}
	return id
}

func seedTitledEvent(t *testing.T, repo events.Repository, title string, start time.Time, categories []string) {
	t.Helper()
	event := events.Event{ID: uuid.New(), UID: title, Summary: title, Start: start, End: start.Add(time.Hour), Categories: categories}
	existing, err := repo.ListWindow(context.Background(), time.Unix(0, 0), time.Now().AddDate(100, 0, 0))
	if err != nil {
		t.Fatal(err)
	}
	if err := repo.ReplaceWindow(context.Background(), repositoryTestCalendarID, start.Add(-time.Minute), start.Add(2*time.Hour), append(existing, event)); err != nil {
		t.Fatal(err)
	}
}

func seedEvent(t *testing.T, repo events.Repository, uid string, start time.Time, categories []string) {
	t.Helper()
	id, err := uuid.NewRandom()
	if err != nil {
		t.Fatal(err)
	}
	event := events.Event{ID: id, UID: uid, Summary: uid, Start: start, End: start.Add(time.Hour), Categories: categories}
	existing, err := repo.ListWindow(context.Background(), time.Unix(0, 0), time.Now().AddDate(100, 0, 0))
	if err != nil {
		t.Fatal(err)
	}
	if err := repo.ReplaceWindow(context.Background(), repositoryTestCalendarID, start.Add(-time.Minute), start.Add(2*time.Hour), append(existing, event)); err != nil {
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
	associate(t, repo, elina, events.ScopeMatched, "#elina")
	associate(t, repo, lana, events.ScopeMatched, "#lana")

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

func TestDisplayableAssociatedFamilyPerson(t *testing.T) {
	db := openDB(t)
	repo := events.NewSQLiteRepository(db)
	service := events.New(repo)
	ctx := context.Background()

	family := seedPerson(t, repo, "Family", 0)
	if err := repo.SetFamily(ctx, family); err != nil {
		t.Fatal(err)
	}
	associate(t, repo, family, events.ScopeMatched, "#homie")
	seedEvent(t, repo, "homie-event", time.Now().Add(time.Hour), []string{"Homie"})

	got, err := service.Displayable(ctx, 21)
	if err != nil || len(got) != 1 || len(got[0].People) != 1 || got[0].People[0].ID != family {
		t.Fatalf("expected the family person: got=%+v err=%v", got, err)
	}
}

func TestDisplayableEventMatchingTwoAssociationsShowsBothPeople(t *testing.T) {
	db := openDB(t)
	repo := events.NewSQLiteRepository(db)
	service := events.New(repo)
	ctx := context.Background()

	elina := seedPerson(t, repo, "Elina", 1)
	family := seedPerson(t, repo, "Family", 0)
	associate(t, repo, elina, events.ScopeMatched, "#elina")
	associate(t, repo, family, events.ScopeMatched, "#homie")
	seedEvent(t, repo, "both", time.Now().Add(time.Hour), []string{"Elina", "Homie"})

	got, err := service.Displayable(ctx, 21)
	if err != nil || len(got) != 1 || len(got[0].People) != 2 || got[0].People[0].ID != family || got[0].People[1].ID != elina {
		t.Fatalf("expected both people ordered by sort_order: got=%+v err=%v", got, err)
	}
}

func TestDisplayablePeopleCanShareATag(t *testing.T) {
	db := openDB(t)
	repo := events.NewSQLiteRepository(db)
	service := events.New(repo)
	ctx := context.Background()

	first := seedPerson(t, repo, "First", 0)
	second := seedPerson(t, repo, "Second", 1)
	associate(t, repo, first, events.ScopeMatched, "#swim")
	associate(t, repo, second, events.ScopeMatched, "#swim")
	seedEvent(t, repo, "swim-meet", time.Now().Add(time.Hour), []string{"swim"})

	got, err := service.Displayable(ctx, 21)
	if err != nil || len(got) != 1 || len(got[0].People) != 2 {
		t.Fatalf("expected both people on the shared tag: got=%+v err=%v", got, err)
	}
}

func TestDisplayableWordTermMatchesWholeWordsInTitle(t *testing.T) {
	db := openDB(t)
	repo := events.NewSQLiteRepository(db)
	service := events.New(repo)
	ctx := context.Background()
	now := time.Now()

	lana := seedPerson(t, repo, "Lana", 0)
	associate(t, repo, lana, events.ScopeMatched, "Lana")
	seedTitledEvent(t, repo, "Lana's swim lesson", now.Add(time.Hour), nil)
	seedTitledEvent(t, repo, "Flight to Atlanta", now.Add(3*time.Hour), nil)
	seedTitledEvent(t, repo, "LANA dentist", now.Add(5*time.Hour), nil)

	got, err := service.Displayable(ctx, 21)
	if err != nil {
		t.Fatal(err)
	}
	titles := map[string]bool{}
	for _, event := range got {
		titles[event.Summary] = true
	}
	if len(got) != 2 || !titles["Lana's swim lesson"] || !titles["LANA dentist"] || titles["Flight to Atlanta"] {
		t.Fatalf("expected whole-word title matches only: got=%+v", got)
	}
}

func TestDisplayableAllScopeClaimsEveryEventAndKeepsTermsDormant(t *testing.T) {
	db := openDB(t)
	repo := events.NewSQLiteRepository(db)
	service := events.New(repo)
	ctx := context.Background()

	lana := seedPerson(t, repo, "Lana", 0)
	associate(t, repo, lana, events.ScopeMatched, "#lana")
	seedEvent(t, repo, "unrelated", time.Now().Add(time.Hour), nil)

	got, err := service.Displayable(ctx, 21)
	if err != nil || len(got) != 0 {
		t.Fatalf("matched scope must not claim an unrelated event: got=%+v err=%v", got, err)
	}

	all := events.ScopeAll
	if _, err := service.UpdateAssociation(ctx, repositoryTestCalendarID, lana, &all, nil); err != nil {
		t.Fatal(err)
	}
	got, err = service.Displayable(ctx, 21)
	if err != nil || len(got) != 1 || got[0].People[0].ID != lana {
		t.Fatalf("all scope must claim every event: got=%+v err=%v", got, err)
	}

	associations, err := repo.ListAssociations(ctx)
	if err != nil || len(associations) != 1 || len(associations[0].Terms) != 1 || associations[0].Terms[0].Term != "lana" {
		t.Fatalf("switching scope must leave the terms alone: got=%+v err=%v", associations, err)
	}
}

func TestDisplayableIgnoresAssociationsOnOtherCalendars(t *testing.T) {
	db := openDB(t)
	repo := events.NewSQLiteRepository(db)
	service := events.New(repo)
	ctx := context.Background()

	other := uuid.MustParse("00000000-0000-7000-8000-000000000021")
	if _, err := db.ExecContext(ctx, `INSERT INTO calendars (id,account_id,provider_ref,name,visible) VALUES (?, ?, '/other/', 'Other', 1)`, other[:], repositoryTestCalendarID[:]); err != nil {
		t.Fatal(err)
	}
	lana := seedPerson(t, repo, "Lana", 0)
	if err := repo.InsertAssociation(ctx, events.Association{CalendarID: other, PersonID: lana, Scope: events.ScopeAll}); err != nil {
		t.Fatal(err)
	}
	seedEvent(t, repo, "on-the-test-calendar", time.Now().Add(time.Hour), []string{"lana"})

	got, err := service.Displayable(ctx, 21)
	if err != nil || len(got) != 0 {
		t.Fatalf("an association is scoped to its own calendar: got=%+v err=%v", got, err)
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
	associate(t, repo, elina, events.ScopeMatched, "#elina")
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

func TestSavePersonAppendsNewPeopleAndKeepsOrderOnUpdate(t *testing.T) {
	db := openDB(t)
	repo := events.NewSQLiteRepository(db)
	service := events.New(repo)
	ctx := context.Background()

	a, err := service.SavePerson(ctx, events.Person{Name: "A", Swatch: "clay"})
	if err != nil {
		t.Fatal(err)
	}
	b, err := service.SavePerson(ctx, events.Person{Name: "B", Swatch: "pine"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := service.ReorderPeople(ctx, []uuid.UUID{b.ID, a.ID}); err != nil {
		t.Fatal(err)
	}
	a.Name = "A Renamed"
	a.Order = 0
	if _, err := service.SavePerson(ctx, a); err != nil {
		t.Fatal(err)
	}
	people, err := repo.ListPeople(ctx)
	if err != nil || len(people) != 2 || people[0].ID != b.ID || people[1].ID != a.ID {
		t.Fatalf("expected a rename to keep the reordered list: got=%+v err=%v", people, err)
	}
}

func TestReorderPeople(t *testing.T) {
	db := openDB(t)
	repo := events.NewSQLiteRepository(db)
	service := events.New(repo)
	ctx := context.Background()

	a, _ := service.SavePerson(ctx, events.Person{Name: "A", Swatch: "clay"})
	b, _ := service.SavePerson(ctx, events.Person{Name: "B", Swatch: "pine"})
	c, _ := service.SavePerson(ctx, events.Person{Name: "C", Swatch: "sky"})

	got, err := service.ReorderPeople(ctx, []uuid.UUID{c.ID, a.ID, b.ID})
	if err != nil || len(got) != 3 || got[0].ID != c.ID || got[1].ID != a.ID || got[2].ID != b.ID {
		t.Fatalf("expected c, a, b: got=%+v err=%v", got, err)
	}

	for name, ids := range map[string][]uuid.UUID{
		"missing":   {a.ID, b.ID},
		"duplicate": {a.ID, a.ID, b.ID},
		"unknown":   {a.ID, b.ID, uuid.New()},
	} {
		if _, err := service.ReorderPeople(ctx, ids); err != events.ErrInvalidPeopleOrder {
			t.Fatalf("%s: expected ErrInvalidPeopleOrder, got %v", name, err)
		}
	}
}

func TestParseTermsInfersKindAndNormalises(t *testing.T) {
	got := events.ParseTerms([]string{" #Lana ", "Lana", "lana", "#lana", "  ", "#"})
	want := []events.AssociationTerm{
		{Kind: events.TermTag, Term: "lana", Label: "#Lana"},
		{Kind: events.TermWord, Term: "lana", Label: "Lana"},
	}
	if len(got) != len(want) {
		t.Fatalf("expected %+v, got %+v", want, got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("term %d: expected %+v, got %+v", i, want[i], got[i])
		}
	}
}

func TestCreateAssociationRejectsDuplicatesAndUpdateRejectsMissing(t *testing.T) {
	db := openDB(t)
	repo := events.NewSQLiteRepository(db)
	service := events.New(repo)
	ctx := context.Background()
	lana := seedPerson(t, repo, "Lana", 0)

	created, err := service.CreateAssociation(ctx, repositoryTestCalendarID, lana, "", []string{"#Lana"})
	if err != nil || created.Scope != events.ScopeMatched || len(created.Terms) != 1 || created.Terms[0].Term != "lana" || created.Terms[0].Label != "#Lana" {
		t.Fatalf("expected a matched association with a normalised tag and preserved label: got=%+v err=%v", created, err)
	}
	if _, err := service.CreateAssociation(ctx, repositoryTestCalendarID, lana, events.ScopeAll, nil); !errors.Is(err, events.ErrAssociationExists) {
		t.Fatalf("expected ErrAssociationExists, got %v", err)
	}
	if _, err := service.UpdateAssociation(ctx, repositoryTestCalendarID, uuid.New(), nil, nil); !errors.Is(err, events.ErrAssociationNotFound) {
		t.Fatalf("expected ErrAssociationNotFound, got %v", err)
	}
}

func TestUpdateAssociationReplacesTermsOnlyWhenGiven(t *testing.T) {
	db := openDB(t)
	repo := events.NewSQLiteRepository(db)
	service := events.New(repo)
	ctx := context.Background()
	lana := seedPerson(t, repo, "Lana", 0)
	associate(t, repo, lana, events.ScopeMatched, "#lana", "Lana")

	replaced := []string{"#swim"}
	got, err := service.UpdateAssociation(ctx, repositoryTestCalendarID, lana, nil, &replaced)
	if err != nil || got.Scope != events.ScopeMatched || len(got.Terms) != 1 || got.Terms[0].Term != "swim" {
		t.Fatalf("expected the terms replaced and the scope kept: got=%+v err=%v", got, err)
	}
	cleared := []string{}
	got, err = service.UpdateAssociation(ctx, repositoryTestCalendarID, lana, nil, &cleared)
	if err != nil || len(got.Terms) != 0 {
		t.Fatalf("expected an empty list to clear the terms: got=%+v err=%v", got, err)
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

func TestRefreshRemovesRowsOutsideANewSnapshot(t *testing.T) {
	repo, service := newEventsService(t)
	ctx := context.Background()
	now := time.Now()
	seedEvent(t, repo, "far-future", now.AddDate(0, 0, 60), nil)
	service.Client = stubFetcher{occurrences: nil}
	if err := service.Refresh(ctx); err != nil {
		t.Fatal(err)
	}
	got, err := repo.ListWindow(ctx, now.AddDate(0, 0, 59), now.AddDate(0, 0, 61))
	if err != nil || len(got) != 0 {
		t.Fatalf("expected stale far-future row to be removed: got=%+v err=%v", got, err)
	}
}

func TestRefreshIsANoOpWithoutAClient(t *testing.T) {
	_, service := newEventsService(t)
	if err := service.Refresh(context.Background()); err != nil {
		t.Fatalf("expected Refresh to be a no-op without a client, got %v", err)
	}
}

func TestCreateEventWithoutAWriterIsReadOnly(t *testing.T) {
	_, service := newEventsService(t)
	service.Client = stubFetcher{}
	start := time.Now().Add(time.Hour)
	personID, err := uuid.NewRandom()
	if err != nil {
		t.Fatal(err)
	}
	_, err = service.CreateEvent(context.Background(), events.Draft{
		Summary: "Dentist", Start: start, End: start.Add(time.Hour), PersonIDs: []uuid.UUID{personID},
	})
	if !errors.Is(err, events.ErrEventsReadOnly) {
		t.Fatalf("expected ErrEventsReadOnly, got %v", err)
	}
}

func TestCreateEventRoundTripsThroughRefresh(t *testing.T) {
	repo, service := newEventsService(t)
	ctx := context.Background()
	personID, err := uuid.NewRandom()
	if err != nil {
		t.Fatal(err)
	}
	if err := repo.UpsertPerson(ctx, events.Person{ID: personID, Name: "Lana", Swatch: "sky"}); err != nil {
		t.Fatal(err)
	}
	var occurrences []caldav.Occurrence
	service.Client = stubWriter{occurrences: &occurrences}

	start := time.Now().Add(time.Hour).Truncate(time.Second)
	end := start.Add(time.Hour)
	got, err := service.CreateEvent(ctx, events.Draft{
		Summary: "Dentist", Location: "123 Main St", Start: start, End: end, PersonIDs: []uuid.UUID{personID},
	})
	if err != nil {
		t.Fatal(err)
	}
	if got.Summary != "Dentist" || got.Location != "123 Main St" {
		t.Fatalf("expected the created event back, got %+v", got)
	}
	if len(got.People) != 1 || got.People[0].ID != personID {
		t.Fatalf("expected the event attributed to Lana, got %+v", got.People)
	}

	associations, err := repo.ListAssociations(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(associations) != 1 || associations[0].PersonID != personID || associations[0].Scope != events.ScopeMatched ||
		len(associations[0].Terms) != 1 || associations[0].Terms[0].Kind != events.TermTag || associations[0].Terms[0].Term != "lana" {
		t.Fatalf("expected a matched association with a lana tag to be created, got %+v", associations)
	}

	stored, err := repo.ListWindow(ctx, start, end.Add(time.Second))
	if err != nil || len(stored) != 1 {
		t.Fatalf("expected the event to survive the sync it triggered: got=%+v err=%v", stored, err)
	}
}

func TestCreateEventReusesAnExistingPersonTag(t *testing.T) {
	repo, service := newEventsService(t)
	ctx := context.Background()
	personID, err := uuid.NewRandom()
	if err != nil {
		t.Fatal(err)
	}
	if err := repo.UpsertPerson(ctx, events.Person{ID: personID, Name: "Lana", Swatch: "sky"}); err != nil {
		t.Fatal(err)
	}
	associate(t, repo, personID, events.ScopeMatched, "#lana-nickname")
	var occurrences []caldav.Occurrence
	service.Client = stubWriter{occurrences: &occurrences}

	start := time.Now().Add(time.Hour).Truncate(time.Second)
	if _, err := service.CreateEvent(ctx, events.Draft{Summary: "Dentist", Start: start, End: start.Add(time.Hour), PersonIDs: []uuid.UUID{personID}}); err != nil {
		t.Fatal(err)
	}

	associations, err := repo.ListAssociations(ctx)
	if err != nil || len(associations) != 1 || len(associations[0].Terms) != 1 {
		t.Fatalf("expected the existing tag to be reused, not a second term added: got=%+v err=%v", associations, err)
	}
}

func TestCreateEventSharesATagAlreadyUsedBySomeoneElse(t *testing.T) {
	repo, service := newEventsService(t)
	ctx := context.Background()
	lana := seedPerson(t, repo, "Lana", 0)
	elina := seedPerson(t, repo, "Elina", 1)
	associate(t, repo, elina, events.ScopeMatched, "#lana")
	var occurrences []caldav.Occurrence
	service.Client = stubWriter{occurrences: &occurrences}

	start := time.Now().Add(time.Hour)
	if _, err := service.CreateEvent(ctx, events.Draft{Summary: "Dentist", Start: start, End: start.Add(time.Hour), PersonIDs: []uuid.UUID{lana}}); err != nil {
		t.Fatalf("a tag another person uses is shareable: %v", err)
	}
	associations, err := repo.ListAssociations(ctx)
	if err != nil || len(associations) != 2 {
		t.Fatalf("expected Lana to get her own association: got=%+v err=%v", associations, err)
	}
}

func TestCreateEventRejectsUnknownPerson(t *testing.T) {
	_, service := newEventsService(t)
	var occurrences []caldav.Occurrence
	service.Client = stubWriter{occurrences: &occurrences}
	personID, err := uuid.NewRandom()
	if err != nil {
		t.Fatal(err)
	}
	start := time.Now().Add(time.Hour)
	_, err = service.CreateEvent(context.Background(), events.Draft{
		Summary: "Dentist", Start: start, End: start.Add(time.Hour), PersonIDs: []uuid.UUID{personID},
	})
	if err == nil {
		t.Fatal("expected an unknown person error")
	}
	if len(occurrences) != 0 {
		t.Fatalf("event was written before validating the person: %+v", occurrences)
	}
}

func TestCreateEventRejectsInvalidDraft(t *testing.T) {
	_, service := newEventsService(t)
	start := time.Now()
	_, err := service.CreateEvent(context.Background(), events.Draft{
		Summary: " ", Start: start, End: start,
	})
	if !errors.Is(err, events.ErrInvalidDraft) {
		t.Fatalf("expected ErrInvalidDraft, got %v", err)
	}
}

func TestRefreshDueIsolatesCalendarsAndHonorsIntervals(t *testing.T) {
	db := openDB(t)
	repo := events.NewSQLiteRepository(db)
	secondID := uuid.New()
	if _, err := db.Exec(`INSERT INTO calendars (id,account_id,provider_ref,name,visible) VALUES (?, ?, '/second/', 'Second', 1)`, secondID[:], repositoryTestCalendarID[:]); err != nil {
		t.Fatal(err)
	}
	failing := &stubCalendarClient{err: errors.New("offline")}
	healthy := &stubCalendarClient{occurrences: []calendars.Occurrence{{UID: "healthy", Summary: "Healthy", Start: time.Now().Add(time.Hour), End: time.Now().Add(2 * time.Hour)}}}
	now := time.Now()
	account := calendars.Account{ID: repositoryTestCalendarID, RefreshInterval: 15 * time.Minute, WindowPast: time.Hour, WindowFuture: 24 * time.Hour}
	provider := &stubSourceProvider{sources: []calendars.Source{{Account: account, Calendar: calendars.Calendar{ID: repositoryTestCalendarID}, Client: failing}, {Account: account, Calendar: calendars.Calendar{ID: secondID}, Client: healthy}}, byID: map[uuid.UUID]calendars.Source{}, successes: map[uuid.UUID]int{}, failures: map[uuid.UUID]int{}}
	service := events.New(repo)
	service.Calendars = provider
	if err := service.RefreshDue(context.Background()); err == nil {
		t.Fatal("expected failed calendar to be reported")
	}
	if failing.reportCount() != 1 || healthy.reportCount() != 1 || provider.successes[secondID] != 1 || provider.failures[repositoryTestCalendarID] != 1 {
		t.Fatalf("sources were not isolated: failing=%d healthy=%d successes=%v failures=%v", failing.reportCount(), healthy.reportCount(), provider.successes, provider.failures)
	}
	stored, err := repo.ListWindow(context.Background(), now, now.Add(3*time.Hour))
	if err != nil || len(stored) != 1 || stored[0].CalendarID != secondID {
		t.Fatalf("healthy calendar was not stored: %+v err=%v", stored, err)
	}
	if err := service.RefreshDue(context.Background()); err != nil {
		t.Fatalf("account refreshed before its interval: %v", err)
	}
	if healthy.reportCount() != 1 {
		t.Fatal("calendar refreshed before its configured interval")
	}
}

func TestManualRefreshWorksForCalendarOutsideListedSources(t *testing.T) {
	repo, service := newEventsService(t)
	client := &stubCalendarClient{}
	source := calendars.Source{Account: calendars.Account{ID: repositoryTestCalendarID, WindowPast: time.Hour, WindowFuture: time.Hour}, Calendar: calendars.Calendar{ID: repositoryTestCalendarID}, Client: client}
	provider := &stubSourceProvider{byID: map[uuid.UUID]calendars.Source{repositoryTestCalendarID: source}, successes: map[uuid.UUID]int{}, failures: map[uuid.UUID]int{}}
	service.Calendars = provider
	if err := service.RefreshCalendar(context.Background(), repositoryTestCalendarID); err != nil {
		t.Fatal(err)
	}
	if client.reportCount() != 1 {
		t.Fatal("manual refresh did not run")
	}
	_ = repo
}

func newEventsService(t *testing.T) (events.Repository, *events.Service) {
	t.Helper()
	db := openDB(t)
	repo := events.NewSQLiteRepository(db)
	return repo, events.New(repo)
}
