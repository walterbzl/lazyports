package ui

import "github.com/charmbracelet/lipgloss"

// Theme holds every semantic color token the UI needs. Concrete themes fill
// these roles; styles are built from the active Theme rather than hardcoded.
type Theme struct {
	Border      lipgloss.Color // table / panel borders, separators
	Accent      lipgloss.Color // primary accent: logo, titles, active borders
	Status      lipgloss.Color // status / toast messages
	Help        lipgloss.Color // footer help text
	InputText   lipgloss.Color // text inside input fields
	TextPrimary lipgloss.Color // panel values, main text
	SelectedFg  lipgloss.Color // selected table row foreground
	SelectedBg  lipgloss.Color // selected table row background
	CPUBar      lipgloss.Color // CPU usage bar fill
	MemBar      lipgloss.Color // memory usage bar fill
	EmptyBar    lipgloss.Color // empty portion of resource bars
}

// catppuccinTheme is the original Catppuccin Mocha palette.
func catppuccinTheme() Theme {
	return Theme{
		Border:      lipgloss.Color("#6c7086"),
		Accent:      lipgloss.Color("#cba6f7"),
		Status:      lipgloss.Color("#f5c2e7"),
		Help:        lipgloss.Color("#6c7086"),
		InputText:   lipgloss.Color("#a6adc8"),
		TextPrimary: lipgloss.Color("#cdd6f4"),
		SelectedFg:  lipgloss.Color("#cdd6f4"),
		SelectedBg:  lipgloss.Color("#313244"),
		CPUBar:      lipgloss.Color("#f38ba8"),
		MemBar:      lipgloss.Color("#fab387"),
		EmptyBar:    lipgloss.Color("#45475a"),
	}
}

// cherryTheme is the Cherry Red palette (PLAYBOOK.md section 7).
func cherryTheme() Theme {
	return Theme{
		Border:      lipgloss.Color("#3d1f26"),
		Accent:      lipgloss.Color("#ff4d6a"),
		Status:      lipgloss.Color("#ff8599"),
		Help:        lipgloss.Color("#8a5a62"),
		InputText:   lipgloss.Color("#c49299"),
		TextPrimary: lipgloss.Color("#e8d0d4"),
		SelectedFg:  lipgloss.Color("#e8d0d4"),
		SelectedBg:  lipgloss.Color("#3d1520"),
		CPUBar:      lipgloss.Color("#ff4d6a"),
		MemBar:      lipgloss.Color("#ff8599"),
		EmptyBar:    lipgloss.Color("#2d1218"),
	}
}

// themeByName resolves a config theme name to a Theme, defaulting to Catppuccin.
func themeByName(name string) Theme {
	switch name {
	case "cherry":
		return cherryTheme()
	default:
		return catppuccinTheme()
	}
}
