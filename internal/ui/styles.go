package ui

import (
	"github.com/charmbracelet/lipgloss"
)

var (
	// Status/Footer styles
	statusStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#f5c2e7"))              // Pink
	helpStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("#6c7086")).MarginTop(1) // Overlay0
	inputStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("#a6adc8")).Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("#cba6f7")).Padding(0, 1).Width(60)

	// Details View Styles
	detailsStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#cba6f7")).
			Padding(1, 2).
			Width(60)
	detailsTitleStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#cba6f7")).
				Bold(true).
				MarginBottom(1)

	// Logo Style
	logoStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#cba6f7")). // Mauve
			Bold(true).
			MarginBottom(1)
)

// makeBaseStyle returns a new lipgloss.Style configured with a normal border
// and the given width and height. It is a pure function — no package-level
// state is mutated when dimensions change.
func makeBaseStyle(w, h int) lipgloss.Style {
	return lipgloss.NewStyle().
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("#6c7086")). // Overlay0
		Width(w - 2).
		Height(h)
}

// makePanelStyle returns the style for the side detail panel.
func makePanelStyle(w, h int) lipgloss.Style {
	return lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#cba6f7")). // Mauve
		Padding(0, 1).
		Width(w - 2).
		Height(h)
}

// panelSectionStyle renders a section header inside the panel.
var panelSectionStyle = lipgloss.NewStyle().
	Foreground(lipgloss.Color("#cba6f7")). // Mauve
	Bold(true)

// panelValueStyle renders values inside the panel.
var panelValueStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#cdd6f4")) // Text

// cpuBarStyle for the CPU fill.
var cpuBarStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#f38ba8")) // Red

// memBarStyle for the MEM fill.
var memBarStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#fab387")) // Peach

// emptyBarStyle for the empty portion of bars.
var emptyBarStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#45475a")) // Surface1
