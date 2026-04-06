package tui

import (
	"fmt"
	"strings"

	"github.com/RockChinQ/civ-tui/game"
	"github.com/charmbracelet/lipgloss"
)

func (m Model) View() string {
	if m.Width == 0 || m.Height == 0 {
		return "Loading..."
	}

	if m.ActiveMenu == MenuHelp {
		return m.renderHelp()
	}

	header := m.renderHeader()
	mapPanel := m.renderMap()
	infoPanel := m.renderInfo()
	msgPanel := m.renderMessages()

	// Combine map + info side by side
	mapAndInfo := lipgloss.JoinHorizontal(lipgloss.Top, mapPanel, infoPanel)

	// Combine everything vertically
	full := lipgloss.JoinVertical(lipgloss.Left, header, mapAndInfo, msgPanel)

	if m.ActiveMenu == MenuBuild || m.ActiveMenu == MenuTech {
		overlay := m.renderMenu()
		full = lipgloss.JoinVertical(lipgloss.Left, full, overlay)
	}

	if m.Game.State != game.StateRunning {
		full = lipgloss.JoinVertical(lipgloss.Left, full, m.renderGameOver())
	}

	return full
}

func (m Model) renderHeader() string {
	playerCiv := m.Game.GetCiv(1)
	if playerCiv == nil {
		return StyleHeader.Render("Civ-TUI")
	}
	gold := playerCiv.Gold
	sci := playerCiv.Science
	goldPT := 0
	for _, c := range m.Game.Cities {
		if c.CivID == 1 {
			goldPT += c.GoldYield(m.Game.Map)
		}
	}
	text := fmt.Sprintf("Turn: %d  Gold: %d (+%d)  Sci: %d  %s",
		m.Game.Turn, gold, goldPT, sci, playerCiv.Name)
	return StyleHeader.Width(m.Width).Render(text)
}

func (m Model) renderMap() string {
	mapW, mapH := m.mapViewSize()
	var sb strings.Builder

	for vy := 0; vy < mapH; vy++ {
		mapY := m.ViewportY + vy
		for vx := 0; vx < mapW; vx++ {
			mapX := m.ViewportX + vx
			cell := m.renderCell(mapX, mapY)
			sb.WriteString(cell)
		}
		if vy < mapH-1 {
			sb.WriteString("\n")
		}
	}

	style := StyleMapBorder.Width(mapW * 2).Height(mapH)
	return style.Render(sb.String())
}

func (m Model) renderCell(x, y int) string {
	if !m.Game.Map.InBounds(x, y) {
		return "  "
	}

	tile := m.Game.Map.GetTile(x, y)
	isCursor := x == m.CursorX && y == m.CursorY

	if !tile.Revealed {
		if isCursor {
			return StyleCursor.Render("░ ")
		}
		return StyleFog.Render("░ ")
	}

	// Determine content
	var ch string
	var style lipgloss.Style

	// Check city first
	city := m.Game.GetCityAt(x, y)
	if city != nil {
		ch = "*"
		if city.CivID == 1 {
			style = StylePlayerCity
		} else {
			style = StyleEnemyCity
		}
	} else {
		// Check unit
		unit := m.Game.GetUnitAt(x, y)
		if unit != nil && (tile.Visible || unit.CivID == 1) {
			ch = game.UnitDefs[unit.Type].Symbol
			if unit.CivID == 1 {
				if m.SelectedUnit != nil && m.SelectedUnit.ID == unit.ID {
					style = StyleSelectedUnit
				} else {
					style = StylePlayerUnit
				}
			} else {
				style = StyleEnemyUnit
			}
		} else {
			// Terrain
			terrain := game.Terrains[tile.Terrain]
			ch = terrain.Symbol
			style = terrainStyle(tile.Terrain)
			if !tile.Visible {
				style = style.Faint(true)
			}
		}
	}

	rendered := style.Render(ch) + " "
	if isCursor {
		rendered = StyleCursor.Render(ch + " ")
	}
	return rendered
}

