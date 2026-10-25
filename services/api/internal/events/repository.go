package events

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

// Person is someone an event can be attributed to.
type Person struct {
	ID       uuid.UUID
	Name     string
	Emoji    string
	Swatch   string
	Order    int
	IsFamily bool
}

// AssociationScope says how much of a calendar belongs to a person.
type AssociationScope string

const (
	// ScopeAll claims every event on the calendar. Terms are kept but ignored.
	ScopeAll AssociationScope = "all"
	// ScopeMatched claims only events that match one of the terms.
	ScopeMatched AssociationScope = "matched"
)

// TermKind is how a term is compared with an event.
type TermKind string

const (
	// TermTag matches a calendar category.
	TermTag TermKind = "tag"
	// TermWord matches a whole word in the event title.
	TermWord TermKind = "word"
)

// AssociationTerm is one thing an event can mention to belong to a person.
type AssociationTerm struct {
	Kind  TermKind
	Term  string // normalised: trimmed, casefolded, without a leading #
	Label string // as the user typed it
}

// Association ties one person to one calendar.
type Association struct {
	CalendarID uuid.UUID
	PersonID   uuid.UUID
	Scope      AssociationScope
	Terms      []AssociationTerm
}

// Event is one concrete calendar occurrence.
type Event struct {
	ID           uuid.UUID
	CalendarID   uuid.UUID
	UID          string
	RecurrenceID string
	Summary      string
	Location     string
	Start        time.Time
	End          time.Time
	AllDay       bool
	Categories   []string // normalised
}

// Repository persists people, calendar associations and synced calendar occurrences.
type Repository interface {
	ListPeople(context.Context) ([]Person, error)
	UpsertPerson(context.Context, Person) error
	SetFamily(context.Context, uuid.UUID) error
	DeletePerson(context.Context, uuid.UUID) error
	ListAssociations(context.Context) ([]Association, error)
	InsertAssociation(context.Context, Association) error
	UpdateAssociation(context.Context, Association) error
	DeleteAssociation(context.Context, uuid.UUID, uuid.UUID) error
	ListWindow(context.Context, time.Time, time.Time) ([]Event, error)
	CountStartingByCalendar(context.Context, time.Time, time.Time) (map[uuid.UUID]int, error)
	DistinctCategories(context.Context, uuid.UUID) ([]string, error)
	ReplaceWindow(context.Context, uuid.UUID, time.Time, time.Time, []Event) error
}

// SQLiteRepository implements Repository with SQLite.
type SQLiteRepository struct {
	db *sqlx.DB
}

// NewSQLiteRepository creates an events repository backed by SQLite.
func NewSQLiteRepository(db *sqlx.DB) *SQLiteRepository {
	return &SQLiteRepository{db: db}
}

type personRow struct {
	ID        []byte `db:"id"`
	Name      string `db:"name"`
	Emoji     string `db:"emoji"`
	Swatch    string `db:"swatch"`
	SortOrder int    `db:"sort_order"`
	IsFamily  int    `db:"is_family"`
}

func (row personRow) person() (Person, error) {
	id, err := uuid.FromBytes(row.ID)
	if err != nil {
		return Person{}, fmt.Errorf("decode person id for %q: %w", row.Name, err)
	}
	return Person{
		ID: id, Name: row.Name, Emoji: row.Emoji, Swatch: row.Swatch,
		Order: row.SortOrder, IsFamily: row.IsFamily == 1,
	}, nil
}

// ListPeople returns every person ordered for display.
func (r *SQLiteRepository) ListPeople(ctx context.Context) ([]Person, error) {
	var rows []personRow
	if err := r.db.SelectContext(ctx, &rows, `SELECT id, name, emoji, swatch, sort_order, is_family FROM people ORDER BY sort_order, name`); err != nil {
		return nil, err
	}
	people := make([]Person, 0, len(rows))
	for _, row := range rows {
		person, err := row.person()
		if err != nil {
			return nil, err
		}
		people = append(people, person)
	}
	return people, nil
}

