// Package queue configures and manages the River job queue.
package queue

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"

	"github.com/kurtschwarz/homie/services/api/internal/store"
	"github.com/riverqueue/river"
	"github.com/riverqueue/river/riverdriver/riversqlite"
	"github.com/riverqueue/river/rivermigrate"
)

type Config struct {
	DBPath       string
	Logger       *slog.Logger
	Workers      *river.Workers
	PeriodicJobs []*river.PeriodicJob
}

type Queue struct {
	db     *sql.DB
	client *river.Client[*sql.Tx]
}

func New(ctx context.Context, config Config) (*Queue, error) {
	if config.Workers == nil {
		return nil, fmt.Errorf("River workers are required")
	}
	db, err := store.OpenSQLite(config.DBPath)
	if err != nil {
		return nil, fmt.Errorf("open River database: %w", err)
	}
	fail := func(err error) (*Queue, error) {
		_ = db.Close()
		return nil, err
	}

	driver := riversqlite.New(db)
	migrator, err := rivermigrate.New(driver, &rivermigrate.Config{Logger: config.Logger})
	if err != nil {
		return fail(fmt.Errorf("configure River migrations: %w", err))
	}
	if _, err := migrator.Migrate(ctx, rivermigrate.DirectionUp, nil); err != nil {
		return fail(fmt.Errorf("apply River migrations: %w", err))
	}
	client, err := river.NewClient(driver, &river.Config{
		Logger: config.Logger,
		Queues: map[string]river.QueueConfig{
			river.QueueDefault: {MaxWorkers: 1},
		},
		Workers:      config.Workers,
		PeriodicJobs: config.PeriodicJobs,
	})
	if err != nil {
		return fail(fmt.Errorf("create River client: %w", err))
	}
	return &Queue{db: db, client: client}, nil
}

func (q *Queue) Start(ctx context.Context) error {
	if err := q.client.Start(ctx); err != nil {
		return fmt.Errorf("start River client: %w", err)
	}
	return nil
}

func (q *Queue) Stop(ctx context.Context) error {
	stopErr := q.client.Stop(ctx)
	closeErr := q.db.Close()
	if stopErr != nil {
		return fmt.Errorf("stop River client: %w", stopErr)
	}
	if closeErr != nil {
		return fmt.Errorf("close River database: %w", closeErr)
	}
	return nil
}
