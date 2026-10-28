package photos_test

import (
	"context"
	"errors"
	"io"
	"io/fs"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/kurtschwarz/homie/services/api/internal/photos"
)

func TestServiceImportStoresFileAndKeepsStableID(t *testing.T) {
	db := openDB(t)
	repo := photos.NewSQLiteRepository(db)
	library, err := photos.NewLibrary(filepath.Join(t.TempDir(), "photos"))
	if err != nil {
		t.Fatal(err)
	}
	service := photos.New(repo, library)
	ctx := context.Background()

	upload := photos.Upload{
		AssetID: "asset", Version: "1", Width: 1400, Height: 1050,
		TakenAt: time.Unix(1_760_640_000, 0), Coordinate: &photos.Coordinate{Latitude: 43.4, Longitude: -80.4},
		Focus: &photos.FocalRegion{X: 0.2, Y: 0.3, Width: 0.4, Height: 0.5, Source: "vision-face"},
	}
	first, err := service.Import(ctx, upload, strings.NewReader("first-bytes"))
	if err != nil {
		t.Fatalf("import: %v", err)
	}
	contents, err := io.ReadAll(mustOpen(t, library.FS(), first.File))
	if err != nil || string(contents) != "first-bytes" {
		t.Fatalf("stored contents=%q err=%v", contents, err)
	}

	upload.Version = "2"
	second, err := service.Import(ctx, upload, strings.NewReader("second-bytes"))
	if err != nil {
		t.Fatalf("re-import: %v", err)
	}
	if second.ID != first.ID || second.File != first.File {
		t.Fatalf("re-import changed identity: first=%+v second=%+v", first, second)
	}
	if second.Focus == nil || *second.Focus != *upload.Focus {
		t.Fatalf("focus did not round-trip through repository: got %+v", second.Focus)
	}
	contents, err = io.ReadAll(mustOpen(t, library.FS(), second.File))
	if err != nil || string(contents) != "second-bytes" {
		t.Fatalf("overwritten contents=%q err=%v", contents, err)
	}
}

func TestServiceImportValidation(t *testing.T) {
	db := openDB(t)
	library, err := photos.NewLibrary(filepath.Join(t.TempDir(), "photos"))
	if err != nil {
		t.Fatal(err)
	}
	service := photos.New(photos.NewSQLiteRepository(db), library)
	ctx := context.Background()
	valid := photos.Upload{AssetID: "asset", Version: "1", Width: 1, Height: 1, TakenAt: time.Unix(1, 0)}

	cases := []struct {
		name   string
		mutate func(*photos.Upload)
	}{
		{"missing asset id", func(u *photos.Upload) { u.AssetID = "" }},
		{"missing version", func(u *photos.Upload) { u.Version = "" }},
		{"zero width", func(u *photos.Upload) { u.Width = 0 }},
		{"zero height", func(u *photos.Upload) { u.Height = 0 }},
		{"missing taken at", func(u *photos.Upload) { u.TakenAt = time.Time{} }},
		{"negative focus origin", func(u *photos.Upload) {
			u.Focus = &photos.FocalRegion{X: -0.1, Y: 0, Width: 0.2, Height: 0.2, Source: "vision-face"}
		}},
		{"focus outside image", func(u *photos.Upload) {
			u.Focus = &photos.FocalRegion{X: 0.8, Y: 0.8, Width: 0.3, Height: 0.2, Source: "vision-face"}
		}},
		{"empty focus source", func(u *photos.Upload) { u.Focus = &photos.FocalRegion{X: 0.1, Y: 0.1, Width: 0.2, Height: 0.2} }},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			upload := valid
			c.mutate(&upload)
			if _, err := service.Import(ctx, upload, strings.NewReader("bytes")); err == nil {
				t.Fatal("expected validation error")
			}
		})
	}
}

func TestServiceImportRemovesOrphanFileOnUpsertFailure(t *testing.T) {
	db := openDB(t)
	library, err := photos.NewLibrary(filepath.Join(t.TempDir(), "photos"))
	if err != nil {
		t.Fatal(err)
	}
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}
	service := photos.New(photos.NewSQLiteRepository(db), library)
	upload := photos.Upload{AssetID: "asset", Version: "1", Width: 1, Height: 1, TakenAt: time.Unix(1, 0)}

	_, err = service.Import(context.Background(), upload, strings.NewReader("bytes"))
	if err == nil {
		t.Fatal("expected upsert against a closed db to fail")
	}
	entries, err := fs.ReadDir(library.FS(), ".")
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 0 {
		t.Fatalf("orphan file left behind: %v", entries)
	}
}

