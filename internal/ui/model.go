package ui

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/v9mirza/lazyports/internal/ports"
)

type model struct {
	scanner         ports.Scanner
	table           table.Model
	textInput       textinput.Model
	entries         []ports.PortEntry
	filteredEntries []ports.PortEntry
	err             error
	status          string
	width           int
	height          int
	isFiltering     bool
	showDetails     bool
	detailsContent  string
	sortMode        ports.SortMode
}

// New builds the initial model with table and textinput configured,
// injecting the Scanner the UI will use for all OS interactions.
func New(scanner ports.Scanner) model {
	columns := []table.Column{
		{Title: "Port", Width: 8},
		{Title: "Proto", Width: 6},
		{Title: "State", Width: 12},
		{Title: "PID", Width: 8},
		{Title: "Address", Width: 22},
		{Title: "Conn", Width: 6},
		{Title: "Process", Width: 20},
	}

	t := table.New(
		table.WithColumns(columns),
		table.WithFocused(true),
		table.WithHeight(10),
	)

	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("#6c7086")).
		BorderBottom(true).
		Bold(true)
	s.Selected = s.Selected.
		Foreground(lipgloss.Color("#cdd6f4")).
		Background(lipgloss.Color("#313244")).
		Bold(false)
	t.SetStyles(s)

	ti := textinput.New()
	ti.Placeholder = "Search ports, processes, pids..."
	ti.CharLimit = 156
	ti.Width = 40

	return model{scanner: scanner, table: t, textInput: ti}
}

// loadPortsCmd returns a tea.Cmd closure bound to the given Scanner.
// The UI never references a concrete scanner or package-global.
func loadPortsCmd(s ports.Scanner) tea.Cmd {
	return func() tea.Msg {
		entries, err := s.GetPorts()
		if err != nil {
			return err
		}
		return entries
	}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(loadPortsCmd(m.scanner), textinput.Blink)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	if m.isFiltering {
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "enter", "esc":
				m.isFiltering = false
				m.table.Focus()
				return m, nil
			}
		}
		m.textInput, cmd = m.textInput.Update(msg)
		m.filterEntries()
		return m, cmd
	}

	if m.showDetails {
		switch msg := msg.(type) {
		case tea.KeyMsg:
			if msg.String() == "esc" || msg.String() == "q" || msg.String() == "enter" {
				m.showDetails = false
				return m, nil
			}
		}
		return m, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "/":
			m.isFiltering = true
			m.textInput.Focus()
			m.textInput.SetValue("")
			return m, textinput.Blink
		case "r":
			m.status = "Refreshing..."
			return m, loadPortsCmd(m.scanner)
		case "s":
			m.sortMode++
			if m.sortMode > ports.SortByPID {
				m.sortMode = ports.SortByPort
			}
			m.sortEntries()
			m.updateTableColumns()
			m.updateTable()
		case "enter":
			if len(m.filteredEntries) > 0 {
				selectedIdx := m.table.Cursor()
				if selectedIdx >= 0 && selectedIdx < len(m.filteredEntries) {
					target := m.filteredEntries[selectedIdx]
					details, err := m.scanner.GetProcessDetails(target.PID)
					if err != nil {
						m.detailsContent = fmt.Sprintf("Error: %v", err)
					} else {
						if target.PID == "-" {
							m.detailsContent = details
						} else {
							m.detailsContent = fmt.Sprintf(
								"Port:      %s/%s\nPID:       %s\nAddress:   %s\nState:     %s\nProcess:   %s\n\n%s",
								target.Port, target.Protocol, target.PID, target.Address, target.State, target.Process, details,
							)
						}
					}
					m.showDetails = true
				}
			}
		case "k":
			if len(m.filteredEntries) > 0 {
				selectedIdx := m.table.Cursor()
				if selectedIdx >= 0 && selectedIdx < len(m.filteredEntries) {
					target := m.filteredEntries[selectedIdx]
					// scanner.KillProcess owns all privilege policy — no os.Geteuid() here.
					if err := m.scanner.KillProcess(target.PID); err != nil {
						m.status = err.Error()
						return m, nil
					}
					m.status = fmt.Sprintf("Killed %s (%s)", target.Process, target.PID)
					return m, loadPortsCmd(m.scanner)
				}
			}
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		availableHeight := m.height - 7
		m.table.SetWidth(m.width - 4)
		tableHeight := availableHeight - 2
		if tableHeight < 2 {
			tableHeight = 2
		}
		m.table.SetHeight(tableHeight)
		m.updateTableColumns()

	case []ports.PortEntry:
		m.entries = msg
		m.sortEntries()
		m.filterEntries()
		m.updateTableColumns()
		m.err = nil
		if m.status == "Refreshing..." {
			m.status = "Refreshed"
		}

	case error:
		m.err = msg
	}

	m.table, cmd = m.table.Update(msg)
	return m, cmd
}

