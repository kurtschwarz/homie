package calendars

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type SQLiteRepository struct{ db *sqlx.DB }

func NewSQLiteRepository(db *sqlx.DB) *SQLiteRepository { return &SQLiteRepository{db: db} }

type accountRow struct {
	ID            []byte `db:"id"`
	Name          string `db:"name"`
	Provider      string `db:"provider"`
	Config        string `db:"config"`
	Credentials   []byte `db:"credentials"`
	Refresh       int64  `db:"refresh_interval_seconds"`
	Past          int64  `db:"window_past_seconds"`
	Future        int64  `db:"window_future_seconds"`
	LastAttempted *int64 `db:"last_attempted_at"`
}
type accountConfig struct {
	URL      string `json:"url"`
	Username string `json:"username"`
	Timezone string `json:"timezone,omitempty"`
}

func (row accountRow) account() (storedAccount, error) {
	id, err := uuid.FromBytes(row.ID)
	if err != nil {
		return storedAccount{}, err
	}
	var config accountConfig
	if err := json.Unmarshal([]byte(row.Config), &config); err != nil {
		return storedAccount{}, fmt.Errorf("decode calendar account %s config: %w", id, err)
	}
	account := Account{ID: id, Name: row.Name, Provider: row.Provider, RefreshInterval: time.Duration(row.Refresh) * time.Second, WindowPast: time.Duration(row.Past) * time.Second, WindowFuture: time.Duration(row.Future) * time.Second}
	if row.Provider == ProviderCalDAV {
		account.CalDAV = &CalDAVConnection{URL: config.URL, Username: config.Username, Timezone: config.Timezone}
	}
	if row.LastAttempted != nil {
		value := time.Unix(*row.LastAttempted, 0)
		account.LastAttemptedAt = &value
	}
	return storedAccount{Account: account, Credentials: row.Credentials}, nil
}

const accountColumns = `id, name, provider, config, credentials, refresh_interval_seconds, window_past_seconds, window_future_seconds, last_attempted_at`

