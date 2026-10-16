package weather

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/kurtschwarz/homie/services/api/internal/weatherkit"
)

type fakeFetcher struct {
	mu       sync.Mutex
	response weatherkit.CurrentConditions
	calls    int
}

func (f *fakeFetcher) CurrentWeather(context.Context, float64, float64) (weatherkit.CurrentConditions, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.calls++
	return f.response, nil
}

func TestCurrentCachesWeatherAndMapsConditions(t *testing.T) {
	fetcher := &fakeFetcher{response: weatherkit.CurrentConditions{
		ConditionCode: "PartlyCloudy",
		Temperature:   12.6,
		Daylight:      boolPointer(false),
		Expiration:    time.Now().Add(30 * time.Minute),
	}}
	service, err := New(fetcher, 43.4422469, -80.4699426)
	if err != nil {
		t.Fatal(err)
	}
	for range 2 {
		current, err := service.Current(context.Background())
		if err != nil {
			t.Fatal(err)
		}
		if current.Condition != ConditionMoon || current.Temperature != 13 {
			t.Fatalf("unexpected current weather: %+v", current)
		}
	}
	if fetcher.calls != 1 {
		t.Fatalf("expected one upstream request, got %d", fetcher.calls)
	}
}

func boolPointer(value bool) *bool { return &value }

func TestMapConditionFallsBackForUnknownCode(t *testing.T) {
	if got := mapCondition("NewAppleCondition", true); got != ConditionCloud {
		t.Fatalf("unexpected condition for unknown code: %s", got)
	}
}
