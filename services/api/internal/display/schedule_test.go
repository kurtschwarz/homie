package display

import (
	"context"
	"errors"
	"testing"
	"time"
)

func defaultSchedule(powerSaving bool) Schedule {
	return Schedule{
		Modes: []ModeSetting{
			{ModeLight, 390, 100},
			{ModeDark, 1140, 60},
			{ModeNight, 1260, 20},
			{ModeOff, 1380, 0},
		},
		Settings: Settings{PowerSavingEnabled: powerSaving},
	}
}

func at(hour, minute int) time.Time {
	return time.Date(2026, 10, 27, hour, minute, 0, 0, time.UTC)
}

func TestResolve(t *testing.T) {
	tests := []struct {
		name        string
		now         time.Time
		powerSaving bool
		override    *Override
		mode        Mode
		brightness  int
		power       bool
	}{
		{"before light wraps to off window", at(3, 0), false, nil, ModeNight, 20, true},
		{"light starts", at(6, 30), false, nil, ModeLight, 100, true},
		{"just before dark", at(18, 59), false, nil, ModeLight, 100, true},
		{"dark", at(19, 0), false, nil, ModeDark, 60, true},
		{"night", at(21, 0), false, nil, ModeNight, 20, true},
		{"off window without power saving keeps night lit", at(23, 30), false, nil, ModeNight, 20, true},
		{"off window with power saving", at(23, 30), true, nil, ModeOff, 20, false},
		{"after midnight with power saving", at(2, 0), true, nil, ModeOff, 20, false},
		{"override wins", at(23, 30), true, &Override{ModeLight, 100, at(23, 40)}, ModeLight, 100, true},
		{"expired override ignored", at(23, 30), true, &Override{ModeLight, 100, at(23, 29)}, ModeOff, 20, false},
		{"off override powers the panel down", at(12, 0), true, &Override{ModeOff, 0, at(12, 1)}, ModeOff, 0, false},
		{"off override powers down even without power saving", at(12, 0), false, &Override{ModeOff, 0, at(12, 1)}, ModeOff, 0, false},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := Resolve(defaultSchedule(test.powerSaving), test.now, test.override)
			if got.Mode != test.mode || got.Brightness != test.brightness || got.Power != test.power {
				t.Fatalf("got %+v, want mode=%s brightness=%d power=%v", got, test.mode, test.brightness, test.power)
			}
			if got.OverrideActive != (test.override != nil && test.now.Before(test.override.EndsAt)) {
				t.Fatalf("override flag wrong: %+v", got)
			}
		})
	}
}

func TestValidate(t *testing.T) {
	if err := defaultSchedule(false).Validate(); err != nil {
		t.Fatalf("default schedule invalid: %v", err)
	}
	// Rotated start (night after midnight) is still ascending around the day.
	rotated := defaultSchedule(false)
	rotated.Modes = []ModeSetting{{ModeLight, 390, 100}, {ModeDark, 1140, 60}, {ModeNight, 1260, 20}, {ModeOff, 30, 0}}
	if err := rotated.Validate(); err != nil {
		t.Fatalf("rotated schedule invalid: %v", err)
	}

	bad := map[string]func(*Schedule){
		"out of order":   func(s *Schedule) { s.Modes[1].StartsAtMinutes = 300 },
		"same time":      func(s *Schedule) { s.Modes[1].StartsAtMinutes = 390 },
		"brightness":     func(s *Schedule) { s.Modes[0].Brightness = 101 },
		"negative start": func(s *Schedule) { s.Modes[0].StartsAtMinutes = -1 },
		"past midnight":  func(s *Schedule) { s.Modes[3].StartsAtMinutes = 1440 },
		"missing mode":   func(s *Schedule) { s.Modes = s.Modes[:3] },
		"duplicate":      func(s *Schedule) { s.Modes[3].Mode = ModeNight },
	}
	for name, mutate := range bad {
		schedule := defaultSchedule(false)
		mutate(&schedule)
		if err := schedule.Validate(); !errors.Is(err, ErrInvalidSchedule) {
			t.Errorf("%s: got %v, want ErrInvalidSchedule", name, err)
		}
	}
}

type recorder struct{ calls []string }

func (r *recorder) SetBrightness(_ context.Context, percent int) error {
	r.calls = append(r.calls, "brightness:"+itoa(percent))
	return nil
}

