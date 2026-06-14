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
