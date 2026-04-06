package tui

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/RockChinQ/civ-tui/game"
	"github.com/RockChinQ/civ-tui/game/model"
	"github.com/RockChinQ/civ-tui/game/worldmap"
	"github.com/charmbracelet/lipgloss"
)

func (m Model) View() string {
	if m.Width == 0 || m.Height == 0 {
		return "Loading..."
	}

	if m.CurrentScreen == ScreenMainMenu {
		if m.InSettings {
			return m.renderSettings()
		}
		return m.renderMainMenu()
	}

	if m.Game == nil {
		return "Loading game..."
	}

	if m.ActiveMenu == MenuHelp {
		return m.renderHelp()
	}

	header := m.renderHeader()
	mapPanel := m.renderMap()
	infoPanel := m.renderInfo()
	msgPanel := m.renderMessages()

	mapAndInfo := lipgloss.JoinHorizontal(lipgloss.Top, mapPanel, infoPanel)
	full := lipgloss.JoinVertical(lipgloss.Left, header, mapAndInfo, msgPanel)

	if m.ActiveMenu == MenuBuild || m.ActiveMenu == MenuTech {
		overlay := m.renderMenu()
		full = lipgloss.JoinVertical(lipgloss.Left, full, overlay)
	}
	if m.ActiveMenu == MenuDiplomacy {
		full = lipgloss.JoinVertical(lipgloss.Left, full, m.renderDiplomacy())
	}
	if m.ActiveMenu == MenuCity {
		city := m.Game.GetCityAt(m.CursorX, m.CursorY)
		if city != nil {
			full = lipgloss.JoinVertical(lipgloss.Left, full, m.renderCityDetails(city))
		}
	}

	if m.Game.State != game.StateRunning {
		full = lipgloss.JoinVertical(lipgloss.Left, full, m.renderGameOver())
	}

	return full
}

func (m Model) renderMainMenu() string {
	var sb strings.Builder
	title := "  CIV-TUI\n  A Terminal Civilization Game\n"
	sb.WriteString(StyleBold.Render(title) + "\n")

	items := []string{"New Game", "Load Game", "Settings", "Quit"}
	for i, item := range items {
		if i == m.MainMenuCursor {
			sb.WriteString(StyleSelectedUnit.Render("> "+item) + "\n")
		} else {
			sb.WriteString("  " + item + "\n")
		}
	}
	sb.WriteString("\n" + StyleDim.Render("[↑/↓] Navigate  [Enter] Select  [Q] Quit"))
	return StyleInfoPanel.Width(m.Width - 4).Height(m.Height - 2).Render(sb.String())
}

