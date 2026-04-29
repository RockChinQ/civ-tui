package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/RockChinQ/civ-tui/game"
	"github.com/RockChinQ/civ-tui/i18n"
	"github.com/RockChinQ/civ-tui/tui"
	"github.com/RockChinQ/civ-tui/web"
	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	webMode := flag.Bool("web", false, "run web UI instead of TUI")
	addr := flag.String("addr", "127.0.0.1:8080", "web server bind address (with -web)")
	flag.Parse()

	i18n.LoadConfig()
	game.MigrateLegacySave()

	if *webMode {
		s := web.NewServer()
		if err := s.Run(*addr); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		return
	}

	m := tui.NewModel()
	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
