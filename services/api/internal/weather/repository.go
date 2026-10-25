package weather

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

// Repository persists weather providers, locations, and cached conditions.
type Repository interface {
	ListProviders(context.Context) ([]storedProvider, error)
	GetProvider(context.Context, uuid.UUID) (storedProvider, error)
	GetActiveProvider(context.Context) (storedProvider, error)
	UpsertProvider(context.Context, storedProvider) error
	DeleteProvider(context.Context, uuid.UUID) error
	ActivateProvider(context.Context, uuid.UUID) error
	RecordAttempt(context.Context, uuid.UUID, time.Time, string) error

	ListLocations(context.Context) ([]Location, error)
	GetLocation(context.Context, uuid.UUID) (Location, error)
	UpsertLocation(context.Context, Location) error
	DeleteLocation(context.Context, uuid.UUID) error

	Load(context.Context, uuid.UUID) (CacheEntry, bool, error)
	Save(context.Context, uuid.UUID, CacheEntry) error
}

// SQLiteRepository implements Repository with SQLite.
type SQLiteRepository struct{ db *sqlx.DB }

// NewSQLiteRepository creates a weather repository backed by SQLite.
func NewSQLiteRepository(db *sqlx.DB) *SQLiteRepository { return &SQLiteRepository{db: db} }

type providerRow struct {
	ID            []byte `db:"id"`
	Provider      string `db:"provider"`
	Config        string `db:"config"`
	Credentials   []byte `db:"credentials"`
	Active        int    `db:"active"`
	Refresh       int64  `db:"refresh_interval_seconds"`
	LastAttempted *int64 `db:"last_attempted_at"`
	LastError     string `db:"last_error"`
}

type weatherKitConfig struct {
	TeamID    string `json:"teamId"`
	ServiceID string `json:"serviceId"`
	KeyID     string `json:"keyId"`
}

const providerColumns = `id, provider, config, credentials, active, refresh_interval_seconds, last_attempted_at, last_error`

func (row providerRow) provider() (storedProvider, error) {
	id, err := uuid.FromBytes(row.ID)
	if err != nil {
		return storedProvider{}, err
	}
	provider := Provider{ID: id, Kind: row.Provider, Active: row.Active == 1, RefreshInterval: time.Duration(row.Refresh) * time.Second, LastError: row.LastError}
	if row.Provider == ProviderWeatherKit {
		var config weatherKitConfig
		if err := json.Unmarshal([]byte(row.Config), &config); err != nil {
			return storedProvider{}, fmt.Errorf("decode weather provider %s config: %w", id, err)
		}
		provider.WeatherKit = &WeatherKitConnection{TeamID: config.TeamID, ServiceID: config.ServiceID, KeyID: config.KeyID}
	}
	if row.LastAttempted != nil {
		value := time.Unix(*row.LastAttempted, 0)
		provider.LastAttemptedAt = &value
	}
	return storedProvider{Provider: provider, Credentials: row.Credentials}, nil
}

func (r *SQLiteRepository) ListProviders(ctx context.Context) ([]storedProvider, error) {
	var rows []providerRow
	if err := r.db.SelectContext(ctx, &rows, `SELECT `+providerColumns+` FROM weather_providers ORDER BY provider`); err != nil {
		return nil, err
	}
	result := make([]storedProvider, 0, len(rows))
	for _, row := range rows {
		provider, err := row.provider()
		if err != nil {
			return nil, err
		}
		result = append(result, provider)
	}
	return result, nil
}

func (r *SQLiteRepository) GetProvider(ctx context.Context, id uuid.UUID) (storedProvider, error) {
	var row providerRow
	if err := r.db.GetContext(ctx, &row, `SELECT `+providerColumns+` FROM weather_providers WHERE id=?`, id[:]); errors.Is(err, sql.ErrNoRows) {
		return storedProvider{}, ErrNotFound
	} else if err != nil {
		return storedProvider{}, err
	}
	return row.provider()
}

func (r *SQLiteRepository) GetActiveProvider(ctx context.Context) (storedProvider, error) {
	var row providerRow
	if err := r.db.GetContext(ctx, &row, `SELECT `+providerColumns+` FROM weather_providers WHERE active=1`); errors.Is(err, sql.ErrNoRows) {
		return storedProvider{}, ErrNotFound
	} else if err != nil {
		return storedProvider{}, err
	}
	return row.provider()
}

// UpsertProvider never changes which provider is active; use ActivateProvider.
func (r *SQLiteRepository) UpsertProvider(ctx context.Context, stored storedProvider) error {
	if stored.WeatherKit == nil {
		return fmt.Errorf("weather provider %s has no provider connection", stored.ID)
	}
	config, err := json.Marshal(weatherKitConfig{TeamID: stored.WeatherKit.TeamID, ServiceID: stored.WeatherKit.ServiceID, KeyID: stored.WeatherKit.KeyID})
	if err != nil {
		return err
	}
	_, err = r.db.ExecContext(ctx, `INSERT INTO weather_providers (id,provider,config,credentials,refresh_interval_seconds) VALUES (?,?,?,?,?) ON CONFLICT(id) DO UPDATE SET config=excluded.config,credentials=excluded.credentials,refresh_interval_seconds=excluded.refresh_interval_seconds`, stored.ID[:], stored.Kind, string(config), stored.Credentials, int64(stored.RefreshInterval/time.Second))
	if err != nil && isUniqueViolation(err) {
		return ErrDuplicateProvider
	}
	return err
}

