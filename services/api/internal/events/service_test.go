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
	if err := repo.UpsertTagMapping(ctx, events.TagMapping{CalendarID: repositoryTestCalendarID, Tag: "elina", Label: "Elina", Kind: events.TagKindPerson, PersonID: &elina}); err != nil {
		t.Fatal(err)
	}
	if err := repo.UpsertTagMapping(ctx, events.TagMapping{CalendarID: repositoryTestCalendarID, Tag: "lana", Label: "Lana", Kind: events.TagKindPerson, PersonID: &lana}); err != nil {
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
	if err := repo.UpsertTagMapping(ctx, events.TagMapping{CalendarID: repositoryTestCalendarID, Tag: "homie", Label: "Homie", Kind: events.TagKindInclude}); err != nil {
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

	if err := repo.UpsertTagMapping(ctx, events.TagMapping{CalendarID: repositoryTestCalendarID, Tag: "homie", Label: "Homie", Kind: events.TagKindInclude}); err != nil {
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
	if err := repo.UpsertTagMapping(ctx, events.TagMapping{CalendarID: repositoryTestCalendarID, Tag: "elina", Label: "Elina", Kind: events.TagKindPerson, PersonID: &elina}); err != nil {
		t.Fatal(err)
	}
	if err := repo.UpsertTagMapping(ctx, events.TagMapping{CalendarID: repositoryTestCalendarID, Tag: "homie", Label: "Homie", Kind: events.TagKindInclude}); err != nil {
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
	if err := repo.UpsertTagMapping(ctx, events.TagMapping{CalendarID: repositoryTestCalendarID, Tag: "elina", Label: "Elina", Kind: events.TagKindPerson, PersonID: &elina}); err != nil {
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

	mappings, err := repo.ListTagMappings(ctx, repositoryTestCalendarID)
	if err != nil {
		t.Fatal(err)
	}
	if len(mappings) != 1 || mappings[0].Tag != "lana" || mappings[0].PersonID == nil || *mappings[0].PersonID != personID {
		t.Fatalf("expected a lana->Lana tag mapping to be created, got %+v", mappings)
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
	if err := repo.UpsertTagMapping(ctx, events.TagMapping{CalendarID: repositoryTestCalendarID, Tag: "lana-nickname", Label: "Lana Nickname", Kind: events.TagKindPerson, PersonID: &personID}); err != nil {
		t.Fatal(err)
	}
	var occurrences []caldav.Occurrence
	service.Client = stubWriter{occurrences: &occurrences}

	start := time.Now().Add(time.Hour).Truncate(time.Second)
	if _, err := service.CreateEvent(ctx, events.Draft{Summary: "Dentist", Start: start, End: start.Add(time.Hour), PersonIDs: []uuid.UUID{personID}}); err != nil {
		t.Fatal(err)
	}

	mappings, err := repo.ListTagMappings(ctx, repositoryTestCalendarID)
	if err != nil || len(mappings) != 1 {
		t.Fatalf("expected the existing mapping to be reused, not a second one created: got=%+v err=%v", mappings, err)
	}
}

func TestCreateEventRejectsATagAlreadyTakenBySomeoneElse(t *testing.T) {
	repo, service := newEventsService(t)
	ctx := context.Background()
	lanaID, err := uuid.NewRandom()
	if err != nil {
		t.Fatal(err)
	}
	elinaID, err := uuid.NewRandom()
	if err != nil {
		t.Fatal(err)
	}
	if err := repo.UpsertPerson(ctx, events.Person{ID: lanaID, Name: "Lana", Swatch: "sky"}); err != nil {
		t.Fatal(err)
	}
	if err := repo.UpsertPerson(ctx, events.Person{ID: elinaID, Name: "Elina", Swatch: "amber"}); err != nil {
		t.Fatal(err)
	}
	if err := repo.UpsertTagMapping(ctx, events.TagMapping{CalendarID: repositoryTestCalendarID, Tag: "lana", Label: "Lana", Kind: events.TagKindPerson, PersonID: &elinaID}); err != nil {
		t.Fatal(err)
	}
	var occurrences []caldav.Occurrence
	service.Client = stubWriter{occurrences: &occurrences}

	start := time.Now().Add(time.Hour)
	_, err = service.CreateEvent(ctx, events.Draft{Summary: "Dentist", Start: start, End: start.Add(time.Hour), PersonIDs: []uuid.UUID{lanaID}})
	if !errors.Is(err, events.ErrTagTaken) {
		t.Fatalf("expected ErrTagTaken, got %v", err)
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
