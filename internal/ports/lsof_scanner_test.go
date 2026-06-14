package ports

import (
	"testing"
)

func TestParseLsofOutput(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []PortEntry
	}{
		{
			name: "TCP LISTEN IPv4",
			input: `COMMAND   PID   USER   FD   TYPE             DEVICE SIZE/OFF NODE NAME
node    12345  alice  25u  IPv4 0xabcdef      0t0  TCP 127.0.0.1:3000 (LISTEN)`,
			want: []PortEntry{
				{Port: "3000", Protocol: "tcp", PID: "12345", Process: "node", State: "LISTEN", Address: "127.0.0.1"},
			},
		},
		{
			name: "TCP LISTEN wildcard port 80",
			input: `COMMAND  PID  USER   FD   TYPE             DEVICE SIZE/OFF NODE NAME
nginx   5678  root   6u  IPv4 0x111111      0t0  TCP *:80 (LISTEN)`,
			want: []PortEntry{
				{Port: "80", Protocol: "tcp", PID: "5678", Process: "nginx", State: "LISTEN", Address: "All Interfaces"},
			},
		},
		{
			name: "IPv6 TCP LISTEN",
			input: `COMMAND  PID  USER   FD   TYPE             DEVICE SIZE/OFF NODE NAME
httpd   9999  root  10u  IPv6 0x222222      0t0  TCP *:443 (LISTEN)`,
			want: []PortEntry{
				{Port: "443", Protocol: "tcp6", PID: "9999", Process: "httpd", State: "LISTEN", Address: "All Interfaces"},
			},
		},
		{
			name:  "empty input returns empty slice",
			input: "",
			want:  nil,
		},
		{
			name:  "header only returns empty slice",
			input: "COMMAND   PID   USER   FD   TYPE   DEVICE SIZE/OFF NODE NAME",
			want:  nil,
		},
		{
			name: "multiple entries",
			input: `COMMAND   PID   USER   FD   TYPE             DEVICE SIZE/OFF NODE NAME
python3  4321  user   12u  IPv4 0xcccccc      0t0  TCP 0.0.0.0:8000 (LISTEN)
redis-se 8888  redis   9u  IPv4 0xdddddd      0t0  TCP 127.0.0.1:6379 (LISTEN)`,
			want: []PortEntry{
				{Port: "8000", Protocol: "tcp", PID: "4321", Process: "python3", State: "LISTEN", Address: "0.0.0.0"},
				{Port: "6379", Protocol: "tcp", PID: "8888", Process: "redis-se", State: "LISTEN", Address: "127.0.0.1"},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := parseLsofOutput(tc.input)
			if len(got) != len(tc.want) {
				t.Fatalf("len=%d, want %d\ngot:  %+v\nwant: %+v", len(got), len(tc.want), got, tc.want)
			}
			for i, e := range got {
				w := tc.want[i]
				if e.Port != w.Port || e.Protocol != w.Protocol || e.PID != w.PID ||
					e.Process != w.Process || e.State != w.State || e.Address != w.Address {
					t.Errorf("entry[%d]:\ngot  %+v\nwant %+v", i, e, w)
				}
			}
		})
	}
}
