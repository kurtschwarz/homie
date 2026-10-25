// Package weather manages weather providers, locations, and cached conditions.
package weather

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"math"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/kurtschwarz/homie/services/api/internal/secrets"
	"github.com/kurtschwarz/homie/services/api/internal/weather/weatherkit"
)

// Condition is the compact set of weather glyphs supported by the UI.
type Condition string

const (
	ConditionSun      Condition = "SUN"
	ConditionCloudSun Condition = "CLOUD_SUN"
	ConditionCloud    Condition = "CLOUD"
	ConditionRain     Condition = "RAIN"
	ConditionMoon     Condition = "MOON"

	ProviderWeatherKit = "weatherkit"
	MinimumRefresh     = 15 * time.Minute
	DefaultRefresh     = 30 * time.Minute
)

var (
	ErrUnavailable       = errors.New("weather cache is unavailable")
	ErrNotFound          = errors.New("weather record not found")
	ErrInvalidProvider   = errors.New("invalid weather provider")
	ErrInvalidLocation   = errors.New("invalid weather location")
	ErrDuplicateProvider = errors.New("a provider of this kind already exists")
	ErrNoProvider        = errors.New("no active weather provider")
)

// WeatherKitConnection identifies an Apple Developer WeatherKit key.
// PrivateKey is write-only: reads never populate it.
type WeatherKitConnection struct {
	TeamID, ServiceID, KeyID, PrivateKey string
}

// Provider is one weather source and the credentials to call it. The
// provider-specific part lives in the field named for the kind.
type Provider struct {
	ID              uuid.UUID
	Kind            string
	Active          bool
	WeatherKit      *WeatherKitConnection
	RefreshInterval time.Duration
	LastAttemptedAt *time.Time
	LastError       string
}

type storedProvider struct {
	Provider
	Credentials []byte
}

// Location is a place to fetch weather for.
type Location struct {
	ID                  uuid.UUID
	Label               string
	Latitude, Longitude float64
	Order               int
}

// WeatherKitPatch carries the WeatherKit fields an update may change; nil leaves a field as is.
type WeatherKitPatch struct {
	TeamID, ServiceID, KeyID, PrivateKey *string
}

// ProviderPatch carries the fields an update may change; nil leaves a field as is.
type ProviderPatch struct {
	ID              uuid.UUID
	WeatherKit      *WeatherKitPatch
	RefreshInterval *time.Duration
}

// LocationPatch carries the fields an update may change; nil leaves a field as is.
type LocationPatch struct {
	ID                  uuid.UUID
	Label               *string
	Latitude, Longitude *float64
	Order               *int
}

// Current is the normalized weather data needed by the home-screen glance.
type Current struct {
	Condition     Condition
	ConditionCode string
	Temperature   int
	FetchedAt     time.Time
	Stale         bool
}

// Fetcher is the portion of a provider client used by Service.
type Fetcher interface {
	CurrentWeather(context.Context, float64, float64) (weatherkit.CurrentConditions, error)
}

// CacheEntry contains upstream weather facts and their fetch time.
type CacheEntry struct {
	ConditionCode string
	Temperature   float64
	Daylight      bool
	FetchedAt     time.Time
}

// Config configures a weather service.
type Config struct {
	Repository Repository
	Box        *secrets.Box
	Logger     *slog.Logger
	Now        func() time.Time
	// NewFetcher overrides how provider clients are built, for tests.
	NewFetcher func(WeatherKitConnection) (Fetcher, error)
}

// Service manages providers and locations and serves cached weather.
type Service struct {
	repo       Repository
	box        *secrets.Box
	logger     *slog.Logger
	now        func() time.Time
	newFetcher func(WeatherKitConnection) (Fetcher, error)

	mu      sync.Mutex
	clients map[uuid.UUID]cachedFetcher
	current map[uuid.UUID]CacheEntry
}

type cachedFetcher struct {
	fingerprint string
	fetcher     Fetcher
}

