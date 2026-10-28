package photos

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

// Coordinate is a photo's optional GPS location.
type Coordinate struct {
	Latitude  float64
	Longitude float64
}

// FocalRegion is a normalized region of interest in top-left-origin coordinates.
type FocalRegion struct {
	X      float64 `json:"x"`
	Y      float64 `json:"y"`
	Width  float64 `json:"width"`
	Height float64 `json:"height"`
	Source string  `json:"source"`
}

// Kind distinguishes where a photo came from.
type Kind string

const (
	KindPhoto   Kind = "PHOTO"
	KindDrawing Kind = "DRAWING"
)

// orPhoto reads the zero Kind as a photo, so a caller that does not care
// about the distinction gets the pre-drawing behaviour.
func (k Kind) orPhoto() Kind {
	if k == "" {
		return KindPhoto
	}
	return k
}

// Photo contains the metadata used to display an imported photo.
type Photo struct {
	ID         uuid.UUID
	AssetID    string
	Version    string
	File       string
	Width      int
	Height     int
	TakenAt    time.Time
	Coordinate *Coordinate
	Place      string
	Focus      *FocalRegion
	Kind       Kind
}

// Showing records one interval during which a photo was displayed.
type Showing struct {
	ID        uuid.UUID
	Photo     Photo
	StartedAt time.Time
	EndsAt    time.Time
	EndedAt   *time.Time
	Source    ShowingSource
}

// ShowingSource describes why the previous display interval ended.
type ShowingSource string

const (
	ShowingSchedule ShowingSource = "SCHEDULE"
	ShowingTap      ShowingSource = "TAP"
)

// ManifestEntry describes the server copy of an Apple photo asset.
type ManifestEntry struct {
	ID      uuid.UUID
	Version string
}

// Repository persists imported photo metadata.
type Repository interface {
	List(context.Context, int) ([]Photo, error)
	Manifest(context.Context) (map[string]ManifestEntry, error)
	IDForAsset(context.Context, string) (uuid.UUID, bool, error)
	Upsert(context.Context, Photo) error
	DeleteMissing(context.Context, []string) ([]string, error)
	Showings(context.Context, int) ([]Showing, error)
	Rotation(context.Context) (*Showing, *Showing, error)
	AppendShowing(context.Context, Showing) error
	NextPhoto(context.Context, uuid.UUID) (Photo, bool, error)
	AdvanceShowing(context.Context, *uuid.UUID, Showing, time.Time, ShowingSource) error
	PruneShowings(context.Context, time.Time) error
	UnpushedDrawings(context.Context) ([]Photo, error)
	MarkDrawingPushed(context.Context, uuid.UUID, string, time.Time) (bool, error)
	PushedAssetIDs(context.Context) ([]string, error)
}

// SQLiteRepository implements Repository with SQLite.
type SQLiteRepository struct {
	db *sqlx.DB
}

// NewSQLiteRepository creates a photo repository backed by SQLite.
func NewSQLiteRepository(db *sqlx.DB) *SQLiteRepository {
	return &SQLiteRepository{db: db}
}

type photoRow struct {
	ID         []byte          `db:"id"`
	AssetID    string          `db:"asset_id"`
	Version    string          `db:"version"`
	File       string          `db:"file"`
	Width      int             `db:"width"`
	Height     int             `db:"height"`
	TakenAt    int64           `db:"taken_at"`
	Latitude   sql.NullFloat64 `db:"latitude"`
	Longitude  sql.NullFloat64 `db:"longitude"`
	Place      sql.NullString  `db:"place"`
	Focus      sql.NullString  `db:"focus"`
	Kind       string          `db:"kind"`
	ImportedAt int64           `db:"imported_at"`
}

type manifestRow struct {
	ID      []byte `db:"id"`
	AssetID string `db:"asset_id"`
	Version string `db:"version"`
}

