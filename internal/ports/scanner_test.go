package ports

import (
	"strings"
	"testing"
)

func TestParseSSOutput(t *testing.T) {
	t.Run("TCP LISTEN with users field", func(t *testing.T) {
		line := `tcp   LISTEN 0      128    0.0.0.0:8080  0.0.0.0:*  users:(("node",pid=1234,fd=5))`
		entries := parseSSOutput(line, false)
		if len(entries) != 1 {
			t.Fatalf("expected 1 entry, got %d", len(entries))
		}
		e := entries[0]
		if e.Port != "8080" {
			t.Errorf("Port: want %q, got %q", "8080", e.Port)
		}
		if e.Protocol != "tcp" {
			t.Errorf("Protocol: want %q, got %q", "tcp", e.Protocol)
		}
		if e.State != "LISTEN" {
			t.Errorf("State: want %q, got %q", "LISTEN", e.State)
		}
		if e.PID != "1234" {
			t.Errorf("PID: want %q, got %q", "1234", e.PID)
		}
		if e.Process != "node" {
			t.Errorf("Process: want %q, got %q", "node", e.Process)
		}
		if e.Address != "All Interfaces" {
			t.Errorf("Address: want %q, got %q", "All Interfaces", e.Address)
		}
	})

	t.Run("UDP line without users field", func(t *testing.T) {
		line := `udp   UNCONN 0      0      0.0.0.0:53    0.0.0.0:*`
		entries := parseSSOutput(line, false)
		if len(entries) != 1 {
			t.Fatalf("expected 1 entry, got %d", len(entries))
		}
		e := entries[0]
		if e.Protocol != "udp" {
			t.Errorf("Protocol: want %q, got %q", "udp", e.Protocol)
		}
		if e.Port != "53" {
			t.Errorf("Port: want %q, got %q", "53", e.Port)
		}
	})

	t.Run("missing PID non-root known port", func(t *testing.T) {
		// Port 80 (HTTP) with no users field, non-root
		line := `tcp   LISTEN 0      128    0.0.0.0:80    0.0.0.0:*`
		entries := parseSSOutput(line, false)
		if len(entries) != 1 {
			t.Fatalf("expected 1 entry, got %d", len(entries))
		}
		e := entries[0]
		if e.PID != "-" {
			t.Errorf("PID: want %q, got %q", "-", e.PID)
		}
		if !strings.Contains(e.Process, "requires sudo") {
			t.Errorf("Process should contain 'requires sudo', got %q", e.Process)
		}
		if !strings.Contains(e.Process, "HTTP") {
			t.Errorf("Process should contain service label 'HTTP', got %q", e.Process)
		}
	})

	t.Run("missing PID root known port", func(t *testing.T) {
		// Port 22 (SSH) with no users field, root user
		line := `tcp   LISTEN 0      128    0.0.0.0:22    0.0.0.0:*`
		entries := parseSSOutput(line, true)
		if len(entries) != 1 {
			t.Fatalf("expected 1 entry, got %d", len(entries))
		}
		e := entries[0]
		if e.PID != "-" {
			t.Errorf("PID: want %q, got %q", "-", e.PID)
		}
		if !strings.Contains(e.Process, "system") {
			t.Errorf("Process should contain 'system', got %q", e.Process)
		}
		if !strings.Contains(e.Process, "SSH") {
			t.Errorf("Process should contain service label 'SSH', got %q", e.Process)
		}
	})

	t.Run("malformed / header-only input", func(t *testing.T) {
		input := "Netid State Recv-Q Send-Q Local Address:Port Peer Address:Port"
		entries := parseSSOutput(input, false)
		if len(entries) != 0 {
			t.Errorf("expected 0 entries for header line, got %d", len(entries))
		}
	})

	t.Run("empty input returns empty slice", func(t *testing.T) {
		entries := parseSSOutput("", false)
		if len(entries) != 0 {
			t.Errorf("expected 0 entries for empty input, got %d", len(entries))
		}
	})
}

func TestGetProcessDetailsSentinel(t *testing.T) {
	s := &SSScanner{}
	result, err := s.GetProcessDetails("-")
	if err != nil {
		t.Errorf("expected nil error for sentinel PID, got %v", err)
	}
	if result == "" {
		t.Errorf("expected non-empty string for sentinel PID '-', got empty")
	}
}