func (r *recorder) SetPower(_ context.Context, on bool) error {
	if on {
		r.calls = append(r.calls, "power:on")
	} else {
		r.calls = append(r.calls, "power:off")
	}
	return nil
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	digits := ""
	for ; n > 0; n /= 10 {
		digits = string(rune('0'+n%10)) + digits
	}
	return digits
}

type memoryRepository struct{ schedule Schedule }

func (m *memoryRepository) Load(context.Context) (Schedule, error) { return m.schedule, nil }
func (m *memoryRepository) Save(_ context.Context, s Schedule) error {
	m.schedule = s
	return nil
}

func TestReconcileWritesOnlyTransitions(t *testing.T) {
	now := at(12, 0)
	repo := &memoryRepository{schedule: defaultSchedule(true)}
	rec := &recorder{}
	service := New(Config{Repository: repo, NewController: func(string) Controller { return rec }, Now: func() time.Time { return now }, WakeSettle: time.Nanosecond})
	ctx := context.Background()

	reconcile := func() {
		t.Helper()
		if err := service.Reconcile(ctx); err != nil {
			t.Fatal(err)
		}
	}
	expect := func(want ...string) {
		t.Helper()
		if len(rec.calls) != len(want) {
			t.Fatalf("calls = %v, want %v", rec.calls, want)
		}
		for i := range want {
			if rec.calls[i] != want[i] {
				t.Fatalf("calls = %v, want %v", rec.calls, want)
			}
		}
		rec.calls = nil
	}

	reconcile()
	expect("power:on", "brightness:100")
	reconcile()
	expect()

	now = at(19, 5)
	reconcile()
	expect("brightness:60")

	now = at(23, 30)
	reconcile()
	expect("brightness:20", "power:off")
	reconcile()
	expect()

	now = at(6, 45)
	reconcile()
	expect("power:on", "brightness:100")
}

func TestDDCUtilPowerOffTimeoutIsSuccess(t *testing.T) {
	hang := func(ctx context.Context, _ string, _ ...string) error {
		<-ctx.Done()
		return ctx.Err()
	}
	ddc := &DDCUtil{Bus: "13", Timeout: 10 * time.Millisecond, Run: hang}
	if err := ddc.SetPower(context.Background(), false); err != nil {
		t.Fatalf("power-off timeout should be success, got %v", err)
	}
	if err := ddc.SetPower(context.Background(), true); err == nil {
		t.Fatal("power-on timeout should be an error")
	}
	if err := ddc.SetBrightness(context.Background(), 50); err == nil {
		t.Fatal("brightness timeout should be an error")
	}
}

func TestDDCUtilArguments(t *testing.T) {
	var got []string
	ddc := &DDCUtil{Bus: "13", Run: func(_ context.Context, name string, args ...string) error {
		got = append([]string{name}, args...)
		return nil
	}}
	if err := ddc.SetBrightness(context.Background(), 40); err != nil {
		t.Fatal(err)
	}
	want := []string{"ddcutil", "--bus", "13", "setvcp", "10", "40"}
	if len(got) != len(want) {
		t.Fatalf("got %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("got %v, want %v", got, want)
		}
	}
}

func TestParseDetect(t *testing.T) {
	output := "Invalid display\n   I2C bus:  /dev/i2c-1\n\nDisplay 1\n   I2C bus:  /dev/i2c-13\n   DRM connector:  card1-HDMI-A-1\n   Monitor:  RTK:FT156ZB:1234\n\nDisplay 2\n   I2C bus:  /dev/i2c-14\n"
	got := parseDetect(output)
	if len(got) != 2 || got[0] != (Bus{"13", "card1-HDMI-A-1", "RTK:FT156ZB:1234"}) || got[1].Number != "14" {
		t.Fatalf("got %+v", got)
	}
	if len(parseDetect("No displays found.\n")) != 0 {
		t.Fatal("expected no buses")
	}
}

func TestParseNMCLI(t *testing.T) {
	status := parseNMCLI("no:Other:40\nyes:My\\:Net:72\n")
	if !status.Connected || status.SSID != "My:Net" || status.SignalPercent != 72 {
		t.Fatalf("got %+v", status)
	}
	if parseNMCLI("no:Other:40\n").Connected {
		t.Fatal("expected not connected")
	}
}
