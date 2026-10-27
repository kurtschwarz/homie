package display

import (
	"context"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

type WifiStatus struct {
	Connected     bool
	SSID          string
	SignalPercent int
}

// WifiReader reports the current Wi-Fi connection.
type WifiReader interface {
	Status(context.Context) (WifiStatus, error)
}

type noWifi struct{}

// NoWifi reports no connection, for machines without NetworkManager.
func NoWifi() WifiReader { return noWifi{} }

func (noWifi) Status(context.Context) (WifiStatus, error) { return WifiStatus{}, nil }

// NMCLI reads the connection from NetworkManager.
type NMCLI struct{}

// NewWifiReader returns an nmcli reader when nmcli is installed, otherwise NoWifi.
func NewWifiReader() WifiReader {
	if _, err := exec.LookPath("nmcli"); err != nil {
		return NoWifi()
	}
	return NMCLI{}
}

func (NMCLI) Status(ctx context.Context) (WifiStatus, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	output, err := exec.CommandContext(ctx, "nmcli", "-t", "-f", "ACTIVE,SSID,SIGNAL", "device", "wifi", "list").Output()
	if err != nil {
		return WifiStatus{}, err
	}
	return parseNMCLI(string(output)), nil
}

func parseNMCLI(output string) WifiStatus {
	for _, line := range strings.Split(output, "\n") {
		fields := splitTerse(line)
		if len(fields) < 3 || fields[0] != "yes" {
			continue
		}
		signal, _ := strconv.Atoi(fields[len(fields)-1])
		return WifiStatus{Connected: true, SSID: strings.Join(fields[1:len(fields)-1], ":"), SignalPercent: signal}
	}
	return WifiStatus{}
}

// splitTerse splits nmcli's terse output on colons that are not backslash-escaped.
func splitTerse(line string) []string {
	var fields []string
	var current strings.Builder
	for i := 0; i < len(line); i++ {
		switch {
		case line[i] == '\\' && i+1 < len(line):
			i++
			current.WriteByte(line[i])
		case line[i] == ':':
			fields = append(fields, current.String())
			current.Reset()
		default:
			current.WriteByte(line[i])
		}
	}
	return append(fields, current.String())
}
