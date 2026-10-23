// Package jobs contains River job definitions for events operations.
package jobs

import (
	"context"
	"log/slog"
	"time"

	"github.com/kurtschwarz/homie/services/api/internal/events"
	"github.com/riverqueue/river"
	"github.com/riverqueue/river/rivertype"
)

const maxAttempts = 3

// Refresher is implemented by the events service.
type Refresher interface {
	RefreshDue(context.Context) error
}

// scanInterval is how often the worker checks which accounts are due.
const scanInterval = time.Minute

var _ Refresher = (*events.Service)(nil)

// RefreshArgs identifies a background CalDAV refresh.
type RefreshArgs struct{}

func (RefreshArgs) Kind() string { return "events_refresh" }

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

// RefreshWorker refreshes the synced calendar window from CalDAV.
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
		w.logger.DebugContext(ctx, "events refresh skipped; CalDAV is not configured")
		return nil
	}
	return w.refresher.RefreshDue(ctx)
}

// PeriodicJob scans for accounts whose refresh interval has elapsed, every
// minute and on startup.
func PeriodicJob() *river.PeriodicJob {
	return river.NewPeriodicJob(
		river.PeriodicInterval(scanInterval),
		func() (river.JobArgs, *river.InsertOpts) { return RefreshArgs{}, nil },
		&river.PeriodicJobOpts{RunOnStart: true},
	)
}
