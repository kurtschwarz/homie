package weather

import (
	"context"
	"errors"
	"math"
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
}

// Fetcher is the portion of the WeatherKit client used by Service.
type Fetcher interface {
	CurrentWeather(context.Context, float64, float64) (weatherkit.CurrentConditions, error)
}

// Service caches WeatherKit's current conditions until Apple's expiration time.
type Service struct {
	client Fetcher
	lat    float64
	lon    float64
	now    func() time.Time

	mu      sync.RWMutex
	current Current
	expires time.Time
	group   singleflight.Group
}

// New creates a weather service for one fixed location.
func New(client Fetcher, latitude, longitude float64) (*Service, error) {
	if client == nil {
		return nil, errors.New("weather client is required")
	}
	if latitude < -90 || latitude > 90 || longitude < -180 || longitude > 180 {
		return nil, errors.New("weather coordinates are outside valid ranges")
	}
	return &Service{client: client, lat: latitude, lon: longitude, now: time.Now}, nil
}

// Current returns cached conditions or fetches fresh data on demand.
func (s *Service) Current(ctx context.Context) (Current, error) {
	now := s.now()
	s.mu.RLock()
	if now.Before(s.expires) {
		current := s.current
		s.mu.RUnlock()
		return current, nil
	}
	s.mu.RUnlock()

	result := s.group.DoChan("current", func() (any, error) {
		if current, ok := s.cached(s.now()); ok {
			return current, nil
		}
		conditions, err := s.client.CurrentWeather(ctx, s.lat, s.lon)
		if err != nil {
			return nil, err
		}
		current := Current{
			Condition:     mapCondition(conditions.ConditionCode, conditions.Daylight != nil && *conditions.Daylight),
			ConditionCode: conditions.ConditionCode,
			Temperature:   int(math.Round(conditions.Temperature)),
		}
		ttl := 10 * time.Minute
		if !conditions.Expiration.IsZero() {
			ttl = conditions.Expiration.Sub(s.now())
		}
		if ttl < 10*time.Minute {
			ttl = 10 * time.Minute
		}
		if ttl > time.Hour {
			ttl = time.Hour
		}
		s.mu.Lock()
		s.current = current
		s.expires = s.now().Add(ttl)
		s.mu.Unlock()
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
