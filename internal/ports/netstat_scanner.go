package ports

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

// NetstatScanner is the Scanner implementation for Windows using netstat + tasklist.
type NetstatScanner struct{}

// Compile-time interface guard.
var _ Scanner = (*NetstatScanner)(nil)

// parseNetstatOutput parses `netstat -ano` output into listening PortEntry values.
// names maps PID -> process name (from tasklist). Pure function — no exec calls.
//
// Parsing is positional (not header-based) so it works regardless of the OS
// display language: every row is "Proto Local Remote [State] PID". TCP rows
// carry a state column; UDP rows do not.
func parseNetstatOutput(out string, names map[string]string) []PortEntry {
	var entries []PortEntry

	for _, line := range strings.Split(out, "\n") {
		fields := strings.Fields(line)
		if len(fields) < 4 {
			continue
		}
		proto := strings.ToLower(fields[0])
		if proto != "tcp" && proto != "udp" {
			continue // skips header and the "Conexiones activas" banner
		}

		isTCP := proto == "tcp"
		localAddr := fields[1]
		pid := fields[len(fields)-1]

		state := ""
		if isTCP {
			state = fields[3]
			// Only listening sockets are shown (parity with `ss -l`).
			if !strings.EqualFold(state, "LISTENING") {
				continue
			}
		}

		port, address := splitHostPort(localAddr)
		if address == "0.0.0.0" || address == "[::]" || address == "*" {
			address = "All Interfaces"
		}

		process := names[pid]
		if process == "" {
			process = "(unknown)"
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

// countNetstatConnections counts ESTABLISHED TCP connections per local port
// from the same `netstat -ano` output. Pure function.
func countNetstatConnections(out string) map[string]int {
	counts := make(map[string]int)
	for _, line := range strings.Split(out, "\n") {
		fields := strings.Fields(line)
		if len(fields) < 4 || strings.ToLower(fields[0]) != "tcp" {
			continue
		}
		if !strings.EqualFold(fields[3], "ESTABLISHED") {
			continue
		}
		if port, _ := splitHostPort(fields[1]); port != "" {
			counts[port]++
		}
	}
	return counts
}

// splitHostPort splits a netstat local-address token ("0.0.0.0:135",
// "[::]:445", "[::1]:5432") into its port and host parts. The port is always
// after the final colon, which holds for both IPv4 and bracketed IPv6.
func splitHostPort(addr string) (port, host string) {
	if i := strings.LastIndex(addr, ":"); i >= 0 {
		return addr[i+1:], addr[:i]
	}
	return "", addr
}

// buildProcessNames returns a PID -> image name map via tasklist.
func buildProcessNames() map[string]string {
	names := make(map[string]string)
	out, err := exec.Command("tasklist", "/FO", "CSV", "/NH").Output()
	if err != nil {
		return names
	}
	for _, line := range strings.Split(string(out), "\n") {
		cols := parseCSVLine(line)
		if len(cols) >= 2 {
			names[cols[1]] = cols[0]
		}
	}
	return names
}

// parseCSVLine splits a simple quoted CSV line ("a","b","c") into fields.
// tasklist never embeds commas/quotes inside fields, so this stays minimal.
func parseCSVLine(line string) []string {
	line = strings.TrimSpace(line)
	if line == "" {
		return nil
	}
	parts := strings.Split(line, "\",\"")
	for i := range parts {
		parts[i] = strings.Trim(parts[i], "\"")
	}
	return parts
}

// GetPorts runs netstat + tasklist and returns listening ports enriched with
// active connection counts.
func (s *NetstatScanner) GetPorts() ([]PortEntry, error) {
	out, err := exec.Command("netstat", "-ano").Output()
	if err != nil {
		return nil, fmt.Errorf("failed to run netstat: %v", err)
	}
	text := string(out)
	entries := parseNetstatOutput(text, buildProcessNames())

	counts := countNetstatConnections(text)
	for i := range entries {
		if c := counts[entries[i].Port]; c > 0 {
			entries[i].Connections = strconv.Itoa(c)
		} else {
			entries[i].Connections = "–"
		}
	}
	return entries, nil
}

// winPID validates a PID string for taskkill.
func winPID(pid string) error {
	if pid == "" || pid == "-" {
		return fmt.Errorf("no PID available for this entry")
	}
	if _, err := strconv.Atoi(pid); err != nil {
		return fmt.Errorf("invalid PID: %s", pid)
	}
	return nil
}

// KillProcess asks the process to close gracefully (taskkill without /F).
func (s *NetstatScanner) KillProcess(pid string) error {
	if err := winPID(pid); err != nil {
		return err
	}
	if out, err := exec.Command("taskkill", "/PID", pid).CombinedOutput(); err != nil {
		return fmt.Errorf("taskkill failed for %s: %s", pid, strings.TrimSpace(string(out)))
	}
	return nil
}

// ForceKillProcess force-terminates the process tree (taskkill /F /T).
func (s *NetstatScanner) ForceKillProcess(pid string) error {
	if err := winPID(pid); err != nil {
		return err
	}
	if out, err := exec.Command("taskkill", "/F", "/T", "/PID", pid).CombinedOutput(); err != nil {
		return fmt.Errorf("force taskkill failed for %s: %s", pid, strings.TrimSpace(string(out)))
	}
	return nil
}

// OpenInBrowser opens http://localhost:PORT via the Windows shell handler.
func (s *NetstatScanner) OpenInBrowser(port string) error {
	url := "http://localhost:" + port
	return exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
}

// GetProcessDetails returns verbose tasklist info for the given pid.
func (s *NetstatScanner) GetProcessDetails(pid string) (string, error) {
	if err := winPID(pid); err != nil {
		return "Process details unavailable for this entry.", nil
	}
	out, err := exec.Command("tasklist", "/FI", "PID eq "+pid, "/V", "/FO", "LIST").Output()
	if err != nil {
		return "", fmt.Errorf("failed to get details: %v", err)
	}
	result := strings.TrimSpace(string(out))
	if result == "" {
		return "No matching process found.", nil
	}
	return result, nil
}

// GetResourceInfo returns memory usage (MB) for the pid via tasklist.
// CPU% is not available from a single tasklist sample on Windows, so it
// is reported as 0.
func (s *NetstatScanner) GetResourceInfo(pid string) (ResourceInfo, error) {
	if err := winPID(pid); err != nil {
		return ResourceInfo{}, nil
	}
	out, err := exec.Command("tasklist", "/FI", "PID eq "+pid, "/FO", "CSV", "/NH").Output()
	if err != nil {
		return ResourceInfo{}, err
	}
	cols := parseCSVLine(strings.Split(string(out), "\n")[0])
	if len(cols) < 5 {
		return ResourceInfo{}, nil
	}
	return ResourceInfo{MemMB: parseWinMemKB(cols[4]) / 1024}, nil
}

// parseWinMemKB extracts the KB value from a tasklist memory field such as
// "121.996 KB" (where "." is a thousands separator in some locales).
func parseWinMemKB(s string) float64 {
	var digits strings.Builder
	for _, r := range s {
		if r >= '0' && r <= '9' {
			digits.WriteRune(r)
		}
	}
	v, _ := strconv.ParseFloat(digits.String(), 64)
	return v
}
