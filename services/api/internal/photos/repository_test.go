package photos_test

import (
	"context"
	"io"
	"io/fs"
	"log/slog"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/kurtschwarz/homie/services/api/internal/photos"
	"github.com/kurtschwarz/homie/services/api/internal/store"
)

func TestSQLiteRepositoryRoundTripAndStableID(t *testing.T) {
	db := openDB(t)
	repo := photos.NewSQLiteRepository(db)
	ctx := context.Background()
	firstID := uuid.MustParse("0192a780-0a00-7000-8000-000000000001")
	photo := photos.Photo{
		ID: firstID, AssetID: "ASSET/L0/001", Version: "1760640000", File: "first.jpg",
		Width: 1400, Height: 1050, TakenAt: time.Unix(1_760_640_000, 0),
		Coordinate: &photos.Coordinate{Latitude: 43.4422, Longitude: -80.4699}, Place: "Kitchener, Ontario",
	}
	if err := repo.Upsert(ctx, photo); err != nil {
		t.Fatalf("upsert: %v", err)
	}

	var storageClass string
	var idLength int
	if err := db.QueryRowxContext(ctx, `SELECT typeof(id), length(id) FROM photos WHERE asset_id = ?`, photo.AssetID).Scan(&storageClass, &idLength); err != nil {
		t.Fatal(err)
	}
	if storageClass != "blob" || idLength != 16 {
		t.Fatalf("stored id: typeof=%q length=%d, want blob and 16", storageClass, idLength)
	}

	got, err := repo.List(ctx, 10)
	if err != nil || len(got) != 1 {
		t.Fatalf("list: got=%+v err=%v", got, err)
	}
	if got[0].ID != photo.ID || got[0].AssetID != photo.AssetID || got[0].Version != photo.Version || got[0].File != photo.File || got[0].Width != photo.Width || got[0].Height != photo.Height || !got[0].TakenAt.Equal(photo.TakenAt) || got[0].Coordinate == nil || *got[0].Coordinate != *photo.Coordinate || got[0].Place != photo.Place {
		t.Fatalf("round trip mismatch: got %+v, want %+v", got[0], photo)
	}

	manifest, err := repo.Manifest(ctx)
	if err != nil || manifest[photo.AssetID] != (photos.ManifestEntry{ID: photo.ID, Version: photo.Version}) {
		t.Fatalf("manifest: got=%+v err=%v", manifest, err)
	}
	if id, found, err := repo.IDForAsset(ctx, photo.AssetID); err != nil || !found || id != photo.ID {
		t.Fatalf("id for asset: id=%s found=%v err=%v", id, found, err)
	}
	if _, found, err := repo.IDForAsset(ctx, "missing"); err != nil || found {
		t.Fatalf("missing id lookup: found=%v err=%v", found, err)
	}

	photo.ID = uuid.MustParse("0192a780-0a00-7000-8000-000000000002")
	photo.Version = "1760640100"
	photo.File = "updated.jpg"
	photo.Coordinate = nil
	photo.Place = ""
	if err := repo.Upsert(ctx, photo); err != nil {
		t.Fatalf("upsert update: %v", err)
	}
	got, err = repo.List(ctx, 10)
	if err != nil || len(got) != 1 {
		t.Fatalf("list after update: got=%+v err=%v", got, err)
	}
	if got[0].ID != firstID || got[0].Version != photo.Version || got[0].File != photo.File || got[0].Coordinate != nil || got[0].Place != "" {
		t.Fatalf("upsert did not preserve id or clear optional fields: %+v", got[0])
	}
}

