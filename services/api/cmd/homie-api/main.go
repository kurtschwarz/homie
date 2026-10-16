package main

import (
	"context"
	"errors"
	"io/fs"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	assets "github.com/kurtschwarz/homie/services/api"
	"github.com/kurtschwarz/homie/services/api/internal/graph"
	"github.com/kurtschwarz/homie/services/api/internal/sdnotify"
	"github.com/kurtschwarz/homie/services/api/internal/server"
)

const version = "dev"

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	addr := envOr("HOMIE_ADDR", ":3000")
	playground := os.Getenv("HOMIE_PLAYGROUND") == "1"

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