func (m Model) renderSettings() string {
	var sb strings.Builder
	sb.WriteString(StyleSectionTitle.Render("SETTINGS") + "\n\n")

	mapSizes := []string{"Small", "Medium", "Large"}
	mapSizeStr := mapSizes[int(m.SettingsMapSize)]

	items := []string{
		"Map Size: " + mapSizeStr,
		"AI Civs:  " + strconv.Itoa(m.SettingsNumAICivs),
		"Difficulty: " + []string{"Easy", "Normal", "Hard"}[m.SettingsDifficulty-1],
		"Back",
	}

	for i, item := range items {
		if i == m.SettingsCursor {
			sb.WriteString(StyleSelectedUnit.Render("> "+item) + "\n")
		} else {
			sb.WriteString("  " + item + "\n")
		}
	}
	sb.WriteString("\n" + StyleDim.Render("[←/→] Change value  [Enter/Esc] Back"))
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
	text := fmt.Sprintf("Turn: %d  Gold: %d (+%d)  Sci: %d  %s",
		m.Game.Turn, gold, goldPT, sci, playerCiv.Name)
	if m.RangeMode {
		text += "  [RANGED MODE - Enter to fire, Esc to cancel]"
	}
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

	rendered := style.Render(ch) + " "
	if isCursor {
		rendered = StyleCursor.Render(ch + " ")
	} else if isInRange {
		rendered = StyleRangeHighlight.Render(ch + " ")
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

func (m Model) renderInfo() string {
	var sb strings.Builder
	w := m.InfoWidth - 4

	sb.WriteString(StyleSectionTitle.Render("SELECTED UNIT") + "\n")
	if m.SelectedUnit != nil && m.SelectedUnit.IsAlive() {
		u := m.SelectedUnit
		stats := model.UnitDefs[u.Type]
		sb.WriteString(fmt.Sprintf("%s (HP %d/%d)\n", stats.Name, u.HP, u.MaxHP))
		sb.WriteString(fmt.Sprintf("Move: %d/%d  XP: %d  Lv: %d\n", u.MovesLeft, u.MaxMoves, u.XP, u.Level))
		sb.WriteString(fmt.Sprintf("Atk: %d  Def: %d\n", u.Attack, u.Defense))
		sb.WriteString(fmt.Sprintf("Pos: (%d, %d)\n", u.X, u.Y))
		tile := m.Game.Map.GetTile(u.X, u.Y)
		if tile != nil {
			t := model.Terrains[tile.Terrain]
			sb.WriteString(fmt.Sprintf("Terrain: %s\n", t.Name))
			if t.DefenseBonus > 0 {
				sb.WriteString(fmt.Sprintf("Defense bonus: +%d%%\n", t.DefenseBonus))
			}
		}
	} else {
		tile := m.Game.Map.GetTile(m.CursorX, m.CursorY)
		sb.WriteString(fmt.Sprintf("Cursor: (%d, %d)\n", m.CursorX, m.CursorY))
		if tile != nil && tile.Revealed {
			t := model.Terrains[tile.Terrain]
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
	if m.SelectedUnit != nil {
		switch m.SelectedUnit.Type {
		case model.UnitSettler:
			sb.WriteString("[F] Found City\n")
		case model.UnitWorker:
			sb.WriteString("[I] Build Improvement\n")
		default:
			if model.UnitDefs[m.SelectedUnit.Type].Range > 0 {
				sb.WriteString("[R] Ranged Attack\n")
			}
		}
	}
	sb.WriteString("[W] Wait/Skip\n")
	sb.WriteString("[N] Next Unit\n")
	sb.WriteString("[B] Build Menu\n")
	sb.WriteString("[T] Tech Menu\n")
	sb.WriteString("[D] Diplomacy\n")
	sb.WriteString("[S] Save Game\n")
	sb.WriteString("[Enter] End Turn\n")
	sb.WriteString("[?] Help\n")

	sb.WriteString(StyleDim.Render(strings.Repeat("─", w)) + "\n")
	sb.WriteString(StyleSectionTitle.Render("RESEARCH") + "\n")
	playerCiv := m.Game.GetCiv(1)
	if playerCiv != nil {
		if playerCiv.Researching != "" {
			tech := model.GetTech(playerCiv.Researching)
			sb.WriteString(fmt.Sprintf("Researching: %s\n", playerCiv.Researching))
			if tech != nil {
				sb.WriteString(fmt.Sprintf("Progress: %d/%d\n", playerCiv.ResearchProgress, tech.Cost))
			}
		} else {
			sb.WriteString(StyleYellow.Render("Press [T] to research\n"))
		}
		done := 0
		for _, t := range model.AllTechs {
			if playerCiv.Techs[t.Name] {
				done++
			}
		}
		sb.WriteString(fmt.Sprintf("Techs: %d/%d\n", done, len(model.AllTechs)))
	}

	sb.WriteString(StyleDim.Render(strings.Repeat("─", w)) + "\n")
	sb.WriteString(StyleSectionTitle.Render("MINIMAP") + "\n")
	sb.WriteString(m.renderMinimap())

	mapH := m.mapViewH()
	style := StyleInfoPanel.Width(m.InfoWidth).Height(mapH)
	return style.Render(sb.String())
}

func (m Model) renderMinimap() string {
	if m.Game == nil {
		return ""
	}
	mmW := 20
	mmH := 10
	if mmW > m.Game.Map.Width {
		mmW = m.Game.Map.Width
	}
	if mmH > m.Game.Map.Height {
		mmH = m.Game.Map.Height
	}
	scaleX := float64(m.Game.Map.Width) / float64(mmW)
	scaleY := float64(m.Game.Map.Height) / float64(mmH)

	var sb strings.Builder
	for my := 0; my < mmH; my++ {
		for mx := 0; mx < mmW; mx++ {
			mapX := int(float64(mx) * scaleX)
			mapY := int(float64(my) * scaleY)
			tile := m.Game.Map.GetTile(mapX, mapY)
			if tile == nil || !tile.Revealed {
				sb.WriteString(StyleFog.Render("░"))
				continue
			}
			// Check city
			hasPlayerCity := false
			hasEnemyCity := false
			for _, c := range m.Game.Cities {
				cx := int(float64(c.X) / scaleX)
				cy := int(float64(c.Y) / scaleY)
				if cx == mx && cy == my {
					if c.CivID == 1 {
						hasPlayerCity = true
					} else {
						hasEnemyCity = true
					}
				}
			}
			if hasPlayerCity {
				sb.WriteString(StylePlayerCity.Render("*"))
			} else if hasEnemyCity {
				sb.WriteString(StyleEnemyCity.Render("*"))
			} else {
				terrain := model.Terrains[tile.Terrain]
				style := terrainStyle(tile.Terrain)
				if !tile.Visible {
					style = style.Faint(true)
				}
				sb.WriteString(style.Render(terrain.Symbol))
			}
		}
		sb.WriteString("\n")
	}
	return sb.String()
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
	playerCiv := m.Game.GetCiv(1)
	var civTechs map[string]bool
	if playerCiv != nil {
		civTechs = playerCiv.Techs
	}

	switch m.ActiveMenu {
	case MenuBuild:
		city := m.Game.GetCityAt(m.CursorX, m.CursorY)
		items := buildMenuItems(city, civTechs)
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
		sb.WriteString(StyleSectionTitle.Render("TECH MENU") + "\n")
		if playerCiv != nil {
			available := model.AvailableTechs(playerCiv.Techs)
			if len(available) == 0 {
				sb.WriteString("No techs available\n")
			}
			for i, t := range available {
				// Show what this tech unlocks
				unlocks := techUnlocks(t.Name)
				line := fmt.Sprintf("%s (cost %d)%s", t.Name, t.Cost, unlocks)
				if i == m.MenuCursor {
					sb.WriteString(StyleSelectedUnit.Render("> "+line) + "\n")
				} else {
					sb.WriteString("  " + line + "\n")
				}
			}
		}
		sb.WriteString(StyleDim.Render("[Enter]=select  [Esc]=cancel"))
	}
	return StyleInfoPanel.Width(50).Render(sb.String())
}

func techUnlocks(techName string) string {
	var unlocks []string
	for _, ud := range model.UnitDefs {
		if ud.RequiresTech == techName {
			unlocks = append(unlocks, ud.Name)
		}
	}
	for _, bd := range model.BuildingDefs {
		if bd.RequiresTech == techName {
			unlocks = append(unlocks, bd.Name)
		}
	}
	if len(unlocks) == 0 {
		return ""
	}
	return " → " + strings.Join(unlocks, ", ")
}

func (m Model) renderCityDetails(city *model.City) string {
	if city == nil {
		return ""
	}
	var sb strings.Builder
	sb.WriteString(StyleSectionTitle.Render("CITY: "+city.Name) + "\n\n")
	sb.WriteString(fmt.Sprintf("Population: %d  HP: %d/%d\n", city.Population, city.HP, city.MaxHP))
	sb.WriteString(fmt.Sprintf("Food: %d/%d  Production: %d\n",
		city.Food, city.FoodNeeded, city.Production))
	tile := m.Game.Map.GetTile(city.X, city.Y)
	sb.WriteString(fmt.Sprintf("Yields: Food %d  Prod %d  Gold %d  Sci %d\n",
		city.FoodYield(tile),
		city.ProductionYield(tile),
		city.GoldYield(tile),
		city.ScienceYield()))
	sb.WriteString("\nBuildings:\n")
	if len(city.Buildings) == 0 {
		sb.WriteString("  (none)\n")
	}
	for bt := range city.Buildings {
		sb.WriteString("  " + model.BuildingDefs[bt].Name + "\n")
	}
	sb.WriteString("\nProduction Queue:\n")
	if len(city.ProductionQ) == 0 {
		sb.WriteString("  (empty)\n")
	}
	for i, item := range city.ProductionQ {
		marker := "  "
		if i == 0 {
			marker = "→ "
		}
		sb.WriteString(fmt.Sprintf("%s%s (%d/%d)\n", marker, item.Name, city.Production, item.Cost))
	}
	sb.WriteString("\n" + StyleDim.Render("[Enter/Esc] Close"))
	return StyleInfoPanel.Width(50).Render(sb.String())
}

func (m Model) renderDiplomacy() string {
	var sb strings.Builder
	sb.WriteString(StyleSectionTitle.Render("DIPLOMACY") + "\n\n")
	civs := m.diplomacyCivs()
	if len(civs) == 0 {
		sb.WriteString("No other civilizations\n")
	}
	for i, c := range civs {
		rel := m.Game.GetRelation(1, c.ID)
		relStr := "Peace"
		relStyle := StyleGreen
		if rel == model.RelationWar {
			relStr = "War"
			relStyle = StyleRed
		}
		line := fmt.Sprintf("%s: %s", c.Name, relStr)
		action := " [Enter=declare war]"
		if rel == model.RelationWar {
			action = " [Enter=make peace]"
		}
		if i == m.MenuCursor {
			sb.WriteString(StyleSelectedUnit.Render("> "+line) + relStyle.Render(action) + "\n")
		} else {
			sb.WriteString("  " + line + "\n")
		}
	}
	sb.WriteString("\n" + StyleDim.Render("[Enter]=toggle war/peace  [Esc]=close"))
	return StyleInfoPanel.Width(50).Render(sb.String())
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
  R - Ranged attack mode (Archers)
  I - Build improvement (Worker)

CITY ACTIONS:
  B - Open build menu (when on city)
  Enter (on own city, no unit) - City details

RESEARCH:
  T - Open tech research menu

DIPLOMACY:
  D - Open diplomacy menu

GAME:
  S - Save game
  Enter - End turn

DISPLAY:
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
