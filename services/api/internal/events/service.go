package events

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/kurtschwarz/homie/services/api/internal/calendars"
	"github.com/kurtschwarz/homie/services/api/internal/calendars/caldav"
)

const (
	// RefreshInterval is the cadence used by the background CalDAV refresh worker.
	RefreshInterval = 15 * time.Minute
	// WindowPast and WindowFuture bound the window of occurrences kept in sync.
	WindowPast   = 24 * time.Hour
	WindowFuture = 21 * 24 * time.Hour
)

// Swatches mirrors packages/ui's palette so savePerson can validate without
// a shared generated file; the ten strings are not worth a build step.
var Swatches = map[string]struct{}{
	"clay": {}, "amber": {}, "moss": {}, "pine": {}, "teal": {},
	"sky": {}, "iris": {}, "orchid": {}, "plum": {}, "slate": {},
}

// ErrUnknownSwatch is returned by savePerson for a swatch outside the palette.
var ErrUnknownSwatch = errors.New("unknown swatch")

// ErrPersonMappingNeedsPerson is returned when a PERSON tag mapping has no person.
var ErrPersonMappingNeedsPerson = errors.New("a person tag mapping requires a person id")

// DisplayEvent is an occurrence attributed to the people it is for.
type DisplayEvent struct {
	Event
	People []Person
}

type mappingKey struct {
	CalendarID uuid.UUID
	Tag        string
}

// Fetcher is the portion of the CalDAV client used by Service.
type Fetcher interface {
	Report(ctx context.Context, from, to time.Time) ([]caldav.Occurrence, error)
}

// Writer is the portion of the CalDAV client used to create events. It is
// separate from Fetcher so a Client that can only read keeps CreateEvent
// returning ErrEventsReadOnly instead of panicking on a failed type assertion.
type Writer interface {
	Put(ctx context.Context, occ caldav.Occurrence) error
}

type SourceProvider interface {
	Sources(context.Context) ([]calendars.Source, error)
	Source(context.Context, uuid.UUID) (calendars.Source, error)
	RecordSuccess(context.Context, uuid.UUID, time.Time) error
	RecordFailure(context.Context, uuid.UUID, string) error
	RecordAttempt(context.Context, uuid.UUID, time.Time) error
}

// ErrEventsReadOnly is returned by CreateEvent when the configured CalDAV
// client cannot write (typically: CalDAV is not configured at all).
var ErrEventsReadOnly = errors.New("events: CalDAV client cannot create events")

// ErrTagTaken is returned when a person has no PERSON tag mapping and their
// normalised name collides with an existing mapping for someone else, or
// with an INCLUDE mapping.
var ErrTagTaken = errors.New("events: tag already mapped to something else")

// ErrInvalidDraft is returned when a draft is missing a title or person, or
// has an invalid time range.
var ErrInvalidDraft = errors.New("events: invalid event draft")

var legacyCalendarID = uuid.MustParse("00000000-0000-7000-8000-000000000001")

// Draft is a locally authored event awaiting a CalDAV UID.
type Draft struct {
	CalendarID uuid.UUID
	Summary    string
	Location   string
	Start      time.Time
	End        time.Time
	AllDay     bool
	PersonIDs  []uuid.UUID
}

// Service serves displayable events and the people/mapping surface backing
// them. Client and Logger are exported so main.go can wire a CalDAV client
// in after construction, without changing New's signature for the tests and
// call sites that only need the repository.
type Service struct {
	Client    Fetcher
	Calendars SourceProvider
	Logger    *slog.Logger

	repo Repository
	now  func() time.Time
}

// New creates an events service over a repository.
func New(repo Repository) *Service {
	return &Service{repo: repo, now: time.Now}
}

// UpcomingCalendarEventCounts returns the number of events beginning on each calendar in
// the next seven days. Hidden calendars are deliberately included.
func (s *Service) UpcomingCalendarEventCounts(ctx context.Context) (map[uuid.UUID]int, error) {
	from := s.now()
	return s.repo.CountStartingByCalendar(ctx, from, from.Add(7*24*time.Hour))
}

