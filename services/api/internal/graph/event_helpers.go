package graph

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/kurtschwarz/homie/services/api/internal/events"
	"github.com/kurtschwarz/homie/services/api/internal/graph/model"
)

// errEventsNotConfigured and eventsPeopleByID live here rather than beside the
// resolvers: gqlgen rewrites the resolver file on every regeneration and
// comments out anything it did not generate.
var errEventsNotConfigured = errors.New("events are not configured")

func (r *Resolver) eventsPeopleByID(ctx context.Context) (map[uuid.UUID]events.Person, error) {
	people, err := r.Events.ListPeople(ctx)
	if err != nil {
		return nil, err
	}
	byID := make(map[uuid.UUID]events.Person, len(people))
	for _, person := range people {
		byID[person.ID] = person
	}
	return byID, nil
}

func toModelPerson(person events.Person) model.Person {
	return model.Person{
		ID: person.ID.String(), Name: person.Name, Emoji: person.Emoji,
		Swatch: person.Swatch, IsFamily: person.IsFamily,
	}
}

func fromModelEventInput(input model.EventInput) (events.Draft, error) {
	calendarID, err := uuid.Parse(input.CalendarID)
	if err != nil {
		return events.Draft{}, err
	}
	personIDs := make([]uuid.UUID, 0, len(input.PersonIds))
	for _, raw := range input.PersonIds {
		id, err := uuid.Parse(raw)
		if err != nil {
			return events.Draft{}, err
		}
		personIDs = append(personIDs, id)
	}
	draft := events.Draft{
		CalendarID: calendarID,
		Summary:    input.Title, Start: time.Unix(int64(input.Start), 0), End: time.Unix(int64(input.End), 0),
		AllDay: input.AllDay, PersonIDs: personIDs,
	}
	if input.Location != nil {
		draft.Location = *input.Location
	}
	return draft, nil
}

func toModelEvent(event events.DisplayEvent) model.Event {
	var location *string
	if event.Location != "" {
		location = &event.Location
	}
	people := make([]model.Person, 0, len(event.People))
	for _, person := range event.People {
		people = append(people, toModelPerson(person))
	}
	return model.Event{
		ID: event.ID.String(), CalendarID: event.CalendarID.String(), Title: event.Summary, Location: location,
		Start: int(event.Start.Unix()), End: int(event.End.Unix()), AllDay: event.AllDay,
		People: people,
	}
}

func toModelAssociation(association events.Association, peopleByID map[uuid.UUID]events.Person) (model.CalendarAssociation, bool) {
	person, ok := peopleByID[association.PersonID]
	if !ok {
		return model.CalendarAssociation{}, false
	}
	modelPerson := toModelPerson(person)
	terms := make([]string, 0, len(association.Terms))
	for _, term := range association.Terms {
		terms = append(terms, term.Label)
	}
	return model.CalendarAssociation{
		CalendarID: association.CalendarID.String(), Person: &modelPerson,
		Scope: toModelScope(association.Scope), Terms: terms,
	}, true
}

func toModelScope(scope events.AssociationScope) model.AssociationScope {
	if scope == events.ScopeAll {
		return model.AssociationScopeAll
	}
	return model.AssociationScopeMatched
}

func fromModelScope(scope model.AssociationScope) events.AssociationScope {
	if scope == model.AssociationScopeAll {
		return events.ScopeAll
	}
	return events.ScopeMatched
}

func parseAssociationIDs(calendarID, personID string) (uuid.UUID, uuid.UUID, error) {
	calendar, err := uuid.Parse(calendarID)
	if err != nil {
		return uuid.Nil, uuid.Nil, err
	}
	person, err := uuid.Parse(personID)
	if err != nil {
		return uuid.Nil, uuid.Nil, err
	}
	return calendar, person, nil
}

func (r *Resolver) modelAssociation(ctx context.Context, association events.Association) (*model.CalendarAssociation, error) {
	peopleByID, err := r.eventsPeopleByID(ctx)
	if err != nil {
		return nil, err
	}
	result, ok := toModelAssociation(association, peopleByID)
	if !ok {
		return nil, errors.New("person does not exist")
	}
	return &result, nil
}
