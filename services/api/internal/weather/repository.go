package weather

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/jmoiron/sqlx"
)

// Repository persists upstream weather conditions across API restarts.
type Repository interface {
	Load(context.Context, string) (CacheEntry, bool, error)
	Save(context.Context, string, CacheEntry) error
}

// SQLiteRepository implements Repository with SQLite.
type SQLiteRepository struct {
	db *sqlx.DB
}

// NewSQLiteRepository creates a weather repository backed by SQLite.
func NewSQLiteRepository(db *sqlx.DB) *SQLiteRepository {
	return &SQLiteRepository{db: db}
}

type currentRow struct {
	Location      string  `db:"location"`
	ConditionCode string  `db:"condition_code"`
	Temperature   float64 `db:"temperature"`
	Daylight      bool    `db:"daylight"`
	FetchedAt     int64   `db:"fetched_at"`
}

// Load reads a cached current-weather entry for a location.
func (r *SQLiteRepository) Load(ctx context.Context, location string) (CacheEntry, bool, error) {
	var row currentRow
	err := r.db.GetContext(ctx, &row, `
		SELECT location, condition_code, temperature, daylight, fetched_at
		FROM weather_current WHERE location = ?`, location,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return CacheEntry{}, false, nil
	}
	if err != nil {
		return CacheEntry{}, false, err
	}
	return CacheEntry{
		ConditionCode: row.ConditionCode,
		Temperature:   row.Temperature,
		Daylight:      row.Daylight,
		FetchedAt:     time.Unix(row.FetchedAt, 0),
	}, true, nil
}

// Save inserts or replaces a cached current-weather entry.
func (r *SQLiteRepository) Save(ctx context.Context, location string, entry CacheEntry) error {
	row := currentRow{
		Location:      location,
		ConditionCode: entry.ConditionCode,
		Temperature:   entry.Temperature,
		Daylight:      entry.Daylight,
		FetchedAt:     entry.FetchedAt.Unix(),
	}
	_, err := r.db.NamedExecContext(ctx, `
		INSERT INTO weather_current (location, condition_code, temperature, daylight, fetched_at)
		VALUES (:location, :condition_code, :temperature, :daylight, :fetched_at)
		ON CONFLICT(location) DO UPDATE SET
			condition_code = excluded.condition_code,
			temperature = excluded.temperature,
			daylight = excluded.daylight,
			fetched_at = excluded.fetched_at`, row,
	)
	return err
}
