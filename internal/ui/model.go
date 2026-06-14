package ui

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/atotto/clipboard"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/v9mirza/lazyports/internal/labels"
	"github.com/v9mirza/lazyports/internal/ports"
)

const autoRefreshInterval = 3 * time.Second

type tickMsg time.Time

// panelMsg carries async-loaded detail + resource info for the side panel.
type panelMsg struct {
	pid      string
	details  string
	resource ports.ResourceInfo
}

type model struct {
	scanner          ports.Scanner
	table            table.Model
	textInput        textinput.Model
	labelInput       textinput.Model
	labelStore       *labels.Store
	entries          []ports.PortEntry
	filteredEntries  []ports.PortEntry
	err              error
	status           string
	width            int
	height           int
	isFiltering      bool
	isLabeling       bool
	showDetails      bool
	detailsContent   string
	sortMode         ports.SortMode
	pendingForceKill string
	autoRefresh      bool
	// side panel
	showSidePanel bool
	panelPID      string
	panelDetails  string
	panelResource ports.ResourceInfo
	panelLoading  bool
}

func New(scanner ports.Scanner) model {
	columns := []table.Column{
		{Title: "Port", Width: 8},
		{Title: "Proto", Width: 6},
		{Title: "State", Width: 12},
		{Title: "PID", Width: 8},
		{Title: "Address", Width: 22},
		{Title: "Conn", Width: 6},
		{Title: "Label", Width: 8},
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

	search := textinput.New()
	search.Placeholder = "Search ports, processes, pids..."
	search.CharLimit = 156
	search.Width = 40

	li := textinput.New()
	li.Placeholder = "Label (empty to clear)..."
	li.CharLimit = 32
	li.Width = 30

	ls, _ := labels.Load()

	return model{
		scanner:    scanner,
		table:      t,
		textInput:  search,
		labelInput: li,
		labelStore: ls,
	}
}

func loadPortsCmd(s ports.Scanner) tea.Cmd {
	return func() tea.Msg {
		entries, err := s.GetPorts()
		if err != nil {
			return err
		}
		return entries
	}
}

func tickCmd() tea.Cmd {
	return tea.Tick(autoRefreshInterval, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func loadPanelCmd(s ports.Scanner, e ports.PortEntry) tea.Cmd {
	pid := e.PID
	return func() tea.Msg {
		details, _ := s.GetProcessDetails(pid)
		res, _ := s.GetResourceInfo(pid)
		return panelMsg{pid: pid, details: details, resource: res}
	}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(loadPortsCmd(m.scanner), textinput.Blink, tickCmd())
}

func (m model) selectedEntry() (ports.PortEntry, bool) {
	idx := m.table.Cursor()
	if idx >= 0 && idx < len(m.filteredEntries) {
		return m.filteredEntries[idx], true
	}
	return ports.PortEntry{}, false
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	// ── Force kill confirmation overlay ──────────────────────────────────────
	if m.pendingForceKill != "" {
		if key, ok := msg.(tea.KeyMsg); ok {
			switch key.String() {
			case "K", "enter":
				pid := m.pendingForceKill
				m.pendingForceKill = ""
				if err := m.scanner.ForceKillProcess(pid); err != nil {
					m.status = err.Error()
					return m, nil
				}
				m.status = fmt.Sprintf("Force-killed PID %s", pid)
				return m, loadPortsCmd(m.scanner)
			case "esc", "q":
				m.pendingForceKill = ""
				m.status = "Force kill cancelled"
				return m, nil
			}
		}
		return m, nil
	}

	// ── Label input overlay ──────────────────────────────────────────────────
	if m.isLabeling {
		if key, ok := msg.(tea.KeyMsg); ok {
			switch key.String() {
			case "enter":
				if e, ok := m.selectedEntry(); ok {
					_ = m.labelStore.Set(e.Port, strings.TrimSpace(m.labelInput.Value()))
				}
				m.isLabeling = false
				m.updateTable()
				return m, nil
			case "esc":
				m.isLabeling = false
				return m, nil
			}
		}
		m.labelInput, cmd = m.labelInput.Update(msg)
		return m, cmd
	}

	// ── Search filter ────────────────────────────────────────────────────────
	if m.isFiltering {
		if key, ok := msg.(tea.KeyMsg); ok {
			if key.String() == "enter" || key.String() == "esc" {
				m.isFiltering = false
				m.table.Focus()
				return m, nil
			}
		}
		m.textInput, cmd = m.textInput.Update(msg)
		m.filterEntries()
		return m, cmd
	}

	// ── Details modal ────────────────────────────────────────────────────────
	if m.showDetails {
		if key, ok := msg.(tea.KeyMsg); ok {
			if key.String() == "esc" || key.String() == "q" || key.String() == "enter" {
				m.showDetails = false
				return m, nil
			}
		}
		return m, nil
	}

	// ── Normal mode ──────────────────────────────────────────────────────────
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

		case "R":
			m.autoRefresh = !m.autoRefresh
			if m.autoRefresh {
				m.status = "Auto-refresh ON (3s)"
			} else {
				m.status = "Auto-refresh OFF"
			}

		case "tab":
			m.showSidePanel = !m.showSidePanel
			m.updateTableColumns()
			if m.showSidePanel {
				if e, ok := m.selectedEntry(); ok {
					m.panelLoading = true
					return m, loadPanelCmd(m.scanner, e)
				}
			}

		case "s":
			m.sortMode++
			if m.sortMode > ports.SortByPID {
				m.sortMode = ports.SortByPort
			}
			m.sortEntries()
			m.updateTableColumns()
			m.updateTable()

		case "enter":
			if e, ok := m.selectedEntry(); ok {
				details, err := m.scanner.GetProcessDetails(e.PID)
				if err != nil {
					m.detailsContent = fmt.Sprintf("Error: %v", err)
				} else if e.PID == "-" {
					m.detailsContent = details
				} else {
					m.detailsContent = fmt.Sprintf(
						"Port:      %s/%s\nPID:       %s\nAddress:   %s\nState:     %s\nProcess:   %s\n\n%s",
						e.Port, e.Protocol, e.PID, e.Address, e.State, e.Process, details,
					)
				}
				m.showDetails = true
			}

		case "k":
			if e, ok := m.selectedEntry(); ok {
				if err := m.scanner.KillProcess(e.PID); err != nil {
					m.status = err.Error()
					return m, nil
				}
				m.status = fmt.Sprintf("Sent SIGTERM to %s (%s)", e.Process, e.PID)
				return m, loadPortsCmd(m.scanner)
			}

		case "K":
			if e, ok := m.selectedEntry(); ok {
				m.pendingForceKill = e.PID
				m.status = fmt.Sprintf("Force kill PID %s? Press K/Enter to confirm, Esc to cancel", e.PID)
			}

		case "y":
			if e, ok := m.selectedEntry(); ok {
				if err := clipboard.WriteAll(e.Port); err != nil {
					m.status = "Clipboard unavailable: " + err.Error()
				} else {
					m.status = fmt.Sprintf("Port %s copied to clipboard", e.Port)
				}
			}

		case "o":
			if e, ok := m.selectedEntry(); ok {
				if err := m.scanner.OpenInBrowser(e.Port); err != nil {
					m.status = "Cannot open browser: " + err.Error()
				} else {
					m.status = fmt.Sprintf("Opening http://localhost:%s...", e.Port)
				}
			}

		case "l":
			if e, ok := m.selectedEntry(); ok {
				m.isLabeling = true
				m.labelInput.SetValue(m.labelStore.Get(e.Port))
				m.labelInput.Focus()
				return m, textinput.Blink
			}
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		availableHeight := m.height - 7
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
		if m.showSidePanel {
			if e, ok := m.selectedEntry(); ok {
				m.panelLoading = true
				return m, loadPanelCmd(m.scanner, e)
			}
		}

	case panelMsg:
		m.panelPID = msg.pid
		m.panelDetails = msg.details
		m.panelResource = msg.resource
		m.panelLoading = false

	case tickMsg:
		if m.autoRefresh {
			return m, tea.Batch(loadPortsCmd(m.scanner), tickCmd())
		}
		return m, tickCmd()

	case error:
		m.err = msg
	}

	prevCursor := m.table.Cursor()
	m.table, cmd = m.table.Update(msg)
	if m.showSidePanel && m.table.Cursor() != prevCursor {
		if e, ok := m.selectedEntry(); ok {
			m.panelLoading = true
			cmd = tea.Batch(cmd, loadPanelCmd(m.scanner, e))
		}
	}
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
			a, _ := strconv.Atoi(m.entries[i].PID)
			b, _ := strconv.Atoi(m.entries[j].PID)
			return a < b
		default:
			a, ea := strconv.Atoi(m.entries[i].Port)
			b, eb := strconv.Atoi(m.entries[j].Port)
			if ea == nil && eb == nil {
				if a == b {
					return m.entries[i].Protocol < m.entries[j].Protocol
				}
				return a < b
			}
			return m.entries[i].Port < m.entries[j].Port
		}
	})
	m.filterEntries()
}