func (r *SQLiteRepository) ListAccounts(ctx context.Context) ([]storedAccount, error) {
	var rows []accountRow
	if err := r.db.SelectContext(ctx, &rows, `SELECT `+accountColumns+` FROM calendar_accounts ORDER BY name`); err != nil {
		return nil, err
	}
	result := make([]storedAccount, 0, len(rows))
	for _, row := range rows {
		account, err := row.account()
		if err != nil {
			return nil, err
		}
		result = append(result, account)
	}
	return result, nil
}
func (r *SQLiteRepository) GetAccount(ctx context.Context, id uuid.UUID) (storedAccount, error) {
	var row accountRow
	if err := r.db.GetContext(ctx, &row, `SELECT `+accountColumns+` FROM calendar_accounts WHERE id=?`, id[:]); errors.Is(err, sql.ErrNoRows) {
		return storedAccount{}, ErrNotFound
	} else if err != nil {
		return storedAccount{}, err
	}
	return row.account()
}
func (r *SQLiteRepository) UpsertAccount(ctx context.Context, stored storedAccount) error {
	if stored.CalDAV == nil {
		return fmt.Errorf("calendar account %s has no provider connection", stored.ID)
	}
	config, err := json.Marshal(accountConfig{URL: stored.CalDAV.URL, Username: stored.CalDAV.Username, Timezone: stored.CalDAV.Timezone})
	if err != nil {
		return err
	}
	_, err = r.db.ExecContext(ctx, `INSERT INTO calendar_accounts (id,name,provider,config,credentials,refresh_interval_seconds,window_past_seconds,window_future_seconds) VALUES (?,?,?,?,?,?,?,?) ON CONFLICT(id) DO UPDATE SET name=excluded.name,provider=excluded.provider,config=excluded.config,credentials=excluded.credentials,refresh_interval_seconds=excluded.refresh_interval_seconds,window_past_seconds=excluded.window_past_seconds,window_future_seconds=excluded.window_future_seconds`, stored.ID[:], stored.Name, stored.Provider, string(config), stored.Credentials, int64(stored.RefreshInterval/time.Second), int64(stored.WindowPast/time.Second), int64(stored.WindowFuture/time.Second))
	return err
}
func (r *SQLiteRepository) DeleteAccount(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM calendar_accounts WHERE id=?`, id[:])
	return err
}

type calendarRow struct {
	ID          []byte `db:"id"`
	AccountID   []byte `db:"account_id"`
	ProviderRef string `db:"provider_ref"`
	Name        string `db:"name"`
	Visible     int    `db:"visible"`
	ReadOnly    int    `db:"read_only"`
	SortOrder   int    `db:"sort_order"`
	LastSynced  *int64 `db:"last_synced_at"`
	LastError   string `db:"last_error"`
}

func (row calendarRow) calendar() (Calendar, error) {
	id, err := uuid.FromBytes(row.ID)
	if err != nil {
		return Calendar{}, err
	}
	accountID, err := uuid.FromBytes(row.AccountID)
	if err != nil {
		return Calendar{}, err
	}
	calendar := Calendar{ID: id, AccountID: accountID, ProviderRef: row.ProviderRef, Name: row.Name, Visible: row.Visible == 1, ReadOnly: row.ReadOnly == 1, Order: row.SortOrder, LastError: row.LastError}
	if row.LastSynced != nil {
		value := time.Unix(*row.LastSynced, 0)
		calendar.LastSyncedAt = &value
	}
	return calendar, nil
}

const calendarColumns = `id, account_id, provider_ref, name, visible, read_only, sort_order, last_synced_at, last_error`

func (r *SQLiteRepository) ListCalendars(ctx context.Context, accountID *uuid.UUID) ([]Calendar, error) {
	query := `SELECT ` + calendarColumns + ` FROM calendars`
	var args []any
	if accountID != nil {
		query += ` WHERE account_id=?`
		args = append(args, accountID[:])
	}
	query += ` ORDER BY sort_order,name`
	var rows []calendarRow
	if err := r.db.SelectContext(ctx, &rows, query, args...); err != nil {
		return nil, err
	}
	result := make([]Calendar, 0, len(rows))
	for _, row := range rows {
		calendar, err := row.calendar()
		if err != nil {
			return nil, err
		}
		result = append(result, calendar)
	}
	return result, nil
}
func (r *SQLiteRepository) GetCalendar(ctx context.Context, id uuid.UUID) (Calendar, error) {
	var row calendarRow
	if err := r.db.GetContext(ctx, &row, `SELECT `+calendarColumns+` FROM calendars WHERE id=?`, id[:]); errors.Is(err, sql.ErrNoRows) {
		return Calendar{}, ErrNotFound
	} else if err != nil {
		return Calendar{}, err
	}
	return row.calendar()
}
func (r *SQLiteRepository) UpsertCalendar(ctx context.Context, calendar Calendar) error {
	_, err := r.db.ExecContext(ctx, `INSERT INTO calendars (id,account_id,provider_ref,name,visible,read_only,sort_order,last_error) VALUES (?,?,?,?,?,?,?,?) ON CONFLICT(id) DO UPDATE SET name=excluded.name,visible=excluded.visible,read_only=excluded.read_only,sort_order=excluded.sort_order`, calendar.ID[:], calendar.AccountID[:], calendar.ProviderRef, calendar.Name, boolInt(calendar.Visible), boolInt(calendar.ReadOnly), calendar.Order, calendar.LastError)
	return err
}
func (r *SQLiteRepository) UpsertDiscoveredCalendar(ctx context.Context, calendar Calendar) error {
	_, err := r.db.ExecContext(ctx, `INSERT INTO calendars (id,account_id,provider_ref,name,visible,read_only,sort_order) VALUES (?,?,?,?,?,?,?) ON CONFLICT(account_id,provider_ref) DO UPDATE SET name=excluded.name`, calendar.ID[:], calendar.AccountID[:], calendar.ProviderRef, calendar.Name, 0, boolInt(calendar.ReadOnly), calendar.Order)
	return err
}
func (r *SQLiteRepository) SetCalendarVisible(ctx context.Context, id uuid.UUID, visible bool) error {
	result, err := r.db.ExecContext(ctx, `UPDATE calendars SET visible=? WHERE id=?`, boolInt(visible), id[:])
	if err != nil {
		return err
	}
	changed, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if changed == 0 {
		return ErrNotFound
	}
	return nil
}
func (r *SQLiteRepository) RecordSync(ctx context.Context, id uuid.UUID, at *time.Time, message string) error {
	var unix *int64
	if at != nil {
		value := at.Unix()
		unix = &value
	}
	_, err := r.db.ExecContext(ctx, `UPDATE calendars SET last_synced_at=COALESCE(?,last_synced_at),last_error=? WHERE id=?`, unix, message, id[:])
	return err
}
func (r *SQLiteRepository) RecordAttempt(ctx context.Context, id uuid.UUID, at time.Time) error {
	_, err := r.db.ExecContext(ctx, `UPDATE calendar_accounts SET last_attempted_at=? WHERE id=?`, at.Unix(), id[:])
	return err
}
func boolInt(value bool) int {
	if value {
		return 1
	}
	return 0
}
