package weather_test

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/kurtschwarz/homie/services/api/internal/secrets"
	"github.com/kurtschwarz/homie/services/api/internal/weather"
	"github.com/kurtschwarz/homie/services/api/internal/weather/weatherkit"
)

type fakeFetcher struct {
	mu       sync.Mutex
	response weatherkit.CurrentConditions
	err      error
	calls    int
}

func (f *fakeFetcher) CurrentWeather(context.Context, float64, float64) (weatherkit.CurrentConditions, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.calls++
	return f.response, f.err
}

type harness struct {
	service *weather.Service
	fetcher *fakeFetcher
	builds  int
	now     time.Time
}

func newHarness(t *testing.T) *harness {
	t.Helper()
	box, err := secrets.New(make([]byte, 32))
	if err != nil {
		t.Fatal(err)
	}
	daylight := true
	h := &harness{now: time.Date(2026, 10, 25, 12, 0, 0, 0, time.UTC), fetcher: &fakeFetcher{response: weatherkit.CurrentConditions{ConditionCode: "PartlyCloudy", Temperature: 12.6, Daylight: &daylight}}}
	h.service, err = weather.New(weather.Config{
		Repository: weather.NewSQLiteRepository(openDB(t)), Box: box,
		Logger: slog.New(slog.NewTextHandler(io.Discard, nil)),
		Now:    func() time.Time { return h.now },
		NewFetcher: func(c weather.WeatherKitConnection) (weather.Fetcher, error) {
			if c.PrivateKey == "bad" {
				return nil, errors.New("bad key")
			}
			h.builds++
			return h.fetcher, nil
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	return h
}

func weatherKit(key string) weather.Provider {
	return weather.Provider{Kind: weather.ProviderWeatherKit, WeatherKit: &weather.WeatherKitConnection{TeamID: "T", ServiceID: "S", KeyID: "K", PrivateKey: key}}
}

func strPtr(s string) *string { return &s }

func TestCreateProviderSealsKeyAndActivatesFirst(t *testing.T) {
	h := newHarness(t)
	ctx := context.Background()
	created, err := h.service.CreateProvider(ctx, weatherKit("pem"))
	if err != nil {
		t.Fatal(err)
	}
	if !created.Active || created.WeatherKit.PrivateKey != "" || created.RefreshInterval != weather.DefaultRefresh {
		t.Fatalf("unexpected created provider: %+v", created)
	}
	listed, _ := h.service.ListProviders(ctx)
	if len(listed) != 1 || listed[0].WeatherKit.PrivateKey != "" {
		t.Fatalf("reads must not expose the private key: %+v", listed)
	}
	if _, err := h.service.CreateProvider(ctx, weatherKit("pem")); !errors.Is(err, weather.ErrDuplicateProvider) {
		t.Fatalf("duplicate kind: got %v", err)
	}
}

func TestCreateProviderValidation(t *testing.T) {
	h := newHarness(t)
	ctx := context.Background()
	if _, err := h.service.CreateProvider(ctx, weatherKit("")); !errors.Is(err, weather.ErrInvalidProvider) {
		t.Fatalf("missing key: got %v", err)
	}
	if _, err := h.service.CreateProvider(ctx, weatherKit("bad")); !errors.Is(err, weather.ErrInvalidProvider) {
		t.Fatalf("unparseable key: got %v", err)
	}
	short := weatherKit("pem")
	short.RefreshInterval = time.Minute
	if _, err := h.service.CreateProvider(ctx, short); !errors.Is(err, weather.ErrInvalidProvider) {
		t.Fatalf("refresh below minimum: got %v", err)
	}
}

func TestUpdateProviderRequiresKeyWhenIdentityChanges(t *testing.T) {
	h := newHarness(t)
	ctx := context.Background()
	created, _ := h.service.CreateProvider(ctx, weatherKit("pem"))
	_, err := h.service.UpdateProvider(ctx, weather.ProviderPatch{ID: created.ID, WeatherKit: &weather.WeatherKitPatch{KeyID: strPtr("other")}})
	if !errors.Is(err, weather.ErrInvalidProvider) {
		t.Fatalf("identity change without key: got %v", err)
	}
	updated, err := h.service.UpdateProvider(ctx, weather.ProviderPatch{ID: created.ID, WeatherKit: &weather.WeatherKitPatch{KeyID: strPtr("other"), PrivateKey: strPtr("pem2")}})
	if err != nil || updated.WeatherKit.KeyID != "other" {
		t.Fatalf("identity change with key: %+v err=%v", updated, err)
	}
	interval := time.Hour
	updated, err = h.service.UpdateProvider(ctx, weather.ProviderPatch{ID: created.ID, RefreshInterval: &interval})
	if err != nil || updated.RefreshInterval != time.Hour || updated.WeatherKit.KeyID != "other" {
		t.Fatalf("interval-only update: %+v err=%v", updated, err)
	}
	if _, err := h.service.UpdateProvider(ctx, weather.ProviderPatch{ID: uuid.New()}); !errors.Is(err, weather.ErrNotFound) {
		t.Fatalf("unknown provider: got %v", err)
	}
}

func TestLocationValidationAndPatch(t *testing.T) {
	h := newHarness(t)
	ctx := context.Background()
	if _, err := h.service.CreateLocation(ctx, weather.Location{Label: "Nowhere", Latitude: 91}); !errors.Is(err, weather.ErrInvalidLocation) {
		t.Fatalf("latitude out of range: got %v", err)
	}
	created, err := h.service.CreateLocation(ctx, weather.Location{Label: "Kitchener", Latitude: 43.45, Longitude: -80.49})
	if err != nil {
		t.Fatal(err)
	}
	lon := 200.0
	if _, err := h.service.UpdateLocation(ctx, weather.LocationPatch{ID: created.ID, Longitude: &lon}); !errors.Is(err, weather.ErrInvalidLocation) {
		t.Fatalf("longitude out of range: got %v", err)
	}
	label := "Home"
	updated, err := h.service.UpdateLocation(ctx, weather.LocationPatch{ID: created.ID, Label: &label})
	if err != nil || updated.Label != "Home" || updated.Latitude != 43.45 {
		t.Fatalf("partial update: %+v err=%v", updated, err)
	}
}

func TestRefreshCachesEveryLocationAndReusesClient(t *testing.T) {
	h := newHarness(t)
	ctx := context.Background()
	if err := h.service.Refresh(ctx); !errors.Is(err, weather.ErrNoProvider) {
		t.Fatalf("refresh without provider: got %v", err)
	}
	if _, err := h.service.CreateProvider(ctx, weatherKit("pem")); err != nil {
		t.Fatal(err)
	}
	buildsAfterCreate := h.builds
	for _, label := range []string{"Kitchener", "Toronto"} {
		if _, err := h.service.CreateLocation(ctx, weather.Location{Label: label, Latitude: 43, Longitude: -80}); err != nil {
			t.Fatal(err)
		}
	}
	for range 2 {
		if err := h.service.Refresh(ctx); err != nil {
			t.Fatal(err)
		}
	}
	if h.fetcher.calls != 4 {
		t.Fatalf("expected 2 locations x 2 refreshes = 4 upstream calls, got %d", h.fetcher.calls)
	}
	if h.builds-buildsAfterCreate != 1 {
		t.Fatalf("client should be built once and reused, built %d times", h.builds-buildsAfterCreate)
	}
	current, err := h.service.Current(ctx)
	if err != nil || current.Condition != weather.ConditionCloudSun || current.Temperature != 13 || current.Stale {
		t.Fatalf("current: %+v err=%v", current, err)
	}
}

func TestRefreshDueHonoursIntervalAndRecordsFailure(t *testing.T) {
	h := newHarness(t)
	ctx := context.Background()
	if err := h.service.RefreshDue(ctx); err != nil {
		t.Fatalf("no provider is not an error: %v", err)
	}
	if _, err := h.service.CreateProvider(ctx, weatherKit("pem")); err != nil {
		t.Fatal(err)
	}
	if _, err := h.service.CreateLocation(ctx, weather.Location{Label: "Kitchener", Latitude: 43, Longitude: -80}); err != nil {
		t.Fatal(err)
	}
	if err := h.service.RefreshDue(ctx); err != nil || h.fetcher.calls != 1 {
		t.Fatalf("first due refresh: calls=%d err=%v", h.fetcher.calls, err)
	}
	h.now = h.now.Add(weather.DefaultRefresh - time.Second)
	if err := h.service.RefreshDue(ctx); err != nil || h.fetcher.calls != 1 {
		t.Fatalf("refresh before interval elapsed: calls=%d err=%v", h.fetcher.calls, err)
	}
	h.now = h.now.Add(time.Second)
	h.fetcher.err = errors.New("upstream down")
	if err := h.service.RefreshDue(ctx); err == nil {
		t.Fatal("expected the upstream failure to be returned")
	}
	providers, _ := h.service.ListProviders(ctx)
	if providers[0].LastError == "" || providers[0].LastAttemptedAt == nil {
		t.Fatalf("failure should be recorded on the provider: %+v", providers[0])
	}
}

func TestCurrentUnavailableWithoutLocationOrCache(t *testing.T) {
	h := newHarness(t)
	ctx := context.Background()
	if _, err := h.service.Current(ctx); !errors.Is(err, weather.ErrUnavailable) {
		t.Fatalf("no locations: got %v", err)
	}
	if _, err := h.service.CreateLocation(ctx, weather.Location{Label: "Kitchener", Latitude: 43, Longitude: -80}); err != nil {
		t.Fatal(err)
	}
	if _, err := h.service.Current(ctx); !errors.Is(err, weather.ErrUnavailable) {
		t.Fatalf("no cache: got %v", err)
	}
}

func TestCurrentMarksStaleAfterTwoRefreshIntervals(t *testing.T) {
	h := newHarness(t)
	ctx := context.Background()
	h.service.CreateProvider(ctx, weatherKit("pem"))
	h.service.CreateLocation(ctx, weather.Location{Label: "Kitchener", Latitude: 43, Longitude: -80})
	if err := h.service.Refresh(ctx); err != nil {
		t.Fatal(err)
	}
	h.now = h.now.Add(2*weather.DefaultRefresh + time.Second)
	current, err := h.service.Current(ctx)
	if err != nil || !current.Stale {
		t.Fatalf("expected stale: %+v err=%v", current, err)
	}
}

func TestTestDraftProviderNeverUsesStoredCredentials(t *testing.T) {
	h := newHarness(t)
	ctx := context.Background()
	if err := h.service.TestDraftProvider(ctx, weather.WeatherKitConnection{TeamID: "T", ServiceID: "S", KeyID: "K"}); !errors.Is(err, weather.ErrInvalidProvider) {
		t.Fatalf("draft without key: got %v", err)
	}
	if err := h.service.TestDraftProvider(ctx, weather.WeatherKitConnection{TeamID: "T", ServiceID: "S", KeyID: "K", PrivateKey: "pem"}); err != nil {
		t.Fatalf("valid draft: %v", err)
	}
}
