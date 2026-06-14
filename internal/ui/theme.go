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

// tokyoNightTheme is the Tokyo Night (Night variant) palette.
func tokyoNightTheme() Theme {
	return Theme{
		Border:      lipgloss.Color("#3b4261"),
		Accent:      lipgloss.Color("#7aa2f7"),
		Status:      lipgloss.Color("#bb9af7"),
		Help:        lipgloss.Color("#565f89"),
		InputText:   lipgloss.Color("#a9b1d6"),
		TextPrimary: lipgloss.Color("#c0caf5"),
		SelectedFg:  lipgloss.Color("#c0caf5"),
		SelectedBg:  lipgloss.Color("#283457"),
		CPUBar:      lipgloss.Color("#f7768e"),
		MemBar:      lipgloss.Color("#e0af68"),
		EmptyBar:    lipgloss.Color("#292e42"),
	}
}

// gruvboxTheme is the Gruvbox Dark palette.
func gruvboxTheme() Theme {
	return Theme{
		Border:      lipgloss.Color("#504945"),
		Accent:      lipgloss.Color("#fabd2f"),
		Status:      lipgloss.Color("#d3869b"),
		Help:        lipgloss.Color("#928374"),
		InputText:   lipgloss.Color("#d5c4a1"),
		TextPrimary: lipgloss.Color("#ebdbb2"),
		SelectedFg:  lipgloss.Color("#fbf1c7"),
		SelectedBg:  lipgloss.Color("#3c3836"),
		CPUBar:      lipgloss.Color("#fb4934"),
		MemBar:      lipgloss.Color("#fe8019"),
		EmptyBar:    lipgloss.Color("#3c3836"),
	}
}

// themeByName resolves a config theme name to a Theme, defaulting to Catppuccin.
func themeByName(name string) Theme {
	switch name {
	case "cherry":
		return cherryTheme()
	case "tokyo-night":
		return tokyoNightTheme()
	case "gruvbox":
		return gruvboxTheme()
	default:
		return catppuccinTheme()
	}
}