// Refresh fetches the configured sync window from CalDAV and replaces it.
// The service stores everything the report returns, tagged or not; the tag
// filter belongs to the read path in Displayable.
func (s *Service) Refresh(ctx context.Context) error {
	if s.Calendars != nil {
		return s.refreshCalendars(ctx, false)
	}
	if s.Client == nil {
		return nil
	}
	from := s.now().Add(-WindowPast)
	to := s.now().Add(WindowFuture)
	occurrences, err := s.Client.Report(ctx, from, to)
	if err != nil {
		if errors.Is(err, caldav.ErrUnauthorized) {
			s.logger().ErrorContext(ctx, "caldav report unauthorized; check the configured app password")
		}
		return err
	}
	newEvents := make([]Event, 0, len(occurrences))
	untagged := 0
	for _, occurrence := range occurrences {
		id := eventID(legacyCalendarID, occurrence.UID, occurrence.RecurrenceID)
		if len(occurrence.Categories) == 0 {
			untagged++
		}
		newEvents = append(newEvents, Event{
			ID: id, CalendarID: legacyCalendarID, UID: occurrence.UID, RecurrenceID: occurrence.RecurrenceID,
			Summary: occurrence.Summary, Location: occurrence.Location,
			Start: occurrence.Start, End: occurrence.End, AllDay: occurrence.AllDay,
			Categories: occurrence.Categories,
		})
	}
	if err := s.repo.ReplaceWindow(ctx, legacyCalendarID, from, to, newEvents); err != nil {
		return err
	}
	s.logger().InfoContext(ctx, "events refresh", "fetched", len(occurrences), "written", len(newEvents), "untagged", untagged)
	return nil
}

// RefreshDue refreshes every calendar whose account's interval has elapsed.
func (s *Service) RefreshDue(ctx context.Context) error {
	if s.Calendars == nil {
		return s.Refresh(ctx)
	}
	return s.refreshCalendars(ctx, true)
}

// RefreshCalendar refreshes one calendar immediately.
func (s *Service) RefreshCalendar(ctx context.Context, id uuid.UUID) error {
	if s.Calendars == nil {
		return ErrEventsReadOnly
	}
	source, err := s.Calendars.Source(ctx, id)
	if err != nil {
		return err
	}
	now := s.now()
	refreshErr := s.refreshSource(ctx, source, now)
	return errors.Join(refreshErr, s.Calendars.RecordAttempt(ctx, source.Account.ID, now))
}

