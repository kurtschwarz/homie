package graph

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/99designs/gqlgen/graphql"
	"github.com/google/uuid"
	"github.com/kurtschwarz/homie/services/api/internal/graph/model"
	"github.com/kurtschwarz/homie/services/api/internal/photos"
	"github.com/kurtschwarz/homie/services/api/internal/store"
)

func TestPhotosResolver(t *testing.T) {
	resolver := &queryResolver{Resolver: &Resolver{}}
	if _, err := resolver.Photos(context.Background(), nil); err == nil || err.Error() != "photos are not configured" {
		t.Fatalf("unconfigured resolver err=%v", err)
	}

	db, err := store.Open(filepath.Join(t.TempDir(), "photos.db"), slog.New(slog.NewTextHandler(io.Discard, nil)))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	repo := photos.NewSQLiteRepository(db)
	photo := photos.Photo{
		ID: uuid.MustParse("0192a780-0a00-7000-8000-000000000001"), AssetID: "asset", Version: "v2", File: "photo.jpg",
		Width: 1400, Height: 1050, TakenAt: time.Unix(1_760_640_000, 0),
		Coordinate: &photos.Coordinate{Latitude: 43.4, Longitude: -80.4},
	}
	if err := repo.Upsert(context.Background(), photo); err != nil {
		t.Fatal(err)
	}
	library, err := photos.NewLibrary(filepath.Join(t.TempDir(), "photos"))
	if err != nil {
		t.Fatal(err)
	}
	resolver.Resolver.Photos = photos.New(repo, library)
	limit := 3
	got, err := resolver.Photos(context.Background(), &limit)
	if err != nil || len(got) != 1 {
		t.Fatalf("photos query: got=%+v err=%v", got, err)
	}
	if got[0].ID != photo.ID.String() || got[0].URL != "/photos/photo.jpg?v=v2" || got[0].Coordinate == nil || got[0].Place != nil {
		t.Fatalf("photo result: %+v", got[0])
	}
}

func TestUploadPhotoResolver(t *testing.T) {
	resolver := &mutationResolver{Resolver: &Resolver{}}
	if _, err := resolver.UploadPhoto(context.Background(), model.PhotoUploadInput{}, graphql.Upload{ContentType: "image/jpeg"}); err == nil || err.Error() != "photos are not configured" {
		t.Fatalf("unconfigured resolver err=%v", err)
	}

	db, err := store.Open(filepath.Join(t.TempDir(), "photos.db"), slog.New(slog.NewTextHandler(io.Discard, nil)))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	library, err := photos.NewLibrary(filepath.Join(t.TempDir(), "photos"))
	if err != nil {
		t.Fatal(err)
	}
	resolver.Resolver.Photos = photos.New(photos.NewSQLiteRepository(db), library)

	latitude, longitude := 43.4, -80.4
	place := "Kitchener, Ontario"
	input := model.PhotoUploadInput{
		AssetID: "asset", Version: "1", TakenAt: 1_760_640_000, Width: 1400, Height: 1050,
		Latitude: &latitude, Longitude: &longitude, Place: &place,
	}
	got, err := resolver.UploadPhoto(context.Background(), input, graphql.Upload{ContentType: "image/jpeg", File: strings.NewReader("jpeg-bytes")})
	if err != nil {
		t.Fatalf("upload: %v", err)
	}
	if got.Width != 1400 || got.Coordinate == nil || got.Place == nil || *got.Place != place {
		t.Fatalf("uploaded photo: %+v", got)
	}
	trimmed := strings.TrimPrefix(got.URL, "/photos/")
	file, err := library.FS().Open(trimmed[:strings.IndexByte(trimmed, '?')])
	if err != nil {
		t.Fatal(err)
	}
	contents, err := io.ReadAll(file)
	if err != nil || string(contents) != "jpeg-bytes" {
		t.Fatalf("stored contents=%q err=%v", contents, err)
	}

	if _, err := resolver.UploadPhoto(context.Background(), input, graphql.Upload{ContentType: "image/png"}); !errors.Is(err, errUnsupportedPhotoContentType) {
		t.Fatalf("content type error=%v, want %v", err, errUnsupportedPhotoContentType)
	}
}