func (m *model) tableWidth() int {
	if m.showSidePanel && m.width >= 80 {
		return m.width * 3 / 5
	}
	return m.width
}

func (m *model) panelWidth() int {
	return m.width - m.tableWidth()
}

func (m *model) updateTableColumns() {
	tw := m.tableWidth()
	m.table.SetWidth(tw - 4)
	fixed := 8 + 6 + 12 + 8 + 22 + 6 + 8 + 14
	procWidth := tw - fixed
	if procWidth < 8 {
		procWidth = 8
	}

	columns := []table.Column{
		{Title: "Port", Width: 8},
		{Title: "Proto", Width: 6},
		{Title: "State", Width: 12},
		{Title: "PID", Width: 8},
		{Title: "Address", Width: 22},
		{Title: "Conn", Width: 6},
		{Title: "Label", Width: 8},
		{Title: "Process", Width: procWidth},
	}

	arrow := " ▼"
	switch m.sortMode {
	case ports.SortByPort:
		columns[0].Title += arrow
	case ports.SortByPID:
		columns[3].Title += arrow
	case ports.SortByProcess:
		columns[7].Title += arrow
	}

	m.table.SetColumns(columns)
}

func (m *model) filterEntries() {
	query := strings.ToLower(m.textInput.Value())
	m.filteredEntries = m.filteredEntries[:0]

	for _, e := range m.entries {
		lbl := strings.ToLower(m.labelStore.Get(e.Port))
		if query == "" ||
			strings.Contains(strings.ToLower(e.Process), query) ||
			strings.Contains(e.Port, query) ||
			strings.Contains(e.PID, query) ||
			strings.Contains(lbl, query) {
			m.filteredEntries = append(m.filteredEntries, e)
		}
	}
	m.updateTable()
}