func TestSQLiteRepositoryListOrderingAndDeleteMissing(t *testing.T) {
	db := openDB(t)
	repo := photos.NewSQLiteRepository(db)
	ctx := context.Background()
	takenAt := time.Unix(1_760_640_000, 0)
	items := []photos.Photo{
		{ID: uuid.MustParse("0192a780-0a00-7000-8000-000000000001"), AssetID: "a", Version: "1", File: "a.jpg", TakenAt: takenAt},
		{ID: uuid.MustParse("0192a780-0a00-7000-8000-000000000002"), AssetID: "b", Version: "1", File: "b.jpg", TakenAt: takenAt},
		{ID: uuid.MustParse("0192a780-0a00-7000-8000-000000000003"), AssetID: "c", Version: "1", File: "c.jpg", TakenAt: takenAt.Add(-time.Hour)},
	}
	for _, photo := range items {
		if err := repo.Upsert(ctx, photo); err != nil {
			t.Fatal(err)
		}
	}
	got, err := repo.List(ctx, 2)
	if err != nil || len(got) != 2 || got[0].AssetID != "b" || got[1].AssetID != "a" {
		t.Fatalf("ordered limited list: got=%+v err=%v", got, err)
	}

	removed, err := repo.DeleteMissing(ctx, []string{"a", "c"})
	if err != nil || len(removed) != 1 || removed[0] != "b.jpg" {
		t.Fatalf("delete missing: removed=%v err=%v", removed, err)
	}
	manifest, err := repo.Manifest(ctx)
	if err != nil || len(manifest) != 2 {
		t.Fatalf("manifest after prune: count=%d err=%v", len(manifest), err)
	}

	largeKeep := make([]string, 1001)
	for i := range largeKeep {
		largeKeep[i] = "keep-" + time.Unix(int64(i), 0).Format("150405")
	}
	// Including all current ids in a manifest larger than SQLite's bind limit
	// should preserve the two remaining rows.
	largeKeep = append(largeKeep, "a", "c")
	removed, err = repo.DeleteMissing(ctx, largeKeep)
	if err != nil || len(removed) != 0 {
		t.Fatalf("large manifest prune: removed=%v err=%v", removed, err)
	}

	removed, err = repo.DeleteMissing(ctx, nil)
	if err != nil || len(removed) != 2 {
		t.Fatalf("empty manifest prune: removed=%v err=%v", removed, err)
	}
	manifest, err = repo.Manifest(ctx)
	if err != nil || len(manifest) != 0 {
		t.Fatalf("manifest after empty prune: count=%d err=%v", len(manifest), err)
	}
}

func TestSQLiteRepositoryShowingsAndCascadeDelete(t *testing.T) {
	db := openDB(t)
	repo := photos.NewSQLiteRepository(db)
	ctx := context.Background()
	photo := photos.Photo{ID: uuid.MustParse("0192a780-0a00-7000-8000-000000000011"), AssetID: "showing", Version: "1", File: "showing.jpg", Width: 10, Height: 10, TakenAt: time.Unix(100, 0)}
	if err := repo.Upsert(ctx, photo); err != nil {
		t.Fatal(err)
	}
	started := time.Unix(200, 0)
	showing := photos.Showing{ID: uuid.MustParse("0192a780-0a00-7000-8000-000000000012"), Photo: photo, StartedAt: started, EndsAt: started.Add(photos.RotationInterval), Source: photos.ShowingSchedule}
	if err := repo.AppendShowing(ctx, showing); err != nil {
		t.Fatal(err)
	}
	if err := repo.AdvanceShowing(ctx, ptrUUID(showing.ID), photos.Showing{ID: uuid.MustParse("0192a780-0a00-7000-8000-000000000013"), Photo: photo, StartedAt: started.Add(5 * time.Second), EndsAt: started.Add(17 * time.Second), Source: photos.ShowingTap}, started.Add(5*time.Second), photos.ShowingTap); err != nil {
		t.Fatal(err)
	}
	current, next, err := repo.Rotation(ctx)
	if err != nil || current == nil || next != nil {
		t.Fatalf("rotation current=%+v next=%+v err=%v", current, next, err)
	}
	history, err := repo.Showings(ctx, 10)
	if err != nil || len(history) != 2 {
		t.Fatalf("history=%+v err=%v", history, err)
	}
	if history[1].Photo.ID != photo.ID || history[1].EndedAt == nil || !history[1].EndedAt.Equal(started.Add(5*time.Second)) {
		t.Fatalf("showing round trip: %+v", history[1])
	}
	if _, err := db.ExecContext(ctx, `DELETE FROM photos WHERE id = ?`, photo.ID[:]); err != nil {
		t.Fatal(err)
	}
	history, err = repo.Showings(ctx, 10)
	if err != nil || len(history) != 0 {
		t.Fatalf("history after photo cascade=%+v err=%v", history, err)
	}
}