// New creates a weather service.
func New(config Config) (*Service, error) {
	if config.Repository == nil {
		return nil, errors.New("weather repository is required")
	}
	if config.Logger == nil {
		config.Logger = slog.Default()
	}
	if config.Now == nil {
		config.Now = time.Now
	}
	if config.NewFetcher == nil {
		config.NewFetcher = func(c WeatherKitConnection) (Fetcher, error) {
			return weatherkit.New(weatherkit.Config{TeamID: c.TeamID, ServiceID: c.ServiceID, KeyID: c.KeyID, PrivateKey: []byte(c.PrivateKey)})
		}
	}
	return &Service{
		repo: config.Repository, box: config.Box, logger: config.Logger, now: config.Now, newFetcher: config.NewFetcher,
		clients: map[uuid.UUID]cachedFetcher{}, current: map[uuid.UUID]CacheEntry{},
	}, nil
}

func (s *Service) ListProviders(ctx context.Context) ([]Provider, error) {
	rows, err := s.repo.ListProviders(ctx)
	if err != nil {
		return nil, err
	}
	result := make([]Provider, 0, len(rows))
	for _, row := range rows {
		result = append(result, row.Provider)
	}
	return result, nil
}

// CreateProvider stores a provider. The first provider created becomes active.
func (s *Service) CreateProvider(ctx context.Context, provider Provider) (Provider, error) {
	provider.ID = uuid.Must(uuid.NewV7())
	if provider.RefreshInterval == 0 {
		provider.RefreshInterval = DefaultRefresh
	}
	if err := validateProvider(provider); err != nil {
		return Provider{}, err
	}
	if provider.WeatherKit.PrivateKey == "" {
		return Provider{}, fmt.Errorf("%w: private key is required", ErrInvalidProvider)
	}
	if _, err := s.newFetcher(*provider.WeatherKit); err != nil {
		return Provider{}, fmt.Errorf("%w: %v", ErrInvalidProvider, err)
	}
	sealed, err := s.seal(provider.WeatherKit.PrivateKey)
	if err != nil {
		return Provider{}, err
	}
	connection := *provider.WeatherKit
	connection.PrivateKey = ""
	provider.WeatherKit = &connection
	if err := s.repo.UpsertProvider(ctx, storedProvider{Provider: provider, Credentials: sealed}); err != nil {
		return Provider{}, err
	}
	if _, err := s.repo.GetActiveProvider(ctx); errors.Is(err, ErrNotFound) {
		if err := s.repo.ActivateProvider(ctx, provider.ID); err != nil {
			return Provider{}, err
		}
		provider.Active = true
	} else if err != nil {
		return Provider{}, err
	}
	return provider, nil
}

// UpdateProvider applies a patch. Changing the team, service, or key ID
// requires a new private key, so stored credentials are never paired with
// identifiers the caller chose without knowing them.
func (s *Service) UpdateProvider(ctx context.Context, patch ProviderPatch) (Provider, error) {
	stored, err := s.repo.GetProvider(ctx, patch.ID)
	if err != nil {
		return Provider{}, err
	}
	provider := stored.Provider
	sealed := stored.Credentials
	if provider.WeatherKit != nil {
		connection := *provider.WeatherKit
		provider.WeatherKit = &connection
	}
	if patch.RefreshInterval != nil {
		provider.RefreshInterval = *patch.RefreshInterval
	}
	if patch.WeatherKit != nil {
		if provider.WeatherKit == nil {
			return Provider{}, fmt.Errorf("%w: provider is not a WeatherKit provider", ErrInvalidProvider)
		}
		changed, identityChanged := patch.WeatherKit, false
		apply := func(target *string, value *string) {
			if value != nil && *value != *target {
				*target, identityChanged = *value, true
			}
		}
		apply(&provider.WeatherKit.TeamID, changed.TeamID)
		apply(&provider.WeatherKit.ServiceID, changed.ServiceID)
		apply(&provider.WeatherKit.KeyID, changed.KeyID)
		if changed.PrivateKey != nil && *changed.PrivateKey != "" {
			candidate := *provider.WeatherKit
			candidate.PrivateKey = *changed.PrivateKey
			if _, err := s.newFetcher(candidate); err != nil {
				return Provider{}, fmt.Errorf("%w: %v", ErrInvalidProvider, err)
			}
			if sealed, err = s.seal(*changed.PrivateKey); err != nil {
				return Provider{}, err
			}
		} else if identityChanged {
			return Provider{}, fmt.Errorf("%w: private key is required when connection details change", ErrInvalidProvider)
		}
	}
	if err := validateProvider(provider); err != nil {
		return Provider{}, err
	}
	if err := s.repo.UpsertProvider(ctx, storedProvider{Provider: provider, Credentials: sealed}); err != nil {
		return Provider{}, err
	}
	return provider, nil
}