func (m *model) updateTable() {
	rows := make([]table.Row, 0, len(m.filteredEntries))
	for _, e := range m.filteredEntries {
		stateIcon := "○"
		if strings.Contains(e.State, "LISTEN") {
			stateIcon = "●"
		} else if strings.Contains(e.State, "ESTAB") {
			stateIcon = "↔"
		}

		addrDisplay := "🌐 " + e.Address
		if strings.HasPrefix(e.Address, "127.") || e.Address == "::1" {
			addrDisplay = "🔒 " + e.Address
		}

		conn := e.Connections
		if conn == "" {
			conn = "–"
		}

		lbl := m.labelStore.Get(e.Port)

		rows = append(rows, table.Row{
			e.Port,
			e.Protocol,
			stateIcon + " " + e.State,
			e.PID,
			addrDisplay,
			conn,
			lbl,
			e.Process,
		})
	}
	m.table.SetRows(rows)
}

// renderBar renders a unicode progress bar for val/max ratio.
func renderBar(val, max float64, width int, fillStyle, emptyStyle lipgloss.Style) string {
	if max <= 0 || width <= 0 {
		return strings.Repeat("░", width)
	}
	ratio := val / max
	if ratio > 1 {
		ratio = 1
	}
	filled := int(ratio * float64(width))
	empty := width - filled
	return fillStyle.Render(strings.Repeat("█", filled)) + emptyStyle.Render(strings.Repeat("░", empty))
}

// renderSidePanel builds the panel content from the current entry and cached panel state.
func (m model) renderSidePanel(e ports.PortEntry, innerWidth int) string {
	var b strings.Builder
	sep := strings.Repeat("─", innerWidth)

	b.WriteString(panelSectionStyle.Render("── CONEXIÓN ──") + "\n")
	b.WriteString(fmt.Sprintf("Puerto:  %s/%s\n", e.Port, e.Protocol))
	b.WriteString(fmt.Sprintf("Estado:  %s\n", e.State))
	addrIcon := "🌐"
	if strings.HasPrefix(e.Address, "127.") || e.Address == "::1" {
		addrIcon = "🔒"
	}
	b.WriteString(fmt.Sprintf("Addr:    %s %s\n", addrIcon, e.Address))
	b.WriteString(panelValueStyle.Render(sep) + "\n")

	b.WriteString(panelSectionStyle.Render("── PROCESO ──") + "\n")
	b.WriteString(fmt.Sprintf("PID:     %s\n", e.PID))
	b.WriteString(fmt.Sprintf("Nombre:  %s\n", e.Process))
	if lbl := m.labelStore.Get(e.Port); lbl != "" {
		b.WriteString(fmt.Sprintf("Etiqueta: %s\n", lbl))
	}
	b.WriteString(panelValueStyle.Render(sep) + "\n")

	b.WriteString(panelSectionStyle.Render("── RED ──") + "\n")
	exposure := "⚠ EXPUESTO"
	if strings.HasPrefix(e.Address, "127.") || e.Address == "::1" {
		exposure = "✓ LOCAL"
	}
	b.WriteString(fmt.Sprintf("Exposición:  %s\n", exposure))
	conn := e.Connections
	if conn == "" || conn == "–" {
		conn = "0"
	}
	b.WriteString(fmt.Sprintf("Conn activas: %s\n", conn))
	b.WriteString(panelValueStyle.Render(sep) + "\n")

	b.WriteString(panelSectionStyle.Render("── RECURSOS ──") + "\n")
	if e.PID == "-" {
		b.WriteString("CPU  –\nMEM  –\n")
	} else if m.panelLoading && m.panelPID != e.PID {
		b.WriteString("Cargando...\n")
	} else if m.panelPID == e.PID {
		barW := innerWidth - 10
		if barW < 4 {
			barW = 4
		}
		cpuBar := renderBar(m.panelResource.CPUPercent, 100, barW, cpuBarStyle, emptyBarStyle)
		memBar := renderBar(m.panelResource.MemMB, 2048, barW, memBarStyle, emptyBarStyle)
		b.WriteString(fmt.Sprintf("CPU  %s %.1f%%\n", cpuBar, m.panelResource.CPUPercent))
		b.WriteString(fmt.Sprintf("MEM  %s %.0fMB\n", memBar, m.panelResource.MemMB))
	} else {
		b.WriteString("CPU  –\nMEM  –\n")
	}
	b.WriteString(panelValueStyle.Render(sep) + "\n")

	if m.panelPID == e.PID && m.panelDetails != "" {
		b.WriteString(panelSectionStyle.Render("── DETALLES ──") + "\n")
		b.WriteString(panelValueStyle.Render(m.panelDetails) + "\n")
		b.WriteString(panelValueStyle.Render(sep) + "\n")
	}

	b.WriteString(panelSectionStyle.Render("── ACCIONES ──") + "\n")
	b.WriteString(helpStyle.Render("[k]Matar [K]Forzar [o]Abrir [y]Copiar [l]Etiqueta"))

	return b.String()
}

