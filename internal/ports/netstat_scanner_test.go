package ports

import "testing"

func TestParseNetstatOutput(t *testing.T) {
	names := map[string]string{
		"1108": "svchost.exe",
		"7900": "postgres.exe",
		"5240": "node.exe",
	}

	tests := []struct {
		name  string
		input string
		want  []PortEntry
	}{
		{
			name: "TCP LISTENING IPv4 with known PID",
			input: "  Proto  Direccion local    Direccion remota    Estado     PID\n" +
				"  TCP    0.0.0.0:135    0.0.0.0:0    LISTENING    1108",
			want: []PortEntry{
				{Port: "135", Protocol: "tcp", PID: "1108", Process: "svchost.exe", State: "LISTENING", Address: "All Interfaces"},
			},
		},
		{
			name:  "TCP LISTENING loopback",
			input: "  TCP    127.0.0.1:5432    0.0.0.0:0    LISTENING    7900",
			want: []PortEntry{
				{Port: "5432", Protocol: "tcp", PID: "7900", Process: "postgres.exe", State: "LISTENING", Address: "127.0.0.1"},
			},
		},
		{
			name:  "IPv6 TCP LISTENING",
			input: "  TCP    [::]:445    [::]:0    LISTENING    5240",
			want: []PortEntry{
				{Port: "445", Protocol: "tcp", PID: "5240", Process: "node.exe", State: "LISTENING", Address: "All Interfaces"},
			},
		},
		{
			name:  "UDP endpoint has no state column",
			input: "  UDP    0.0.0.0:5353    *:*    9999",
			want: []PortEntry{
				{Port: "5353", Protocol: "udp", PID: "9999", Process: "(unknown)", State: "", Address: "All Interfaces"},
			},
		},
		{
			name:  "ESTABLISHED TCP is excluded from listening list",
			input: "  TCP    127.0.0.1:5432    127.0.0.1:54321    ESTABLISHED    7900",
			want:  nil,
		},
		{
			name:  "header and banner lines are ignored",
			input: "Conexiones activas\n\n  Proto  Direccion local  Direccion remota  Estado  PID",
			want:  nil,
		},
		{
			name:  "empty input",
			input: "",
			want:  nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := parseNetstatOutput(tc.input, names)
			if len(got) != len(tc.want) {
				t.Fatalf("len=%d, want %d\ngot:  %+v\nwant: %+v", len(got), len(tc.want), got, tc.want)
			}
			for i, e := range got {
				if e != tc.want[i] {
					t.Errorf("entry[%d]:\ngot  %+v\nwant %+v", i, e, tc.want[i])
				}
			}
		})
	}
}

func TestCountNetstatConnections(t *testing.T) {
	input := "  TCP    0.0.0.0:5432    0.0.0.0:0    LISTENING    7900\n" +
		"  TCP    127.0.0.1:5432    127.0.0.1:60001    ESTABLISHED    7900\n" +
		"  TCP    127.0.0.1:5432    127.0.0.1:60002    ESTABLISHED    7900\n" +
		"  TCP    0.0.0.0:135    0.0.0.0:0    LISTENING    1108\n" +
		"  UDP    0.0.0.0:5353    *:*    9999"

	counts := countNetstatConnections(input)
	if counts["5432"] != 2 {
		t.Errorf("port 5432: got %d established, want 2", counts["5432"])
	}
	if counts["135"] != 0 {
		t.Errorf("port 135: got %d established, want 0", counts["135"])
	}
}

func TestParseWinMemKB(t *testing.T) {
	tests := map[string]float64{
		"5.516 KB":   5516,
		"121.996 KB": 121996,
		"8 KB":       8,
		"":           0,
	}
	for in, want := range tests {
		if got := parseWinMemKB(in); got != want {
			t.Errorf("parseWinMemKB(%q) = %v, want %v", in, got, want)
		}
	}
}

func TestSplitHostPort(t *testing.T) {
	tests := []struct {
		addr, port, host string
	}{
		{"0.0.0.0:135", "135", "0.0.0.0"},
		{"127.0.0.1:5432", "5432", "127.0.0.1"},
		{"[::]:445", "445", "[::]"},
		{"[::1]:5432", "5432", "[::1]"},
	}
	for _, tc := range tests {
		port, host := splitHostPort(tc.addr)
		if port != tc.port || host != tc.host {
			t.Errorf("splitHostPort(%q) = (%q,%q), want (%q,%q)", tc.addr, port, host, tc.port, tc.host)
		}
	}
}