func (s *Service) DeleteProvider(ctx context.Context, id uuid.UUID) error {
	if err := s.repo.DeleteProvider(ctx, id); err != nil {
		return err
	}
	s.mu.Lock()
	delete(s.clients, id)
	s.mu.Unlock()
	return nil
}

func (s *Service) ActivateProvider(ctx context.Context, id uuid.UUID) error {
	return s.repo.ActivateProvider(ctx, id)
}

// TestStoredProvider checks a saved provider with its stored credentials.
func (s *Service) TestStoredProvider(ctx context.Context, id uuid.UUID) error {
	stored, err := s.repo.GetProvider(ctx, id)
	if err != nil {
		return err
	}
	fetcher, err := s.fetcher(stored)
	if err != nil {
		return err
	}
	return s.probe(ctx, fetcher)
}

// TestDraftProvider checks a connection that has not been saved. It never
// uses stored credentials, so the caller must supply the private key.
func (s *Service) TestDraftProvider(ctx context.Context, connection WeatherKitConnection) error {
	draft := Provider{Kind: ProviderWeatherKit, WeatherKit: &connection, RefreshInterval: DefaultRefresh}
	if err := validateProvider(draft); err != nil {
		return err
	}
	if connection.PrivateKey == "" {
		return fmt.Errorf("%w: private key is required", ErrInvalidProvider)
	}
	fetcher, err := s.newFetcher(connection)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidProvider, err)
	}
	return s.probe(ctx, fetcher)
}

// probe requests weather for the first location, or the origin when none exist.
func (s *Service) probe(ctx context.Context, fetcher Fetcher) error {
	var latitude, longitude float64
	locations, err := s.repo.ListLocations(ctx)
	if err != nil {
		return err
	}
	if len(locations) > 0 {
		latitude, longitude = locations[0].Latitude, locations[0].Longitude
	}
	_, err = fetcher.CurrentWeather(ctx, latitude, longitude)
	return err
}

func (s *Service) ListLocations(ctx context.Context) ([]Location, error) {
	return s.repo.ListLocations(ctx)
}

func (s *Service) CreateLocation(ctx context.Context, location Location) (Location, error) {
	location.ID = uuid.Must(uuid.NewV7())
	if err := validateLocation(location); err != nil {
		return Location{}, err
	}
	if err := s.repo.UpsertLocation(ctx, location); err != nil {
		return Location{}, err
	}
	return location, nil
}

func (s *Service) UpdateLocation(ctx context.Context, patch LocationPatch) (Location, error) {
	location, err := s.repo.GetLocation(ctx, patch.ID)
	if err != nil {
		return Location{}, err
	}
	if patch.Label != nil {
		location.Label = *patch.Label
	}
	if patch.Latitude != nil {
		location.Latitude = *patch.Latitude
	}
	if patch.Longitude != nil {
		location.Longitude = *patch.Longitude
	}
	if patch.Order != nil {
		location.Order = *patch.Order
	}
	if err := validateLocation(location); err != nil {
		return Location{}, err
	}
	if err := s.repo.UpsertLocation(ctx, location); err != nil {
		return Location{}, err
	}
	return location, nil
}

