package photos_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/kurtschwarz/homie/services/api/internal/photos"
)

func TestRotationServiceSchedulesAndAdvances(t *testing.T) {
	db := openDB(t)
	repo := photos.NewSQLiteRepository(db)
	ctx := context.Background()
	first := photos.Photo{ID: uuid.MustParse("0192a780-0a00-7000-8000-000000000031"), AssetID: "first", Version: "1", File: "first.jpg", Width: 1, Height: 1, TakenAt: time.Unix(100, 0)}
	second := photos.Photo{ID: uuid.MustParse("0192a780-0a00-7000-8000-000000000032"), AssetID: "second", Version: "1", File: "second.jpg", Width: 1, Height: 1, TakenAt: time.Unix(200, 0)}
	for _, photo := range []photos.Photo{first, second} {
		if err := repo.Upsert(ctx, photo); err != nil {
			t.Fatal(err)
		}
	}
	now := time.Unix(1_800_000_000, 0)
	service := photos.NewRotationServiceWithClock(repo, nil, func() time.Time { return now })
	current, next, err := service.Advance(ctx, photos.ShowingSchedule)
	if err != nil || current == nil || current.Photo.ID != second.ID || next != nil {
		t.Fatalf("initial current=%+v next=%+v err=%v", current, next, err)
	}
	current, next, err = service.Rotation(ctx)
	if err != nil || current == nil || current.Photo.ID != second.ID || next == nil || next.Photo.ID != first.ID {
		t.Fatalf("rotation should pre-schedule next: current=%+v next=%+v err=%v", current, next, err)
	}
	current, next, err = service.Advance(ctx, photos.ShowingSchedule)
	if err != nil || current == nil || current.Photo.ID != second.ID || next == nil || next.Photo.ID != first.ID {
		t.Fatalf("pre-schedule current=%+v next=%+v err=%v", current, next, err)
	}
	if got, _, err := service.Advance(ctx, photos.ShowingSchedule); err != nil || got == nil || got.Photo.ID != second.ID {
		t.Fatalf("early scheduled advance should be a no-op: current=%+v err=%v", got, err)
	}
	now = current.EndsAt
	current, next, err = service.Advance(ctx, photos.ShowingSchedule)
	if err != nil || current == nil || current.Photo.ID != first.ID || next == nil || next.Photo.ID != second.ID {
		t.Fatalf("scheduled transition current=%+v next=%+v err=%v", current, next, err)
	}
	now = now.Add(time.Second)
	current, next, err = service.Advance(ctx, photos.ShowingTap)
	if err != nil || current == nil || current.Photo.ID != second.ID || next == nil || next.Photo.ID != first.ID {
		t.Fatalf("tap transition current=%+v next=%+v err=%v", current, next, err)
	}
	history, err := service.Showings(ctx, 10)
	if err != nil || len(history) != 4 {
		t.Fatalf("history count=%d err=%v", len(history), err)
	}
}

func TestRotationServiceEmptyLibrary(t *testing.T) {
	service := photos.NewRotationService(photos.NewSQLiteRepository(openDB(t)), nil)
	if _, _, err := service.Advance(context.Background(), photos.ShowingSchedule); err != photos.ErrNoPhotos {
		t.Fatalf("empty rotation err=%v", err)
	}
}
