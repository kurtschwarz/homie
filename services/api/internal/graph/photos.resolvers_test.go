package graph

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"path/filepath"
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
	resolver.Resolver.Photos = photos.New(repo)
	limit := 3
	got, err := resolver.Photos(context.Background(), &limit)
	if err != nil || len(got) != 1 {
		t.Fatalf("photos query: got=%+v err=%v", got, err)
	}
	if got[0].ID != photo.ID.String() || got[0].URL != "/photos/photo.jpg?v=v2" || got[0].Coordinate == nil || got[0].Place != nil {
		t.Fatalf("photo result: %+v", got[0])
	}
}

func TestUploadPhotoResolverIsExplicitlyStubbed(t *testing.T) {
	resolver := &mutationResolver{Resolver: &Resolver{}}
	_, err := resolver.UploadPhoto(context.Background(), model.PhotoUploadInput{}, graphql.Upload{})
	if !errors.Is(err, errPhotoUploadsNotImplemented) {
		t.Fatalf("upload stub error=%v, want %v", err, errPhotoUploadsNotImplemented)
	}
}
