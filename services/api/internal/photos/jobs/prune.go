package jobs

import (
	"context"
	"log/slog"
	"time"

	"github.com/kurtschwarz/homie/services/api/internal/photos"
	"github.com/riverqueue/river"
)

const pruneInterval = 24 * time.Hour

type Pruner interface{ Prune(context.Context) error }

var _ Pruner = (*photos.RotationService)(nil)

type PruneArgs struct{}

func (PruneArgs) Kind() string { return "photo_showings_prune" }

type PruneWorker struct {
	river.WorkerDefaults[PruneArgs]
	pruner Pruner
	logger *slog.Logger
}

func NewPruneWorker(pruner Pruner, logger *slog.Logger) *PruneWorker {
	if logger == nil {
		logger = slog.Default()
	}
	return &PruneWorker{pruner: pruner, logger: logger}
}

func (w *PruneWorker) Timeout(*river.Job[PruneArgs]) time.Duration { return 30 * time.Second }

func (w *PruneWorker) Work(ctx context.Context, _ *river.Job[PruneArgs]) error {
	if w.pruner == nil {
		w.logger.DebugContext(ctx, "photo showing prune skipped; photos are not configured")
		return nil
	}
	return w.pruner.Prune(ctx)
}

func PeriodicJob() *river.PeriodicJob {
	return river.NewPeriodicJob(river.PeriodicInterval(pruneInterval), func() (river.JobArgs, *river.InsertOpts) {
		return PruneArgs{}, nil
	}, &river.PeriodicJobOpts{RunOnStart: true})
}