func (s *Service) DeleteLocation(ctx context.Context, id uuid.UUID) error {
	if err := s.repo.DeleteLocation(ctx, id); err != nil {
		return err
	}
	s.mu.Lock()
	delete(s.current, id)
	s.mu.Unlock()
	return nil
}

// Refresh fetches every location from the active provider, ignoring the schedule.
func (s *Service) Refresh(ctx context.Context) error {
	stored, err := s.repo.GetActiveProvider(ctx)
	if errors.Is(err, ErrNotFound) {
		return ErrNoProvider
	} else if err != nil {
		return err
	}
	return s.refresh(ctx, stored)
}

// RefreshDue refreshes the active provider when its refresh interval has
// elapsed since the last attempt. Having no active provider is not an error.
func (s *Service) RefreshDue(ctx context.Context) error {
	stored, err := s.repo.GetActiveProvider(ctx)
	if errors.Is(err, ErrNotFound) {
		return nil
	} else if err != nil {
		return err
	}
	if stored.LastAttemptedAt != nil && s.now().Sub(*stored.LastAttemptedAt) < stored.RefreshInterval {
		return nil
	}
	return s.refresh(ctx, stored)
}

func (s *Service) refresh(ctx context.Context, stored storedProvider) error {
	if err := s.repo.RecordAttempt(ctx, stored.ID, s.now(), ""); err != nil {
		return err
	}
	err := s.fetchAll(ctx, stored)
	if err != nil {
		s.logger.WarnContext(ctx, "weather refresh failed", "provider_id", stored.ID, "error", err)
		if recordErr := s.repo.RecordAttempt(ctx, stored.ID, s.now(), err.Error()); recordErr != nil {
			s.logger.WarnContext(ctx, "record weather failure", "error", recordErr)
		}
	}
	return err
}

func (s *Service) fetchAll(ctx context.Context, stored storedProvider) error {
	fetcher, err := s.fetcher(stored)
	if err != nil {
		return err
	}
	locations, err := s.repo.ListLocations(ctx)
	if err != nil {
		return err
	}
	var errs []error
	for _, location := range locations {
		conditions, err := fetcher.CurrentWeather(ctx, location.Latitude, location.Longitude)
		if err != nil {
			errs = append(errs, fmt.Errorf("%s: %w", location.Label, err))
			continue
		}
		entry := CacheEntry{
			ConditionCode: conditions.ConditionCode,
			Temperature:   conditions.Temperature,
			Daylight:      conditions.Daylight != nil && *conditions.Daylight,
			FetchedAt:     s.now(),
		}
		s.mu.Lock()
		s.current[location.ID] = entry
		s.mu.Unlock()
		if err := s.repo.Save(ctx, location.ID, entry); err != nil {
			s.logger.WarnContext(ctx, "save refreshed weather", "location_id", location.ID, "error", err)
		}
	}
	return errors.Join(errs...)
}

// Current returns cached conditions for the first location without making an upstream request.
func (s *Service) Current(ctx context.Context) (Current, error) {
	locations, err := s.repo.ListLocations(ctx)
	if err != nil {
		return Current{}, err
	}
	if len(locations) == 0 {
		return Current{}, ErrUnavailable
	}
	id := locations[0].ID
	s.mu.Lock()
	entry, ok := s.current[id]
	s.mu.Unlock()
	if !ok {
		var found bool
		entry, found, err = s.repo.Load(ctx, id)
		if err != nil {
			return Current{}, fmt.Errorf("load cached weather: %w", err)
		}
		if !found {
			return Current{}, ErrUnavailable
		}
		s.mu.Lock()
		s.current[id] = entry
		s.mu.Unlock()
	}
	return s.currentFromEntry(ctx, entry), nil
}

func (s *Service) currentFromEntry(ctx context.Context, entry CacheEntry) Current {
	staleAfter := 2 * DefaultRefresh
	if active, err := s.repo.GetActiveProvider(ctx); err == nil {
		staleAfter = 2 * active.RefreshInterval
	}
	return Current{
		Condition:     mapCondition(entry.ConditionCode, entry.Daylight),
		ConditionCode: entry.ConditionCode,
		Temperature:   int(math.Round(entry.Temperature)),
		FetchedAt:     entry.FetchedAt,
		Stale:         s.now().Sub(entry.FetchedAt) > staleAfter,
	}
}

