package photos_test

import (
	"bytes"
	"context"
	"database/sql"
	"errors"
	"image"
	"image/jpeg"
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
	first, err := service.SaveDrawing(ctx, photos.Drawing{Width: 20, Height: 10, DrawnAt: drawnAt}, bytes.NewReader(tinyJPEG(t)))
	if err != nil {
		t.Fatal(err)
	}
	second, err := service.SaveDrawing(ctx, photos.Drawing{Width: 20, Height: 10, DrawnAt: drawnAt}, bytes.NewReader(tinyJPEG(t)))
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

	drawing, err := service.SaveDrawing(ctx, photos.Drawing{Width: 20, Height: 10, DrawnAt: time.Unix(1_761_000_000, 0)}, bytes.NewReader(tinyJPEG(t)))
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

func TestServiceSaveDrawingHomeLocation(t *testing.T) {
	jpegBytes := tinyJPEG(t)
	newService := func(home photos.HomeFunc) *photos.Service {
		library, err := photos.NewLibrary(filepath.Join(t.TempDir(), "photos"))
		if err != nil {
			t.Fatal(err)
		}
		service := photos.New(photos.NewSQLiteRepository(openDB(t)), library)
		service.Home = home
		return service
	}
	ctx := context.Background()
	drawing := photos.Drawing{Width: 8, Height: 8, DrawnAt: time.Unix(1_761_000_000, 0)}

	withHome := newService(func(context.Context) (photos.Coordinate, string, bool, error) {
		return photos.Coordinate{Latitude: 45.4215, Longitude: -75.6972}, "Ottawa", true, nil
	})
	photo, err := withHome.SaveDrawing(ctx, drawing, bytes.NewReader(jpegBytes))
	if err != nil {
		t.Fatal(err)
	}
	if photo.Coordinate == nil || photo.Coordinate.Latitude != 45.4215 || photo.Place != "Ottawa" {
		t.Fatalf("home not applied: %+v", photo)
	}

	noHome := newService(func(context.Context) (photos.Coordinate, string, bool, error) {
		return photos.Coordinate{}, "", false, nil
	})
	photo, err = noHome.SaveDrawing(ctx, drawing, bytes.NewReader(jpegBytes))
	if err != nil {
		t.Fatalf("saving without a weather location must work: %v", err)
	}
	if photo.Coordinate != nil || photo.Place != "" {
		t.Fatalf("unexpected location: %+v", photo)
	}
}

func tinyJPEG(t *testing.T) []byte {
	t.Helper()
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, image.NewRGBA(image.Rect(0, 0, 8, 8)), nil); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

func TestServiceSaveDrawingArtist(t *testing.T) {
	db := openDB(t)
	library, err := photos.NewLibrary(filepath.Join(t.TempDir(), "photos"))
	if err != nil {
		t.Fatal(err)
	}
	service := photos.New(photos.NewSQLiteRepository(db), library)
	ada, family := uuid.New(), uuid.New()
	for i, row := range []struct {
		id     uuid.UUID
		name   string
		family int
	}{{ada, "Ada", 0}, {family, "Family", 1}} {
		if _, err := db.Exec(`INSERT INTO people (id, name, emoji, swatch, sort_order, is_family) VALUES (?, ?, '', 'slate', ?, ?)`, row.id[:], row.name, i, row.family); err != nil {
			t.Fatal(err)
		}
	}
	service.Artist = func(ctx context.Context, id uuid.UUID) (string, bool, error) {
		var name string
		err := db.GetContext(ctx, &name, `SELECT name FROM people WHERE id = ? AND is_family = 0`, id[:])
		if err == sql.ErrNoRows {
			return "", false, nil
		}
		return name, err == nil, err
	}
	ctx := context.Background()
	drawing := photos.Drawing{Width: 8, Height: 8, DrawnAt: time.Unix(1_761_000_000, 0)}

	drawing.ArtistID = &ada
	photo, err := service.SaveDrawing(ctx, drawing, bytes.NewReader(tinyJPEG(t)))
	if err != nil {
		t.Fatal(err)
	}
	if photo.Artist == nil || photo.Artist.Name != "Ada" {
		t.Fatalf("artist = %+v", photo.Artist)
	}
	saved, err := fs.ReadFile(library.FS(), photo.File)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(saved, []byte("Ada\x00")) {
		t.Fatal("EXIF artist missing from saved file")
	}
	listed, err := service.List(ctx, 10)
	if err != nil || len(listed) != 1 || listed[0].Artist == nil || listed[0].Artist.ID != ada {
		t.Fatalf("read back = %+v err=%v", listed, err)
	}

	drawing.ArtistID = &family
	if _, err := service.SaveDrawing(ctx, drawing, bytes.NewReader(tinyJPEG(t))); !errors.Is(err, photos.ErrUnknownArtist) {
		t.Fatalf("family err = %v", err)
	}

	drawing.ArtistID = nil
	anon, err := service.SaveDrawing(ctx, drawing, bytes.NewReader(tinyJPEG(t)))
	if err != nil || anon.Artist != nil {
		t.Fatalf("anonymous = %+v err=%v", anon, err)
	}
	saved, _ = fs.ReadFile(library.FS(), anon.File)
	if bytes.Contains(saved, []byte("Ada")) {
		t.Fatal("unexpected artist in EXIF")
	}

	if _, err := db.Exec(`DELETE FROM people WHERE id = ?`, ada[:]); err != nil {
		t.Fatal(err)
	}
	listed, err = service.List(ctx, 10)
	if err != nil || len(listed) != 2 {
		t.Fatalf("drawings must survive deleting the artist: %+v err=%v", listed, err)
	}
	for _, p := range listed {
		if p.Artist != nil {
			t.Fatalf("artist must be cleared: %+v", p.Artist)
		}
	}
}
