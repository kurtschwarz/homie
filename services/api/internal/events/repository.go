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

// TagKind distinguishes a calendar tag that names a person from one that
// merely brings an event onto the kiosk.
type TagKind string

const (
	TagKindPerson  TagKind = "person"
	TagKindInclude TagKind = "include"
)

// TagMapping connects a normalised calendar tag to what it means.
type TagMapping struct {
	Tag      string // normalised: trimmed, casefolded
	Label    string // as the user typed it
	Kind     TagKind
	PersonID *uuid.UUID // nil when Kind == TagKindInclude
}

// Event is one concrete calendar occurrence.
type Event struct {
	ID           uuid.UUID
	UID          string
	RecurrenceID string
	Summary      string
	Location     string
	Start        time.Time
	End          time.Time
	AllDay       bool
	Categories   []string // normalised
}

// Repository persists people, tag mappings and synced calendar occurrences.
type Repository interface {
	ListPeople(context.Context) ([]Person, error)
	UpsertPerson(context.Context, Person) error
	SetFamily(context.Context, uuid.UUID) error
	DeletePerson(context.Context, uuid.UUID) error
	ListTagMappings(context.Context) ([]TagMapping, error)
	UpsertTagMapping(context.Context, TagMapping) error
	DeleteTagMapping(context.Context, string) error
	ListWindow(context.Context, time.Time, time.Time) ([]Event, error)
	DistinctCategories(context.Context) ([]string, error)
	ReplaceWindow(context.Context, time.Time, time.Time, []Event) error
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

// DeletePerson removes a person. ON DELETE CASCADE drops their tag mappings.
func (r *SQLiteRepository) DeletePerson(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM people WHERE id = ?`, id[:])
	return err
}

type tagMappingRow struct {
	Tag      string  `db:"tag"`
	Label    string  `db:"label"`
	Kind     string  `db:"kind"`
	PersonID *[]byte `db:"person_id"`
}

func (row tagMappingRow) mapping() (TagMapping, error) {
	mapping := TagMapping{Tag: row.Tag, Label: row.Label, Kind: TagKind(row.Kind)}
	if row.PersonID != nil {
		id, err := uuid.FromBytes(*row.PersonID)
		if err != nil {
			return TagMapping{}, fmt.Errorf("decode person id for tag %q: %w", row.Tag, err)
		}
		mapping.PersonID = &id
	}
	return mapping, nil
}

// ListTagMappings returns every stored tag mapping.
func (r *SQLiteRepository) ListTagMappings(ctx context.Context) ([]TagMapping, error) {
	var rows []tagMappingRow
	if err := r.db.SelectContext(ctx, &rows, `SELECT tag, label, kind, person_id FROM tag_mappings ORDER BY tag`); err != nil {
		return nil, err
	}
	mappings := make([]TagMapping, 0, len(rows))
	for _, row := range rows {
		mapping, err := row.mapping()
		if err != nil {
			return nil, err
		}
		mappings = append(mappings, mapping)
	}
	return mappings, nil
}

// UpsertTagMapping inserts or updates a tag mapping. The caller normalises Tag.
func (r *SQLiteRepository) UpsertTagMapping(ctx context.Context, mapping TagMapping) error {
	var personID *[]byte
	if mapping.PersonID != nil {
		b := (*mapping.PersonID)[:]
		personID = &b
	}
	row := tagMappingRow{Tag: mapping.Tag, Label: mapping.Label, Kind: string(mapping.Kind), PersonID: personID}
	_, err := r.db.NamedExecContext(ctx, `
		INSERT INTO tag_mappings (tag, label, kind, person_id)
		VALUES (:tag, :label, :kind, :person_id)
		ON CONFLICT(tag) DO UPDATE SET
			label = excluded.label,
			kind = excluded.kind,
			person_id = excluded.person_id`, row)
	return err
}

// DeleteTagMapping removes a tag mapping by its normalised tag.
func (r *SQLiteRepository) DeleteTagMapping(ctx context.Context, tag string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM tag_mappings WHERE tag = ?`, tag)
	return err
}

type eventRow struct {
	ID           []byte `db:"id"`
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
	return Event{
		ID: id, UID: row.UID, RecurrenceID: row.RecurrenceID, Summary: row.Summary, Location: row.Location,
		Start: time.Unix(row.StartAt, 0), End: time.Unix(row.EndAt, 0), AllDay: row.AllDay == 1,
		Categories: splitCategories(row.Categories),
	}, nil
}

// ListWindow returns events overlapping [from, to), ordered by start time.
// The overlap comparison (rather than BETWEEN) includes an event that began
// before the window but is still running.
func (r *SQLiteRepository) ListWindow(ctx context.Context, from, to time.Time) ([]Event, error) {
	var rows []eventRow
	if err := r.db.SelectContext(ctx, &rows, `
		SELECT id, uid, recurrence_id, summary, location, start_at, end_at, all_day, categories, synced_at
		FROM events
		WHERE start_at < ? AND end_at > ?
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

// DistinctCategories returns every normalised tag seen across stored events.
func (r *SQLiteRepository) DistinctCategories(ctx context.Context) ([]string, error) {
	var raw []string
	if err := r.db.SelectContext(ctx, &raw, `SELECT categories FROM events WHERE categories != ''`); err != nil {
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

// ReplaceWindow deletes rows overlapping [from, to) and inserts the given
// set, in one transaction. An empty events slice empties the window. Ids are
// preserved across the replace for an unchanged (uid, recurrence_id).
func (r *SQLiteRepository) ReplaceWindow(ctx context.Context, from, to time.Time, newEvents []Event) error {
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
		SELECT id, uid, recurrence_id FROM events WHERE start_at < ? AND end_at > ?`, to.Unix(), from.Unix()); err != nil {
		return err
	}
	for _, row := range existingRows {
		existing[row.UID+"\x00"+row.RecurrenceID] = row.ID
	}

	if _, err := tx.ExecContext(ctx, `DELETE FROM events WHERE start_at < ? AND end_at > ?`, to.Unix(), from.Unix()); err != nil {
		return err
	}

	syncedAt := time.Now().Unix()
	for _, event := range newEvents {
		id := event.ID[:]
		if existingID, ok := existing[event.UID+"\x00"+event.RecurrenceID]; ok {
			id = existingID
		}
		row := eventRow{
			ID: id, UID: event.UID, RecurrenceID: event.RecurrenceID, Summary: event.Summary, Location: event.Location,
			StartAt: event.Start.Unix(), EndAt: event.End.Unix(), AllDay: boolToInt(event.AllDay),
			Categories: strings.Join(event.Categories, ","), SyncedAt: syncedAt,
		}
		// UNIQUE (uid, recurrence_id) is global but the delete above is
		// window-scoped, so an occurrence that moved into the window from
		// outside it still has a surviving row. Upsert onto that row rather
		// than colliding with it; this also absorbs a duplicated key within
		// one report, last occurrence winning.
		if _, err := tx.NamedExecContext(ctx, `
			INSERT INTO events (id, uid, recurrence_id, summary, location, start_at, end_at, all_day, categories, synced_at)
			VALUES (:id, :uid, :recurrence_id, :summary, :location, :start_at, :end_at, :all_day, :categories, :synced_at)
			ON CONFLICT(uid, recurrence_id) DO UPDATE SET
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
