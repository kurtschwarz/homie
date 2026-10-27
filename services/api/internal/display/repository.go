package display

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"
)

// Repository persists the display schedule and settings.
type Repository interface {
	Load(context.Context) (Schedule, error)
	Save(context.Context, Schedule) error
}

// SQLiteRepository implements Repository with SQLite.
type SQLiteRepository struct{ db *sqlx.DB }

// NewSQLiteRepository creates a display repository backed by SQLite.
func NewSQLiteRepository(db *sqlx.DB) *SQLiteRepository { return &SQLiteRepository{db: db} }

type modeRow struct {
	Mode            string `db:"mode"`
	StartsAtMinutes int    `db:"starts_at_minutes"`
	Brightness      int    `db:"brightness"`
}

type settingsRow struct {
	PowerSavingEnabled bool   `db:"power_saving_enabled"`
	WakeForAlerts      bool   `db:"wake_for_alerts"`
	DDCBus             string `db:"ddc_bus"`
}

func (r *SQLiteRepository) Load(ctx context.Context) (Schedule, error) {
	var rows []modeRow
	if err := r.db.SelectContext(ctx, &rows, `SELECT mode, starts_at_minutes, brightness FROM display_modes ORDER BY starts_at_minutes`); err != nil {
		return Schedule{}, err
	}
	var settings settingsRow
	if err := r.db.GetContext(ctx, &settings, `SELECT power_saving_enabled, wake_for_alerts, ddc_bus FROM display_settings WHERE id=1`); err != nil {
		return Schedule{}, err
	}
	schedule := Schedule{Settings: Settings{PowerSavingEnabled: settings.PowerSavingEnabled, WakeForAlerts: settings.WakeForAlerts, DDCBus: settings.DDCBus}}
	for _, row := range rows {
		schedule.Modes = append(schedule.Modes, ModeSetting{Mode: Mode(row.Mode), StartsAtMinutes: row.StartsAtMinutes, Brightness: row.Brightness})
	}
	return schedule, nil
}

func (r *SQLiteRepository) Save(ctx context.Context, schedule Schedule) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	for _, setting := range schedule.Modes {
		if _, err := tx.ExecContext(ctx, `UPDATE display_modes SET starts_at_minutes=?, brightness=? WHERE mode=?`, setting.StartsAtMinutes, setting.Brightness, string(setting.Mode)); err != nil {
			return fmt.Errorf("save display mode %s: %w", setting.Mode, err)
		}
	}
	if _, err := tx.ExecContext(ctx, `UPDATE display_settings SET power_saving_enabled=?, wake_for_alerts=?, ddc_bus=? WHERE id=1`, schedule.PowerSavingEnabled, schedule.WakeForAlerts, schedule.DDCBus); err != nil {
		return err
	}
	return tx.Commit()
}