func (m *model) sortEntries() {
	sort.Slice(m.entries, func(i, j int) bool {
		switch m.sortMode {
		case ports.SortByProcess:
			return strings.ToLower(m.entries[i].Process) < strings.ToLower(m.entries[j].Process)
		case ports.SortByPID:
			if m.entries[i].PID == "-" {
				return false
			}
			if m.entries[j].PID == "-" {
				return true
			}
			pid1, _ := strconv.Atoi(m.entries[i].PID)
			pid2, _ := strconv.Atoi(m.entries[j].PID)
			return pid1 < pid2
		default:
			p1, err1 := strconv.Atoi(m.entries[i].Port)
			p2, err2 := strconv.Atoi(m.entries[j].Port)
			if err1 == nil && err2 == nil {
				if p1 == p2 {
					return m.entries[i].Protocol < m.entries[j].Protocol
				}
				return p1 < p2
			}
			return m.entries[i].Port < m.entries[j].Port
		}
	})
	m.filterEntries()
}

func (m *model) updateTableColumns() {
	usedWidth := (8 + 6 + 12 + 8 + 22 + 6) + 14
	remainingWidth := m.width - usedWidth

	columns := []table.Column{
		{Title: "Port", Width: 8},
		{Title: "Proto", Width: 6},
		{Title: "State", Width: 12},
		{Title: "PID", Width: 8},
		{Title: "Address", Width: 22},
		{Title: "Conn", Width: 6},
		{Title: "Process", Width: remainingWidth},
	}

	arrow := " ▼"
	switch m.sortMode {
	case ports.SortByPort:
		columns[0].Title += arrow
	case ports.SortByPID:
		columns[3].Title += arrow
	case ports.SortByProcess:
		columns[5].Title += arrow
	}

	m.table.SetColumns(columns)
}

func (m *model) filterEntries() {
	query := strings.ToLower(m.textInput.Value())
	m.filteredEntries = []ports.PortEntry{}

	for _, e := range m.entries {
		if query == "" ||
			strings.Contains(strings.ToLower(e.Process), query) ||
			strings.Contains(e.Port, query) ||
			strings.Contains(e.PID, query) {
			m.filteredEntries = append(m.filteredEntries, e)
		}
	}
	m.updateTable()
}

func (m *model) updateTable() {
	rows := []table.Row{}
	for _, e := range m.filteredEntries {
		stateIcon := "○"
		if strings.Contains(e.State, "LISTEN") {
			stateIcon = "●"
		} else if strings.Contains(e.State, "ESTAB") {
			stateIcon = "↔"
		}

		// #2: local vs exposed indicator
		addrDisplay := e.Address
		if strings.HasPrefix(e.Address, "127.") || e.Address == "::1" {
			addrDisplay = "🔒 " + e.Address
		} else {
			addrDisplay = "🌐 " + e.Address
		}

		// #3: active connections count
		conn := e.Connections
		if conn == "" {
			conn = "–"
		}

		rows = append(rows, table.Row{
			e.Port,
			e.Protocol,
			stateIcon + " " + e.State,
			e.PID,
			addrDisplay,
			conn,
			e.Process,
		})
	}
	m.table.SetRows(rows)
}

func (m model) View() string {
	if m.err != nil {
		return fmt.Sprintf("Error: %v\nPress 'q' to quit", m.err)
	}

	if m.showDetails {
		content := detailsTitleStyle.Render("Connection Details") + "\n" + m.detailsContent
		content += "\n\n" + helpStyle.Render("Press Esc/Enter to close")
		box := detailsStyle.Render(content)
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, box)
	}

	logo := logoStyle.Render("⚡ LazyPorts")
	availableHeight := m.height - 7
	tableView := makeBaseStyle(m.width, availableHeight).Render(m.table.View())

	controls := "↑/↓: Navigate • /: Filter • Enter: Details • k: Kill • r: Refresh • s: Sort • q: Quit"
	if m.isFiltering {
		controls = "Type to search • Esc/Enter: Done"
		return fmt.Sprintf("%s\n%s\n%s\n%s", logo, tableView, inputStyle.Render(m.textInput.View()), helpStyle.Render(controls))
	}

	if m.status != "" {
		controls = statusStyle.Render(m.status) + " • " + controls
	}
	return fmt.Sprintf("%s\n%s\n%s", logo, tableView, helpStyle.Render(controls))
}