// UpsertPerson inserts or updates a person. is_family is always written as 0
// and excluded from the update: it can only change via SetFamily, which is
// the only path that satisfies the partial unique index.
func (r *SQLiteRepository) UpsertPerson(ctx context.Context, person Person) error {
	row := personRow{
		ID: person.ID[:], Name: person.Name, Emoji: person.Emoji, Swatch: person.Swatch,
		SortOrder: person.Order, IsFamily: 0,
	}
	_, err := r.db.NamedExecContext(ctx, `
		INSERT INTO people (id, name, emoji, swatch, sort_order, is_family)
		VALUES (:id, :name, :emoji, :swatch, :sort_order, :is_family)
		ON CONFLICT(id) DO UPDATE SET
			name = excluded.name,
			emoji = excluded.emoji,
			swatch = excluded.swatch,
			sort_order = excluded.sort_order`, row)
	return err
}

// SetFamily moves the family flag to the given person in one transaction.
func (r *SQLiteRepository) SetFamily(ctx context.Context, id uuid.UUID) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	if _, err := tx.ExecContext(ctx, `UPDATE people SET is_family = 0 WHERE is_family = 1`); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `UPDATE people SET is_family = 1 WHERE id = ?`, id[:]); err != nil {
		return err
	}
	return tx.Commit()
}

// DeletePerson removes a person. ON DELETE CASCADE drops their associations.
func (r *SQLiteRepository) DeletePerson(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM people WHERE id = ?`, id[:])
	return err
}

type associationRow struct {
	CalendarID []byte `db:"calendar_id"`
	PersonID   []byte `db:"person_id"`
	Scope      string `db:"scope"`
}

type associationTermRow struct {
	CalendarID []byte `db:"calendar_id"`
	PersonID   []byte `db:"person_id"`
	Kind       string `db:"kind"`
	Term       string `db:"term"`
	Label      string `db:"label"`
}

type associationKey struct {
	calendarID uuid.UUID
	personID   uuid.UUID
}

// ListAssociations returns every association with its terms, ordered by
// calendar, person and term so results are stable.
func (r *SQLiteRepository) ListAssociations(ctx context.Context) ([]Association, error) {
	var rows []associationRow
	if err := r.db.SelectContext(ctx, &rows, `SELECT calendar_id, person_id, scope FROM calendar_associations ORDER BY calendar_id, person_id`); err != nil {
		return nil, err
	}
	var termRows []associationTermRow
	if err := r.db.SelectContext(ctx, &termRows, `SELECT calendar_id, person_id, kind, term, label FROM calendar_association_terms ORDER BY calendar_id, person_id, kind, term`); err != nil {
		return nil, err
	}
	associations := make([]Association, 0, len(rows))
	index := make(map[associationKey]int, len(rows))
	for _, row := range rows {
		calendarID, err := uuid.FromBytes(row.CalendarID)
		if err != nil {
			return nil, fmt.Errorf("decode calendar id for association: %w", err)
		}
		personID, err := uuid.FromBytes(row.PersonID)
		if err != nil {
			return nil, fmt.Errorf("decode person id for association: %w", err)
		}
		index[associationKey{calendarID, personID}] = len(associations)
		associations = append(associations, Association{CalendarID: calendarID, PersonID: personID, Scope: AssociationScope(row.Scope)})
	}
	for _, row := range termRows {
		calendarID, err := uuid.FromBytes(row.CalendarID)
		if err != nil {
			return nil, fmt.Errorf("decode calendar id for association term: %w", err)
		}
		personID, err := uuid.FromBytes(row.PersonID)
		if err != nil {
			return nil, fmt.Errorf("decode person id for association term: %w", err)
		}
		i, ok := index[associationKey{calendarID, personID}]
		if !ok {
			continue
		}
		associations[i].Terms = append(associations[i].Terms, AssociationTerm{Kind: TermKind(row.Kind), Term: row.Term, Label: row.Label})
	}
	return associations, nil
}

// InsertAssociation stores a new association and its terms. It returns
// ErrAssociationExists when the calendar and person are already associated.
func (r *SQLiteRepository) InsertAssociation(ctx context.Context, association Association) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	var exists bool
	if err := tx.GetContext(ctx, &exists, `SELECT EXISTS (SELECT 1 FROM calendar_associations WHERE calendar_id=? AND person_id=?)`, association.CalendarID[:], association.PersonID[:]); err != nil {
		return err
	}
	if exists {
		return ErrAssociationExists
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO calendar_associations (calendar_id, person_id, scope) VALUES (?, ?, ?)`, association.CalendarID[:], association.PersonID[:], string(association.Scope)); err != nil {
		return err
	}
	if err := insertTerms(ctx, tx, association); err != nil {
		return err
	}
	return tx.Commit()
}