type showingRow struct {
	ID        []byte          `db:"id"`
	PhotoID   []byte          `db:"photo_id"`
	StartedAt int64           `db:"started_at"`
	EndsAt    int64           `db:"ends_at"`
	EndedAt   sql.NullInt64   `db:"ended_at"`
	Source    string          `db:"source"`
	AssetID   string          `db:"asset_id"`
	Version   string          `db:"version"`
	File      string          `db:"file"`
	Width     int             `db:"width"`
	Height    int             `db:"height"`
	TakenAt   int64           `db:"taken_at"`
	Latitude  sql.NullFloat64 `db:"latitude"`
	Longitude sql.NullFloat64 `db:"longitude"`
	Place     sql.NullString  `db:"place"`
	Focus     sql.NullString  `db:"focus"`
	Kind      string          `db:"kind"`
}

var _ Repository = (*SQLiteRepository)(nil)

// List returns photos newest first, with a stable UUID tie-breaker.
func (r *SQLiteRepository) List(ctx context.Context, limit int) ([]Photo, error) {
	var rows []photoRow
	if err := r.db.SelectContext(ctx, &rows, `
		SELECT id, asset_id, version, file, width, height, taken_at, latitude, longitude, place, focus, kind, imported_at
		FROM photos
		ORDER BY taken_at DESC, id DESC
		LIMIT ?`, limit); err != nil {
		return nil, err
	}
	photos := make([]Photo, 0, len(rows))
	for _, row := range rows {
		photo, err := row.photo()
		if err != nil {
			return nil, err
		}
		photos = append(photos, photo)
	}
	return photos, nil
}

// Manifest returns the stored version and stable id for each Apple asset id.
func (r *SQLiteRepository) Manifest(ctx context.Context) (map[string]ManifestEntry, error) {
	var rows []manifestRow
	if err := r.db.SelectContext(ctx, &rows, `SELECT id, asset_id, version FROM photos WHERE kind = 'PHOTO'`); err != nil {
		return nil, err
	}
	manifest := make(map[string]ManifestEntry, len(rows))
	for _, row := range rows {
		id, err := uuid.FromBytes(row.ID)
		if err != nil {
			return nil, fmt.Errorf("decode photo id for asset %q: %w", row.AssetID, err)
		}
		manifest[row.AssetID] = ManifestEntry{ID: id, Version: row.Version}
	}
	return manifest, nil
}

// IDForAsset returns the stable server id for an Apple asset, if it exists.
func (r *SQLiteRepository) IDForAsset(ctx context.Context, assetID string) (uuid.UUID, bool, error) {
	var raw []byte
	err := r.db.GetContext(ctx, &raw, `SELECT id FROM photos WHERE asset_id = ? AND kind = 'PHOTO'`, assetID)
	if err == sql.ErrNoRows {
		return uuid.Nil, false, nil
	}
	if err != nil {
		return uuid.Nil, false, err
	}
	id, err := uuid.FromBytes(raw)
	if err != nil {
		return uuid.Nil, false, fmt.Errorf("decode photo id for asset %q: %w", assetID, err)
	}
	return id, true, nil
}

// Upsert inserts or updates an asset without changing its stable id.
func (r *SQLiteRepository) Upsert(ctx context.Context, photo Photo) error {
	var latitude, longitude sql.NullFloat64
	if photo.Coordinate != nil {
		latitude = sql.NullFloat64{Float64: photo.Coordinate.Latitude, Valid: true}
		longitude = sql.NullFloat64{Float64: photo.Coordinate.Longitude, Valid: true}
	}
	var focus sql.NullString
	if photo.Focus != nil {
		encoded, err := json.Marshal(photo.Focus)
		if err != nil {
			return fmt.Errorf("encode focal region: %w", err)
		}
		focus = sql.NullString{String: string(encoded), Valid: true}
	}
	row := photoRow{
		ID: photo.ID[:], AssetID: photo.AssetID, Version: photo.Version, File: photo.File,
		Width: photo.Width, Height: photo.Height, TakenAt: photo.TakenAt.Unix(),
		Latitude:   latitude,
		Longitude:  longitude,
		Place:      sql.NullString{String: photo.Place, Valid: photo.Place != ""},
		Focus:      focus,
		Kind:       string(photo.Kind.orPhoto()),
		ImportedAt: time.Now().Unix(),
	}
	_, err := r.db.NamedExecContext(ctx, `
		INSERT INTO photos (id, asset_id, version, file, width, height, taken_at, latitude, longitude, place, focus, kind, imported_at)
		VALUES (:id, :asset_id, :version, :file, :width, :height, :taken_at, :latitude, :longitude, :place, :focus, :kind, :imported_at)
		ON CONFLICT(asset_id) DO UPDATE SET
			version = excluded.version,
			file = excluded.file,
			width = excluded.width,
			height = excluded.height,
			taken_at = excluded.taken_at,
			latitude = excluded.latitude,
			longitude = excluded.longitude,
			place = excluded.place,
			focus = excluded.focus,
			imported_at = excluded.imported_at`, row)
	return err
}

