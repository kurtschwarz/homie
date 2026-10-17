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
