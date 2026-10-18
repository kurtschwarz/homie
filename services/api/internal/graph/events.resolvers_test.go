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

func newEventsRepoAndService(t *testing.T) (*events.SQLiteRepository, *events.Service) {
	t.Helper()
	db, err := store.Open(filepath.Join(t.TempDir(), "events.db"), slog.New(slog.NewTextHandler(io.Discard, nil)))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	repo := events.NewSQLiteRepository(db)
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
	if got, err := resolver.TagMappings(ctx); err != nil || len(got) != 0 {
		t.Fatalf("tagMappings: got=%+v err=%v", got, err)
	}
	if got, err := resolver.CalendarTags(ctx); err != nil || len(got) != 0 {
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

func TestSaveTagMappingResolverNormalisesAndRejectsMissingPerson(t *testing.T) {
	_, service := newEventsRepoAndService(t)
	resolver := &mutationResolver{Resolver: &Resolver{Events: service}}
	ctx := context.Background()

	person, err := service.SavePerson(ctx, events.Person{Name: "Elina", Swatch: "clay"})
	if err != nil {
		t.Fatal(err)
	}
	personID := person.ID.String()

	mapping, err := resolver.SaveTagMapping(ctx, model.TagMappingInput{Tag: " Elina ", Kind: model.TagKindPerson, PersonID: &personID})
	if err != nil || mapping.Tag != "elina" || mapping.Label != "Elina" || mapping.Person == nil || mapping.Person.ID != personID {
		t.Fatalf("save mapping: got=%+v err=%v", mapping, err)
	}

	if _, err := resolver.SaveTagMapping(ctx, model.TagMappingInput{Tag: "homie", Kind: model.TagKindPerson}); err != events.ErrPersonMappingNeedsPerson {
		t.Fatalf("expected ErrPersonMappingNeedsPerson, got %v", err)
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
	if _, err := service.SaveTagMapping(ctx, "elina", events.TagKindPerson, &personID); err != nil {
		t.Fatal(err)
	}

	now := time.Now()
	later := events.Event{ID: uuid.New(), UID: "later", Summary: "Later", Start: now.Add(2 * time.Hour), End: now.Add(3 * time.Hour), Categories: []string{"elina"}}
	sooner := events.Event{ID: uuid.New(), UID: "sooner", Summary: "Sooner", Start: now.Add(time.Hour), End: now.Add(2 * time.Hour), Categories: []string{"elina"}}
	if err := repo.ReplaceWindow(ctx, now, now.Add(4*time.Hour), []events.Event{later, sooner}); err != nil {
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
	if err := repo.ReplaceWindow(ctx, now, now.Add(time.Hour), []events.Event{event}); err != nil {
		t.Fatal(err)
	}
	got, err := resolver.CalendarTags(ctx)
	if err != nil || len(got) != 1 || got[0] != "yellow" {
		t.Fatalf("expected yellow in calendarTags even though unmapped: got=%+v err=%v", got, err)
	}
}
