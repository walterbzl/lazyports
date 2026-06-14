package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/v9mirza/lazyports/internal/ports"
	"github.com/v9mirza/lazyports/internal/ui"
)

func main() {
	scanner := &ports.SSScanner{}
	m := ui.New(scanner)

	if _, err := tea.NewProgram(m, tea.WithAltScreen()).Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}
