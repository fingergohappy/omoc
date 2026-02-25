package main

import (
	"fmt"
	"os"

	"omoc/internal/config"
	"omoc/internal/models"
	"omoc/internal/tui"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load config: %v\n", err)
		os.Exit(1)
	}

	available, err := models.Fetch()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to fetch models (is opencode installed?): %v\n", err)
		os.Exit(1)
	}

	m := tui.New(cfg, available)
	p := tea.NewProgram(m, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
