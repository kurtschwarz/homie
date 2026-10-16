// Package sdnotify sends systemd service notifications (READY=1,
// WATCHDOG=1) over the NOTIFY_SOCKET unix datagram socket. It is a no-op
// when NOTIFY_SOCKET is unset, so the binary runs unchanged outside systemd.
package sdnotify

import (
	"net"
	"os"
)

func notify(state string) error {
	addr := os.Getenv("NOTIFY_SOCKET")
	if addr == "" {
		return nil
	}
	conn, err := net.Dial("unixgram", addr)
	if err != nil {
		return err
	}
	defer conn.Close()
	_, err = conn.Write([]byte(state))
	return err
}

// Ready tells systemd the service has finished starting up.
func Ready() error { return notify("READY=1") }

// Watchdog pings systemd's watchdog timer to signal the service is alive.
func Watchdog() error { return notify("WATCHDOG=1") }

// Stopping tells systemd the service is beginning a graceful shutdown.
func Stopping() error { return notify("STOPPING=1") }
