// Package jobs contains River job definitions for weather operations.
package jobs

import (
	"context"
	"log/slog"
	"time"

	"github.com/kurtschwarz/homie/services/api/internal/weather"
	"github.com/riverqueue/river"
	"github.com/riverqueue/river/rivertype"
)

const maxAttempts = 3

// Refresher is implemented by the weather service.
type Refresher interface {
	Refresh(context.Context) error
}

var _ Refresher = (*weather.Service)(nil)

// RefreshArgs identifies a background weather refresh.
type RefreshArgs struct{}

func (RefreshArgs) Kind() string { return "weather_refresh" }

func (RefreshArgs) InsertOpts() river.InsertOpts {
	return river.InsertOpts{
		MaxAttempts: maxAttempts,
		UniqueOpts: river.UniqueOpts{
			ByPeriod: weather.RefreshInterval,
			ByState: []rivertype.JobState{
				rivertype.JobStateAvailable,
				rivertype.JobStatePending,
				rivertype.JobStateRunning,
				rivertype.JobStateRetryable,
				rivertype.JobStateScheduled,
			},
		},
	}
}

// RefreshWorker refreshes cached weather conditions.
type RefreshWorker struct {
	river.WorkerDefaults[RefreshArgs]
	refresher Refresher
	logger    *slog.Logger
}

func NewRefreshWorker(refresher Refresher, logger *slog.Logger) *RefreshWorker {
	if logger == nil {
		logger = slog.Default()
	}
	return &RefreshWorker{refresher: refresher, logger: logger}
}

func (w *RefreshWorker) Timeout(*river.Job[RefreshArgs]) time.Duration { return 30 * time.Second }

func (w *RefreshWorker) Work(ctx context.Context, _ *river.Job[RefreshArgs]) error {
	if w.refresher == nil {
		w.logger.DebugContext(ctx, "weather refresh skipped; WeatherKit is not configured")
		return nil
	}
	return w.refresher.Refresh(ctx)
}

// PeriodicJob schedules a weather refresh every ten minutes and on startup.
func PeriodicJob() *river.PeriodicJob {
	return river.NewPeriodicJob(
		river.PeriodicInterval(weather.RefreshInterval),
		func() (river.JobArgs, *river.InsertOpts) { return RefreshArgs{}, nil },
		&river.PeriodicJobOpts{RunOnStart: true},
	)
}
