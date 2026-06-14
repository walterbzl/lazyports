package ui

import (
	"testing"

	"github.com/v9mirza/lazyports/internal/config"
	"github.com/v9mirza/lazyports/internal/ports"
)

// mockScanner satisfies ports.Scanner without any OS calls.
type mockScanner struct{}

func (m *mockScanner) GetPorts() ([]ports.PortEntry, error)       { return nil, nil }
func (m *mockScanner) KillProcess(_ string) error                 { return nil }
func (m *mockScanner) ForceKillProcess(_ string) error            { return nil }
func (m *mockScanner) GetProcessDetails(_ string) (string, error) { return "", nil }
func (m *mockScanner) GetResourceInfo(_ string) (ports.ResourceInfo, error) {
	return ports.ResourceInfo{}, nil
}
func (m *mockScanner) OpenInBrowser(_ string) error { return nil }

func newTestModel(entries []ports.PortEntry) model {
	m := New(&mockScanner{}, config.Defaults())
	m.entries = entries
	m.filteredEntries = make([]ports.PortEntry, 0, len(entries))
	return m
}

func TestFilterEntries(t *testing.T) {
	fixtures := []ports.PortEntry{
		{Port: "3000", PID: "1234", Process: "node", State: "LISTEN", Protocol: "tcp", Address: "127.0.0.1"},
		{Port: "5432", PID: "5678", Process: "postgres", State: "LISTEN", Protocol: "tcp", Address: "127.0.0.1"},
		{Port: "6379", PID: "9012", Process: "redis-server", State: "LISTEN", Protocol: "tcp", Address: "127.0.0.1"},
		{Port: "8080", PID: "-", Process: "HTTP-Alt (requires sudo)", State: "LISTEN", Protocol: "tcp", Address: "0.0.0.0"},
	}

	tests := []struct {
		name      string
		query     string
		wantN     int
		wantPorts []string
	}{
		{
			name:      "empty query returns all",
			query:     "",
			wantN:     4,
			wantPorts: []string{"3000", "5432", "6379", "8080"},
		},
		{
			name:      "filter by exact port",
			query:     "3000",
			wantN:     1,
			wantPorts: []string{"3000"},
		},
		{
			name:      "filter by partial port",
			query:     "54",
			wantN:     1,
			wantPorts: []string{"5432"},
		},
		{
			name:      "filter by process name exact",
			query:     "node",
			wantN:     1,
			wantPorts: []string{"3000"},
		},
		{
			name:      "filter by process name case insensitive",
			query:     "REDIS",
			wantN:     1,
			wantPorts: []string{"6379"},
		},
		{
			name:      "filter by process name partial",
			query:     "post",
			wantN:     1,
			wantPorts: []string{"5432"},
		},
		{
			name:      "filter by PID",
			query:     "5678",
			wantN:     1,
			wantPorts: []string{"5432"},
		},
		{
			name:      "filter by PID sentinel matches sudo entry",
			query:     "-",
			wantN:     1,
			wantPorts: []string{"8080"},
		},
		{
			name:      "no match returns empty",
			query:     "zzznomatch",
			wantN:     0,
			wantPorts: []string{},
		},
		{
			name:      "whitespace-only query returns all",
			query:     " ",
			wantN:     0, // " " won't match any field literally (space doesn't appear in port/pid)
			wantPorts: []string{},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m := newTestModel(fixtures)
			m.textInput.SetValue(tc.query)
			m.filterEntries()

			if len(m.filteredEntries) != tc.wantN {
				t.Fatalf("query=%q got %d entries, want %d\nentries: %+v",
					tc.query, len(m.filteredEntries), tc.wantN, m.filteredEntries)
			}
			for i, port := range tc.wantPorts {
				if m.filteredEntries[i].Port != port {
					t.Errorf("entry[%d].Port=%q, want %q", i, m.filteredEntries[i].Port, port)
				}
			}
		})
	}
}

func TestFilterEntriesWithLabel(t *testing.T) {
	entries := []ports.PortEntry{
		{Port: "3000", PID: "1234", Process: "node"},
		{Port: "5432", PID: "5678", Process: "postgres"},
	}
	m := newTestModel(entries)
	// Manually set a label on port 3000
	_ = m.labelStore.Set("3000", "frontend")

	m.textInput.SetValue("frontend")
	m.filterEntries()

	if len(m.filteredEntries) != 1 {
		t.Fatalf("expected 1 entry matching label 'frontend', got %d", len(m.filteredEntries))
	}
	if m.filteredEntries[0].Port != "3000" {
		t.Errorf("expected port 3000, got %s", m.filteredEntries[0].Port)
	}
}

func TestFilterEntriesLabelCaseInsensitive(t *testing.T) {
	entries := []ports.PortEntry{
		{Port: "8080", PID: "999", Process: "httpd"},
	}
	m := newTestModel(entries)
	_ = m.labelStore.Set("8080", "MyAPI")

	m.textInput.SetValue("myapi")
	m.filterEntries()

	if len(m.filteredEntries) != 1 {
		t.Fatalf("label search should be case-insensitive, got %d results", len(m.filteredEntries))
	}
}
