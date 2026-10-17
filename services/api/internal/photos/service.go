package photos

import "context"

// Service provides photo metadata to callers.
type Service struct {
	repository Repository
}

// New creates a photo service backed by repository.
func New(repository Repository) *Service {
	return &Service{repository: repository}
}

// List returns the newest photos.
func (s *Service) List(ctx context.Context, limit int) ([]Photo, error) {
	return s.repository.List(ctx, limit)
}
