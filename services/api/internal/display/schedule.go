// Package display owns the theme schedule, backlight control, and Wi-Fi status.
package display

import (
	"errors"
	"fmt"
	"sort"
	"time"
)

type Mode string

const (
	ModeLight Mode = "light"
	ModeDark  Mode = "dark"
	ModeNight Mode = "night"
	ModeOff   Mode = "off"
)

// allModes is the order the day cycles through them.
var allModes = []Mode{ModeLight, ModeDark, ModeNight, ModeOff}

const minutesPerDay = 24 * 60

var ErrInvalidSchedule = errors.New("invalid display schedule")

type ModeSetting struct {
	Mode            Mode
	StartsAtMinutes int
	Brightness      int
}

type Settings struct {
	PowerSavingEnabled bool
	WakeForAlerts      bool
	// DDCBus is the I2C bus number ddcutil talks to. Empty means no display control.
	DDCBus string
}

type Schedule struct {
	Modes []ModeSetting
	Settings
}

// Validate requires every mode exactly once, in range, and in ascending order around the day.
func (s Schedule) Validate() error {
	if len(s.Modes) != len(allModes) {
		return fmt.Errorf("%w: expected %d modes, got %d", ErrInvalidSchedule, len(allModes), len(s.Modes))
	}
	byMode := map[Mode]ModeSetting{}
	for _, setting := range s.Modes {
		known := false
		for _, mode := range allModes {
			known = known || mode == setting.Mode
		}
		if !known {
			return fmt.Errorf("%w: unknown mode %q", ErrInvalidSchedule, setting.Mode)
		}
		if _, dup := byMode[setting.Mode]; dup {
			return fmt.Errorf("%w: mode %q appears twice", ErrInvalidSchedule, setting.Mode)
		}
		if setting.StartsAtMinutes < 0 || setting.StartsAtMinutes >= minutesPerDay {
			return fmt.Errorf("%w: %s starts outside the day", ErrInvalidSchedule, setting.Mode)
		}
		if setting.Brightness < 0 || setting.Brightness > 100 {
			return fmt.Errorf("%w: %s brightness must be 0-100", ErrInvalidSchedule, setting.Mode)
		}
		byMode[setting.Mode] = setting
	}
	for _, r := range s.DDCBus {
		if r < '0' || r > '9' {
			return fmt.Errorf("%w: DDC bus must be a number", ErrInvalidSchedule)
		}
	}
	descents := 0
	for i, mode := range allModes {
		next := byMode[allModes[(i+1)%len(allModes)]]
		if next.StartsAtMinutes == byMode[mode].StartsAtMinutes {
			return fmt.Errorf("%w: %s and %s start at the same time", ErrInvalidSchedule, mode, next.Mode)
		}
		if next.StartsAtMinutes < byMode[mode].StartsAtMinutes {
			descents++
		}
	}
	if descents != 1 {
		return fmt.Errorf("%w: modes must run light, dark, night, off in order around the day", ErrInvalidSchedule)
	}
	return nil
}

// Override is a live preview that replaces the schedule until it expires.
type Override struct {
	Mode       Mode
	Brightness int
	EndsAt     time.Time
}

// Current is what the panel and display should be doing right now.
type Current struct {
	Mode       Mode
	Brightness int
	// Power is false only while the panel should be off.
	Power          bool
	OverrideActive bool
	OverrideEndsAt time.Time
}

// Resolve picks the current mode from the clock, the schedule, and any live override.
func Resolve(schedule Schedule, now time.Time, override *Override) Current {
	if override != nil && now.Before(override.EndsAt) {
		return Current{Mode: override.Mode, Brightness: override.Brightness, Power: override.Mode != ModeOff, OverrideActive: true, OverrideEndsAt: override.EndsAt}
	}
	sorted := append([]ModeSetting(nil), schedule.Modes...)
	if len(sorted) == 0 {
		return Current{Mode: ModeLight, Brightness: 100, Power: true}
	}
	sort.Slice(sorted, func(i, j int) bool { return sorted[i].StartsAtMinutes < sorted[j].StartsAtMinutes })
	minutes := now.Hour()*60 + now.Minute()
	index := len(sorted) - 1
	for i, setting := range sorted {
		if setting.StartsAtMinutes <= minutes {
			index = i
		}
	}
	current := sorted[index]
	if current.Mode != ModeOff {
		return Current{Mode: current.Mode, Brightness: current.Brightness, Power: true}
	}
	lit := sorted[index]
	for i := 1; i <= len(sorted); i++ {
		candidate := sorted[(index-i+len(sorted))%len(sorted)]
		if candidate.Mode != ModeOff {
			lit = candidate
			break
		}
	}
	if schedule.PowerSavingEnabled {
		return Current{Mode: ModeOff, Brightness: lit.Brightness, Power: false}
	}
	return Current{Mode: lit.Mode, Brightness: lit.Brightness, Power: true}
}
