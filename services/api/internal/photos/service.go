package photos

import (
	"context"
	"errors"
	"io"
	"math"
	"strconv"
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
	Focus      *FocalRegion
}

var (
	errMissingAssetID = errors.New("assetID is required")
	errMissingVersion = errors.New("version is required")
	errInvalidSize    = errors.New("width and height must be positive")
	errMissingTakenAt = errors.New("takenAt is required")
	errInvalidFocus   = errors.New("focus must be a non-empty normalized region within the image")
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
	if upload.Focus != nil {
		f := upload.Focus
		if !finite(f.X) || !finite(f.Y) || !finite(f.Width) || !finite(f.Height) || f.X < 0 || f.Y < 0 || f.Width <= 0 || f.Height <= 0 || f.X+f.Width > 1 || f.Y+f.Height > 1 || f.Source == "" {
			return Photo{}, errInvalidFocus
		}
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
		Coordinate: upload.Coordinate, Place: upload.Place, Focus: upload.Focus,
	}
	if err := s.repository.Upsert(ctx, photo); err != nil {
		if !existed {
			_ = s.library.Remove(name)
		}
		return Photo{}, err
	}
	return photo, nil
}

// Drawing is the metadata accompanying a saved drawing's bytes.
type Drawing struct {
	Width   int
	Height  int
	DrawnAt time.Time
}

// SaveDrawing stores a drawing as a new photo of kind drawing, with no focal
// region or coordinate.
func (s *Service) SaveDrawing(ctx context.Context, drawing Drawing, src io.Reader) (Photo, error) {
	if drawing.Width <= 0 || drawing.Height <= 0 {
		return Photo{}, errInvalidSize
	}

	if drawing.DrawnAt.IsZero() {
		return Photo{}, errMissingTakenAt
	}

	id, err := uuid.NewV7()
	if err != nil {
		return Photo{}, err
	}

	name := s.library.Name(id)
	if err := s.library.Save(name, src); err != nil {
		return Photo{}, err
	}

	photo := Photo{
		ID: id, AssetID: "drawing:" + id.String(), Version: strconv.FormatInt(drawing.DrawnAt.Unix(), 10), File: name,
		Width: drawing.Width, Height: drawing.Height, TakenAt: drawing.DrawnAt, Kind: KindDrawing,
	}

	if err := s.repository.Upsert(ctx, photo); err != nil {
		_ = s.library.Remove(name)
		return Photo{}, err
	}

	return photo, nil
}

// ErrUnknownDrawing is returned when a drawing id does not match a saved drawing.
var ErrUnknownDrawing = errors.New("unknown drawing")

// UnpushedDrawings returns drawings that have not yet been pushed to Photos.
func (s *Service) UnpushedDrawings(ctx context.Context) ([]Photo, error) {
	return s.repository.UnpushedDrawings(ctx)
}

// MarkDrawingPushed records the Photos asset created for a drawing. Repeating
// the call is safe and keeps the first recorded asset.
func (s *Service) MarkDrawingPushed(ctx context.Context, id uuid.UUID, assetID string) error {
	if assetID == "" {
		return errMissingAssetID
	}
	found, err := s.repository.MarkDrawingPushed(ctx, id, assetID, time.Now())
	if err != nil {
		return err
	}
	if !found {
		return ErrUnknownDrawing
	}
	return nil
}

// PushedAssetIDs returns the Photos asset identifiers known to be drawings.
func (s *Service) PushedAssetIDs(ctx context.Context) ([]string, error) {
	return s.repository.PushedAssetIDs(ctx)
}

func finite(value float64) bool { return !math.IsNaN(value) && !math.IsInf(value, 0) }
