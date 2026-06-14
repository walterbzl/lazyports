package ports

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"syscall"
)

// sharedKill sends SIGTERM to the process (graceful shutdown). Works on any Unix.
func sharedKill(pid string) error {
	n, err := pidGuard(pid)
	if err != nil {
		return err
	}
	proc, err := os.FindProcess(n)
	if err != nil {
		return err
	}
	if err := proc.Signal(syscall.SIGTERM); err != nil {
		return fmt.Errorf("SIGTERM failed for %s (try sudo): %v", pid, err)
	}
	return nil
}

// sharedForceKill sends SIGKILL to the process (immediate termination). Works on any Unix.
func sharedForceKill(pid string) error {
	n, err := pidGuard(pid)
	if err != nil {
		return err
	}
	proc, err := os.FindProcess(n)
	if err != nil {
		return err
	}
	if err := proc.Kill(); err != nil {
		return fmt.Errorf("SIGKILL failed for %s (try sudo): %v", pid, err)
	}
	return nil
}

// sharedGetProcessDetails returns process info using ps, with OS-aware flags.
func sharedGetProcessDetails(pid string) (string, error) {
	if pid == "-" {
		if os.Geteuid() == 0 {
			return "System process (no detailed information available).", nil
		}
		return "Process details require sudo privileges.", nil
	}

	var out []byte
	var err error
	if runtime.GOOS == "darwin" {
		out, err = exec.Command("ps", "-p", pid, "-o", "user=", "-o", "lstart=", "-o", "command=").Output()
	} else {
		out, err = exec.Command("ps", "-p", pid, "-o", "user,lstart,cmd", "--no-headers").Output()
	}
	if err != nil {
		return "", fmt.Errorf("failed to get details: %v", err)
	}
	result := strings.TrimSpace(string(out))
	if proj := detectProject(pid); proj != "" {
		result += "\n\n── PROJECT ──\n" + proj
	}
	return result, nil
}

// sharedGetResourceInfo returns CPU% and RSS memory (in MB) for the given pid via ps.
func sharedGetResourceInfo(pid string) (ResourceInfo, error) {
	if pid == "-" {
		return ResourceInfo{}, nil
	}

	var out []byte
	var err error
	if runtime.GOOS == "darwin" {
		out, err = exec.Command("ps", "-p", pid, "-o", "%cpu=", "-o", "rss=").Output()
	} else {
		out, err = exec.Command("ps", "-p", pid, "-o", "%cpu,rss", "--no-headers").Output()
	}
	if err != nil {
		return ResourceInfo{}, err
	}
	fields := strings.Fields(strings.TrimSpace(string(out)))
	if len(fields) < 2 {
		return ResourceInfo{}, fmt.Errorf("unexpected ps output")
	}
	cpu, _ := strconv.ParseFloat(fields[0], 64)
	rssKB, _ := strconv.ParseFloat(fields[1], 64)
	return ResourceInfo{CPUPercent: cpu, MemMB: rssKB / 1024}, nil
}

// sharedOpenInBrowser opens http://localhost:PORT in the default browser.
func sharedOpenInBrowser(port string) error {
	url := "http://localhost:" + port
	var cmd string
	if runtime.GOOS == "darwin" {
		cmd = "open"
	} else {
		cmd = "xdg-open"
	}
	return exec.Command(cmd, url).Start()
}
