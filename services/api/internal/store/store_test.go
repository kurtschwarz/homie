package store

import (
	"context"
	"path/filepath"
	"testing"
)

func TestOpenAbsolutePath(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nested", "homie.db")
	db, err := Open(path, nil)
	if err != nil {
		t.Fatal(err)
	}
	var table string
	if err := db.Get(&table, `SELECT name FROM sqlite_master WHERE type='table' AND name='weather_current'`); err != nil {
		t.Fatalf("expected goose migration to create weather_current: %v", err)
	}
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}

	db, err = Open(path, nil)
	if err != nil {
		t.Fatalf("reopen migrated database: %v", err)
	}
	defer db.Close()
	var count int
	if err := db.Get(&count, `SELECT COUNT(*) FROM goose_db_version`); err != nil {
		t.Fatal(err)
	}
	if count == 0 {
		t.Fatal("expected goose migration history to be recorded")
	}
}

func TestOpenSQLiteDoesNotApplyApplicationMigrations(t *testing.T) {
	path := filepath.Join(t.TempDir(), "river.db")
	db, err := OpenSQLite(path)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	var count int
	if err := db.QueryRowContext(context.Background(), `SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='weather_current'`).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Fatal("OpenSQLite unexpectedly applied application migrations")
	}
}
