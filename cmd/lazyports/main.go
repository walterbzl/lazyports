package main

import (
	"fmt"
	"log"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/v9mirza/lazyports/internal/config"
	"github.com/v9mirza/lazyports/internal/ports"
	"github.com/v9mirza/lazyports/internal/ui"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Printf("warning: config parse error (%v), using defaults", err)
	}

	scanner := &ports.SSScanner{}
	m := ui.New(scanner, cfg)

	if _, err := tea.NewProgram(m, tea.WithAltScreen()).Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}
