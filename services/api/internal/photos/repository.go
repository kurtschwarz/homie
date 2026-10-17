package photos

import (
	"context"
	"database/sql"
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
}

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
	ImportedAt int64           `db:"imported_at"`
}

type manifestRow struct {
	ID      []byte `db:"id"`
	AssetID string `db:"asset_id"`
	Version string `db:"version"`
}

// List returns photos newest first, with a stable UUID tie-breaker.
func (r *SQLiteRepository) List(ctx context.Context, limit int) ([]Photo, error) {
	var rows []photoRow
	if err := r.db.SelectContext(ctx, &rows, `
		SELECT id, asset_id, version, file, width, height, taken_at, latitude, longitude, place, imported_at
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
	if err := r.db.SelectContext(ctx, &rows, `SELECT id, asset_id, version FROM photos`); err != nil {
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
	err := r.db.GetContext(ctx, &raw, `SELECT id FROM photos WHERE asset_id = ?`, assetID)
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
	row := photoRow{
		ID: photo.ID[:], AssetID: photo.AssetID, Version: photo.Version, File: photo.File,
		Width: photo.Width, Height: photo.Height, TakenAt: photo.TakenAt.Unix(),
		Latitude:   latitude,
		Longitude:  longitude,
		Place:      sql.NullString{String: photo.Place, Valid: photo.Place != ""},
		ImportedAt: time.Now().Unix(),
	}
	_, err := r.db.NamedExecContext(ctx, `
		INSERT INTO photos (id, asset_id, version, file, width, height, taken_at, latitude, longitude, place, imported_at)
		VALUES (:id, :asset_id, :version, :file, :width, :height, :taken_at, :latitude, :longitude, :place, :imported_at)
		ON CONFLICT(asset_id) DO UPDATE SET
			version = excluded.version,
			file = excluded.file,
			width = excluded.width,
			height = excluded.height,
			taken_at = excluded.taken_at,
			latitude = excluded.latitude,
			longitude = excluded.longitude,
			place = excluded.place,
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
		if err := tx.SelectContext(ctx, &files, `SELECT file FROM photos`); err != nil {
			return nil, err
		}
		if _, err := tx.ExecContext(ctx, `DELETE FROM photos`); err != nil {
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
	if err := tx.SelectContext(ctx, &files, `SELECT file FROM photos WHERE asset_id NOT IN (SELECT asset_id FROM photos_keep)`); err != nil {
		return nil, err
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM photos WHERE asset_id NOT IN (SELECT asset_id FROM photos_keep)`); err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return files, nil
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
	return Photo{
		ID: id, AssetID: row.AssetID, Version: row.Version, File: row.File,
		Width: row.Width, Height: row.Height, TakenAt: time.Unix(row.TakenAt, 0),
		Coordinate: coordinate, Place: row.Place.String,
	}, nil
}
