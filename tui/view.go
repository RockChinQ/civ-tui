package tui

import (
	"fmt"
	"strings"

	"github.com/RockChinQ/civ-tui/game"
	"github.com/RockChinQ/civ-tui/game/model"
	"github.com/RockChinQ/civ-tui/game/worldmap"
	"github.com/RockChinQ/civ-tui/i18n"
	"github.com/charmbracelet/lipgloss"
)

func (m Model) View() string {
	if m.Width == 0 || m.Height == 0 {
		return i18n.T("Loading...")
	}

	if m.CurrentScreen == ScreenMainMenu {
		if m.InSettings {
			return m.renderSettings()
		}
		if m.InNewGame {
			return m.renderNewGame()
		}
		return m.renderMainMenu()
	}

	if m.Game == nil {
		return i18n.T("Loading game...")
	}

	if m.ActiveMenu == MenuHelp {
		return m.renderHelp()
	}

	// Popup menus
	switch m.ActiveMenu {
	case MenuBuild, MenuTech:
		return m.renderAsPopup(m.renderMenu(), 52)
	case MenuDiplomacy:
		return m.renderAsPopup(m.renderDiplomacy(), 52)
	case MenuCity:
		city := m.Game.GetCityAt(m.CursorX, m.CursorY)
		if city != nil {
			return m.renderAsPopup(m.renderCityDetails(city), 52)
		}
	case MenuInspect:
		return m.renderAsPopup(m.renderInspect(), 52)
	}

	header := m.renderHeader()
	mapPanel := m.renderMap()
	infoPanel := m.renderInfo()
	msgPanel := m.renderMessages()

	mapAndInfo := lipgloss.JoinHorizontal(lipgloss.Top, mapPanel, infoPanel)
	full := lipgloss.JoinVertical(lipgloss.Left, header, mapAndInfo, msgPanel)

	if m.Game.State != game.StateRunning {
		full = lipgloss.JoinVertical(lipgloss.Left, full, m.renderGameOver())
	}

	return full
}

func (m Model) renderMainMenu() string {
	var sb strings.Builder
	title := "  CIV-TUI\n  " + i18n.T("A Terminal Civilization Game") + "\n"
	sb.WriteString(StyleBold.Render(title) + "\n")

	items := []string{i18n.T("New Game"), i18n.T("Load Game"), i18n.T("Settings"), i18n.T("Quit")}
	for i, item := range items {
		if i == m.MainMenuCursor {
			sb.WriteString(StyleSelectedUnit.Render("> "+item) + "\n")
		} else {
			sb.WriteString("  " + item + "\n")
		}
	}
	sb.WriteString("\n" + StyleDim.Render(i18n.T("[↑/↓] Navigate  [Enter] Select  [Q] Quit")))
	return StyleInfoPanel.Width(m.Width - 4).Height(m.Height - 2).Render(sb.String())
}

func (m Model) renderSettings() string {
	var sb strings.Builder
	sb.WriteString(StyleSectionTitle.Render(i18n.T("SETTINGS")) + "\n\n")

	items := []string{
		i18n.Tf("Language: %s", i18n.LangName(i18n.GetLang())),
		i18n.T("Back"),
	}

	for i, item := range items {
		if i == m.SettingsCursor {
			sb.WriteString(StyleSelectedUnit.Render("> "+item) + "\n")
		} else {
			sb.WriteString("  " + item + "\n")
		}
	}
	sb.WriteString("\n" + StyleDim.Render(i18n.T("[←/→] Change value  [Enter/Esc] Back")))
	return StyleInfoPanel.Width(m.Width - 4).Height(m.Height - 2).Render(sb.String())
}

func (m Model) renderNewGame() string {
	var sb strings.Builder
	sb.WriteString(StyleSectionTitle.Render(i18n.T("NEW GAME")) + "\n\n")

	mapSizes := []string{i18n.T("Small"), i18n.T("Medium"), i18n.T("Large")}
	mapSizeStr := mapSizes[int(m.SettingsMapSize)]

	items := []string{
		i18n.Tf("Map Size: %s", mapSizeStr),
		i18n.Tf("AI Civs: %d", m.SettingsNumAICivs),
		i18n.Tf("Difficulty: %s", i18n.T([]string{"Easy", "Normal", "Hard"}[m.SettingsDifficulty-1])),
		i18n.T("Start Game"),
		i18n.T("Back"),
	}

	for i, item := range items {
		if i == m.NewGameCursor {
			sb.WriteString(StyleSelectedUnit.Render("> "+item) + "\n")
		} else {
			sb.WriteString("  " + item + "\n")
		}
	}
	sb.WriteString("\n" + StyleDim.Render(i18n.T("[←/→] Change value  [Enter] Select  [Esc] Back")))
	return StyleInfoPanel.Width(m.Width - 4).Height(m.Height - 2).Render(sb.String())
}

func (m Model) renderHeader() string {
	if m.Game == nil {
		return StyleHeader.Render("Civ-TUI")
	}
	playerCiv := m.Game.GetCiv(1)
	if playerCiv == nil {
		return StyleHeader.Width(m.Width).Render("Civ-TUI")
	}
	gold := playerCiv.Gold
	sci := playerCiv.Science
	goldPT := 0
	for _, c := range m.Game.Cities {
		if c.CivID == 1 {
			goldPT += c.GoldYield(m.Game.Map.GetTile(c.X, c.Y))
		}
	}
	text := i18n.Tf("Turn: %d  Gold: %d (+%d)  Sci: %d  %s",
		m.Game.Turn, gold, goldPT, sci, i18n.T(playerCiv.Name))
	if m.RangeMode {
		text += i18n.T("  [RANGED MODE - Enter to fire, Esc to cancel]")
	} else if m.DestMode {
		text += i18n.T("  [GOTO MODE - Move cursor to destination, Enter to confirm, Esc to cancel]")
	}
	return StyleHeader.Width(m.Width).Render(text)
}