func terrainStyle(t game.TerrainType) lipgloss.Style {
	switch t {
	case game.TerrainOcean:
		return StyleOcean
	case game.TerrainCoast:
		return StyleCoast
	case game.TerrainGrassland:
		return StyleGrassland
	case game.TerrainPlains:
		return StylePlains
	case game.TerrainHills:
		return StyleHills
	case game.TerrainMountains:
		return StyleMountains
	case game.TerrainForest:
		return StyleForest
	case game.TerrainDesert:
		return StyleDesert
	case game.TerrainTundra:
		return StyleTundra
	}
	return StyleBase
}

func (m Model) renderInfo() string {
	var sb strings.Builder
	w := m.InfoWidth - 4

	sb.WriteString(StyleSectionTitle.Render("SELECTED UNIT") + "\n")
	if m.SelectedUnit != nil && m.SelectedUnit.IsAlive() {
		u := m.SelectedUnit
		stats := game.UnitDefs[u.Type]
		sb.WriteString(fmt.Sprintf("%s (HP %d/%d)\n", stats.Name, u.HP, u.MaxHP))
		sb.WriteString(fmt.Sprintf("Move: %d/%d  Civ: %d\n", u.MovesLeft, u.MaxMoves, u.CivID))
		sb.WriteString(fmt.Sprintf("Pos: (%d, %d)\n", u.X, u.Y))
		tile := m.Game.Map.GetTile(u.X, u.Y)
		if tile != nil {
			t := game.Terrains[tile.Terrain]
			sb.WriteString(fmt.Sprintf("Terrain: %s\n", t.Name))
			sb.WriteString(fmt.Sprintf("Yields: F%d P%d G%d\n", t.Food, t.Production, t.Gold))
		}
	} else {
		tile := m.Game.Map.GetTile(m.CursorX, m.CursorY)
		sb.WriteString(fmt.Sprintf("Cursor: (%d, %d)\n", m.CursorX, m.CursorY))
		if tile != nil && tile.Revealed {
			t := game.Terrains[tile.Terrain]
			sb.WriteString(fmt.Sprintf("Terrain: %s\n", t.Name))
			sb.WriteString(fmt.Sprintf("Yields: F%d P%d G%d\n", t.Food, t.Production, t.Gold))
		} else {
			sb.WriteString("Unexplored\n")
		}
		city := m.Game.GetCityAt(m.CursorX, m.CursorY)
		if city != nil {
			sb.WriteString(fmt.Sprintf("City: %s\n", city.Name))
			sb.WriteString(fmt.Sprintf("Pop: %d  HP: %d/%d\n", city.Population, city.HP, city.MaxHP))
		}
	}

	sb.WriteString(StyleDim.Render(strings.Repeat("─", w)) + "\n")

	sb.WriteString(StyleSectionTitle.Render("ACTIONS") + "\n")
	if m.SelectedUnit != nil && m.SelectedUnit.Type == game.UnitSettler {
		sb.WriteString("[F] Found City\n")
	}
	sb.WriteString("[W] Wait/Skip\n")
	sb.WriteString("[N] Next Unit\n")
	sb.WriteString("[B] Build Menu\n")
	sb.WriteString("[T] Tech Menu\n")
	sb.WriteString("[Enter] End Turn\n")
	sb.WriteString("[?] Help\n")

	sb.WriteString(StyleDim.Render(strings.Repeat("─", w)) + "\n")

	sb.WriteString(StyleSectionTitle.Render("RESEARCH") + "\n")
	playerCiv := m.Game.GetCiv(1)
	if playerCiv != nil {
		if playerCiv.Researching != "" {
			tech := game.GetTech(playerCiv.Researching)
			sb.WriteString(fmt.Sprintf("Researching: %s\n", playerCiv.Researching))
			if tech != nil {
				sb.WriteString(fmt.Sprintf("Progress: %d/%d\n", playerCiv.ResearchProgress, tech.Cost))
			}
		} else {
			sb.WriteString(StyleYellow.Render("Press [T] to research\n"))
		}
		done := 0
		for _, t := range game.AllTechs {
			if playerCiv.Techs[t.Name] {
				done++
			}
		}
		sb.WriteString(fmt.Sprintf("Techs: %d/%d\n", done, len(game.AllTechs)))
	}

	mapH := m.mapViewH()
	style := StyleInfoPanel.Width(m.InfoWidth).Height(mapH)
	return style.Render(sb.String())
}

