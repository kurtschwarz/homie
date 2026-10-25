package events

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sort"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

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

// ErrAssociationExists is returned when a person is already associated with a calendar.
var ErrAssociationExists = errors.New("events: person is already associated with this calendar")

// ErrAssociationNotFound is returned when a person has no association with a calendar.
var ErrAssociationNotFound = errors.New("events: person is not associated with this calendar")

// DisplayEvent is an occurrence attributed to the people it is for.
type DisplayEvent struct {
	Event
	People []Person
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

// tagsForPeople resolves each person to a calendar tag, creating an
// association with one from their normalised name when they have none.
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
	associations, err := s.repo.ListAssociations(ctx)
	if err != nil {
		return nil, err
	}
	byPerson := make(map[uuid.UUID]Association)
	for _, association := range associations {
		if association.CalendarID == calendarID {
			byPerson[association.PersonID] = association
		}
	}

	var tags []string
	seenPeople := make(map[uuid.UUID]struct{}, len(personIDs))
	for _, personID := range personIDs {
		if _, duplicate := seenPeople[personID]; duplicate {
			continue
		}
		seenPeople[personID] = struct{}{}
		association, associated := byPerson[personID]
		if tag := firstTag(association); tag != "" {
			tags = append(tags, tag)
			continue
		}
		person := peopleByID[personID]
		term := AssociationTerm{Kind: TermTag, Term: normalizeTag(person.Name), Label: person.Name}
		if associated {
			association.Terms = append(association.Terms, term)
			err = s.repo.UpdateAssociation(ctx, association)
		} else {
			association = Association{CalendarID: calendarID, PersonID: personID, Scope: ScopeMatched, Terms: []AssociationTerm{term}}
			err = s.repo.InsertAssociation(ctx, association)
		}
		if err != nil {
			return nil, err
		}
		tags = append(tags, term.Term)
	}
	return tags, nil
}

func firstTag(association Association) string {
	for _, term := range association.Terms {
		if term.Kind == TermTag {
			return term.Term
		}
	}
	return ""
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

// Displayable returns events in the next `days` days that belong to at least
// one person through a calendar association.
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
	for _, p := range people {
		peopleByID[p.ID] = p
	}
	associations, err := s.repo.ListAssociations(ctx)
	if err != nil {
		return nil, err
	}
	byCalendar := make(map[uuid.UUID][]Association)
	for _, association := range associations {
		byCalendar[association.CalendarID] = append(byCalendar[association.CalendarID], association)
	}

	result := make([]DisplayEvent, 0, len(events))
	for _, event := range events {
		var matched []Person
		for _, association := range byCalendar[event.CalendarID] {
			if !association.matches(event) {
				continue
			}
			if person, ok := peopleByID[association.PersonID]; ok {
				matched = append(matched, person)
			}
		}
		if len(matched) == 0 {
			continue
		}
		sort.Slice(matched, func(i, j int) bool { return matched[i].Order < matched[j].Order })
		result = append(result, DisplayEvent{Event: event, People: matched})
	}
	return result, nil
}

// ListPeople returns every stored person.
func (s *Service) ListPeople(ctx context.Context) ([]Person, error) { return s.repo.ListPeople(ctx) }

// ListAssociations returns every calendar association with its terms.
func (s *Service) ListAssociations(ctx context.Context) ([]Association, error) {
	return s.repo.ListAssociations(ctx)
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

// DeletePerson removes a person; their associations cascade with them.
func (s *Service) DeletePerson(ctx context.Context, id uuid.UUID) error {
	return s.repo.DeletePerson(ctx, id)
}

// CreateAssociation associates a person with a calendar. Terms are labels as
// typed; see ParseTerms. An empty scope means ScopeMatched.
func (s *Service) CreateAssociation(ctx context.Context, calendarID, personID uuid.UUID, scope AssociationScope, labels []string) (Association, error) {
	if scope == "" {
		scope = ScopeMatched
	}
	association := Association{CalendarID: calendarID, PersonID: personID, Scope: scope, Terms: ParseTerms(labels)}
	if err := s.repo.InsertAssociation(ctx, association); err != nil {
		return Association{}, err
	}
	return association, nil
}

// UpdateAssociation changes an association. A nil scope or nil labels keeps the
// stored value, so switching scope never disturbs the terms.
func (s *Service) UpdateAssociation(ctx context.Context, calendarID, personID uuid.UUID, scope *AssociationScope, labels *[]string) (Association, error) {
	associations, err := s.repo.ListAssociations(ctx)
	if err != nil {
		return Association{}, err
	}
	var association Association
	found := false
	for _, candidate := range associations {
		if candidate.CalendarID == calendarID && candidate.PersonID == personID {
			association, found = candidate, true
			break
		}
	}
	if !found {
		return Association{}, ErrAssociationNotFound
	}
	if scope != nil {
		association.Scope = *scope
	}
	if labels != nil {
		association.Terms = ParseTerms(*labels)
	}
	if err := s.repo.UpdateAssociation(ctx, association); err != nil {
		return Association{}, err
	}
	return association, nil
}

// DeleteAssociation removes a person's association with a calendar.
func (s *Service) DeleteAssociation(ctx context.Context, calendarID, personID uuid.UUID) error {
	return s.repo.DeleteAssociation(ctx, calendarID, personID)
}

// ParseTerms turns labels as typed into terms. A leading # marks a calendar
// tag; anything else is a word to look for in event titles. Blank labels and
// repeats are dropped.
func ParseTerms(labels []string) []AssociationTerm {
	type key struct {
		kind TermKind
		term string
	}
	seen := make(map[key]struct{}, len(labels))
	terms := make([]AssociationTerm, 0, len(labels))
	for _, raw := range labels {
		label := strings.TrimSpace(raw)
		kind := TermWord
		text := label
		if strings.HasPrefix(label, "#") {
			kind = TermTag
			text = strings.TrimSpace(strings.TrimPrefix(label, "#"))
		}
		term := normalizeTag(text)
		if term == "" {
			continue
		}
		if _, duplicate := seen[key{kind, term}]; duplicate {
			continue
		}
		seen[key{kind, term}] = struct{}{}
		terms = append(terms, AssociationTerm{Kind: kind, Term: term, Label: label})
	}
	return terms
}

func (a Association) matches(event Event) bool {
	if a.Scope == ScopeAll {
		return true
	}
	for _, term := range a.Terms {
		switch term.Kind {
		case TermTag:
			for _, category := range event.Categories {
				if normalizeTag(category) == term.Term {
					return true
				}
			}
		case TermWord:
			if containsWord(event.Summary, term.Term) {
				return true
			}
		}
	}
	return false
}

// containsWord reports whether word appears in text bounded by non-letters,
// so "lana" matches "Lana's swim" but not "Atlanta".
func containsWord(text, word string) bool {
	if word == "" {
		return false
	}
	text = strings.ToLower(text)
	for start := 0; start < len(text); {
		i := strings.Index(text[start:], word)
		if i < 0 {
			return false
		}
		i += start
		end := i + len(word)
		before, _ := utf8.DecodeLastRuneInString(text[:i])
		after, _ := utf8.DecodeRuneInString(text[end:])
		if (i == 0 || !isWordRune(before)) && (end == len(text) || !isWordRune(after)) {
			return true
		}
		_, size := utf8.DecodeRuneInString(text[i:])
		start = i + size
	}
	return false
}

func isWordRune(r rune) bool {
	return unicode.IsLetter(r) || unicode.IsDigit(r)
}

func normalizeTag(tag string) string {
	return strings.ToLower(strings.TrimSpace(tag))
}
