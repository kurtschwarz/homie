package graph

import (
	"context"
	"errors"

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
		ID: event.ID.String(), Title: event.Summary, Location: location,
		Start: int(event.Start.Unix()), End: int(event.End.Unix()), AllDay: event.AllDay,
		People: people,
	}
}

func toModelTagMapping(mapping events.TagMapping, peopleByID map[uuid.UUID]events.Person) model.TagMapping {
	result := model.TagMapping{Tag: mapping.Tag, Label: mapping.Label, Kind: toModelTagKind(mapping.Kind)}
	if mapping.PersonID != nil {
		if person, ok := peopleByID[*mapping.PersonID]; ok {
			modelPerson := toModelPerson(person)
			result.Person = &modelPerson
		}
	}
	return result
}

func toModelTagKind(kind events.TagKind) model.TagKind {
	if kind == events.TagKindInclude {
		return model.TagKindInclude
	}
	return model.TagKindPerson
}

func fromModelTagKind(kind model.TagKind) events.TagKind {
	if kind == model.TagKindInclude {
		return events.TagKindInclude
	}
	return events.TagKindPerson
}