func (m Model) mapViewH() int {
	_, h := m.mapViewSize()
	return h
}

func (m Model) renderMessages() string {
	var sb strings.Builder
	msgs := m.Game.Messages
	start := len(msgs) - 8
	if start < 0 {
		start = 0
	}
	msgs = msgs[start:]
	sb.WriteString(StyleSectionTitle.Render("Messages:") + "\n")
	for _, msg := range msgs {
		if len(msg) > m.Width-6 {
			msg = msg[:m.Width-6]
		}
		sb.WriteString(StyleDim.Render(msg) + "\n")
	}
	w := m.Width - 2
	if w < 10 {
		w = 10
	}
	return StyleMsgPanel.Width(w).Height(m.MsgHeight).Render(sb.String())
}

func (m Model) renderMenu() string {
	var sb strings.Builder
	switch m.ActiveMenu {
	case MenuBuild:
		city := m.Game.GetCityAt(m.CursorX, m.CursorY)
		items := buildMenuItems(city)
		sb.WriteString(StyleSectionTitle.Render("BUILD MENU") + "\n")
		for i, item := range items {
			line := fmt.Sprintf("%s (cost %d)", item.Name, item.Cost)
			if i == m.MenuCursor {
				sb.WriteString(StyleSelectedUnit.Render("> "+line) + "\n")
			} else {
				sb.WriteString("  " + line + "\n")
			}
		}
		sb.WriteString(StyleDim.Render("[Enter]=select  [Esc]=cancel"))
	case MenuTech:
		playerCiv := m.Game.GetCiv(1)
		sb.WriteString(StyleSectionTitle.Render("TECH MENU") + "\n")
		if playerCiv != nil {
			available := game.AvailableTechs(playerCiv.Techs)
			if len(available) == 0 {
				sb.WriteString("No techs available\n")
			}
			for i, t := range available {
				line := fmt.Sprintf("%s (cost %d)", t.Name, t.Cost)
				if i == m.MenuCursor {
					sb.WriteString(StyleSelectedUnit.Render("> "+line) + "\n")
				} else {
					sb.WriteString("  " + line + "\n")
				}
			}
		}
		sb.WriteString(StyleDim.Render("[Enter]=select  [Esc]=cancel"))
	}
	return StyleInfoPanel.Width(40).Render(sb.String())
}

func (m Model) renderHelp() string {
	help := `
CIV-TUI HELP
============

MOVEMENT:
  Arrow keys / hjkl - Move cursor / selected unit

UNIT ACTIONS:
  F - Found City (Settler only)
  W - Wait/Skip unit turn
  N - Select next unit with moves

CITY ACTIONS:
  B - Open build menu (when on city)

RESEARCH:
  T - Open tech research menu

TURN:
  Enter - End turn

DISPLAY:
  Cursor moves over map
  Blue units/cities = yours
  Red units/cities = enemy
  Dimmed = explored but not visible
  Dark tiles = fog of war

VICTORY:
  Domination: eliminate all enemies
  Science: research all technologies
  200 turn limit

Press ? or Esc to close help
`
	return StyleInfoPanel.Width(m.Width).Height(m.Height).Render(help)
}

func (m Model) renderGameOver() string {
	var msg string
	switch m.Game.State {
	case game.StateVictory:
		msg = StyleGreen.Render("VICTORY! Press Q to quit.")
	case game.StateDefeat:
		msg = StyleRed.Render("DEFEAT! Press Q to quit.")
	case game.StateDraw:
		msg = StyleYellow.Render("DRAW - Turn limit reached! Press Q to quit.")
	}
	return StyleBold.Render(msg)
}
