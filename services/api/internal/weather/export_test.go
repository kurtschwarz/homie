package weather

import "github.com/google/uuid"

// NewStoredProviderForTest builds a stored WeatherKit provider for repository tests.
func NewStoredProviderForTest(id uuid.UUID, credentials []byte) storedProvider {
	return storedProvider{
		Provider:    Provider{ID: id, Kind: ProviderWeatherKit, WeatherKit: &WeatherKitConnection{TeamID: "T", ServiceID: "S", KeyID: "K"}, RefreshInterval: DefaultRefresh},
		Credentials: credentials,
	}
}
