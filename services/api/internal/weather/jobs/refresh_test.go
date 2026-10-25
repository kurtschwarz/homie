package jobs

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/riverqueue/river"
)

type fakeRefresher struct{ err error }

func (f fakeRefresher) RefreshDue(context.Context) error { return f.err }

func TestRefreshWorkerDelegatesAndPropagatesError(t *testing.T) {
	wantErr := errors.New("refresh failed")
	worker := NewRefreshWorker(fakeRefresher{err: wantErr}, slog.New(slog.NewTextHandler(io.Discard, nil)))
	if err := worker.Work(context.Background(), &river.Job[RefreshArgs]{}); !errors.Is(err, wantErr) {
		t.Fatalf("Work error = %v, want %v", err, wantErr)
	}
	if got := worker.Timeout(nil); got != 30*time.Second {
		t.Fatalf("worker timeout = %v, want 30s", got)
	}
}

func TestRefreshWorkerNoOpsWhenUnconfigured(t *testing.T) {
	worker := NewRefreshWorker(nil, slog.New(slog.NewTextHandler(io.Discard, nil)))
	if err := worker.Work(context.Background(), &river.Job[RefreshArgs]{}); err != nil {
		t.Fatalf("Work error = %v, want nil", err)
	}
}

func TestRefreshArgsAndPeriodicJobUseScanInterval(t *testing.T) {
	opts := (RefreshArgs{}).InsertOpts()
	if opts.MaxAttempts != maxAttempts || opts.UniqueOpts.ByPeriod != scanInterval {
		t.Fatalf("unexpected refresh options: %+v", opts)
	}
	if PeriodicJob() == nil {
		t.Fatal("PeriodicJob returned nil")
	}
}
