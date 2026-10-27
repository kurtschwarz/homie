package sdlisten

import "testing"

func TestFdIndex(t *testing.T) {
	const pid = 4242

	tests := []struct {
		name      string
		want      string
		listenPID string
		fds       string
		names     string
		index     int
		wantErr   bool
	}{
		{name: "not activated", want: "lan", index: -1},
		{name: "meant for another process", want: "lan", listenPID: "1", fds: "1", names: "lan", index: -1},
		{name: "single socket", want: "lan", listenPID: "4242", fds: "1", names: "lan", index: 0},
		{name: "second of two", want: "lan", listenPID: "4242", fds: "2", names: "other:lan", index: 1},
		{name: "name missing", want: "lan", listenPID: "4242", fds: "1", names: "other", index: -1, wantErr: true},
		{name: "name beyond passed count", want: "lan", listenPID: "4242", fds: "1", names: "other:lan", index: -1, wantErr: true},
		{name: "bad count", want: "lan", listenPID: "4242", fds: "zero", names: "lan", index: -1, wantErr: true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			index, err := fdIndex(test.want, test.listenPID, test.fds, test.names, pid)
			if (err != nil) != test.wantErr {
				t.Fatalf("error = %v, wantErr %v", err, test.wantErr)
			}
			if index != test.index {
				t.Fatalf("index = %d, want %d", index, test.index)
			}
		})
	}
}
