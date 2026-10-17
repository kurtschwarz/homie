package photos

import (
	"context"
	"errors"
	"io"
	"time"

	"github.com/google/uuid"
)

// Service provides photo metadata to callers.
type Service struct {
	repository Repository
	library    *Library
}

// New creates a photo service backed by repository and library.
func New(repository Repository, library *Library) *Service {
	return &Service{repository: repository, library: library}
}

// List returns the newest photos.
func (s *Service) List(ctx context.Context, limit int) ([]Photo, error) {
	return s.repository.List(ctx, limit)
}

// Upload is the metadata accompanying an uploaded photo's bytes.
type Upload struct {
	AssetID    string
	Version    string
	Width      int
	Height     int
	TakenAt    time.Time
	Coordinate *Coordinate
	Place      string
}

var (
	errMissingAssetID = errors.New("assetID is required")
	errMissingVersion = errors.New("version is required")
	errInvalidSize    = errors.New("width and height must be positive")
	errMissingTakenAt = errors.New("takenAt is required")
)

// Import stores an uploaded photo's bytes on disk and persists its metadata,
// keeping the same id across re-uploads of a known asset.
func (s *Service) Import(ctx context.Context, upload Upload, src io.Reader) (Photo, error) {
	if upload.AssetID == "" {
		return Photo{}, errMissingAssetID
	}
	if upload.Version == "" {
		return Photo{}, errMissingVersion
	}
	if upload.Width <= 0 || upload.Height <= 0 {
		return Photo{}, errInvalidSize
	}
	if upload.TakenAt.IsZero() {
		return Photo{}, errMissingTakenAt
	}

	id, existed, err := s.repository.IDForAsset(ctx, upload.AssetID)
	if err != nil {
		return Photo{}, err
	}
	if !existed {
		id, err = uuid.NewV7()
		if err != nil {
			return Photo{}, err
		}
	}

	name := s.library.Name(id)
	if err := s.library.Save(name, src); err != nil {
		return Photo{}, err
	}

	photo := Photo{
		ID: id, AssetID: upload.AssetID, Version: upload.Version, File: name,
		Width: upload.Width, Height: upload.Height, TakenAt: upload.TakenAt,
		Coordinate: upload.Coordinate, Place: upload.Place,
	}
	if err := s.repository.Upsert(ctx, photo); err != nil {
		if !existed {
			_ = s.library.Remove(name)
		}
		return Photo{}, err
	}
	return photo, nil
}
