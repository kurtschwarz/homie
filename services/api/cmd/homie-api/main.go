package main

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"syscall"
	"time"

	assets "github.com/kurtschwarz/homie/services/api"
	"github.com/kurtschwarz/homie/services/api/internal/graph"
	"github.com/kurtschwarz/homie/services/api/internal/sdnotify"
	"github.com/kurtschwarz/homie/services/api/internal/server"
	"github.com/kurtschwarz/homie/services/api/internal/weather"
	"github.com/kurtschwarz/homie/services/api/internal/weatherkit"
)

const version = "dev"

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	addr := envOr("HOMIE_ADDR", ":3000")
	playground := os.Getenv("HOMIE_PLAYGROUND") == "1"
	weatherService, err := configureWeather()
	if err != nil {
		logger.Error("configure WeatherKit; weather query will be unavailable", "error", err)
	}

	display, err := fs.Sub(assets.DisplayFS, "web/display")
	if err != nil {
		logger.Error("mount display assets", "error", err)
		os.Exit(1)
	}
	manager, err := fs.Sub(assets.ManagerFS, "web/manager")
	if err != nil {
		logger.Error("mount manager assets", "error", err)
		os.Exit(1)
	}

	handler := server.New(server.Config{
		Resolver: &graph.Resolver{
			Version:   version,
			StartedAt: time.Now(),
			Weather:   weatherService,
		},
		Display:    display,
		Manager:    manager,
		Playground: playground,
	})

	httpServer := &http.Server{
		Addr:    addr,
		Handler: handler,
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		logger.Info("listening", "addr", addr, "version", version)
		if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("listen", "error", err)
			os.Exit(1)
		}
	}()

	if err := sdnotify.Ready(); err != nil {
		logger.Warn("sdnotify ready", "error", err)
	}
	go watchdogLoop(ctx, logger)

	<-ctx.Done()
	logger.Info("shutting down")
	_ = sdnotify.Stopping()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		logger.Error("shutdown", "error", err)
	}
}

func configureWeather() (*weather.Service, error) {
	teamID := os.Getenv("HOMIE_WEATHERKIT_TEAM_ID")
	serviceID := os.Getenv("HOMIE_WEATHERKIT_SERVICE_ID")
	keyID := os.Getenv("HOMIE_WEATHERKIT_KEY_ID")
	keyPath := os.Getenv("HOMIE_WEATHERKIT_KEY_PATH")
	latitude := os.Getenv("HOMIE_WEATHER_LAT")
	longitude := os.Getenv("HOMIE_WEATHER_LON")
	if teamID == "" && serviceID == "" && keyID == "" && keyPath == "" && latitude == "" && longitude == "" {
		return nil, nil
	}
	if teamID == "" || serviceID == "" || keyID == "" || keyPath == "" || latitude == "" || longitude == "" {
		return nil, errors.New("all HOMIE_WEATHERKIT_* and HOMIE_WEATHER_LAT/LON values are required")
	}
	if filepath.Clean(keyPath) != "/var/lib/homie/weatherkit.p8" {
		return nil, errors.New("HOMIE_WEATHERKIT_KEY_PATH must be /var/lib/homie/weatherkit.p8")
	}
	lat, err := strconv.ParseFloat(latitude, 64)
	if err != nil {
		return nil, errors.New("HOMIE_WEATHER_LAT must be a decimal number")
	}
	lon, err := strconv.ParseFloat(longitude, 64)
	if err != nil {
		return nil, errors.New("HOMIE_WEATHER_LON must be a decimal number")
	}
	privateKey, err := os.ReadFile(keyPath)
	if err != nil {
		return nil, fmt.Errorf("read WeatherKit private key: %w", err)
	}
	client, err := weatherkit.New(weatherkit.Config{
		TeamID: teamID, ServiceID: serviceID, KeyID: keyID, PrivateKey: privateKey,
	})
	if err != nil {
		return nil, err
	}
	return weather.New(client, lat, lon)
}

// watchdogLoop pings systemd's WatchdogSec at roughly half its interval, as
// systemd recommends, so a hung process (not just a crashed one) gets
// detected and restarted. It is a no-op if the unit has no watchdog set.
func watchdogLoop(ctx context.Context, logger *slog.Logger) {
	interval := watchdogInterval()
	if interval <= 0 {
		return
	}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := sdnotify.Watchdog(); err != nil {
				logger.Warn("sdnotify watchdog", "error", err)
			}
		}
	}
}

func watchdogInterval() time.Duration {
	usec := os.Getenv("WATCHDOG_USEC")
	if usec == "" {
		return 0
	}
	d, err := time.ParseDuration(usec + "us")
	if err != nil {
		return 0
	}
	return d / 2
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
