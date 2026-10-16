package queue

import (
	"context"
	"io"
	"log/slog"
	"path/filepath"
	"testing"

	"github.com/kurtschwarz/homie/services/api/internal/weather/jobs"
	"github.com/riverqueue/river"
)

func TestNewMigratesSeparateRiverDatabase(t *testing.T) {
	workers := river.NewWorkers()
	river.AddWorker(workers, jobs.NewRefreshWorker(nil, slog.New(slog.NewTextHandler(io.Discard, nil))))
	q, err := New(context.Background(), Config{
		DBPath:  filepath.Join(t.TempDir(), "river.db"),
		Logger:  slog.New(slog.NewTextHandler(io.Discard, nil)),
		Workers: workers,
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := q.Stop(context.Background()); err != nil {
		t.Fatal(err)
	}
}
