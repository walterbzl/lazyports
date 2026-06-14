package ports

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

// Scanner is the interface for port discovery and process management.
type Scanner interface {
	GetPorts() ([]PortEntry, error)
	KillProcess(pid string) error
	GetProcessDetails(pid string) (string, error)
}

// SSScanner is the concrete Scanner implementation that uses the ss utility.
type SSScanner struct{}

// Compile-time interface guard.
var _ Scanner = (*SSScanner)(nil)

// commonPorts maps well-known port numbers to service names.
var commonPorts = map[string]string{
	"21":    "FTP",
	"22":    "SSH",
	"23":    "Telnet",
	"25":    "SMTP",
	"53":    "DNS",
	"80":    "HTTP",
	"110":   "POP3",
	"143":   "IMAP",
	"443":   "HTTPS",
	"3306":  "MySQL",
	"5432":  "PostgreSQL",
	"6379":  "Redis",
	"8080":  "HTTP-Alt",
	"27017": "MongoDB",
}

// parseSSOutput parses the text output of `ss -tulnp` into a slice of PortEntry.
// isRoot controls the suffix used when PID information is unavailable.
// This is a pure function — no exec calls, making it fully unit-testable.
func parseSSOutput(out string, isRoot bool) []PortEntry {
	lines := strings.Split(out, "\n")
	var entries []PortEntry

	for _, line := range lines {
		if line == "" {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 5 || fields[0] == "Netid" {
			continue
		}

		proto := fields[0]
		state := fields[1]
		localAddr := fields[4]
		address := localAddr
		port := ""

		lastColon := strings.LastIndex(localAddr, ":")
		if lastColon != -1 {
			port = localAddr[lastColon+1:]
			address = localAddr[:lastColon]
		}
		if address == "*" || address == "0.0.0.0" || address == "[::]" {
			address = "All Interfaces"
		}

		pid := ""
		process := ""
		for _, f := range fields {
			if strings.Contains(f, "users:((") {
				content := strings.TrimPrefix(f, "users:((")
				content = strings.TrimSuffix(content, "))")
				content = strings.TrimSuffix(content, ")")
				parts := strings.Split(content, ",")
				for _, p := range parts {
					if strings.HasPrefix(p, "\"") {
						process = strings.Trim(p, "\"")
					}
					if strings.HasPrefix(p, "pid=") {
						pid = strings.TrimPrefix(p, "pid=")
					}
				}
			}
		}

		if pid == "" {
			pid = "-"
			suffix := "(requires sudo)"
			if isRoot {
				suffix = "(system)"
			}
			if service, ok := commonPorts[port]; ok {
				process = fmt.Sprintf("%s %s", service, suffix)
			} else {
				process = suffix
			}
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

// GetPorts executes `ss -tulnp` and returns the parsed port entries.
func (s *SSScanner) GetPorts() ([]PortEntry, error) {
	cmd := exec.Command("ss", "-tulnp")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to run ss: %v", err)
	}
	isRoot := os.Geteuid() == 0
	return parseSSOutput(string(output), isRoot), nil
}

// KillProcess terminates the process identified by pid.
// When pid is "-", it means the process details are unavailable (system process or requires sudo);
// KillProcess returns a descriptive error so the UI can display it without any OS calls.
func (s *SSScanner) KillProcess(pid string) error {
	if pid == "-" {
		if os.Geteuid() == 0 {
			return fmt.Errorf("Cannot kill system process")
		}
		return fmt.Errorf("Run as sudo to kill this process")
	}
	pidInt, err := strconv.Atoi(pid)
	if err != nil {
		return err
	}
	proc, err := os.FindProcess(pidInt)
	if err != nil {
		return err
	}
	return proc.Kill()
}

// GetProcessDetails returns human-readable details for the given pid.
// When pid is "-", it returns a descriptive string without invoking ps.
func (s *SSScanner) GetProcessDetails(pid string) (string, error) {
	if pid == "-" {
		if os.Geteuid() == 0 {
			return "System process (no detailed information available).", nil
		}
		return "Process details require sudo privileges.", nil
	}
	cmd := exec.Command("ps", "-p", pid, "-o", "user,lstart,cmd", "--no-headers")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get details: %v", err)
	}
	return strings.TrimSpace(string(output)), nil
}
