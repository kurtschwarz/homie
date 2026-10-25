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

// scanInterval is how often the worker checks whether the active provider is due.
const scanInterval = time.Minute

// Refresher is implemented by the weather service.
type Refresher interface {
	RefreshDue(context.Context) error
}

var _ Refresher = (*weather.Service)(nil)

// RefreshArgs identifies a background weather refresh.
type RefreshArgs struct{}

func (RefreshArgs) Kind() string { return "weather_refresh" }

func (RefreshArgs) InsertOpts() river.InsertOpts {
	return river.InsertOpts{
		MaxAttempts: maxAttempts,
		UniqueOpts: river.UniqueOpts{
			ByPeriod: scanInterval,
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
		w.logger.DebugContext(ctx, "weather refresh skipped; weather is not configured")
		return nil
	}
	return w.refresher.RefreshDue(ctx)
}

// PeriodicJob scans for a due weather refresh every minute and on startup.
func PeriodicJob() *river.PeriodicJob {
	return river.NewPeriodicJob(
		river.PeriodicInterval(scanInterval),
		func() (river.JobArgs, *river.InsertOpts) { return RefreshArgs{}, nil },
		&river.PeriodicJobOpts{RunOnStart: true},
	)
}
