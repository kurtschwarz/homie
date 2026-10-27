// Package sdlisten adopts listening sockets that systemd bound on the
// service's behalf (socket activation), so the process can serve a
// privileged port without holding CAP_NET_BIND_SERVICE. It is a no-op when
// the process was not socket-activated, so the binary runs unchanged outside
// systemd.
package sdlisten

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
)

// firstFD is SD_LISTEN_FDS_START: systemd passes sockets from fd 3 upward.
const firstFD = 3

// Listener returns the socket systemd bound under the given
// FileDescriptorName, or nil when the process was not socket-activated.
func Listener(name string) (net.Listener, error) {
	index, err := fdIndex(
		name,
		os.Getenv("LISTEN_PID"),
		os.Getenv("LISTEN_FDS"),
		os.Getenv("LISTEN_FDNAMES"),
		os.Getpid(),
	)
	if err != nil || index < 0 {
		return nil, err
	}

	// Unset so child processes do not mistake these descriptors for theirs.
	_ = os.Unsetenv("LISTEN_PID")
	_ = os.Unsetenv("LISTEN_FDS")
	_ = os.Unsetenv("LISTEN_FDNAMES")

	file := os.NewFile(uintptr(firstFD+index), name)
	defer file.Close()
	return net.FileListener(file)
}

// fdIndex resolves the position of the named socket among the passed
// descriptors. It returns -1 with no error when nothing was passed to this
// process.
func fdIndex(name, listenPID, listenFDs, listenFDNames string, pid int) (int, error) {
	if listenPID == "" || listenFDs == "" {
		return -1, nil
	}
	target, err := strconv.Atoi(listenPID)
	if err != nil || target != pid {
		return -1, nil
	}
	count, err := strconv.Atoi(listenFDs)
	if err != nil || count < 1 {
		return -1, fmt.Errorf("sdlisten: invalid LISTEN_FDS %q", listenFDs)
	}

	for index, candidate := range strings.Split(listenFDNames, ":") {
		if index >= count {
			break
		}
		if candidate == name {
			return index, nil
		}
	}
	return -1, fmt.Errorf("sdlisten: no socket named %q among %d passed descriptors", name, count)
}
