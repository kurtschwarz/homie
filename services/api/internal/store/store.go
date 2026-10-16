package store

import (
	"embed"
	"fmt"
	"log/slog"
	"net/url"
	"os"
	"path/filepath"

	"github.com/jmoiron/sqlx"
	"github.com/pressly/goose/v3"
	_ "modernc.org/sqlite"
)

//go:embed migrations/*.sql
var migrations embed.FS

func init() {
	sqlx.BindDriver("sqlite", sqlx.QUESTION)
}

// Open opens the SQLite database, configures it, and applies embedded migrations.
func Open(path string, logger *slog.Logger) (*sqlx.DB, error) {
	if path == "" {
		return nil, fmt.Errorf("SQLite database path is required")
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
		return nil, fmt.Errorf("create SQLite database directory: %w", err)
	}
	dsn := "file:" + url.PathEscape(path) + "?_pragma=journal_mode(WAL)&_pragma=busy_timeout(5000)&_pragma=synchronous(NORMAL)&_pragma=foreign_keys(1)"
	db, err := sqlx.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("open SQLite database: %w", err)
	}
	// SQLite allows concurrent readers but only one writer. A single connection
	// also keeps PRAGMA user_version migrations serialized.
	db.SetMaxOpenConns(1)
	if logger == nil {
		logger = slog.Default()
	}
	goose.SetBaseFS(migrations)
	if err := goose.SetDialect("sqlite3"); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("configure SQLite migrations: %w", err)
	}
	goose.SetLogger(gooseSlogLogger{logger: logger})
	if err := goose.Up(db.DB, "migrations"); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("apply SQLite migrations: %w", err)
	}
	return db, nil
}

type gooseSlogLogger struct {
	logger *slog.Logger
}

func (l gooseSlogLogger) Printf(format string, args ...any) {
	l.logger.Debug(fmt.Sprintf(format, args...))
}

func (l gooseSlogLogger) Fatalf(format string, args ...any) {
	l.logger.Error(fmt.Sprintf(format, args...))
}