// DeleteMissing deletes photos not in keep and returns their file names.
func (r *SQLiteRepository) DeleteMissing(ctx context.Context, keep []string) ([]string, error) {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()
	if len(keep) == 0 {
		var files []string
		if err := tx.SelectContext(ctx, &files, `SELECT file FROM photos WHERE kind = 'PHOTO'`); err != nil {
			return nil, err
		}
		if _, err := tx.ExecContext(ctx, `DELETE FROM photos WHERE kind = 'PHOTO'`); err != nil {
			return nil, err
		}
		if err := tx.Commit(); err != nil {
			return nil, err
		}
		return files, nil
	}

	if _, err := tx.ExecContext(ctx, `CREATE TEMP TABLE IF NOT EXISTS photos_keep (asset_id TEXT PRIMARY KEY) WITHOUT ROWID`); err != nil {
		return nil, err
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM photos_keep`); err != nil {
		return nil, err
	}
	for start := 0; start < len(keep); start += 900 {
		end := start + 900
		if end > len(keep) {
			end = len(keep)
		}
		values := make([]string, end-start)
		args := make([]any, end-start)
		for i, assetID := range keep[start:end] {
			values[i] = "(?)"
			args[i] = assetID
		}
		if _, err := tx.ExecContext(ctx, `INSERT OR IGNORE INTO photos_keep (asset_id) VALUES `+strings.Join(values, ","), args...); err != nil {
			return nil, err
		}
	}
	var files []string
	if err := tx.SelectContext(ctx, &files, `SELECT file FROM photos WHERE kind = 'PHOTO' AND asset_id NOT IN (SELECT asset_id FROM photos_keep)`); err != nil {
		return nil, err
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM photos WHERE kind = 'PHOTO' AND asset_id NOT IN (SELECT asset_id FROM photos_keep)`); err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return files, nil
}

// Rotation returns the active showing and its scheduled successor, if present.
func (r *SQLiteRepository) Rotation(ctx context.Context) (*Showing, *Showing, error) {
	rows, err := r.queryShowings(ctx, `
		SELECT s.id, s.photo_id, s.started_at, s.ends_at, s.ended_at, s.source,
		       p.asset_id, p.version, p.file, p.width, p.height, p.taken_at,
		       p.latitude, p.longitude, p.place, p.focus, p.kind
		FROM photo_showings s JOIN photos p ON p.id = s.photo_id
		WHERE s.ended_at IS NULL ORDER BY s.started_at ASC, s.id ASC LIMIT 2`)
	if err != nil {
		return nil, nil, err
	}
	if len(rows) == 0 {
		return nil, nil, nil
	}
	current := &rows[0]
	var next *Showing
	if len(rows) > 1 {
		next = &rows[1]
	}
	return current, next, nil
}

// Showings returns recent showing history, newest first.
func (r *SQLiteRepository) Showings(ctx context.Context, limit int) ([]Showing, error) {
	if limit <= 0 {
		limit = 50
	}
	return r.queryShowings(ctx, `
		SELECT s.id, s.photo_id, s.started_at, s.ends_at, s.ended_at, s.source,
		       p.asset_id, p.version, p.file, p.width, p.height, p.taken_at,
		       p.latitude, p.longitude, p.place, p.focus, p.kind
		FROM photo_showings s JOIN photos p ON p.id = s.photo_id
		ORDER BY s.started_at DESC, s.id DESC LIMIT ?`, limit)
}

