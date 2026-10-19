package weather_test

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"math"
	"sync"
	"testing"
	"time"

	"github.com/kurtschwarz/homie/services/api/internal/store"
	"github.com/kurtschwarz/homie/services/api/internal/weather"
	"github.com/kurtschwarz/homie/services/api/internal/weather/weatherkit"
)

type fakeFetcher struct {
	mu       sync.Mutex
	response weatherkit.CurrentConditions
	err      error
	calls    int
}

type memoryCache struct {
	entry weather.CacheEntry
	found bool
}

func (c *memoryCache) Load(context.Context, string) (weather.CacheEntry, bool, error) {
	return c.entry, c.found, nil
}

func (c *memoryCache) Save(_ context.Context, _ string, entry weather.CacheEntry) error {
	c.entry = entry
	c.found = true
	return nil
}

func (f *fakeFetcher) CurrentWeather(context.Context, float64, float64) (weatherkit.CurrentConditions, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.calls++
	return f.response, f.err
}

func TestCurrentLoadsFreshWeatherFromSQLite(t *testing.T) {
	db, err := store.Open(t.TempDir()+"/weather.db", slog.New(slog.NewTextHandler(io.Discard, nil)))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	now := time.Date(2026, 10, 16, 12, 0, 0, 0, time.UTC)
	entry := weather.CacheEntry{
		ConditionCode: "PartlyCloudy",
		Temperature:   12.6,
		Daylight:      true,
		FetchedAt:     now.Add(-time.Minute),
	}
	repository := weather.NewSQLiteRepository(db)
	if err := repository.Save(context.Background(), "43.4422469,-80.4699426", entry); err != nil {
		t.Fatal(err)
	}
	fetcher := &fakeFetcher{err: errors.New("WeatherKit should not be called")}
	service, err := weather.New(weather.Config{Client: fetcher, Latitude: 43.4422469, Longitude: -80.4699426, Repository: repository, Now: func() time.Time { return now }})
	if err != nil {
		t.Fatal(err)
	}
	current, err := service.Current(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if fetcher.calls != 0 {
		t.Fatalf("expected SQLite cache hit without upstream call, got %d calls", fetcher.calls)
	}
	if current.Condition != weather.ConditionCloudSun || current.Temperature != 13 || current.Stale || !current.FetchedAt.Equal(entry.FetchedAt) {
		t.Fatalf("unexpected cached current weather: %+v", current)
	}
}

func TestCurrentServesStaleSQLiteEntry(t *testing.T) {
	db, err := store.Open(t.TempDir()+"/weather.db", slog.New(slog.NewTextHandler(io.Discard, nil)))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	now := time.Date(2026, 10, 16, 12, 0, 0, 0, time.UTC)
	entry := weather.CacheEntry{
		ConditionCode: "Rain",
		Temperature:   8,
		Daylight:      true,
		FetchedAt:     now.Add(-time.Hour),
	}
	repository := weather.NewSQLiteRepository(db)
	if err := repository.Save(context.Background(), "43.4422469,-80.4699426", entry); err != nil {
		t.Fatal(err)
	}
	service, err := weather.New(weather.Config{Client: &fakeFetcher{}, Latitude: 43.4422469, Longitude: -80.4699426, Repository: repository, Logger: slog.New(slog.NewTextHandler(io.Discard, nil)), Now: func() time.Time { return now }})
	if err != nil {
		t.Fatal(err)
	}
	current, err := service.Current(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if !current.Stale || current.Condition != weather.ConditionRain || current.Temperature != 8 || !current.FetchedAt.Equal(entry.FetchedAt) {
		t.Fatalf("unexpected stale current weather: %+v", current)
	}
}

func TestCurrentFailsWithoutCache(t *testing.T) {
	service, err := weather.New(weather.Config{Client: &fakeFetcher{}, Latitude: 43.4422469, Longitude: -80.4699426})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := service.Current(context.Background()); !errors.Is(err, weather.ErrUnavailable) {
		t.Fatalf("expected unavailable error, got %v", err)
	}
}

func TestCurrentRemapsConditionFromPersistedUpstreamCode(t *testing.T) {
	db, err := store.Open(t.TempDir()+"/weather.db", slog.New(slog.NewTextHandler(io.Discard, nil)))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	now := time.Now().UTC()
	entry := weather.CacheEntry{ConditionCode: "NewAppleCondition", Temperature: math.Pi, Daylight: true, FetchedAt: now}
	repository := weather.NewSQLiteRepository(db)
	if err := repository.Save(context.Background(), "43.4422469,-80.4699426", entry); err != nil {
		t.Fatal(err)
	}
	service, err := weather.New(weather.Config{Client: &fakeFetcher{}, Latitude: 43.4422469, Longitude: -80.4699426, Repository: repository})
	if err != nil {
		t.Fatal(err)
	}
	current, err := service.Current(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if current.Condition != weather.ConditionCloud || current.ConditionCode != "NewAppleCondition" {
		t.Fatalf("persisted condition was not mapped at read time: %+v", current)
	}
}

func TestRefreshCachesWeatherAndMapsConditions(t *testing.T) {
	now := time.Date(2026, 10, 16, 12, 0, 0, 0, time.UTC)
	fetcher := &fakeFetcher{response: weatherkit.CurrentConditions{
		ConditionCode: "PartlyCloudy",
		Temperature:   12.6,
		Daylight:      boolPointer(false),
	}}
	service, err := weather.New(weather.Config{Client: fetcher, Latitude: 43.4422469, Longitude: -80.4699426, Now: func() time.Time { return now }})
	if err != nil {
		t.Fatal(err)
	}
	if err := service.Refresh(context.Background()); err != nil {
		t.Fatal(err)
	}
	for range 2 {
		current, err := service.Current(context.Background())
		if err != nil {
			t.Fatal(err)
		}
		if current.Condition != weather.ConditionMoon || current.Temperature != 13 {
			t.Fatalf("unexpected current weather: %+v", current)
		}
	}
	if fetcher.calls != 1 {
		t.Fatalf("expected one upstream request, got %d", fetcher.calls)
	}
}

func TestMapConditionFallsBackForUnknownCode(t *testing.T) {
	cache := &memoryCache{
		entry: weather.CacheEntry{
			ConditionCode: "NewAppleCondition",
			Temperature:   12,
			Daylight:      true,
			FetchedAt:     time.Now(),
		},
		found: true,
	}
	service, err := weather.New(weather.Config{Client: &fakeFetcher{}, Latitude: 43.4422469, Longitude: -80.4699426, Repository: cache})
	if err != nil {
		t.Fatal(err)
	}
	if got, err := service.Current(context.Background()); err != nil || got.Condition != weather.ConditionCloud {
		t.Fatalf("unexpected condition for unknown code: %+v, %v", got, err)
	}
}

func TestCurrentMarksCacheStaleAfterTwoRefreshIntervals(t *testing.T) {
	now := time.Date(2026, 10, 16, 12, 0, 0, 0, time.UTC)
	cache := &memoryCache{entry: weather.CacheEntry{
		ConditionCode: "Clear", Temperature: 10, Daylight: true,
		FetchedAt: now.Add(-2*weather.RefreshInterval - time.Second),
	}, found: true}
	service, err := weather.New(weather.Config{
		Client: &fakeFetcher{}, Latitude: 43.4422469, Longitude: -80.4699426,
		Repository: cache, Now: func() time.Time { return now },
	})
	if err != nil {
		t.Fatal(err)
	}
	current, err := service.Current(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if !current.Stale {
		t.Fatalf("expected stale weather, got %+v", current)
	}
}

func boolPointer(value bool) *bool { return &value }
