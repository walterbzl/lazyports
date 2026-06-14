package ports

// SortMode represents the field by which ports are sorted.
type SortMode int

const (
	SortByPort    SortMode = iota
	SortByProcess SortMode = iota
	SortByPID     SortMode = iota
)

// PortEntry represents a single process listening on a port.
type PortEntry struct {
	Port        string
	Protocol    string
	PID         string
	Process     string
	State       string
	Address     string
	Connections string // active ESTABLISHED connections for this port ("–" if unavailable)
}

// ResourceInfo holds live CPU and memory metrics for a process.
type ResourceInfo struct {
	CPUPercent float64
	MemMB      float64
}