func (r *SQLiteRepository) queryShowings(ctx context.Context, query string, args ...any) ([]Showing, error) {
	var rows []showingRow
	if err := r.db.SelectContext(ctx, &rows, query, args...); err != nil {
		return nil, err
	}
	result := make([]Showing, 0, len(rows))
	for _, row := range rows {
		photo, err := (photoRow{ID: row.PhotoID, AssetID: row.AssetID, Version: row.Version, File: row.File,
			Width: row.Width, Height: row.Height, TakenAt: row.TakenAt, Latitude: row.Latitude,
			Longitude: row.Longitude, Place: row.Place, Focus: row.Focus, Kind: row.Kind}).photo()
		if err != nil {
			return nil, err
		}
		id, err := uuid.FromBytes(row.ID)
		if err != nil {
			return nil, fmt.Errorf("decode showing id: %w", err)
		}
		showing := Showing{ID: id, Photo: photo, StartedAt: time.Unix(row.StartedAt, 0), EndsAt: time.Unix(row.EndsAt, 0), Source: ShowingSource(row.Source)}
		if row.EndedAt.Valid {
			ended := time.Unix(row.EndedAt.Int64, 0)
			showing.EndedAt = &ended
		}
		result = append(result, showing)
	}
	return result, nil
}

// AppendShowing persists a scheduled showing.
func (r *SQLiteRepository) AppendShowing(ctx context.Context, showing Showing) error {
	_, err := r.db.ExecContext(ctx, `INSERT INTO photo_showings (id, photo_id, started_at, ends_at, ended_at, source) VALUES (?, ?, ?, ?, ?, ?)`,
		showing.ID[:], showing.Photo.ID[:], showing.StartedAt.Unix(), showing.EndsAt.Unix(), nullableUnix(showing.EndedAt), showing.Source)
	return err
}

// AdvanceShowing atomically closes the current row when needed and appends a
// new showing. expectedCurrent nil means the timeline is still empty.
func (r *SQLiteRepository) AdvanceShowing(ctx context.Context, expectedCurrent *uuid.UUID, showing Showing, endedAt time.Time, source ShowingSource) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	var currentID []byte
	if expectedCurrent != nil {
		err = tx.GetContext(ctx, &currentID, `SELECT id FROM photo_showings WHERE id = ? AND ended_at IS NULL`, expectedCurrent[:])
	} else {
		err = tx.GetContext(ctx, &currentID, `SELECT id FROM photo_showings WHERE ended_at IS NULL LIMIT 1`)
	}
	if err == sql.ErrNoRows {
		if expectedCurrent != nil {
			return fmt.Errorf("photo rotation changed while advancing")
		}
	} else if err != nil {
		return err
	} else if expectedCurrent == nil || string(currentID) != string(expectedCurrent[:]) {
		return fmt.Errorf("photo rotation changed while advancing")
	} else {
		if _, err := tx.ExecContext(ctx, `UPDATE photo_showings SET ended_at = ? WHERE id = ? AND ended_at IS NULL`, endedAt.Unix(), currentID); err != nil {
			return err
		}
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO photo_showings (id, photo_id, started_at, ends_at, ended_at, source) VALUES (?, ?, ?, ?, ?, ?)`,
		showing.ID[:], showing.Photo.ID[:], showing.StartedAt.Unix(), showing.EndsAt.Unix(), nullableUnix(showing.EndedAt), showing.Source); err != nil {
		return err
	}
	return tx.Commit()
}

// NextPhoto returns the next photo in the existing newest-first ordering, wrapping at the end.
func (r *SQLiteRepository) NextPhoto(ctx context.Context, last uuid.UUID) (Photo, bool, error) {
	var row photoRow
	err := r.db.GetContext(ctx, &row, `
		SELECT id, asset_id, version, file, width, height, taken_at, latitude, longitude, place, focus, kind, imported_at
		FROM photos WHERE (taken_at, id) < (SELECT taken_at, id FROM photos WHERE id = ?)
		ORDER BY taken_at DESC, id DESC LIMIT 1`, last[:])
	if err == sql.ErrNoRows {
		err = r.db.GetContext(ctx, &row, `SELECT id, asset_id, version, file, width, height, taken_at, latitude, longitude, place, focus, kind, imported_at FROM photos ORDER BY taken_at DESC, id DESC LIMIT 1`)
	}
	if err == sql.ErrNoRows {
		return Photo{}, false, nil
	}
	if err != nil {
		return Photo{}, false, err
	}
	photo, err := row.photo()
	return photo, err == nil, err
}

// PruneShowings removes history older than before.
func (r *SQLiteRepository) PruneShowings(ctx context.Context, before time.Time) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM photo_showings WHERE started_at < ?`, before.Unix())
	return err
}

