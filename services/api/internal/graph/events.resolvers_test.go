package graph

import (
	"context"
	"io"
	"log/slog"
	"path/filepath"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/kurtschwarz/homie/services/api/internal/events"
	"github.com/kurtschwarz/homie/services/api/internal/graph/model"
	"github.com/kurtschwarz/homie/services/api/internal/store"
)

const testCalendarID = "00000000-0000-7000-8000-000000000001"

func newEventsRepoAndService(t *testing.T) (*events.SQLiteRepository, *events.Service) {
	t.Helper()
	db, err := store.Open(filepath.Join(t.TempDir(), "events.db"), slog.New(slog.NewTextHandler(io.Discard, nil)))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	repo := events.NewSQLiteRepository(db)
	calendarID := uuid.MustParse(testCalendarID)
	if _, err := db.Exec(`INSERT INTO calendar_accounts (id,name,provider,config,credentials) VALUES (?, 'Test', 'caldav', '{}', x'00')`, calendarID[:]); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO calendars (id,account_id,provider_ref,name,visible) VALUES (?, ?, '/test/', 'Test', 1)`, calendarID[:], calendarID[:]); err != nil {
		t.Fatal(err)
	}
	return repo, events.New(repo)
}

func TestEventsQueriesReturnEmptyWhenNotConfigured(t *testing.T) {
	resolver := &queryResolver{Resolver: &Resolver{}}
	ctx := context.Background()
	if got, err := resolver.Events(ctx, nil); err != nil || len(got) != 0 {
		t.Fatalf("events: got=%+v err=%v", got, err)
	}
	if got, err := resolver.People(ctx); err != nil || len(got) != 0 {
		t.Fatalf("people: got=%+v err=%v", got, err)
	}
	if got, err := resolver.CalendarAssociations(ctx, nil, nil); err != nil || len(got) != 0 {
		t.Fatalf("calendarAssociations: got=%+v err=%v", got, err)
	}
	if got, err := resolver.CalendarTags(ctx, testCalendarID); err != nil || len(got) != 0 {
		t.Fatalf("calendarTags: got=%+v err=%v", got, err)
	}
}

func TestSavePersonResolverCreateUpdateAndRejectsUnknownSwatch(t *testing.T) {
	_, service := newEventsRepoAndService(t)
	resolver := &mutationResolver{Resolver: &Resolver{Events: service}}
	ctx := context.Background()

	created, err := resolver.SavePerson(ctx, model.PersonInput{Name: "Elina", Emoji: "🦊", Swatch: "clay", IsFamily: false})
	if err != nil || created.ID == "" {
		t.Fatalf("create: got=%+v err=%v", created, err)
	}

	id := created.ID
	updated, err := resolver.SavePerson(ctx, model.PersonInput{ID: &id, Name: "Elina Renamed", Emoji: "🦊", Swatch: "pine", IsFamily: true})
	if err != nil || updated.Name != "Elina Renamed" || !updated.IsFamily {
		t.Fatalf("update: got=%+v err=%v", updated, err)
	}

	if _, err := resolver.SavePerson(ctx, model.PersonInput{Name: "Bad", Swatch: "scarlet"}); err != events.ErrUnknownSwatch {
		t.Fatalf("expected ErrUnknownSwatch, got %v", err)
	}
}

func TestReorderPeopleResolver(t *testing.T) {
	_, service := newEventsRepoAndService(t)
	resolver := &mutationResolver{Resolver: &Resolver{Events: service}}
	ctx := context.Background()

	a, err := resolver.SavePerson(ctx, model.PersonInput{Name: "A", Swatch: "clay"})
	if err != nil {
		t.Fatal(err)
	}
	b, err := resolver.SavePerson(ctx, model.PersonInput{Name: "B", Swatch: "pine"})
	if err != nil {
		t.Fatal(err)
	}

	got, err := resolver.ReorderPeople(ctx, []string{b.ID, a.ID})
	if err != nil || len(got) != 2 || got[0].ID != b.ID || got[1].ID != a.ID {
		t.Fatalf("expected b, a: got=%+v err=%v", got, err)
	}
	if _, err := resolver.ReorderPeople(ctx, []string{b.ID}); err != events.ErrInvalidPeopleOrder {
		t.Fatalf("expected ErrInvalidPeopleOrder, got %v", err)
	}
	if _, err := resolver.ReorderPeople(ctx, []string{"nope", a.ID}); err == nil {
		t.Fatal("expected an error for a malformed id")
	}
}

func TestCalendarAssociationResolversCreateUpdateQueryAndDelete(t *testing.T) {
	_, service := newEventsRepoAndService(t)
	mutation := &mutationResolver{Resolver: &Resolver{Events: service}}
	query := &queryResolver{Resolver: &Resolver{Events: service}}
	ctx := context.Background()

	person, err := service.SavePerson(ctx, events.Person{Name: "Elina", Swatch: "clay"})
	if err != nil {
		t.Fatal(err)
	}
	personID := person.ID.String()

	created, err := mutation.CreateCalendarAssociation(ctx, model.CreateCalendarAssociationInput{
		CalendarID: testCalendarID, PersonID: personID, Scope: model.AssociationScopeMatched, Terms: []string{" #Elina ", "Ellie"},
	})
	if err != nil || created.Person == nil || created.Person.ID != personID || created.Scope != model.AssociationScopeMatched ||
		len(created.Terms) != 2 || created.Terms[0] != "#Elina" || created.Terms[1] != "Ellie" {
		t.Fatalf("create: got=%+v err=%v", created, err)
	}
	if _, err := mutation.CreateCalendarAssociation(ctx, model.CreateCalendarAssociationInput{CalendarID: testCalendarID, PersonID: personID, Scope: model.AssociationScopeAll}); err != events.ErrAssociationExists {
		t.Fatalf("expected ErrAssociationExists, got %v", err)
	}

	all := model.AssociationScopeAll
	updated, err := mutation.UpdateCalendarAssociation(ctx, model.UpdateCalendarAssociationInput{CalendarID: testCalendarID, PersonID: personID, Scope: &all})
	if err != nil || updated.Scope != model.AssociationScopeAll || len(updated.Terms) != 2 {
		t.Fatalf("update must keep the terms when none are given: got=%+v err=%v", updated, err)
	}
	if _, err := mutation.UpdateCalendarAssociation(ctx, model.UpdateCalendarAssociationInput{CalendarID: testCalendarID, PersonID: uuid.NewString()}); err != events.ErrAssociationNotFound {
		t.Fatalf("expected ErrAssociationNotFound, got %v", err)
	}

	calendar := testCalendarID
	listed, err := query.CalendarAssociations(ctx, &calendar, &personID)
	if err != nil || len(listed) != 1 {
		t.Fatalf("query: got=%+v err=%v", listed, err)
	}
	other := uuid.NewString()
	if listed, err = query.CalendarAssociations(ctx, nil, &other); err != nil || len(listed) != 0 {
		t.Fatalf("a person filter must exclude other people: got=%+v err=%v", listed, err)
	}

	if ok, err := mutation.DeleteCalendarAssociation(ctx, testCalendarID, personID); err != nil || !ok {
		t.Fatalf("delete: ok=%v err=%v", ok, err)
	}
	if listed, err = query.CalendarAssociations(ctx, nil, nil); err != nil || len(listed) != 0 {
		t.Fatalf("expected no associations after delete: got=%+v err=%v", listed, err)
	}
}

func TestEventsResolverOrdersByStart(t *testing.T) {
	repo, service := newEventsRepoAndService(t)
	resolver := &queryResolver{Resolver: &Resolver{Events: service}}
	ctx := context.Background()

	personID, err := uuid.NewRandom()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := service.SavePerson(ctx, events.Person{ID: personID, Name: "Elina", Swatch: "clay"}); err != nil {
		t.Fatal(err)
	}
	if _, err := service.CreateAssociation(ctx, uuid.MustParse(testCalendarID), personID, events.ScopeMatched, []string{"#elina"}); err != nil {
		t.Fatal(err)
	}

	now := time.Now()
	later := events.Event{ID: uuid.New(), UID: "later", Summary: "Later", Start: now.Add(2 * time.Hour), End: now.Add(3 * time.Hour), Categories: []string{"elina"}}
	sooner := events.Event{ID: uuid.New(), UID: "sooner", Summary: "Sooner", Start: now.Add(time.Hour), End: now.Add(2 * time.Hour), Categories: []string{"elina"}}
	if err := repo.ReplaceWindow(ctx, uuid.MustParse(testCalendarID), now, now.Add(4*time.Hour), []events.Event{later, sooner}); err != nil {
		t.Fatal(err)
	}

	got, err := resolver.Events(ctx, nil)
	if err != nil || len(got) != 2 || got[0].Title != "Sooner" || got[1].Title != "Later" {
		t.Fatalf("expected ordering by start: got=%+v err=%v", got, err)
	}
}

func TestCalendarTagsResolverReturnsUnmappedTags(t *testing.T) {
	repo, service := newEventsRepoAndService(t)
	resolver := &queryResolver{Resolver: &Resolver{Events: service}}
	ctx := context.Background()
	now := time.Now()
	event := events.Event{ID: uuid.New(), UID: "a", Summary: "A", Start: now, End: now.Add(time.Hour), Categories: []string{"yellow"}}
	if err := repo.ReplaceWindow(ctx, uuid.MustParse(testCalendarID), now, now.Add(time.Hour), []events.Event{event}); err != nil {
		t.Fatal(err)
	}
	got, err := resolver.CalendarTags(ctx, testCalendarID)
	if err != nil || len(got) != 1 || got[0] != "yellow" {
		t.Fatalf("expected yellow in calendarTags even though unmapped: got=%+v err=%v", got, err)
	}
}