func (s *Service) refreshCalendars(ctx context.Context, dueOnly bool) error {
	sources, err := s.Calendars.Sources(ctx)
	if err != nil {
		return err
	}
	now := s.now()
	byAccount := make(map[uuid.UUID][]calendars.Source)
	for _, source := range sources {
		byAccount[source.Account.ID] = append(byAccount[source.Account.ID], source)
	}
	results := make(chan error, len(byAccount))
	started := 0
	for _, accountSources := range byAccount {
		account := accountSources[0].Account
		if dueOnly && account.LastAttemptedAt != nil && account.LastAttemptedAt.Add(account.RefreshInterval).After(now) {
			continue
		}
		started++
		go func(sources []calendars.Source) { results <- s.refreshAccount(ctx, sources, now) }(accountSources)
	}
	var errs []error
	for range started {
		if err := <-results; err != nil {
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}

func (s *Service) refreshAccount(ctx context.Context, sources []calendars.Source, now time.Time) error {
	results := make(chan error, len(sources))
	for _, source := range sources {
		go func(source calendars.Source) { results <- s.refreshSource(ctx, source, now) }(source)
	}
	var errs []error
	for range sources {
		if err := <-results; err != nil {
			errs = append(errs, err)
		}
	}
	if err := s.Calendars.RecordAttempt(ctx, sources[0].Account.ID, now); err != nil {
		errs = append(errs, err)
	}
	return errors.Join(errs...)
}

func (s *Service) refreshSource(ctx context.Context, source calendars.Source, now time.Time) error {
	from, to := now.Add(-source.Account.WindowPast), now.Add(source.Account.WindowFuture)
	occurrences, err := source.Client.Report(ctx, from, to)
	if err != nil {
		_ = s.Calendars.RecordFailure(ctx, source.Calendar.ID, err.Error())
		return fmt.Errorf("refresh calendar %s: %w", source.Calendar.ID, err)
	}
	newEvents := make([]Event, 0, len(occurrences))
	for _, occurrence := range occurrences {
		id := eventID(source.Calendar.ID, occurrence.UID, occurrence.RecurrenceID)
		newEvents = append(newEvents, Event{ID: id, CalendarID: source.Calendar.ID, UID: occurrence.UID, RecurrenceID: occurrence.RecurrenceID, Summary: occurrence.Summary, Location: occurrence.Location, Start: occurrence.Start, End: occurrence.End, AllDay: occurrence.AllDay, Categories: occurrence.Categories})
	}
	if err := s.repo.ReplaceWindow(ctx, source.Calendar.ID, from, to, newEvents); err != nil {
		_ = s.Calendars.RecordFailure(ctx, source.Calendar.ID, err.Error())
		return err
	}
	if err := s.Calendars.RecordSuccess(ctx, source.Calendar.ID, now); err != nil {
		return err
	}
	s.logger().InfoContext(ctx, "calendar refresh", "calendar_id", source.Calendar.ID, "fetched", len(occurrences))
	return nil
}

// CreateEvent writes a new VEVENT to CalDAV, syncs it back in, and returns
// it attributed to its people. CalDAV stays the only writer to the events
// table, so the periodic ReplaceWindow sync can never clobber it.
func (s *Service) CreateEvent(ctx context.Context, draft Draft) (DisplayEvent, error) {
	if strings.TrimSpace(draft.Summary) == "" || !draft.End.After(draft.Start) || len(draft.PersonIDs) == 0 {
		return DisplayEvent{}, ErrInvalidDraft
	}
	if s.Calendars != nil {
		return s.createCalendarEvent(ctx, draft)
	}
	if s.Client == nil {
		return DisplayEvent{}, ErrEventsReadOnly
	}
	writer, ok := s.Client.(Writer)
	if !ok {
		return DisplayEvent{}, ErrEventsReadOnly
	}

	categories, err := s.tagsForPeople(ctx, draft.PersonIDs)
	if err != nil {
		return DisplayEvent{}, err
	}

	id, err := uuid.NewV7()
	if err != nil {
		return DisplayEvent{}, err
	}
	occurrence := caldav.Occurrence{
		UID:        id.String() + "@homie",
		Summary:    draft.Summary,
		Location:   draft.Location,
		Start:      draft.Start,
		End:        draft.End,
		AllDay:     draft.AllDay,
		Categories: categories,
	}
	if err := writer.Put(ctx, occurrence); err != nil {
		return DisplayEvent{}, err
	}

	if err := s.Refresh(ctx); err != nil {
		// The event is saved; a failed sync just means it will not be
		// visible until the next periodic refresh. Return it as created
		// rather than reporting a save failure that did not happen.
		s.logger().WarnContext(ctx, "events refresh after create failed", "error", err)
	}

	events, err := s.repo.ListWindow(ctx, draft.Start, draft.Start.Add(time.Second))
	if err == nil {
		for _, event := range events {
			if event.UID == occurrence.UID {
				return DisplayEvent{Event: event, People: s.peopleForIDs(ctx, draft.PersonIDs)}, nil
			}
		}
	}

	return DisplayEvent{
		Event: Event{
			ID: id, CalendarID: legacyCalendarID, UID: occurrence.UID, Summary: draft.Summary, Location: draft.Location,
			Start: draft.Start, End: draft.End, AllDay: draft.AllDay, Categories: categories,
		},
		People: s.peopleForIDs(ctx, draft.PersonIDs),
	}, nil
}

func (s *Service) createCalendarEvent(ctx context.Context, draft Draft) (DisplayEvent, error) {
	if draft.CalendarID == uuid.Nil {
		return DisplayEvent{}, ErrInvalidDraft
	}
	sources, err := s.Calendars.Sources(ctx)
	if err != nil {
		return DisplayEvent{}, err
	}
	var source *calendars.Source
	for i := range sources {
		if sources[i].Calendar.ID == draft.CalendarID {
			source = &sources[i]
			break
		}
	}
	if source == nil {
		return DisplayEvent{}, ErrEventsReadOnly
	}
	categories, err := s.tagsForPeopleInCalendar(ctx, draft.CalendarID, draft.PersonIDs)
	if err != nil {
		return DisplayEvent{}, err
	}
	id, err := uuid.NewV7()
	if err != nil {
		return DisplayEvent{}, err
	}
	occurrence := calendars.Occurrence{UID: id.String() + "@homie", Summary: draft.Summary, Location: draft.Location, Start: draft.Start, End: draft.End, AllDay: draft.AllDay, Categories: categories}
	if err := source.Client.Put(ctx, occurrence); err != nil {
		return DisplayEvent{}, err
	}
	if err := s.RefreshCalendar(ctx, draft.CalendarID); err != nil {
		s.logger().WarnContext(ctx, "events refresh after create failed", "error", err)
	}
	stored, err := s.repo.ListWindow(ctx, draft.Start, draft.Start.Add(time.Second))
	if err == nil {
		for _, event := range stored {
			if event.CalendarID == draft.CalendarID && event.UID == occurrence.UID {
				return DisplayEvent{Event: event, People: s.peopleForIDs(ctx, draft.PersonIDs)}, nil
			}
		}
	}
	return DisplayEvent{Event: Event{ID: eventID(draft.CalendarID, occurrence.UID, occurrence.RecurrenceID), CalendarID: draft.CalendarID, UID: occurrence.UID, Summary: draft.Summary, Location: draft.Location, Start: draft.Start, End: draft.End, AllDay: draft.AllDay, Categories: categories}, People: s.peopleForIDs(ctx, draft.PersonIDs)}, nil
}

func eventID(calendarID uuid.UUID, uid, recurrenceID string) uuid.UUID {
	return uuid.NewSHA1(calendarID, []byte(uid+"\x00"+recurrenceID))
}

// tagsForPeople resolves each person to a PERSON tag, creating one from
// their normalised name when they have none. It errors rather than steal a
// tag already mapped to someone else.
func (s *Service) tagsForPeople(ctx context.Context, personIDs []uuid.UUID) ([]string, error) {
	return s.tagsForPeopleInCalendar(ctx, legacyCalendarID, personIDs)
}

func (s *Service) tagsForPeopleInCalendar(ctx context.Context, calendarID uuid.UUID, personIDs []uuid.UUID) ([]string, error) {
	if len(personIDs) == 0 {
		return nil, nil
	}
	people, err := s.repo.ListPeople(ctx)
	if err != nil {
		return nil, err
	}
	peopleByID := make(map[uuid.UUID]Person, len(people))
	for _, p := range people {
		peopleByID[p.ID] = p
	}
	for _, personID := range personIDs {
		if _, ok := peopleByID[personID]; !ok {
			return nil, fmt.Errorf("events: person %s does not exist", personID)
		}
	}
	mappings, err := s.repo.ListTagMappings(ctx, calendarID)
	if err != nil {
		return nil, err
	}
	mappingByTag := make(map[string]TagMapping, len(mappings))
	tagByPerson := make(map[uuid.UUID]string, len(mappings))
	for _, m := range mappings {
		mappingByTag[m.Tag] = m
		if m.Kind == TagKindPerson && m.PersonID != nil {
			if _, ok := tagByPerson[*m.PersonID]; !ok {
				tagByPerson[*m.PersonID] = m.Tag
			}
		}
	}

	var tags []string
	seenPeople := make(map[uuid.UUID]struct{}, len(personIDs))
	for _, personID := range personIDs {
		if _, duplicate := seenPeople[personID]; duplicate {
			continue
		}
		seenPeople[personID] = struct{}{}
		if tag, ok := tagByPerson[personID]; ok {
			tags = append(tags, tag)
			continue
		}
		person, ok := peopleByID[personID]
		if !ok {
			continue // validated above
		}
		tag := normalizeTag(person.Name)
		if existing, ok := mappingByTag[tag]; ok && (existing.Kind != TagKindPerson || existing.PersonID == nil || *existing.PersonID != personID) {
			return nil, fmt.Errorf("%w: %q", ErrTagTaken, tag)
		}
		mapping := TagMapping{CalendarID: calendarID, Tag: tag, Label: person.Name, Kind: TagKindPerson, PersonID: &personID}
		if err := s.repo.UpsertTagMapping(ctx, mapping); err != nil {
			return nil, err
		}
		mappingByTag[tag] = mapping
		tagByPerson[personID] = tag
		tags = append(tags, tag)
	}
	return tags, nil
}

func (s *Service) peopleForIDs(ctx context.Context, personIDs []uuid.UUID) []Person {
	if len(personIDs) == 0 {
		return nil
	}
	people, err := s.repo.ListPeople(ctx)
	if err != nil {
		return nil
	}
	peopleByID := make(map[uuid.UUID]Person, len(people))
	for _, p := range people {
		peopleByID[p.ID] = p
	}
	var result []Person
	seen := make(map[uuid.UUID]struct{}, len(personIDs))
	for _, id := range personIDs {
		if _, duplicate := seen[id]; duplicate {
			continue
		}
		seen[id] = struct{}{}
		if p, ok := peopleByID[id]; ok {
			result = append(result, p)
		}
	}
	sort.Slice(result, func(i, j int) bool { return result[i].Order < result[j].Order })
	return result
}

func (s *Service) logger() *slog.Logger {
	if s.Logger != nil {
		return s.Logger
	}
	return slog.Default()
}

// Displayable returns events in the next `days` days that qualify for the
// kiosk: at least one PERSON tag or at least one INCLUDE tag.
func (s *Service) Displayable(ctx context.Context, days int) ([]DisplayEvent, error) {
	from := s.now()
	to := from.AddDate(0, 0, days)
	events, err := s.repo.ListWindow(ctx, from, to)
	if err != nil {
		return nil, err
	}
	people, err := s.repo.ListPeople(ctx)
	if err != nil {
		return nil, err
	}
	peopleByID := make(map[uuid.UUID]Person, len(people))
	var family *Person
	for _, p := range people {
		peopleByID[p.ID] = p
		if p.IsFamily {
			family = &p
		}
	}
	mappingByCalendarTag := make(map[mappingKey]TagMapping)
	calendarIDs := make(map[uuid.UUID]struct{})
	for _, event := range events {
		calendarIDs[event.CalendarID] = struct{}{}
	}
	for calendarID := range calendarIDs {
		mappings, err := s.repo.ListTagMappings(ctx, calendarID)
		if err != nil {
			return nil, err
		}
		for _, m := range mappings {
			mappingByCalendarTag[mappingKey{CalendarID: m.CalendarID, Tag: m.Tag}] = m
		}
	}

	result := make([]DisplayEvent, 0, len(events))
	for _, event := range events {
		var matched []Person
		include := false
		seen := make(map[uuid.UUID]struct{})
		for _, tag := range event.Categories {
			mapping, ok := mappingByCalendarTag[mappingKey{CalendarID: event.CalendarID, Tag: normalizeTag(tag)}]
			if !ok {
				continue
			}
			switch mapping.Kind {
			case TagKindPerson:
				if mapping.PersonID == nil {
					continue
				}
				if _, dup := seen[*mapping.PersonID]; dup {
					continue
				}
				person, ok := peopleByID[*mapping.PersonID]
				if !ok {
					continue
				}
				seen[*mapping.PersonID] = struct{}{}
				matched = append(matched, person)
			case TagKindInclude:
				include = true
			}
		}
		if len(matched) == 0 && !include {
			continue
		}
		if len(matched) == 0 && include && family != nil {
			matched = []Person{*family}
		}
		sort.Slice(matched, func(i, j int) bool { return matched[i].Order < matched[j].Order })
		result = append(result, DisplayEvent{Event: event, People: matched})
	}
	return result, nil
}

// ListPeople returns every stored person.
func (s *Service) ListPeople(ctx context.Context) ([]Person, error) { return s.repo.ListPeople(ctx) }

// ListTagMappings returns every stored tag mapping.
func (s *Service) ListTagMappings(ctx context.Context) ([]TagMapping, error) {
	return s.repo.ListTagMappings(ctx, legacyCalendarID)
}

func (s *Service) ListCalendarTagMappings(ctx context.Context, calendarID uuid.UUID) ([]TagMapping, error) {
	return s.repo.ListTagMappings(ctx, calendarID)
}

// CalendarTags returns every distinct category seen in the synced window.
func (s *Service) CalendarTags(ctx context.Context) ([]string, error) {
	return s.repo.DistinctCategories(ctx, legacyCalendarID)
}

func (s *Service) CalendarTagsFor(ctx context.Context, calendarID uuid.UUID) ([]string, error) {
	return s.repo.DistinctCategories(ctx, calendarID)
}

// SavePerson validates the swatch, then creates or updates the person.
// IsFamily routes through SetFamily so the partial unique index is satisfied
// by moving the flag rather than colliding with it.
func (s *Service) SavePerson(ctx context.Context, person Person) (Person, error) {
	if _, ok := Swatches[person.Swatch]; !ok {
		return Person{}, ErrUnknownSwatch
	}
	if person.ID == uuid.Nil {
		id, err := uuid.NewV7()
		if err != nil {
			return Person{}, err
		}
		person.ID = id
	}
	if err := s.repo.UpsertPerson(ctx, person); err != nil {
		return Person{}, err
	}
	if person.IsFamily {
		if err := s.repo.SetFamily(ctx, person.ID); err != nil {
			return Person{}, err
		}
	}
	return person, nil
}

// DeletePerson removes a person; their tag mappings cascade with them.
func (s *Service) DeletePerson(ctx context.Context, id uuid.UUID) error {
	return s.repo.DeletePerson(ctx, id)
}

// SaveTagMapping normalises the tag and stores the original spelling as Label.
func (s *Service) SaveTagMapping(ctx context.Context, rawTag string, kind TagKind, personID *uuid.UUID) (TagMapping, error) {
	return s.SaveCalendarTagMapping(ctx, legacyCalendarID, rawTag, kind, personID)
}

func (s *Service) SaveCalendarTagMapping(ctx context.Context, calendarID uuid.UUID, rawTag string, kind TagKind, personID *uuid.UUID) (TagMapping, error) {
	if kind == TagKindPerson && personID == nil {
		return TagMapping{}, ErrPersonMappingNeedsPerson
	}
	if kind == TagKindInclude {
		personID = nil
	}
	mapping := TagMapping{CalendarID: calendarID, Tag: normalizeTag(rawTag), Label: strings.TrimSpace(rawTag), Kind: kind, PersonID: personID}
	if err := s.repo.UpsertTagMapping(ctx, mapping); err != nil {
		return TagMapping{}, err
	}
	return mapping, nil
}

// DeleteTagMapping removes a mapping. The caller passes the raw tag; it is normalised here.
func (s *Service) DeleteTagMapping(ctx context.Context, rawTag string) error {
	return s.repo.DeleteTagMapping(ctx, legacyCalendarID, normalizeTag(rawTag))
}

func (s *Service) DeleteCalendarTagMapping(ctx context.Context, calendarID uuid.UUID, rawTag string) error {
	return s.repo.DeleteTagMapping(ctx, calendarID, normalizeTag(rawTag))
}

func normalizeTag(tag string) string {
	return strings.ToLower(strings.TrimSpace(tag))
}
