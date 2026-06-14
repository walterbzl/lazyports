package ports

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// LsofScanner is the Scanner implementation for macOS using lsof.
type LsofScanner struct{}

// Compile-time interface guard.
var _ Scanner = (*LsofScanner)(nil)

// parseLsofOutput parses the text output of `lsof -nP -iTCP -iUDP -sTCP:LISTEN`
// into a slice of PortEntry. Pure function — no exec calls, fully unit-testable.
func parseLsofOutput(out string) []PortEntry {
	lines := strings.Split(out, "\n")
	var entries []PortEntry

	for _, line := range lines {
		if line == "" {
			continue
		}
		fields := strings.Fields(line)
		// lsof columns: COMMAND PID USER FD TYPE DEVICE SIZE/OFF NODE NAME
		if len(fields) < 9 || fields[0] == "COMMAND" {
			continue
		}

		process := fields[0]
		pid := fields[1]
		proto := strings.ToLower(fields[7]) // NODE column holds TCP/UDP on macOS

		// NAME column: "127.0.0.1:3000 (LISTEN)" or "*:80 (LISTEN)"
		name := fields[8]
		state := ""
		if idx := strings.Index(name, " ("); idx >= 0 {
			state = strings.Trim(name[idx+2:], "()")
			name = name[:idx]
		}

		// Split host:port from the NAME
		port := ""
		address := name
		if i := strings.LastIndex(name, ":"); i >= 0 {
			port = name[i+1:]
			address = name[:i]
		}
		if address == "*" {
			address = "All Interfaces"
		}

		// Map protocol from lsof type field (IPv4/IPv6 + proto from NODE)
		typeField := strings.ToLower(fields[4]) // IPv4 or IPv6
		if typeField == "ipv6" {
			proto = "tcp6"
			if strings.EqualFold(fields[7], "UDP") {
				proto = "udp6"
			}
		}

		if state == "" {
			state = "LISTEN"
		}

		entries = append(entries, PortEntry{
			Port:     port,
			Protocol: proto,
			PID:      pid,
			Process:  process,
			State:    state,
			Address:  address,
		})
	}
	return entries
}

// GetPorts executes lsof and returns listening ports.
func (s *LsofScanner) GetPorts() ([]PortEntry, error) {
	out, err := exec.Command("lsof", "-nP", "-iTCP", "-iUDP", "-sTCP:LISTEN").Output()
	if err != nil {
		return nil, fmt.Errorf("failed to run lsof: %v", err)
	}
	isRoot := os.Geteuid() == 0
	entries := parseLsofOutput(string(out))

	// Populate Connections field via a second lsof call for ESTABLISHED TCP
	counts := countLsofConnections()
	for i := range entries {
		if counts != nil {
			if c := counts[entries[i].Port]; c > 0 {
				entries[i].Connections = fmt.Sprintf("%d", c)
			} else {
				entries[i].Connections = "–"
			}
		} else {
			entries[i].Connections = "–"
		}
		// Annotate system ports when PID info is missing
		if entries[i].PID == "" {
			entries[i].PID = "-"
			suffix := "(requires sudo)"
			if isRoot {
				suffix = "(system)"
			}
			if entries[i].Process == "" {
				if svc, ok := commonPorts[entries[i].Port]; ok {
					entries[i].Process = svc + " " + suffix
				} else {
					entries[i].Process = suffix
				}
			}
		}
	}
	return entries, nil
}

// countLsofConnections counts ESTABLISHED TCP connections per local port via lsof.
func countLsofConnections() map[string]int {
	out, err := exec.Command("lsof", "-nP", "-iTCP", "-sTCP:ESTABLISHED").Output()
	if err != nil {
		return nil
	}
	counts := make(map[string]int)
	for _, line := range strings.Split(string(out), "\n") {
		fields := strings.Fields(line)
		if len(fields) < 9 || fields[0] == "COMMAND" {
			continue
		}
		name := fields[8]
		if idx := strings.Index(name, " ("); idx >= 0 {
			name = name[:idx]
		}
		if i := strings.LastIndex(name, ":"); i >= 0 {
			localPart := name[:i]
			if j := strings.Index(localPart, "->"); j < 0 {
				// name is "local->peer"; take the local port
				// actually lsof ESTABLISHED shows "local:port->peer:port"
				_ = localPart
			}
			// The NAME in ESTABLISHED form is "addr:port->peeraddr:peerport"
			// We want the local port which is before ->
			arrow := strings.Index(name, "->")
			if arrow >= 0 {
				local := name[:arrow]
				if li := strings.LastIndex(local, ":"); li >= 0 {
					counts[local[li+1:]]++
				}
			}
		}
	}
	return counts
}

// Delegate all process-management methods to shared helpers.
func (s *LsofScanner) KillProcess(pid string) error      { return sharedKill(pid) }
func (s *LsofScanner) ForceKillProcess(pid string) error { return sharedForceKill(pid) }
func (s *LsofScanner) OpenInBrowser(port string) error   { return sharedOpenInBrowser(port) }
func (s *LsofScanner) GetResourceInfo(pid string) (ResourceInfo, error) {
	return sharedGetResourceInfo(pid)
}
func (s *LsofScanner) GetProcessDetails(pid string) (string, error) {
	return sharedGetProcessDetails(pid)
}