func (m model) View() string {
	if m.err != nil {
		return fmt.Sprintf("Error: %v\nPress 'q' to quit", m.err)
	}

	logo := logoStyle.Render("⚡ LazyPorts")
	availableHeight := m.height - 7

	// ── Force kill confirmation ──────────────────────────────────────────────
	if m.pendingForceKill != "" {
		prompt := detailsStyle.Render(
			detailsTitleStyle.Render("⚠ Force Kill") + "\n" +
				fmt.Sprintf("Send SIGKILL to PID %s?\n\n", m.pendingForceKill) +
				helpStyle.Render("K / Enter: confirm  •  Esc: cancel"),
		)
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, prompt)
	}

	// ── Details modal ────────────────────────────────────────────────────────
	if m.showDetails {
		content := detailsTitleStyle.Render("Connection Details") + "\n" + m.detailsContent
		content += "\n\n" + helpStyle.Render("Press Esc/Enter to close")
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, detailsStyle.Render(content))
	}

	// ── Label input ──────────────────────────────────────────────────────────
	if m.isLabeling {
		if e, ok := m.selectedEntry(); ok {
			prompt := detailsStyle.Render(
				detailsTitleStyle.Render(fmt.Sprintf("Label port %s", e.Port)) + "\n" +
					inputStyle.Render(m.labelInput.View()) + "\n" +
					helpStyle.Render("Enter: save  •  Esc: cancel"),
			)
			return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, prompt)
		}
	}

	// ── Search filter ────────────────────────────────────────────────────────
	if m.isFiltering {
		tw := m.tableWidth()
		tableView := makeBaseStyle(tw, availableHeight).Render(m.table.View())
		return fmt.Sprintf("%s\n%s\n%s\n%s",
			logo, tableView,
			inputStyle.Render(m.textInput.View()),
			helpStyle.Render("Type to search  •  Esc/Enter: done"),
		)
	}

	// ── Normal / split view ──────────────────────────────────────────────────
	tw := m.tableWidth()
	tableView := makeBaseStyle(tw, availableHeight).Render(m.table.View())

	var mainContent string
	if m.showSidePanel && m.width >= 80 {
		pw := m.panelWidth()
		innerW := pw - 6
		if innerW < 10 {
			innerW = 10
		}
		if e, ok := m.selectedEntry(); ok {
			panelContent := m.renderSidePanel(e, innerW)
			panelView := makePanelStyle(pw, availableHeight).Render(panelContent)
			mainContent = lipgloss.JoinHorizontal(lipgloss.Top, tableView, panelView)
		} else {
			mainContent = tableView
		}
	} else {
		mainContent = tableView
	}

	refresh := ""
	if m.autoRefresh {
		refresh = " • auto:3s"
	}
	panelHint := "Tab:Panel"
	if m.showSidePanel {
		panelHint = "Tab:Panel↑"
	}
	controls := fmt.Sprintf("↑/↓ Nav • / Filter • Enter Details • k Kill • K Force • y Copy • o Open • l Label • R Auto%s • %s • s Sort • q Quit", refresh, panelHint)
	if m.status != "" {
		controls = statusStyle.Render(m.status) + " • " + controls
	}
	return fmt.Sprintf("%s\n%s\n%s", logo, mainContent, helpStyle.Render(controls))
}