// fetcher returns a client for the stored provider, reusing it while its
// credentials are unchanged so the client's signed-token cache survives.
func (s *Service) fetcher(stored storedProvider) (Fetcher, error) {
	if stored.WeatherKit == nil {
		return nil, fmt.Errorf("%w: unsupported provider %q", ErrInvalidProvider, stored.Kind)
	}
	digest := sha256.Sum256(stored.Credentials)
	fingerprint := fmt.Sprintf("%s|%s|%s|%s", stored.WeatherKit.TeamID, stored.WeatherKit.ServiceID, stored.WeatherKit.KeyID, hex.EncodeToString(digest[:]))
	s.mu.Lock()
	defer s.mu.Unlock()
	if cached, ok := s.clients[stored.ID]; ok && cached.fingerprint == fingerprint {
		return cached.fetcher, nil
	}
	privateKey, err := s.open(stored.Credentials)
	if err != nil {
		return nil, err
	}
	connection := *stored.WeatherKit
	connection.PrivateKey = privateKey
	fetcher, err := s.newFetcher(connection)
	if err != nil {
		return nil, err
	}
	s.clients[stored.ID] = cachedFetcher{fingerprint: fingerprint, fetcher: fetcher}
	return fetcher, nil
}

func (s *Service) seal(privateKey string) ([]byte, error) {
	if s.box == nil {
		return nil, errors.New("weather credential encryption is not configured")
	}
	plain, err := json.Marshal(struct {
		PrivateKey string `json:"privateKey"`
	}{privateKey})
	if err != nil {
		return nil, err
	}
	return s.box.Seal(plain)
}

func (s *Service) open(sealed []byte) (string, error) {
	if s.box == nil {
		return "", errors.New("weather credential encryption is not configured")
	}
	plain, err := s.box.Open(sealed)
	if err != nil {
		return "", err
	}
	var credentials struct {
		PrivateKey string `json:"privateKey"`
	}
	if err := json.Unmarshal(plain, &credentials); err != nil {
		return "", err
	}
	return credentials.PrivateKey, nil
}

func validateProvider(provider Provider) error {
	if provider.Kind != ProviderWeatherKit || provider.WeatherKit == nil || provider.RefreshInterval < MinimumRefresh {
		return ErrInvalidProvider
	}
	if provider.WeatherKit.TeamID == "" || provider.WeatherKit.ServiceID == "" || provider.WeatherKit.KeyID == "" {
		return ErrInvalidProvider
	}
	return nil
}

func validateLocation(location Location) error {
	if location.Label == "" || math.IsNaN(location.Latitude) || math.IsNaN(location.Longitude) ||
		location.Latitude < -90 || location.Latitude > 90 || location.Longitude < -180 || location.Longitude > 180 {
		return ErrInvalidLocation
	}
	return nil
}

func mapCondition(code string, daylight bool) Condition {
	var condition Condition
	switch code {
	case "Clear", "MostlyClear", "Hot":
		condition = ConditionSun
	case "PartlyCloudy", "MostlyCloudy", "Breezy", "Windy":
		condition = ConditionCloudSun
	case "Cloudy", "Foggy", "Haze", "Smoky", "BlowingDust", "BlowingSand", "BlowingSnow":
		condition = ConditionCloud
	case "Drizzle", "Rain", "HeavyRain", "Hail", "Sleet", "FreezingDrizzle", "FreezingRain", "MixedRainAndSleet", "MixedRainAndSnow", "MixedSnowAndSleet", "Snow", "HeavySnow", "Flurries", "Blizzard", "Thunderstorms", "StrongStorms", "Hurricane", "TropicalStorm":
		condition = ConditionRain
	default:
		condition = ConditionCloud
	}
	if condition == ConditionCloudSun && !daylight {
		return ConditionMoon
	}
	return condition
}
