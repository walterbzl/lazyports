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
	Port     string
	Protocol string
	PID      string
	Process  string
	State    string
	Address  string
}
