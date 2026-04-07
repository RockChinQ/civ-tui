package main

import (
	"fmt"
	"os"

	"github.com/RockChinQ/civ-tui/game"
	"github.com/RockChinQ/civ-tui/i18n"
	"github.com/RockChinQ/civ-tui/tui"
	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	i18n.LoadConfig()
	game.MigrateLegacySave()
	m := tui.NewModel()
	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
