package main

import (
	"fmt"
	"log"
	"os"
	"runtime"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/walterbzl/lazyports/internal/config"
	"github.com/walterbzl/lazyports/internal/ports"
	"github.com/walterbzl/lazyports/internal/ui"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Printf("warning: config parse error (%v), using defaults", err)
	}

	var scanner ports.Scanner
	switch runtime.GOOS {
	case "windows":
		scanner = &ports.NetstatScanner{}
	case "darwin":
		scanner = &ports.LsofScanner{}
	default:
		scanner = &ports.SSScanner{}
	}

	m := ui.New(scanner, cfg)

	if _, err := tea.NewProgram(m, tea.WithAltScreen()).Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}
