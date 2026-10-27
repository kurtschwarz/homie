// Package jobs contains River job definitions for display operations.
package jobs

import (
	"context"
	"time"

	"github.com/riverqueue/river"
	"github.com/riverqueue/river/rivertype"
)

const scanInterval = time.Minute

// Nudger asks the display controller goroutine to reconcile.
type Nudger interface {
	Nudge()
}

// ReconcileArgs identifies a scheduled display reconcile.
type ReconcileArgs struct{}

func (ReconcileArgs) Kind() string { return "display_reconcile" }

func (ReconcileArgs) InsertOpts() river.InsertOpts {
	return river.InsertOpts{
		MaxAttempts: 1,
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

// ReconcileWorker nudges the display controller once a minute.
type ReconcileWorker struct {
	river.WorkerDefaults[ReconcileArgs]
	nudger Nudger
}

func NewReconcileWorker(nudger Nudger) *ReconcileWorker { return &ReconcileWorker{nudger: nudger} }

func (w *ReconcileWorker) Work(context.Context, *river.Job[ReconcileArgs]) error {
	if w.nudger != nil {
		w.nudger.Nudge()
	}
	return nil
}

// PeriodicJob reconciles the display every minute and on startup.
func PeriodicJob() *river.PeriodicJob {
	return river.NewPeriodicJob(
		river.PeriodicInterval(scanInterval),
		func() (river.JobArgs, *river.InsertOpts) { return ReconcileArgs{}, nil },
		&river.PeriodicJobOpts{RunOnStart: true},
	)
}