func (m Model) renderMap() string {
	mapW, mapH := m.mapViewSize()
	var sb strings.Builder

	// X-axis coordinate labels (top row)
	sb.WriteString(strings.Repeat(" ", mapLabelW))
	for vx := 0; vx < mapW; vx++ {
		mapX := m.ViewportX + vx
		if mapX%mapLabelStep == 0 {
			label := fmt.Sprintf("%-2d", mapX)
			sb.WriteString(StyleDim.Render(label))
		} else {
			sb.WriteString("  ")
		}
	}
	sb.WriteString("\n")

	// Map rows with Y-axis labels
	for vy := 0; vy < mapH; vy++ {
		mapY := m.ViewportY + vy
		if mapY%mapLabelStep == 0 {
			sb.WriteString(StyleDim.Render(fmt.Sprintf("%2d ", mapY)))
		} else {
			sb.WriteString(strings.Repeat(" ", mapLabelW))
		}
		for vx := 0; vx < mapW; vx++ {
			mapX := m.ViewportX + vx
			cell := m.renderCell(mapX, mapY)
			sb.WriteString(cell)
		}
		if vy < mapH-1 {
			sb.WriteString("\n")
		}
	}

	style := StyleMapBorder.Width(mapLabelW + mapW*2).Height(mapH + mapLabelH)
	return style.Render(sb.String())
}

func (m Model) renderCell(x, y int) string {
	if m.Game == nil || !m.Game.Map.InBounds(x, y) {
		return "  "
	}

	tile := m.Game.Map.GetTile(x, y)
	isCursor := x == m.CursorX && y == m.CursorY

	// Highlight range targets
	isInRange := false
	if m.RangeMode && m.SelectedUnit != nil {
		stats := model.UnitDefs[m.SelectedUnit.Type]
		dist := worldmap.AbsDist(m.SelectedUnit.X, m.SelectedUnit.Y, x, y)
		if dist <= stats.Range && dist > 0 {
			isInRange = true
		}
	}

	// Check if tile is in movement range
	isReachable := false
	if m.ReachableTiles != nil && !m.RangeMode {
		_, isReachable = m.ReachableTiles[[2]int{x, y}]
	}

	if !tile.Revealed {
		if isCursor {
			return StyleCursor.Render("░ ")
		}
		return StyleFog.Render("░ ")
	}

	var ch string
	var style lipgloss.Style

	city := m.Game.GetCityAt(x, y)
	if city != nil {
		ch = "*"
		if city.CivID == 1 {
			style = StylePlayerCity
		} else {
			style = StyleEnemyCity
		}
	} else {
		unit := m.Game.GetUnitAt(x, y)
		if unit != nil && (tile.Visible || unit.CivID == 1) {
			ch = model.UnitDefs[unit.Type].Symbol
			if unit.CivID == 1 {
				if m.SelectedUnit != nil && m.SelectedUnit.ID == unit.ID {
					style = StyleSelectedUnit
				} else if unit.IsBusy() {
					style = StyleBusyUnit
				} else if unit.IsMovingToDest() {
					style = StyleMovingUnit
				} else {
					style = StylePlayerUnit
				}
			} else {
				style = StyleEnemyUnit
			}
		} else {
			terrain := model.Terrains[tile.Terrain]
			// Show improvement symbol if present
			if tile.Improvement != model.ImprovementNone && tile.Visible {
				imp := model.Improvements[tile.Improvement]
				ch = imp.Symbol
			} else {
				ch = terrain.Symbol
			}
			style = terrainStyle(tile.Terrain)
			if !tile.Visible {
				style = style.Faint(true)
			}
		}
	}

	// Destination marker for the selected unit's movement destination
	isDestMarker := m.SelectedUnit != nil && m.SelectedUnit.HasDest &&
		m.SelectedUnit.DestX == x && m.SelectedUnit.DestY == y

	rendered := style.Render(ch) + " "
	if isCursor {
		rendered = StyleCursor.Render(ch + " ")
	} else if isDestMarker {
		rendered = StyleDestMarker.Render(ch + " ")
	} else if isInRange {
		rendered = StyleRangeHighlight.Render(ch + " ")
	} else if isReachable {
		rendered = StyleMoveHighlight.Render(ch + " ")
	}
	return rendered
}

func terrainStyle(t model.TerrainType) lipgloss.Style {
	switch t {
	case model.TerrainOcean:
		return StyleOcean
	case model.TerrainCoast:
		return StyleCoast
	case model.TerrainGrassland:
		return StyleGrassland
	case model.TerrainPlains:
		return StylePlains
	case model.TerrainHills:
		return StyleHills
	case model.TerrainMountains:
		return StyleMountains
	case model.TerrainForest:
		return StyleForest
	case model.TerrainDesert:
		return StyleDesert
	case model.TerrainTundra:
		return StyleTundra
	}
	return StyleBase
}

func (m Model) mapViewH() int {
	_, h := m.mapViewSize()
	return h + mapLabelH
}