func TestServiceSaveDrawing(t *testing.T) {
	db := openDB(t)
	library, err := photos.NewLibrary(filepath.Join(t.TempDir(), "photos"))
	if err != nil {
		t.Fatal(err)
	}
	service := photos.New(photos.NewSQLiteRepository(db), library)
	ctx := context.Background()
	drawnAt := time.Unix(1_761_000_000, 0)

	if _, err := service.SaveDrawing(ctx, photos.Drawing{Width: 0, Height: 10, DrawnAt: drawnAt}, strings.NewReader("x")); err == nil {
		t.Fatal("expected size error")
	}
	first, err := service.SaveDrawing(ctx, photos.Drawing{Width: 20, Height: 10, DrawnAt: drawnAt}, strings.NewReader("jpeg"))
	if err != nil {
		t.Fatal(err)
	}
	second, err := service.SaveDrawing(ctx, photos.Drawing{Width: 20, Height: 10, DrawnAt: drawnAt}, strings.NewReader("jpeg"))
	if err != nil {
		t.Fatal(err)
	}
	if first.ID == second.ID {
		t.Fatal("each drawing needs its own identity")
	}
	list, err := service.List(ctx, 10)
	if err != nil || len(list) != 2 {
		t.Fatalf("list: %+v err=%v", list, err)
	}
	for _, p := range list {
		if p.Kind != photos.KindDrawing || p.Focus != nil || p.Coordinate != nil || !p.TakenAt.Equal(drawnAt) {
			t.Fatalf("drawing mismatch: %+v", p)
		}
	}
}

func TestServiceDrawingPushTracking(t *testing.T) {
	db := openDB(t)
	library, err := photos.NewLibrary(filepath.Join(t.TempDir(), "photos"))
	if err != nil {
		t.Fatal(err)
	}
	service := photos.New(photos.NewSQLiteRepository(db), library)
	ctx := context.Background()

	if pending, err := service.UnpushedDrawings(ctx); err != nil || len(pending) != 0 {
		t.Fatalf("empty unpushed: %+v err=%v", pending, err)
	}
	if ids, err := service.PushedAssetIDs(ctx); err != nil || ids == nil || len(ids) != 0 {
		t.Fatalf("empty pushed ids: %#v err=%v", ids, err)
	}

	drawing, err := service.SaveDrawing(ctx, photos.Drawing{Width: 20, Height: 10, DrawnAt: time.Unix(1_761_000_000, 0)}, strings.NewReader("jpeg"))
	if err != nil {
		t.Fatal(err)
	}
	if pending, err := service.UnpushedDrawings(ctx); err != nil || len(pending) != 1 || pending[0].ID != drawing.ID {
		t.Fatalf("unpushed after save: %+v err=%v", pending, err)
	}

	if err := service.MarkDrawingPushed(ctx, drawing.ID, "asset-1"); err != nil {
		t.Fatal(err)
	}
	if err := service.MarkDrawingPushed(ctx, drawing.ID, "asset-2"); err != nil {
		t.Fatalf("repeat mark: %v", err)
	}
	if pending, err := service.UnpushedDrawings(ctx); err != nil || len(pending) != 0 {
		t.Fatalf("unpushed after mark: %+v err=%v", pending, err)
	}
	if ids, err := service.PushedAssetIDs(ctx); err != nil || len(ids) != 1 || ids[0] != "asset-1" {
		t.Fatalf("pushed ids: %v err=%v", ids, err)
	}

	if err := service.MarkDrawingPushed(ctx, uuid.New(), "asset-x"); !errors.Is(err, photos.ErrUnknownDrawing) {
		t.Fatalf("unknown drawing err=%v", err)
	}
	if err := service.MarkDrawingPushed(ctx, drawing.ID, ""); err == nil {
		t.Fatal("expected error for empty asset id")
	}
}
