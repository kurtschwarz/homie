package main

import (
	"context"
	"errors"
	"io/fs"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	_ "time/tzdata" // a minimal Raspberry Pi image may have no system zoneinfo

	assets "github.com/kurtschwarz/homie/services/api"
	"github.com/kurtschwarz/homie/services/api/internal/accounts"
	"github.com/kurtschwarz/homie/services/api/internal/auth"
	"github.com/kurtschwarz/homie/services/api/internal/calendars"
	"github.com/kurtschwarz/homie/services/api/internal/display"
	displayjobs "github.com/kurtschwarz/homie/services/api/internal/display/jobs"
	"github.com/kurtschwarz/homie/services/api/internal/events"
	eventsjobs "github.com/kurtschwarz/homie/services/api/internal/events/jobs"
	"github.com/kurtschwarz/homie/services/api/internal/graph"
	"github.com/kurtschwarz/homie/services/api/internal/photos"
	photosjobs "github.com/kurtschwarz/homie/services/api/internal/photos/jobs"
	"github.com/kurtschwarz/homie/services/api/internal/queue"
	"github.com/kurtschwarz/homie/services/api/internal/sdlisten"
	"github.com/kurtschwarz/homie/services/api/internal/sdnotify"
	"github.com/kurtschwarz/homie/services/api/internal/secrets"
	"github.com/kurtschwarz/homie/services/api/internal/server"
	"github.com/kurtschwarz/homie/services/api/internal/store"
	"github.com/kurtschwarz/homie/services/api/internal/weather"
	weatherjobs "github.com/kurtschwarz/homie/services/api/internal/weather/jobs"
	"github.com/riverqueue/river"
)