func TestSQLiteRepositoryNextPhotoWraps(t *testing.T) {
	db := openDB(t)
	repo := photos.NewSQLiteRepository(db)
	ctx := context.Background()
	takenAt := time.Unix(100, 0)
	a := photos.Photo{ID: uuid.MustParse("0192a780-0a00-7000-8000-000000000021"), AssetID: "a", Version: "1", File: "a.jpg", Width: 1, Height: 1, TakenAt: takenAt}
	b := photos.Photo{ID: uuid.MustParse("0192a780-0a00-7000-8000-000000000022"), AssetID: "b", Version: "1", File: "b.jpg", Width: 1, Height: 1, TakenAt: takenAt}
	for _, photo := range []photos.Photo{a, b} {
		if err := repo.Upsert(ctx, photo); err != nil {
			t.Fatal(err)
		}
	}
	next, found, err := repo.NextPhoto(ctx, b.ID)
	if err != nil || !found || next.ID != a.ID {
		t.Fatalf("next=%+v found=%v err=%v", next, found, err)
	}
	next, found, err = repo.NextPhoto(ctx, a.ID)
	if err != nil || !found || next.ID != b.ID {
		t.Fatalf("wrap next=%+v found=%v err=%v", next, found, err)
	}
}

func TestLibrarySaveNameAndRemove(t *testing.T) {
	library, err := photos.NewLibrary(filepath.Join(t.TempDir(), "photos"))
	if err != nil {
		t.Fatal(err)
	}
	id := uuid.MustParse("0192a780-0a00-7000-8000-000000000001")
	name := library.Name(id)
	if name != "0192a7800a0070008000000000000001.jpg" || name != library.Name(id) {
		t.Fatalf("name is not stable: %q", name)
	}
	if err := library.Save(name, strings.NewReader("jpeg")); err != nil {
		t.Fatal(err)
	}
	contents, err := io.ReadAll(mustOpen(t, library.FS(), name))
	if err != nil || string(contents) != "jpeg" {
		t.Fatalf("saved contents %q, err=%v", contents, err)
	}
	if _, err := library.FS().Open(name + ".tmp"); err == nil {
		t.Fatal("temporary file survived successful save")
	}
	if err := library.Remove(name); err != nil {
		t.Fatal(err)
	}
	if err := library.Remove(name); err != nil {
		t.Fatalf("remove missing file should be harmless: %v", err)
	}
}

func openDB(t *testing.T) *sqlx.DB {
	t.Helper()
	db, err := store.Open(filepath.Join(t.TempDir(), "photos.db"), slog.New(slog.NewTextHandler(io.Discard, nil)))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return db
}

func ptrUUID(id uuid.UUID) *uuid.UUID { return &id }

func mustOpen(t *testing.T, filesystem interface {
	Open(string) (fs.File, error)
}, name string) fs.File {
	t.Helper()
	f, err := filesystem.Open(name)
	if err != nil {
		t.Fatal(err)
	}
	return f
}

func TestDeleteMissingAndHelpersIgnoreDrawings(t *testing.T) {
	db := openDB(t)
	repo := photos.NewSQLiteRepository(db)
	ctx := context.Background()
	drawing := photos.Photo{ID: uuid.MustParse("0192a780-0a00-7000-8000-000000000051"), AssetID: "drawing:1", Version: "1", File: "drawing.jpg", Width: 10, Height: 10, TakenAt: time.Unix(100, 0), Kind: photos.KindDrawing}
	apple := photos.Photo{ID: uuid.MustParse("0192a780-0a00-7000-8000-000000000052"), AssetID: "apple", Version: "1", File: "apple.jpg", Width: 10, Height: 10, TakenAt: time.Unix(90, 0)}
	for _, p := range []photos.Photo{drawing, apple} {
		if err := repo.Upsert(ctx, p); err != nil {
			t.Fatal(err)
		}
	}
	manifest, err := repo.Manifest(ctx)
	if err != nil || len(manifest) != 1 || manifest["apple"].ID != apple.ID {
		t.Fatalf("manifest must hold only apple photos: %+v err=%v", manifest, err)
	}
	if _, found, err := repo.IDForAsset(ctx, drawing.AssetID); err != nil || found {
		t.Fatalf("drawing must not resolve as an apple asset: found=%v err=%v", found, err)
	}
	for _, keep := range [][]string{nil, {"apple"}, {"other"}} {
		if _, err := repo.DeleteMissing(ctx, keep); err != nil {
			t.Fatal(err)
		}
		list, err := repo.List(ctx, 10)
		if err != nil {
			t.Fatal(err)
		}
		found := false
		for _, p := range list {
			if p.ID == drawing.ID {
				found = p.Kind == photos.KindDrawing
			}
		}
		if !found {
			t.Fatalf("drawing deleted by reconcile keeping %v", keep)
		}
	}
}
