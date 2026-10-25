package graph

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/kurtschwarz/homie/services/api/internal/graph/model"
	"github.com/kurtschwarz/homie/services/api/internal/weather"
)

// These live here rather than beside the resolvers because gqlgen comments
// out anything it did not generate when it rewrites the resolver file.
var errWeatherNotConfigured = errors.New("weather is not configured")

func weatherKitFromInput(input model.WeatherKitConnectionInput) weather.WeatherKitConnection {
	return weather.WeatherKitConnection{TeamID: input.TeamID, ServiceID: input.ServiceID, KeyID: input.KeyID, PrivateKey: input.PrivateKey}
}

func weatherProviderFromCreateInput(input model.CreateWeatherProviderInput) (weather.Provider, error) {
	if input.Connection == nil || input.Connection.Weatherkit == nil {
		return weather.Provider{}, errConnectionProvider
	}
	connection := weatherKitFromInput(*input.Connection.Weatherkit)
	return weather.Provider{Kind: weather.ProviderWeatherKit, WeatherKit: &connection, RefreshInterval: seconds(input.RefreshIntervalSeconds)}, nil
}

func weatherProviderPatchFromInput(input model.UpdateWeatherProviderInput) (weather.ProviderPatch, error) {
	id, err := uuid.Parse(input.ID)
	if err != nil {
		return weather.ProviderPatch{}, err
	}
	patch := weather.ProviderPatch{ID: id, RefreshInterval: secondsPtr(input.RefreshIntervalSeconds)}
	if input.Connection != nil {
		if input.Connection.Weatherkit == nil {
			return weather.ProviderPatch{}, errConnectionProvider
		}
		changed := input.Connection.Weatherkit
		patch.WeatherKit = &weather.WeatherKitPatch{TeamID: changed.TeamID, ServiceID: changed.ServiceID, KeyID: changed.KeyID, PrivateKey: changed.PrivateKey}
	}
	return patch, nil
}

func weatherProviderToModel(provider weather.Provider) model.WeatherProvider {
	result := model.WeatherProvider{
		ID: provider.ID.String(), Kind: model.WeatherProviderKindWeatherkit, Active: provider.Active,
		RefreshIntervalSeconds: int(provider.RefreshInterval / time.Second), LastError: provider.LastError,
	}
	if provider.WeatherKit != nil {
		result.Connection = model.WeatherKitConnection{
			TeamID: provider.WeatherKit.TeamID, ServiceID: provider.WeatherKit.ServiceID, KeyID: provider.WeatherKit.KeyID,
			HasPrivateKey: true,
		}
	}
	if provider.LastAttemptedAt != nil {
		value := int(provider.LastAttemptedAt.Unix())
		result.LastAttemptedAt = &value
	}
	return result
}

func weatherLocationToModel(location weather.Location) model.WeatherLocation {
	return model.WeatherLocation{ID: location.ID.String(), Label: location.Label, Latitude: location.Latitude, Longitude: location.Longitude, Order: location.Order}
}