// UpdateAssociation replaces an association's scope and terms. It returns
// ErrAssociationNotFound when the calendar and person are not associated.
func (r *SQLiteRepository) UpdateAssociation(ctx context.Context, association Association) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	result, err := tx.ExecContext(ctx, `UPDATE calendar_associations SET scope=? WHERE calendar_id=? AND person_id=?`, string(association.Scope), association.CalendarID[:], association.PersonID[:])
	if err != nil {
		return err
	}
	changed, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if changed == 0 {
		return ErrAssociationNotFound
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM calendar_association_terms WHERE calendar_id=? AND person_id=?`, association.CalendarID[:], association.PersonID[:]); err != nil {
		return err
	}
	if err := insertTerms(ctx, tx, association); err != nil {
		return err
	}
	return tx.Commit()
}

func insertTerms(ctx context.Context, tx *sqlx.Tx, association Association) error {
	for _, term := range association.Terms {
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO calendar_association_terms (calendar_id, person_id, kind, term, label)
			VALUES (?, ?, ?, ?, ?)
			ON CONFLICT(calendar_id, person_id, kind, term) DO UPDATE SET label = excluded.label`,
			association.CalendarID[:], association.PersonID[:], string(term.Kind), term.Term, term.Label); err != nil {
			return err
		}
	}
	return nil
}