func (r *SQLiteRepository) DeleteProvider(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM weather_providers WHERE id=?`, id[:])
	return err
}

// ActivateProvider makes one provider the only active one.
func (r *SQLiteRepository) ActivateProvider(ctx context.Context, id uuid.UUID) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	if _, err := tx.ExecContext(ctx, `UPDATE weather_providers SET active=0 WHERE active=1 AND id<>?`, id[:]); err != nil {
		return err
	}
	result, err := tx.ExecContext(ctx, `UPDATE weather_providers SET active=1 WHERE id=?`, id[:])
	if err != nil {
		return err
	}
	if changed, err := result.RowsAffected(); err != nil {
		return err
	} else if changed == 0 {
		return ErrNotFound
	}
	return tx.Commit()
}

func (r *SQLiteRepository) RecordAttempt(ctx context.Context, id uuid.UUID, at time.Time, message string) error {
	_, err := r.db.ExecContext(ctx, `UPDATE weather_providers SET last_attempted_at=?,last_error=? WHERE id=?`, at.Unix(), message, id[:])
	return err
}

type locationRow struct {
	ID        []byte  `db:"id"`
	Label     string  `db:"label"`
	Latitude  float64 `db:"latitude"`
	Longitude float64 `db:"longitude"`
	SortOrder int     `db:"sort_order"`
}

func (row locationRow) location() (Location, error) {
	id, err := uuid.FromBytes(row.ID)
	if err != nil {
		return Location{}, err
	}
	return Location{ID: id, Label: row.Label, Latitude: row.Latitude, Longitude: row.Longitude, Order: row.SortOrder}, nil
}

const locationColumns = `id, label, latitude, longitude, sort_order`

func (r *SQLiteRepository) ListLocations(ctx context.Context) ([]Location, error) {
	var rows []locationRow
	if err := r.db.SelectContext(ctx, &rows, `SELECT `+locationColumns+` FROM weather_locations ORDER BY sort_order,label`); err != nil {
		return nil, err
	}
	result := make([]Location, 0, len(rows))
	for _, row := range rows {
		location, err := row.location()
		if err != nil {
			return nil, err
		}
		result = append(result, location)
	}
	return result, nil
}

func (r *SQLiteRepository) GetLocation(ctx context.Context, id uuid.UUID) (Location, error) {
	var row locationRow
	if err := r.db.GetContext(ctx, &row, `SELECT `+locationColumns+` FROM weather_locations WHERE id=?`, id[:]); errors.Is(err, sql.ErrNoRows) {
		return Location{}, ErrNotFound
	} else if err != nil {
		return Location{}, err
	}
	return row.location()
}

func (r *SQLiteRepository) UpsertLocation(ctx context.Context, location Location) error {
	_, err := r.db.ExecContext(ctx, `INSERT INTO weather_locations (id,label,latitude,longitude,sort_order) VALUES (?,?,?,?,?) ON CONFLICT(id) DO UPDATE SET label=excluded.label,latitude=excluded.latitude,longitude=excluded.longitude,sort_order=excluded.sort_order`, location.ID[:], location.Label, location.Latitude, location.Longitude, location.Order)
	return err
}

func (r *SQLiteRepository) DeleteLocation(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM weather_locations WHERE id=?`, id[:])
	return err
}

type currentRow struct {
	LocationID    []byte  `db:"location_id"`
	ConditionCode string  `db:"condition_code"`
	Temperature   float64 `db:"temperature"`
	Daylight      bool    `db:"daylight"`
	FetchedAt     int64   `db:"fetched_at"`
}

// Load reads the cached current weather for a location.
func (r *SQLiteRepository) Load(ctx context.Context, locationID uuid.UUID) (CacheEntry, bool, error) {
	var row currentRow
	err := r.db.GetContext(ctx, &row, `SELECT location_id, condition_code, temperature, daylight, fetched_at FROM weather_current WHERE location_id = ?`, locationID[:])
	if errors.Is(err, sql.ErrNoRows) {
		return CacheEntry{}, false, nil
	}
	if err != nil {
		return CacheEntry{}, false, err
	}
	return CacheEntry{ConditionCode: row.ConditionCode, Temperature: row.Temperature, Daylight: row.Daylight, FetchedAt: time.Unix(row.FetchedAt, 0)}, true, nil
}

// Save inserts or replaces the cached current weather for a location.
func (r *SQLiteRepository) Save(ctx context.Context, locationID uuid.UUID, entry CacheEntry) error {
	row := currentRow{LocationID: locationID[:], ConditionCode: entry.ConditionCode, Temperature: entry.Temperature, Daylight: entry.Daylight, FetchedAt: entry.FetchedAt.Unix()}
	_, err := r.db.NamedExecContext(ctx, `
		INSERT INTO weather_current (location_id, condition_code, temperature, daylight, fetched_at)
		VALUES (:location_id, :condition_code, :temperature, :daylight, :fetched_at)
		ON CONFLICT(location_id) DO UPDATE SET
			condition_code = excluded.condition_code,
			temperature = excluded.temperature,
			daylight = excluded.daylight,
			fetched_at = excluded.fetched_at`, row)
	return err
}

func isUniqueViolation(err error) bool {
	return err != nil && strings.Contains(err.Error(), "UNIQUE constraint failed")
}
