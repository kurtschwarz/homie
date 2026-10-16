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

	"golang.org/x/sync/singleflight"

	"github.com/kurtschwarz/homie/services/api/internal/weatherkit"
)

// Condition is the compact set of weather glyphs supported by the UI.
type Condition string

const (
	ConditionSun      Condition = "SUN"
	ConditionCloudSun Condition = "CLOUD_SUN"
	ConditionCloud    Condition = "CLOUD"
	ConditionRain     Condition = "RAIN"
	ConditionMoon     Condition = "MOON"
)

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

// CacheEntry contains upstream weather facts and their freshness window.
type CacheEntry struct {
	ConditionCode string
	Temperature   float64
	Daylight      bool
	FetchedAt     time.Time
	ExpiresAt     time.Time
}

// Config configures a weather service for one fixed location.
type Config struct {
	Client     Fetcher
	Repository Repository
	Latitude   float64
	Longitude  float64
	Logger     *slog.Logger
	Now        func() time.Time
}

// Service caches WeatherKit's current conditions until Apple's expiration time.
type Service struct {
	client   Fetcher
	repo     Repository
	logger   *slog.Logger
	now      func() time.Time
	lat      float64
	lon      float64
	location string

	mu      sync.RWMutex
	current Current
	expires time.Time
	group   singleflight.Group
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
	return &Service{
		client:   config.Client,
		repo:     config.Repository,
		logger:   config.Logger,
		lat:      config.Latitude,
		lon:      config.Longitude,
		location: fmt.Sprintf("%s,%s", coordinate(config.Latitude), coordinate(config.Longitude)),
		now:      config.Now,
	}, nil
}

// Current returns cached conditions or fetches fresh data on demand.
func (s *Service) Current(ctx context.Context) (Current, error) {
	now := s.currentTime()
	s.mu.RLock()
	if now.Before(s.expires) {
		current := s.current
		s.mu.RUnlock()
		return current, nil
	}
	s.mu.RUnlock()

	result := s.group.DoChan("current", func() (any, error) {
		if current, ok := s.cached(s.currentTime()); ok {
			return current, nil
		}
		var stale *Current
		if s.repo != nil {
			entry, found, err := s.repo.Load(ctx, s.location)
			if err != nil {
				s.logger.Warn("load cached weather", "error", err)
			} else if found {
				current := currentFromEntry(entry, !s.currentTime().Before(entry.ExpiresAt))
				if !current.Stale {
					s.store(current, entry.ExpiresAt)
					return current, nil
				}
				stale = &current
			}
		}
		conditions, err := s.client.CurrentWeather(ctx, s.lat, s.lon)
		if err != nil {
			if stale != nil {
				s.logger.Warn("WeatherKit refresh failed; serving stale weather", "error", err, "fetched_at", stale.FetchedAt, "age", s.currentTime().Sub(stale.FetchedAt))
				return *stale, nil
			}
			return nil, err
		}
		fetchedAt := s.currentTime()
		ttl := 10 * time.Minute
		if !conditions.Expiration.IsZero() {
			ttl = conditions.Expiration.Sub(fetchedAt)
		}
		if ttl < 10*time.Minute {
			ttl = 10 * time.Minute
		}
		if ttl > time.Hour {
			ttl = time.Hour
		}
		expiresAt := fetchedAt.Add(ttl)
		entry := CacheEntry{
			ConditionCode: conditions.ConditionCode,
			Temperature:   conditions.Temperature,
			Daylight:      conditions.Daylight != nil && *conditions.Daylight,
			FetchedAt:     fetchedAt,
			ExpiresAt:     expiresAt,
		}
		current := currentFromEntry(entry, false)
		s.store(current, expiresAt)
		if s.repo != nil {
			if err := s.repo.Save(ctx, s.location, entry); err != nil {
				s.logger.Warn("save cached weather", "error", err)
			}
		}
		return current, nil
	})
	select {
	case <-ctx.Done():
		return Current{}, ctx.Err()
	case value := <-result:
		if value.Err != nil {
			return Current{}, value.Err
		}
		return value.Val.(Current), nil
	}
}

func currentFromEntry(entry CacheEntry, stale bool) Current {
	return Current{
		Condition:     mapCondition(entry.ConditionCode, entry.Daylight),
		ConditionCode: entry.ConditionCode,
		Temperature:   int(math.Round(entry.Temperature)),
		FetchedAt:     entry.FetchedAt,
		Stale:         stale,
	}
}

func (s *Service) store(current Current, expires time.Time) {
	s.mu.Lock()
	s.current = current
	s.expires = expires
	s.mu.Unlock()
}

func (s *Service) currentTime() time.Time { return s.now() }

func coordinate(value float64) string {
	return strconv.FormatFloat(value, 'f', -1, 64)
}

func (s *Service) cached(now time.Time) (Current, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.current, now.Before(s.expires)
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
