package display

import (
	"context"
	"os/exec"
	"strings"
	"time"
)

const detectTimeout = 30 * time.Second

// Bus is a display ddcutil found, addressed by its I2C bus number.
type Bus struct {
	Number    string
	Connector string
	Monitor   string
}

// BusLister finds the displays that can be controlled over DDC/CI.
type BusLister interface {
	ListBuses(context.Context) ([]Bus, error)
}

type noBuses struct{}

// NoBuses lists nothing, for machines without ddcutil.
func NoBuses() BusLister { return noBuses{} }

func (noBuses) ListBuses(context.Context) ([]Bus, error) { return nil, nil }

// DDCDetector lists buses with `ddcutil detect`.
type DDCDetector struct {
	// Run returns a command's output; tests replace it.
	Run func(ctx context.Context, name string, args ...string) (string, error)
}

// NewBusLister returns a ddcutil detector when ddcutil is installed, otherwise NoBuses.
func NewBusLister() BusLister {
	if _, err := exec.LookPath("ddcutil"); err != nil {
		return NoBuses()
	}
	return DDCDetector{}
}

func (d DDCDetector) ListBuses(ctx context.Context) ([]Bus, error) {
	ctx, cancel := context.WithTimeout(ctx, detectTimeout)
	defer cancel()
	run := d.Run
	if run == nil {
		run = execOutput
	}
	output, err := run(ctx, "ddcutil", "detect", "--brief")
	if err != nil {
		return nil, err
	}
	return parseDetect(output), nil
}

// parseDetect reads the valid "Display N" blocks; "Invalid display" blocks cannot be controlled.
func parseDetect(output string) []Bus {
	var buses []Bus
	var current *Bus
	flush := func() {
		if current != nil && current.Number != "" {
			buses = append(buses, *current)
		}
		current = nil
	}
	for _, line := range strings.Split(output, "\n") {
		trimmed := strings.TrimSpace(line)
		switch {
		case strings.HasPrefix(line, "Display "):
			flush()
			current = &Bus{}
		case strings.HasPrefix(line, "Invalid display"), strings.HasPrefix(line, "Phantom display"):
			flush()
		case current == nil:
		case strings.HasPrefix(trimmed, "I2C bus:"):
			current.Number = strings.TrimPrefix(strings.TrimSpace(strings.TrimPrefix(trimmed, "I2C bus:")), "/dev/i2c-")
		case strings.HasPrefix(trimmed, "DRM connector:"):
			current.Connector = strings.TrimSpace(strings.TrimPrefix(trimmed, "DRM connector:"))
		case strings.HasPrefix(trimmed, "Monitor:"):
			current.Monitor = strings.TrimSpace(strings.TrimPrefix(trimmed, "Monitor:"))
		}
	}
	flush()
	return buses
}
