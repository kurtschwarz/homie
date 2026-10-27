package display

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strconv"
	"sync"
	"time"
)

// ddcMu serializes every ddcutil invocation: one monitor, one I2C bus.
var ddcMu sync.Mutex

// Controller drives the panel. Implementations only ever write; they never read the monitor back.
type Controller interface {
	SetBrightness(ctx context.Context, percent int) error
	SetPower(ctx context.Context, on bool) error
}

type noopController struct{}

// NoopController is used when no DDC bus is configured.
func NoopController() Controller { return noopController{} }

func (noopController) SetBrightness(context.Context, int) error { return nil }
func (noopController) SetPower(context.Context, bool) error     { return nil }

const defaultDDCTimeout = 5 * time.Second

// DDCUtil controls a panel over DDC/CI with the ddcutil command.
type DDCUtil struct {
	Bus     string
	Timeout time.Duration
	// Run executes a command; tests replace it.
	Run func(ctx context.Context, name string, args ...string) error
}

func (d *DDCUtil) SetBrightness(ctx context.Context, percent int) error {
	_, err := d.setvcp(ctx, "10", strconv.Itoa(percent))
	return err
}

// SetPower treats a timeout on the power-off write as success: the panel turns
// off and ddcutil then never returns.
func (d *DDCUtil) SetPower(ctx context.Context, on bool) error {
	if on {
		_, err := d.setvcp(ctx, "D6", "01")
		return err
	}
	timedOut, err := d.setvcp(ctx, "D6", "04")
	if timedOut {
		return nil
	}
	return err
}

func (d *DDCUtil) setvcp(ctx context.Context, code, value string) (timedOut bool, err error) {
	timeout := d.Timeout
	if timeout <= 0 {
		timeout = defaultDDCTimeout
	}
	run := d.Run
	if run == nil {
		run = execRun
	}
	callCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	err = run(callCtx, "ddcutil", "--bus", d.Bus, "setvcp", code, value)
	if err == nil {
		return false, nil
	}
	if ctx.Err() == nil && errors.Is(callCtx.Err(), context.DeadlineExceeded) {
		return true, fmt.Errorf("ddcutil setvcp %s %s timed out after %s: %w", code, value, timeout, err)
	}
	return false, fmt.Errorf("ddcutil setvcp %s %s: %w", code, value, err)
}

func execRun(ctx context.Context, name string, args ...string) error {
	_, err := execOutput(ctx, name, args...)
	return err
}

func execOutput(ctx context.Context, name string, args ...string) (string, error) {
	ddcMu.Lock()
	defer ddcMu.Unlock()
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.WaitDelay = time.Second
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("%w: %s", err, output)
	}
	return string(output), nil
}
