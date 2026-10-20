package photos

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)

const (
	// RotationInterval is how long a photo is scheduled to remain on screen.
	RotationInterval = 12 * time.Second
	// ShowingRetention is how long individual photo-showing records are kept.
	ShowingRetention = 90 * 24 * time.Hour
)

var ErrNoPhotos = errors.New("no photos available for rotation")

// Selector chooses a photo for a new display slot. It can be replaced by a
// future service that ranks photos using history or other signals.
type Selector interface {
	Next(context.Context, uuid.UUID) (Photo, bool, error)
}

type sequentialSelector struct{ repository Repository }

func (s sequentialSelector) Next(ctx context.Context, last uuid.UUID) (Photo, bool, error) {
	return s.repository.NextPhoto(ctx, last)
}

// RotationService maintains the persistent display timeline.
type RotationService struct {
	repository Repository
	selector   Selector
	now        func() time.Time
}

// NewRotationServiceWithClock creates a rotation service with an injected clock.
func NewRotationServiceWithClock(repository Repository, selector Selector, now func() time.Time) *RotationService {
	service := NewRotationService(repository, selector)
	if now != nil {
		service.now = now
	}
	return service
}

// NewRotationService creates a persistent rotation service.
func NewRotationService(repository Repository, selector Selector) *RotationService {
	if selector == nil {
		selector = sequentialSelector{repository: repository}
	}
	return &RotationService{repository: repository, selector: selector, now: time.Now}
}

// Rotation reads current and pre-scheduled next showings without changing state.
func (s *RotationService) Rotation(ctx context.Context) (*Showing, *Showing, error) {
	current, next, err := s.repository.Rotation(ctx)
	if err != nil || current == nil || next != nil {
		return current, next, err
	}
	photo, found, err := s.selector.Next(ctx, current.Photo.ID)
	if err != nil || !found {
		return current, nil, err
	}
	queued, err := s.newShowing(photo, current.EndsAt, ShowingSchedule)
	if err != nil {
		return nil, nil, err
	}
	if err := s.repository.AppendShowing(ctx, *queued); err != nil {
		return nil, nil, err
	}
	return current, queued, nil
}

// Showings returns recent display history, newest first.
func (s *RotationService) Showings(ctx context.Context, limit int) ([]Showing, error) {
	return s.repository.Showings(ctx, limit)
}

// Advance closes the current slot (if any), promotes the scheduled slot, and
// appends a fresh next slot. Calls are serialized by SQLite's transaction lock.
func (s *RotationService) Advance(ctx context.Context, source ShowingSource) (*Showing, *Showing, error) {
	if source != ShowingSchedule && source != ShowingTap {
		return nil, nil, errors.New("invalid photo showing source")
	}
	now := s.now()
	current, next, err := s.repository.Rotation(ctx)
	if err != nil {
		return nil, nil, err
	}
	if current == nil {
		photo, found, err := s.selector.Next(ctx, uuid.Nil)
		if err != nil {
			return nil, nil, err
		}
		if !found {
			return nil, nil, ErrNoPhotos
		}
		showing, err := s.newShowing(photo, now, source)
		if err != nil {
			return nil, nil, err
		}
		if err := s.repository.AdvanceShowing(ctx, nil, *showing, now, source); err != nil {
			return nil, nil, err
		}
		return showing, nil, nil
	}
	var expectedCurrent *uuid.UUID
	if source == ShowingSchedule && now.Before(current.EndsAt) {
		return current, next, nil
	}
	if next == nil {
		photo, found, err := s.selector.Next(ctx, current.Photo.ID)
		if err != nil {
			return nil, nil, err
		}
		if !found {
			return nil, nil, ErrNoPhotos
		}
		newShowing, err := s.newShowing(photo, now, source)
		if err != nil {
			return nil, nil, err
		}
		id := current.ID
		if err := s.repository.AdvanceShowing(ctx, &id, *newShowing, now, source); err != nil {
			return nil, nil, err
		}
		return newShowing, nil, nil
	}
	id := current.ID
	expectedCurrent = &id
	if next != nil {
		current = next
		next = nil
	}

	lastID := uuid.Nil
	if expectedCurrent != nil {
		lastID = current.Photo.ID
	}
	photo, found, err := s.selector.Next(ctx, lastID)
	if err != nil {
		return nil, nil, err
	}
	if !found {
		return nil, nil, ErrNoPhotos
	}
	newShowing, err := s.newShowing(photo, now, source)
	if err != nil {
		return nil, nil, err
	}
	next = newShowing
	if err := s.repository.AdvanceShowing(ctx, expectedCurrent, *newShowing, now, source); err != nil {
		return nil, nil, err
	}
	return current, next, nil
}

func (s *RotationService) newShowing(photo Photo, startedAt time.Time, source ShowingSource) (*Showing, error) {
	id, err := uuid.NewV7()
	if err != nil {
		return nil, err
	}
	return &Showing{ID: id, Photo: photo, StartedAt: startedAt, EndsAt: startedAt.Add(RotationInterval), Source: source}, nil
}

// Prune removes showing history outside the retention window.
func (s *RotationService) Prune(ctx context.Context) error {
	return s.repository.PruneShowings(ctx, s.now().Add(-ShowingRetention))
}
