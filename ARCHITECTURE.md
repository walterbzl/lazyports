# LazyPorts — Architecture

## Package Layout

```
lazyports/
├── cmd/
│   └── lazyports/
│       └── main.go          Entry point — wires Scanner + UI, launches tea.Program
├── internal/
│   ├── ports/
│   │   ├── entry.go         PortEntry struct, SortMode type + constants
│   │   ├── scanner.go       Scanner interface + SSScanner (Linux ss/ps impl)
│   │   └── scanner_test.go  Unit tests for ss output parser (no exec calls)
│   └── ui/
│       ├── model.go         Bubble Tea model — Init/Update/View + helpers
│       └── styles.go        Lipgloss style declarations + makeBaseStyle factory
├── go.mod / go.sum
└── install.sh
```

## Dependency Direction

```
cmd/lazyports/main.go
  └─► internal/ui     (model, styles)
        └─► internal/ports  (Scanner interface, PortEntry, SortMode)
              └─► stdlib (os, os/exec, strconv, strings, fmt)
```

One-way only. `internal/ports` has zero project imports.

## Data Flow

```
[Linux: ss -tulnp]
       │
       ▼
 SSScanner.GetPorts()        ← exec.Command + parseSSOutput()
       │
       ▼  tea.Msg ([]ports.PortEntry)
 model.Update() [internal/ui]
  sortEntries() → filterEntries() → updateTable()
       │
       ▼
 model.View()  → rendered string
```

## Key Types

```go
// internal/ports/entry.go
type PortEntry struct { Port, Protocol, PID, Process, State, Address string }
type SortMode int  // SortByPort | SortByProcess | SortByPID

// internal/ports/scanner.go
type Scanner interface {
    GetPorts() ([]PortEntry, error)
    KillProcess(pid string) error
    GetProcessDetails(pid string) (string, error)
}

// internal/ui/model.go
type model struct {
    scanner         ports.Scanner
    table           table.Model
    textInput       textinput.Model
    entries         []ports.PortEntry
    filteredEntries []ports.PortEntry
    width, height   int
    // ...
}
```

## Design Principles

- **Scanner interface**: all OS syscalls (`exec.Command`, `os.FindProcess`, `os.Geteuid`) live in `internal/ports`. The UI never calls the OS directly.
- **makeBaseStyle(w, h)**: pure factory — no mutable package-level style vars.
- **loadPortsCmd(s Scanner)**: closure factory injected at construction time, not a global.
- **Testability**: `parseSSOutput(out string, isRoot bool)` is a pure function — unit-tested without forking any process.

## Platform

v0.1.0 is **Linux-only** (`ss -tulnp`, `ps`). macOS support (`lsof`) is on the roadmap.
