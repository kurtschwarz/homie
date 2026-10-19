package weather

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"math"
	"strconv"
	"sync"
	"time"

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

	// RefreshInterval is the cadence used by the background refresh worker.
	RefreshInterval   = 10 * time.Minute
	defaultStaleAfter = 2 * RefreshInterval
)

var ErrUnavailable = errors.New("weather cache is unavailable")

// Current is the normalized weather data needed by the home-screen glance.
type Current struct {
	Condition     Condition
	ConditionCode string
	Temperature   int
	FetchedAt     time.Time
	Stale         bool
}

// Fetcher is the portion of the WeatherKit client used by Service.
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

// Config configures a weather service for one fixed location.
type Config struct {
	Client     Fetcher
	Repository Repository
	Latitude   float64
	Longitude  float64
	Logger     *slog.Logger
	Now        func() time.Time
	StaleAfter time.Duration
}

// Service serves cached weather and refreshes it in the background.
type Service struct {
	client     Fetcher
	repo       Repository
	logger     *slog.Logger
	now        func() time.Time
	lat        float64
	lon        float64
	location   string
	staleAfter time.Duration

	mu      sync.RWMutex
	current *CacheEntry
}

// New creates a weather service for one fixed location.
func New(config Config) (*Service, error) {
	if config.Client == nil {
		return nil, errors.New("weather client is required")
	}
	if config.Latitude < -90 || config.Latitude > 90 || config.Longitude < -180 || config.Longitude > 180 {
		return nil, errors.New("weather coordinates are outside valid ranges")
	}
	if config.Logger == nil {
		config.Logger = slog.Default()
	}
	if config.Now == nil {
		config.Now = time.Now
	}
	if config.StaleAfter <= 0 {
		config.StaleAfter = defaultStaleAfter
	}
	return &Service{
		client: config.Client, repo: config.Repository, logger: config.Logger,
		lat: config.Latitude, lon: config.Longitude,
		location: fmt.Sprintf("%s,%s", coordinate(config.Latitude), coordinate(config.Longitude)),
		now:      config.Now, staleAfter: config.StaleAfter,
	}, nil
}

// Refresh fetches current conditions from WeatherKit and updates both caches.
func (s *Service) Refresh(ctx context.Context) error {
	conditions, err := s.client.CurrentWeather(ctx, s.lat, s.lon)
	if err != nil {
		return err
	}
	entry := CacheEntry{
		ConditionCode: conditions.ConditionCode,
		Temperature:   conditions.Temperature,
		Daylight:      conditions.Daylight != nil && *conditions.Daylight,
		FetchedAt:     s.currentTime(),
	}
	s.store(entry)
	if s.repo != nil {
		if err := s.repo.Save(ctx, s.location, entry); err != nil {
			s.logger.WarnContext(ctx, "save refreshed weather", "error", err)
		}
	}
	return nil
}

// Current returns cached conditions without making an upstream request.
func (s *Service) Current(ctx context.Context) (Current, error) {
	if entry, ok := s.cached(); ok {
		return s.currentFromEntry(entry), nil
	}
	if s.repo != nil {
		entry, found, err := s.repo.Load(ctx, s.location)
		if err != nil {
			return Current{}, fmt.Errorf("load cached weather: %w", err)
		}
		if found {
			s.store(entry)
			return s.currentFromEntry(entry), nil
		}
	}
	return Current{}, ErrUnavailable
}

func (s *Service) currentFromEntry(entry CacheEntry) Current {
	return Current{
		Condition:     mapCondition(entry.ConditionCode, entry.Daylight),
		ConditionCode: entry.ConditionCode,
		Temperature:   int(math.Round(entry.Temperature)),
		FetchedAt:     entry.FetchedAt,
		Stale:         s.currentTime().Sub(entry.FetchedAt) > s.staleAfter,
	}
}

func (s *Service) store(entry CacheEntry) {
	s.mu.Lock()
	s.current = &entry
	s.mu.Unlock()
}

func (s *Service) cached() (CacheEntry, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.current == nil {
		return CacheEntry{}, false
	}
	return *s.current, true
}

func (s *Service) currentTime() time.Time { return s.now() }

func coordinate(value float64) string { return strconv.FormatFloat(value, 'f', -1, 64) }

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