// Set at link time by the Dockerfile via -ldflags -X.
var (
	version = "dev"
	commit  = "unknown"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	addr := envOr("HOMIE_ADDR", ":3000")
	playground := os.Getenv("HOMIE_PLAYGROUND") == "1"
	db, err := store.Open(envOr("HOMIE_DB_PATH", "/var/lib/homie/homie.db"), logger)
	if err != nil {
		logger.Error("open SQLite cache; weather will use memory only", "error", err)
	}
	var photosService *photos.Service
	var photoRotation *photos.RotationService
	var photoRepository photos.Repository
	var photosFS fs.FS
	if db != nil {
		library, err := photos.NewLibrary(envOr("HOMIE_PHOTOS_PATH", "/var/lib/homie/photos"))
		if err != nil {
			logger.Error("open photo library; photos will be unavailable", "error", err)
		} else {
			photoRepository = photos.NewSQLiteRepository(db)
			photosService = photos.New(photoRepository, library)
			photoRotation = photos.NewRotationService(photoRepository, nil)
			photosFS = library.FS()
		}
	}
	var eventsService *events.Service
	var calendarsService *calendars.Service
	var weatherService *weather.Service
	var displayService *display.Service
	if db != nil {
		newController := func(bus string) display.Controller {
			if bus == "" {
				return display.NoopController()
			}
			return &display.DDCUtil{Bus: bus}
		}
		displayService = display.New(display.Config{Repository: display.NewSQLiteRepository(db), NewController: newController, Wifi: display.NewWifiReader(), Buses: display.NewBusLister(), Logger: logger})
	}
	if db != nil {
		secretBox, err := secrets.Load(envOr("HOMIE_SECRET_KEY_PATH", secrets.DefaultKeyPath))
		if err != nil {
			logger.Error("load calendar credential encryption key", "error", err)
			os.Exit(1)
		}
		calendarsService = calendars.New(calendars.NewSQLiteRepository(db), secretBox, logger)
		eventsService = events.New(events.NewSQLiteRepository(db))
		eventsService.Calendars = calendarsService
		eventsService.Logger = logger
		weatherService, err = weather.New(weather.Config{Repository: weather.NewSQLiteRepository(db), Box: secretBox, Logger: logger})
		if err != nil {
			logger.Error("configure weather; weather will be unavailable", "error", err)
		}
		if weatherService != nil && photosService != nil {
			photosService.Home = func(ctx context.Context) (photos.Coordinate, string, bool, error) {
				home, ok, err := weatherService.Home(ctx)
				if err != nil || !ok {
					return photos.Coordinate{}, "", false, err
				}
				return photos.Coordinate{Latitude: home.Latitude, Longitude: home.Longitude}, home.Label, true, nil
			}
		}
	}

	var refresher weatherjobs.Refresher
	if weatherService != nil {
		refresher = weatherService
	}
	var eventsRefresher eventsjobs.Refresher
	if eventsService != nil {
		eventsRefresher = eventsService
	}
	workers := river.NewWorkers()
	river.AddWorker(workers, weatherjobs.NewRefreshWorker(refresher, logger))
	river.AddWorker(workers, eventsjobs.NewRefreshWorker(eventsRefresher, logger))
	river.AddWorker(workers, photosjobs.NewPruneWorker(photoRotation, logger))
	var displayNudger displayjobs.Nudger
	if displayService != nil {
		displayNudger = displayService
	}
	river.AddWorker(workers, displayjobs.NewReconcileWorker(displayNudger))

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
	localAddr := os.Getenv("HOMIE_LOCAL_ADDR")
	var token string
	var signer *auth.Signer
	var accountsService *accounts.Service
	token, err = auth.Load(envOr("HOMIE_AUTH_TOKEN_PATH", auth.DefaultTokenPath))
	if err != nil {
		logger.Error("load API auth token", "error", err)
		os.Exit(1)
	}
	logger.Info("API bearer token", "token", token)
	signingSecret, err := auth.LoadSigningSecret(envOr("HOMIE_SIGNING_SECRET_PATH", auth.DefaultSigningSecretPath))
	if err != nil {
		logger.Error("load photo URL signing secret", "error", err)
		os.Exit(1)
	}
	signer = auth.NewSigner(signingSecret)
	if db != nil {
		accountsService, err = accounts.New(context.Background(), accounts.NewSQLiteRepository(db), logger)
		if err != nil {
			logger.Error("configure user accounts", "error", err)
			os.Exit(1)
		}
		if code := accountsService.SetupCode(); code != "" {
			logger.Info("user setup code", "code", code)
		}
	}

	serverConfig := server.Config{
		Resolver: &graph.Resolver{
			Version:       version,
			StartedAt:     time.Now(),
			Weather:       weatherService,
			Photos:        photosService,
			PhotoRotation: photoRotation,
			Events:        eventsService,
			Calendars:     calendarsService,
			PhotoSigner:   signer,
			Accounts:      accountsService,
			Display:       displayService,
		},
		Display:    display,
		Manager:    manager,
		Photos:     photosFS,
		Playground: playground,
		Token:      token,
		Signer:     signer,
	}
	apiHandler := server.New(serverConfig)
	if localAddr != "" {
		apiHandler = server.NewLAN(serverConfig)
	}

	httpServer := &http.Server{
		Addr:    addr,
		Handler: apiHandler,
	}
	var localServer *http.Server
	if localAddr != "" {
		localServer = &http.Server{Addr: localAddr, Handler: server.NewLocal(serverConfig)}
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	backgroundQueue, err := queue.New(ctx, queue.Config{
		DBPath:       envOr("HOMIE_RIVER_DB_PATH", "/var/lib/homie/river.db"),
		Logger:       logger,
		Workers:      workers,
		PeriodicJobs: []*river.PeriodicJob{weatherjobs.PeriodicJob(), eventsjobs.PeriodicJob(), photosjobs.PeriodicJob(), displayjobs.PeriodicJob()},
	})
	if err != nil {
		logger.Error("configure background queue", "error", err)
		if db != nil {
			_ = db.Close()
		}
		os.Exit(1)
	}
	if displayService != nil {
		go displayService.Run(ctx)
	}
	if err := backgroundQueue.Start(ctx); err != nil {
		logger.Error("start background queue", "error", err)
		_ = backgroundQueue.Stop(context.Background())
		if db != nil {
			_ = db.Close()
		}
		os.Exit(1)
	}

	lanListener, err := sdlisten.Listener("lan")
	if err == nil && lanListener == nil {
		lanListener, err = net.Listen("tcp", addr)
	}
	if err != nil {
		logger.Error("listen", "error", err)
		_ = backgroundQueue.Stop(context.Background())
		if db != nil {
			_ = db.Close()
		}
		os.Exit(1)
	}

	go func() {
		logger.Info("listening", "addr", lanListener.Addr().String(), "version", version, "commit", commit)
		if err := httpServer.Serve(lanListener); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("listen", "error", err)
			os.Exit(1)
		}
	}()
	if localServer != nil {
		go func() {
			logger.Info("listening for local UI", "addr", localAddr, "version", version)
			if err := localServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
				logger.Error("listen for local UI", "error", err)
				os.Exit(1)
			}
		}()
	}

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
	if localServer != nil {
		if err := localServer.Shutdown(shutdownCtx); err != nil {
			logger.Error("shutdown local UI", "error", err)
		}
	}
	if err := backgroundQueue.Stop(shutdownCtx); err != nil {
		logger.Error("stop background queue", "error", err)
	}
	if db != nil {
		if err := db.Close(); err != nil {
			logger.Error("close SQLite cache", "error", err)
		}
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
