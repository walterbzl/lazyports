package ui

import (
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/lipgloss"
)

// Styles holds every lipgloss.Style the UI renders with, built from a Theme.
// Building once (in New) keeps View() allocation-free and theme-driven.
type Styles struct {
	theme Theme

	status       lipgloss.Style
	help         lipgloss.Style
	input        lipgloss.Style
	details      lipgloss.Style
	detailsTitle lipgloss.Style
	logo         lipgloss.Style
	panelSection lipgloss.Style
	panelValue   lipgloss.Style
	cpuBar       lipgloss.Style
	memBar       lipgloss.Style
	emptyBar     lipgloss.Style
}

// newStyles builds the full style set from the given theme.
func newStyles(t Theme) Styles {
	return Styles{
		theme:  t,
		status: lipgloss.NewStyle().Foreground(t.Status),
		help:   lipgloss.NewStyle().Foreground(t.Help).MarginTop(1),
		input: lipgloss.NewStyle().
			Foreground(t.InputText).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(t.Accent).
			Padding(0, 1).
			Width(60),
		details: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(t.Accent).
			Padding(1, 2).
			Width(60),
		detailsTitle: lipgloss.NewStyle().
			Foreground(t.Accent).
			Bold(true).
			MarginBottom(1),
		logo: lipgloss.NewStyle().
			Foreground(t.Accent).
			Bold(true).
			MarginBottom(1),
		panelSection: lipgloss.NewStyle().Foreground(t.Accent).Bold(true),
		panelValue:   lipgloss.NewStyle().Foreground(t.TextPrimary),
		cpuBar:       lipgloss.NewStyle().Foreground(t.CPUBar),
		memBar:       lipgloss.NewStyle().Foreground(t.MemBar),
		emptyBar:     lipgloss.NewStyle().Foreground(t.EmptyBar),
	}
}

// base returns the bordered container style for the table region.
func (s Styles) base(w, h int) lipgloss.Style {
	return lipgloss.NewStyle().
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(s.theme.Border).
		Width(w - 2).
		Height(h)
}

// panel returns the bordered container style for the side detail panel.
func (s Styles) panel(w, h int) lipgloss.Style {
	return lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(s.theme.Accent).
		Padding(0, 1).
		Width(w - 2).
		Height(h)
}

// tableStyles builds the bubbles/table style set from the theme.
func (s Styles) tableStyles() table.Styles {
	ts := table.DefaultStyles()
	ts.Header = ts.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(s.theme.Border).
		BorderBottom(true).
		Bold(true)
	ts.Selected = ts.Selected.
		Foreground(s.theme.SelectedFg).
		Background(s.theme.SelectedBg).
		Bold(false)
	return ts
}
