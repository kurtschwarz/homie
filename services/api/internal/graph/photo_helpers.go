package graph

import (
	"errors"
	"github.com/kurtschwarz/homie/services/api/internal/graph/model"
	"github.com/kurtschwarz/homie/services/api/internal/photos"
)

var errUnsupportedPhotoContentType = errors.New("only image/jpeg uploads are supported")

func toModelPhoto(photo photos.Photo) model.Photo {
	var coordinate *model.Coordinate
	if photo.Coordinate != nil {
		coordinate = &model.Coordinate{Latitude: photo.Coordinate.Latitude, Longitude: photo.Coordinate.Longitude}
	}
	var place *string
	if photo.Place != "" {
		place = &photo.Place
	}
	var focus *model.FocalRegion
	if photo.Focus != nil {
		focus = &model.FocalRegion{X: photo.Focus.X, Y: photo.Focus.Y, Width: photo.Focus.Width, Height: photo.Focus.Height}
	}
	return model.Photo{
		ID: photo.ID.String(), URL: "/photos/" + photo.File + "?v=" + photo.Version,
		Width: photo.Width, Height: photo.Height, TakenAt: int(photo.TakenAt.Unix()),
		Coordinate: coordinate, Place: place, Focus: focus,
	}
}

func toModelShowing(showing photos.Showing) model.PhotoShowing {
	var endedAt *int
	if showing.EndedAt != nil {
		value := int(showing.EndedAt.Unix())
		endedAt = &value
	}
	photo := toModelPhoto(showing.Photo)
	return model.PhotoShowing{
		ID: showing.ID.String(), Photo: &photo,
		StartedAt: int(showing.StartedAt.Unix()), EndsAt: int(showing.EndsAt.Unix()),
		EndedAt: endedAt, Source: model.PhotoShowingSource(showing.Source),
	}
}

func toModelRotation(current, next *photos.Showing) *model.PhotoShowingState {
	result := &model.PhotoShowingState{}
	if current != nil {
		showing := toModelShowing(*current)
		result.Current = &showing
	}
	if next != nil {
		showing := toModelShowing(*next)
		result.Next = &showing
	}
	return result
}