// DeleteAssociation removes an association; its terms cascade with it.
func (r *SQLiteRepository) DeleteAssociation(ctx context.Context, calendarID, personID uuid.UUID) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM calendar_associations WHERE calendar_id=? AND person_id=?`, calendarID[:], personID[:])
	return err
}

type eventRow struct {
	ID           []byte `db:"id"`
	CalendarID   []byte `db:"calendar_id"`
	UID          string `db:"uid"`
	RecurrenceID string `db:"recurrence_id"`
	Summary      string `db:"summary"`
	Location     string `db:"location"`
	StartAt      int64  `db:"start_at"`
	EndAt        int64  `db:"end_at"`
	AllDay       int    `db:"all_day"`
	Categories   string `db:"categories"`
	SyncedAt     int64  `db:"synced_at"`
}

func (row eventRow) event() (Event, error) {
	id, err := uuid.FromBytes(row.ID)
	if err != nil {
		return Event{}, fmt.Errorf("decode event id for uid %q: %w", row.UID, err)
	}
	calendarID, err := uuid.FromBytes(row.CalendarID)
	if err != nil {
		return Event{}, fmt.Errorf("decode calendar id for uid %q: %w", row.UID, err)
	}
	return Event{
		ID: id, CalendarID: calendarID, UID: row.UID, RecurrenceID: row.RecurrenceID, Summary: row.Summary, Location: row.Location,
		Start: time.Unix(row.StartAt, 0), End: time.Unix(row.EndAt, 0), AllDay: row.AllDay == 1,
		Categories: splitCategories(row.Categories),
	}, nil
}

// ListWindow returns events from visible calendars overlapping [from, to),
// ordered by start time. The overlap comparison (rather than BETWEEN)
// includes an event that began before the window but is still running.
func (r *SQLiteRepository) ListWindow(ctx context.Context, from, to time.Time) ([]Event, error) {
	var rows []eventRow
	if err := r.db.SelectContext(ctx, &rows, `
		SELECT id, calendar_id, uid, recurrence_id, summary, location, start_at, end_at, all_day, categories, synced_at
		FROM events
		WHERE calendar_id IN (SELECT id FROM calendars WHERE visible = 1)
		  AND start_at < ? AND end_at > ?
		ORDER BY start_at, id`, to.Unix(), from.Unix()); err != nil {
		return nil, err
	}
	result := make([]Event, 0, len(rows))
	for _, row := range rows {
		event, err := row.event()
		if err != nil {
			return nil, err
		}
		result = append(result, event)
	}
	return result, nil
}

// CountStartingByCalendar counts events whose start falls in [from, to),
// including events from calendars hidden from the display.
func (r *SQLiteRepository) CountStartingByCalendar(ctx context.Context, from, to time.Time) (map[uuid.UUID]int, error) {
	var rows []struct {
		CalendarID []byte `db:"calendar_id"`
		Count      int    `db:"event_count"`
	}
	if err := r.db.SelectContext(ctx, &rows, `
		SELECT calendar_id, COUNT(*) AS event_count
		FROM events
		WHERE start_at >= ? AND start_at < ?
		GROUP BY calendar_id`, from.Unix(), to.Unix()); err != nil {
		return nil, err
	}
	counts := make(map[uuid.UUID]int, len(rows))
	for _, row := range rows {
		id, err := uuid.FromBytes(row.CalendarID)
		if err != nil {
			return nil, fmt.Errorf("decode calendar id for event count: %w", err)
		}
		counts[id] = row.Count
	}
	return counts, nil
}

// DistinctCategories returns every normalised tag seen across stored events.
func (r *SQLiteRepository) DistinctCategories(ctx context.Context, calendarID uuid.UUID) ([]string, error) {
	var raw []string
	if err := r.db.SelectContext(ctx, &raw, `SELECT categories FROM events WHERE calendar_id=? AND categories != ''`, calendarID[:]); err != nil {
		return nil, err
	}
	seen := map[string]struct{}{}
	var result []string
	for _, categories := range raw {
		for _, tag := range splitCategories(categories) {
			if _, ok := seen[tag]; ok {
				continue
			}
			seen[tag] = struct{}{}
			result = append(result, tag)
		}
	}
	return result, nil
}

// ReplaceWindow replaces every occurrence owned by one calendar with the given
// bounded report, in one transaction. An empty events slice empties the calendar. Ids are
// preserved across the replace for an unchanged (uid, recurrence_id).
func (r *SQLiteRepository) ReplaceWindow(ctx context.Context, calendarID uuid.UUID, from, to time.Time, newEvents []Event) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	existing := make(map[string][]byte)
	var existingRows []struct {
		ID           []byte `db:"id"`
		UID          string `db:"uid"`
		RecurrenceID string `db:"recurrence_id"`
	}
	if err := tx.SelectContext(ctx, &existingRows, `
		SELECT id, uid, recurrence_id FROM events WHERE calendar_id=?`, calendarID[:]); err != nil {
		return err
	}
	for _, row := range existingRows {
		existing[row.UID+"\x00"+row.RecurrenceID] = row.ID
	}

	if _, err := tx.ExecContext(ctx, `DELETE FROM events WHERE calendar_id=?`, calendarID[:]); err != nil {
		return err
	}

	syncedAt := time.Now().Unix()
	for _, event := range newEvents {
		id := event.ID[:]
		if existingID, ok := existing[event.UID+"\x00"+event.RecurrenceID]; ok {
			id = existingID
		}
		row := eventRow{
			ID: id, CalendarID: calendarID[:], UID: event.UID, RecurrenceID: event.RecurrenceID, Summary: event.Summary, Location: event.Location,
			StartAt: event.Start.Unix(), EndAt: event.End.Unix(), AllDay: boolToInt(event.AllDay),
			Categories: strings.Join(event.Categories, ","), SyncedAt: syncedAt,
		}
		// UNIQUE (uid, recurrence_id) is global but the delete above is
		// window-scoped, so an occurrence that moved into the window from
		// outside it still has a surviving row. Upsert onto that row rather
		// than colliding with it; this also absorbs a duplicated key within
		// one report, last occurrence winning.
		if _, err := tx.NamedExecContext(ctx, `
			INSERT INTO events (id, calendar_id, uid, recurrence_id, summary, location, start_at, end_at, all_day, categories, synced_at)
			VALUES (:id, :calendar_id, :uid, :recurrence_id, :summary, :location, :start_at, :end_at, :all_day, :categories, :synced_at)
			ON CONFLICT(calendar_id, uid, recurrence_id) DO UPDATE SET
				summary = excluded.summary,
				location = excluded.location,
				start_at = excluded.start_at,
				end_at = excluded.end_at,
				all_day = excluded.all_day,
				categories = excluded.categories,
				synced_at = excluded.synced_at`, row); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func splitCategories(raw string) []string {
	if raw == "" {
		return nil
	}
	return strings.Split(raw, ",")
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}
