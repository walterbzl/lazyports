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

// countConnections returns a map of local port → number of ESTABLISHED TCP connections.
// Non-fatal: returns nil if ss fails (caller treats nil as "unavailable").
func countConnections() map[string]int {
	out, err := exec.Command("ss", "-tn", "state", "established").Output()
	if err != nil {
		return nil
	}
	counts := make(map[string]int)
	for _, line := range strings.Split(string(out), "\n") {
		fields := strings.Fields(line)
		// ss -tn columns: Netid State Recv-Q Send-Q LocalAddr:Port PeerAddr:Port
		if len(fields) < 5 || fields[0] == "Netid" {
			continue
		}
		localAddr := fields[4]
		if i := strings.LastIndex(localAddr, ":"); i >= 0 {
			counts[localAddr[i+1:]]++
		}
	}
	return counts
}

// detectProject inspects the CWD of a process via /proc and returns
// a short human-readable project/framework string, or "" if unavailable.
func detectProject(pid string) string {
	cwd, err := os.Readlink("/proc/" + pid + "/cwd")
	if err != nil {
		return ""
	}
	var lines []string
	if data, err := os.ReadFile(cwd + "/package.json"); err == nil {
		content := string(data)
		if name := extractJSONString(content, "name"); name != "" {
			lines = append(lines, "Project:   "+name)
		}
		if fw := detectNodeFramework(content); fw != "" {
			lines = append(lines, "Framework: "+fw)
		}
	} else if data, err := os.ReadFile(cwd + "/go.mod"); err == nil {
		parts := strings.Split(strings.TrimSpace(strings.TrimPrefix(strings.Split(string(data), "\n")[0], "module ")), "/")
		lines = append(lines, "Project:   "+parts[len(parts)-1])
		lines = append(lines, "Framework: Go")
	} else if _, err := os.ReadFile(cwd + "/Cargo.toml"); err == nil {
		lines = append(lines, "Framework: Rust")
	} else if _, err := os.ReadFile(cwd + "/pyproject.toml"); err == nil {
		lines = append(lines, "Framework: Python")
	}
	lines = append(lines, "Directory: "+cwd)
	return strings.Join(lines, "\n")
}

func extractJSONString(content, key string) string {
	search := `"` + key + `"`
	idx := strings.Index(content, search)
	if idx < 0 {
		return ""
	}
	rest := content[idx+len(search):]
	colon := strings.Index(rest, ":")
	if colon < 0 {
		return ""
	}
	rest = strings.TrimSpace(rest[colon+1:])
	if len(rest) == 0 || rest[0] != '"' {
		return ""
	}
	end := strings.Index(rest[1:], `"`)
	if end < 0 {
		return ""
	}
	return rest[1 : end+1]
}

func detectNodeFramework(pkg string) string {
	for _, dep := range []struct{ key, name string }{
		{`"next"`, "Next.js"}, {`"vite"`, "Vite"}, {`"express"`, "Express"},
		{`"fastify"`, "Fastify"}, {`"nuxt"`, "Nuxt.js"}, {`"remix"`, "Remix"},
		{`"@angular/core"`, "Angular"}, {`"vue"`, "Vue.js"}, {`"svelte"`, "Svelte"},
	} {
		if strings.Contains(pkg, dep.key) {
			return dep.name
		}
	}
	return ""
}

// GetPorts executes `ss -tulnp`, parses listening ports, and enriches each
// entry with its active connection count via a second ss call.
func (s *SSScanner) GetPorts() ([]PortEntry, error) {
	cmd := exec.Command("ss", "-tulnp")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to run ss: %v", err)
	}
	isRoot := os.Geteuid() == 0
	entries := parseSSOutput(string(output), isRoot)

	counts := countConnections()
	for i := range entries {
		if counts != nil {
			if c := counts[entries[i].Port]; c > 0 {
				entries[i].Connections = strconv.Itoa(c)
			} else {
				entries[i].Connections = "–"
			}
		} else {
			entries[i].Connections = "–"
		}
	}
	return entries, nil
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

// GetProcessDetails returns human-readable details for the given pid,
// enriched with project/framework detection from the process CWD.
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
	result := strings.TrimSpace(string(output))
	if proj := detectProject(pid); proj != "" {
		result += "\n\n── PROJECT ──\n" + proj
	}
	return result, nil
}
