package graph

import (
	"errors"
	"net/url"
	"strconv"
	"time"

	"github.com/kurtschwarz/homie/services/api/internal/auth"
	"github.com/kurtschwarz/homie/services/api/internal/graph/model"
	"github.com/kurtschwarz/homie/services/api/internal/photos"
)

var errUnsupportedPhotoContentType = errors.New("only image/jpeg uploads are supported")

func toModelPhoto(photo photos.Photo, signer *auth.Signer) model.Photo {
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
	var artist *model.Artist
	if photo.Artist != nil {
		artist = &model.Artist{ID: photo.Artist.ID.String(), Name: photo.Artist.Name}
	}
	expiry := auth.Expiry(time.Now())
	query := url.Values{"v": {photo.Version}, "exp": {strconv.FormatInt(expiry, 10)}}
	if signer != nil {
		query.Set("sig", signer.Sign(photo.File, expiry))
	}
	return model.Photo{
		ID: photo.ID.String(), URL: "/photos/" + photo.File + "?" + query.Encode(),
		Width: photo.Width, Height: photo.Height, TakenAt: int(photo.TakenAt.Unix()),
		Coordinate: coordinate, Place: place, Focus: focus, Artist: artist,
	}
}

func toModelShowing(showing photos.Showing, signer *auth.Signer) model.PhotoShowing {
	var endedAt *int
	if showing.EndedAt != nil {
		value := int(showing.EndedAt.Unix())
		endedAt = &value
	}
	photo := toModelPhoto(showing.Photo, signer)
	return model.PhotoShowing{
		ID: showing.ID.String(), Photo: &photo,
		StartedAt: int(showing.StartedAt.Unix()), EndsAt: int(showing.EndsAt.Unix()),
		EndedAt: endedAt, Source: model.PhotoShowingSource(showing.Source),
	}
}

func toModelRotation(current, next *photos.Showing, signer *auth.Signer) *model.PhotoShowingState {
	result := &model.PhotoShowingState{}
	if current != nil {
		showing := toModelShowing(*current, signer)
		result.Current = &showing
	}
	if next != nil {
		showing := toModelShowing(*next, signer)
		result.Next = &showing
	}
	return result
}