// UnpushedDrawings returns drawings not yet pushed to Photos, oldest first.
func (r *SQLiteRepository) UnpushedDrawings(ctx context.Context) ([]Photo, error) {
	var rows []photoRow
	if err := r.db.SelectContext(ctx, &rows, `
		SELECT id, asset_id, version, file, width, height, taken_at, latitude, longitude, place, focus, kind, imported_at
		FROM photos
		WHERE kind = 'DRAWING' AND pushed_asset_id IS NULL
		ORDER BY taken_at ASC, id ASC`); err != nil {
		return nil, err
	}
	result := make([]Photo, 0, len(rows))
	for _, row := range rows {
		photo, err := row.photo()
		if err != nil {
			return nil, err
		}
		result = append(result, photo)
	}
	return result, nil
}

// MarkDrawingPushed records the Photos asset created for a drawing. The first
// mark wins; repeating it changes nothing. It reports whether the drawing exists.
func (r *SQLiteRepository) MarkDrawingPushed(ctx context.Context, id uuid.UUID, assetID string, at time.Time) (bool, error) {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return false, err
	}
	defer func() { _ = tx.Rollback() }()
	var exists int
	if err := tx.GetContext(ctx, &exists, `SELECT COUNT(*) FROM photos WHERE id = ? AND kind = 'DRAWING'`, id[:]); err != nil {
		return false, err
	}
	if exists == 0 {
		return false, nil
	}
	if _, err := tx.ExecContext(ctx, `
		UPDATE photos SET pushed_asset_id = ?, pushed_at = ?
		WHERE id = ? AND kind = 'DRAWING' AND pushed_asset_id IS NULL`, assetID, at.Unix(), id[:]); err != nil {
		return false, err
	}
	return true, tx.Commit()
}

// PushedAssetIDs returns the Photos asset identifiers recorded for drawings.
func (r *SQLiteRepository) PushedAssetIDs(ctx context.Context) ([]string, error) {
	ids := []string{}
	if err := r.db.SelectContext(ctx, &ids, `SELECT pushed_asset_id FROM photos WHERE pushed_asset_id IS NOT NULL ORDER BY pushed_at ASC, id ASC`); err != nil {
		return nil, err
	}
	return ids, nil
}

func nullableUnix(value *time.Time) any {
	if value == nil {
		return nil
	}
	return value.Unix()
}

func (row photoRow) photo() (Photo, error) {
	id, err := uuid.FromBytes(row.ID)
	if err != nil {
		return Photo{}, fmt.Errorf("decode photo id for asset %q: %w", row.AssetID, err)
	}
	var coordinate *Coordinate
	if row.Latitude.Valid != row.Longitude.Valid {
		return Photo{}, fmt.Errorf("photo %q has incomplete coordinates", row.AssetID)
	}
	if row.Latitude.Valid {
		coordinate = &Coordinate{Latitude: row.Latitude.Float64, Longitude: row.Longitude.Float64}
	}
	var focus *FocalRegion
	if row.Focus.Valid {
		focus = &FocalRegion{}
		if err := json.Unmarshal([]byte(row.Focus.String), focus); err != nil {
			return Photo{}, fmt.Errorf("decode focal region for photo %q: %w", row.AssetID, err)
		}
	}
	return Photo{
		ID: id, AssetID: row.AssetID, Version: row.Version, File: row.File,
		Width: row.Width, Height: row.Height, TakenAt: time.Unix(row.TakenAt, 0),
		Coordinate: coordinate, Place: row.Place.String, Focus: focus, Kind: Kind(row.Kind),
	}, nil
}
