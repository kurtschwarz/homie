package display

import (
	"context"
	"errors"
	"log/slog"
	"sync"
	"time"
)

var ErrInvalidPreview = errors.New("invalid display preview")

const (
	maxPreview   = time.Hour
	wakeSettle   = 500 * time.Millisecond
	reconcileBuf = 1
)

type Config struct {
	Repository Repository
	// NewController builds the panel controller for a DDC bus; it is called again when the stored bus changes.
	NewController func(bus string) Controller
	Wifi          WifiReader
	Buses         BusLister
	Logger        *slog.Logger
	Now           func() time.Time
	// WakeSettle is the pause between powering the panel on and setting brightness.
	WakeSettle time.Duration
}

// Service owns the schedule and the single goroutine that writes to the panel.
type Service struct {
	repo          Repository
	newController func(bus string) Controller
	controller    Controller
	bus           string
	wifi          WifiReader
	buses         BusLister
	logger        *slog.Logger
	now           func() time.Time
	settle        time.Duration

	mu       sync.Mutex
	override *Override

	nudge chan struct{}

	// last* are touched only by the reconcile goroutine.
	written        bool
	lastPower      bool
	lastBrightness int
}

func New(config Config) *Service {
	s := &Service{
		repo:          config.Repository,
		newController: config.NewController,
		wifi:          config.Wifi,
		buses:         config.Buses,
		logger:        config.Logger,
		now:           config.Now,
		settle:        config.WakeSettle,
		nudge:         make(chan struct{}, reconcileBuf),
	}
	if s.newController == nil {
		s.newController = func(string) Controller { return NoopController() }
	}
	if s.wifi == nil {
		s.wifi = NoWifi()
	}
	if s.buses == nil {
		s.buses = NoBuses()
	}
	if s.logger == nil {
		s.logger = slog.Default()
	}
	if s.now == nil {
		s.now = time.Now
	}
	if s.settle == 0 {
		s.settle = wakeSettle
	}
	return s
}

func (s *Service) Schedule(ctx context.Context) (Schedule, error) { return s.repo.Load(ctx) }

func (s *Service) UpdateSchedule(ctx context.Context, schedule Schedule) (Schedule, error) {
	if err := schedule.Validate(); err != nil {
		return Schedule{}, err
	}
	if err := s.repo.Save(ctx, schedule); err != nil {
		return Schedule{}, err
	}
	s.Nudge()
	return s.repo.Load(ctx)
}

// Status resolves the current mode without touching the hardware.
func (s *Service) Status(ctx context.Context) (Current, error) {
	schedule, err := s.repo.Load(ctx)
	if err != nil {
		return Current{}, err
	}
	return Resolve(schedule, s.now(), s.liveOverride()), nil
}

// Preview shows a mode for a while, ignoring the schedule.
func (s *Service) Preview(ctx context.Context, mode Mode, duration time.Duration) (Current, error) {
	if duration <= 0 || duration > maxPreview {
		return Current{}, ErrInvalidPreview
	}
	schedule, err := s.repo.Load(ctx)
	if err != nil {
		return Current{}, err
	}
	setting, ok := modeSetting(schedule, mode)
	if !ok {
		return Current{}, ErrInvalidPreview
	}
	s.mu.Lock()
	s.override = &Override{Mode: mode, Brightness: setting.Brightness, EndsAt: s.now().Add(duration)}
	s.mu.Unlock()
	s.Nudge()
	return s.Status(ctx)
}

// Buses lists the displays that can be controlled over DDC/CI.
func (s *Service) Buses(ctx context.Context) ([]Bus, error) { return s.buses.ListBuses(ctx) }

func (s *Service) Wifi(ctx context.Context) (WifiStatus, error) { return s.wifi.Status(ctx) }

func (s *Service) liveOverride() *Override {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.override != nil && !s.now().Before(s.override.EndsAt) {
		s.override = nil
	}
	return s.override
}

// Nudge asks the controller goroutine to reconcile and returns immediately.
func (s *Service) Nudge() {
	select {
	case s.nudge <- struct{}{}:
	default:
	}
}

// Run reconciles on every nudge until ctx ends. It is the only caller of the Controller.
func (s *Service) Run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-s.nudge:
			if err := s.Reconcile(ctx); err != nil {
				s.logger.WarnContext(ctx, "reconcile display", "error", err)
			}
		}
	}
}

// Reconcile writes the desired panel state when it differs from the last write.
func (s *Service) Reconcile(ctx context.Context) error {
	schedule, err := s.repo.Load(ctx)
	if err != nil {
		return err
	}
	current := Resolve(schedule, s.now(), s.liveOverride())
	if s.controller == nil || schedule.DDCBus != s.bus {
		s.controller, s.bus, s.written = s.newController(schedule.DDCBus), schedule.DDCBus, false
	}
	if s.written && s.lastPower == current.Power && (!current.Power || s.lastBrightness == current.Brightness) {
		return nil
	}
	if !current.Power {
		return s.sleep(ctx, current)
	}
	return s.wake(ctx, current)
}

func (s *Service) sleep(ctx context.Context, current Current) error {
	if s.lastPower || !s.written {
		if !s.written || s.lastBrightness != current.Brightness {
			if err := s.controller.SetBrightness(ctx, current.Brightness); err != nil {
				return err
			}
		}
	}
	if err := s.controller.SetPower(ctx, false); err != nil {
		return err
	}
	s.written, s.lastPower, s.lastBrightness = true, false, current.Brightness
	return nil
}

func (s *Service) wake(ctx context.Context, current Current) error {
	if !s.written || !s.lastPower {
		if err := s.controller.SetPower(ctx, true); err != nil {
			return err
		}
		select {
		case <-time.After(s.settle):
		case <-ctx.Done():
			return ctx.Err()
		}
		s.written, s.lastPower = true, true
		s.lastBrightness = -1
	}
	if err := s.controller.SetBrightness(ctx, current.Brightness); err != nil {
		return err
	}
	s.lastBrightness = current.Brightness
	return nil
}

func modeSetting(schedule Schedule, mode Mode) (ModeSetting, bool) {
	for _, setting := range schedule.Modes {
		if setting.Mode == mode {
			return setting, true
		}
	}
	return ModeSetting{}, false
}
