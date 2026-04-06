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

	mapSizes := []string{i18n.T("Small"), i18n.T("Medium"), i18n.T("Large")}
	mapSizeStr := mapSizes[int(m.SettingsMapSize)]

	items := []string{
		i18n.Tf("Language: %s", i18n.LangName(i18n.GetLang())),
		i18n.Tf("Map Size: %s", mapSizeStr),
		i18n.Tf("AI Civs: %d", m.SettingsNumAICivs),
		i18n.Tf("Difficulty: %s", i18n.T([]string{"Easy", "Normal", "Hard"}[m.SettingsDifficulty-1])),
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

func (m Model) renderInfo() string {
	var sb strings.Builder
	w := m.InfoWidth - 4

	sb.WriteString(StyleSectionTitle.Render(i18n.T("SELECTED UNIT")) + "\n")
	if m.SelectedUnit != nil && m.SelectedUnit.IsAlive() {
		u := m.SelectedUnit
		stats := model.UnitDefs[u.Type]
		sb.WriteString(fmt.Sprintf("%s (HP %d/%d)\n", i18n.T(stats.Name), u.HP, u.MaxHP))
		sb.WriteString(i18n.Tf("Move: %d/%d  XP: %d  Lv: %d", u.MovesLeft, u.MaxMoves, u.XP, u.Level) + "\n")
		sb.WriteString(i18n.Tf("Atk: %d  Def: %d", u.Attack, u.Defense) + "\n")
		sb.WriteString(i18n.Tf("Pos: (%d, %d)", u.X, u.Y) + "\n")
		tile := m.Game.Map.GetTile(u.X, u.Y)
		if tile != nil {
			t := model.Terrains[tile.Terrain]
			sb.WriteString(i18n.Tf("Terrain: %s", i18n.T(t.Name)) + "\n")
			if t.DefenseBonus > 0 {
				sb.WriteString(i18n.Tf("Defense bonus: +%d%%", t.DefenseBonus) + "\n")
			}
		}
		// Show terrain info at cursor if cursor is not on unit
		if m.CursorX != u.X || m.CursorY != u.Y {
			curTile := m.Game.Map.GetTile(m.CursorX, m.CursorY)
			if curTile != nil && curTile.Revealed {
				ct := model.Terrains[curTile.Terrain]
				sb.WriteString(i18n.Tf("Cursor: %s (cost %d)", i18n.T(ct.Name), ct.MoveCost) + "\n")
			}
		}
	} else {
		tile := m.Game.Map.GetTile(m.CursorX, m.CursorY)
		sb.WriteString(i18n.Tf("Cursor: (%d, %d)", m.CursorX, m.CursorY) + "\n")
		if tile != nil && tile.Revealed {
			t := model.Terrains[tile.Terrain]
			sb.WriteString(i18n.Tf("Terrain: %s", i18n.T(t.Name)) + "\n")
			sb.WriteString(i18n.Tf("Yields: F%d P%d G%d", t.Food, t.Production, t.Gold) + "\n")
		} else {
			sb.WriteString(i18n.T("Unexplored") + "\n")
		}
		city := m.Game.GetCityAt(m.CursorX, m.CursorY)
		if city != nil {
			sb.WriteString(i18n.Tf("City: %s", city.Name) + "\n")
			sb.WriteString(i18n.Tf("Pop: %d  HP: %d/%d", city.Population, city.HP, city.MaxHP) + "\n")
		}
	}

	sb.WriteString(StyleDim.Render(strings.Repeat("─", w)) + "\n")
	sb.WriteString(StyleSectionTitle.Render(i18n.T("ACTIONS")) + "\n")
	if m.SelectedUnit != nil {
		switch m.SelectedUnit.Type {
		case model.UnitSettler:
			sb.WriteString(i18n.T("[F] Found City") + "\n")
		case model.UnitWorker:
			sb.WriteString(i18n.T("[I] Build Improvement") + "\n")
		default:
			if model.UnitDefs[m.SelectedUnit.Type].Range > 0 {
				sb.WriteString(i18n.T("[R] Ranged Attack") + "\n")
			}
		}
	}
	sb.WriteString(i18n.T("[W] Wait/Skip") + "\n")
	sb.WriteString(i18n.T("[N] Next Unit") + "\n")
	sb.WriteString(i18n.T("[B] Build Menu") + "\n")
	sb.WriteString(i18n.T("[T] Tech Menu") + "\n")
	sb.WriteString(i18n.T("[D] Diplomacy") + "\n")
	sb.WriteString(i18n.T("[S] Save Game") + "\n")
	sb.WriteString(i18n.T("[V] Inspect Tile") + "\n")
	sb.WriteString(i18n.T("[Enter] End Turn") + "\n")
	sb.WriteString(i18n.T("[?] Help") + "\n")

	sb.WriteString(StyleDim.Render(strings.Repeat("─", w)) + "\n")
	sb.WriteString(StyleSectionTitle.Render(i18n.T("RESEARCH")) + "\n")
	playerCiv := m.Game.GetCiv(1)
	if playerCiv != nil {
		if playerCiv.Researching != "" {
			tech := model.GetTech(playerCiv.Researching)
			sb.WriteString(i18n.Tf("Researching: %s", i18n.T(playerCiv.Researching)) + "\n")
			if tech != nil {
				sb.WriteString(i18n.Tf("Progress: %d/%d", playerCiv.ResearchProgress, tech.Cost) + "\n")
			}
		} else {
			sb.WriteString(StyleYellow.Render(i18n.T("Press [T] to research")) + "\n")
		}
		done := 0
		for _, t := range model.AllTechs {
			if playerCiv.Techs[t.Name] {
				done++
			}
		}
		sb.WriteString(i18n.Tf("Techs: %d/%d", done, len(model.AllTechs)) + "\n")
	}

	sb.WriteString(StyleDim.Render(strings.Repeat("─", w)) + "\n")
	sb.WriteString(StyleSectionTitle.Render(i18n.T("MINIMAP")) + "\n")
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
	sb.WriteString(StyleSectionTitle.Render(i18n.T("Messages:")) + "\n")
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
		sb.WriteString(StyleSectionTitle.Render(i18n.T("BUILD MENU")) + "\n")
		for i, item := range items {
			line := i18n.Tf("%s (cost %d)", i18n.T(item.Name), item.Cost)
			if i == m.MenuCursor {
				sb.WriteString(StyleSelectedUnit.Render("> "+line) + "\n")
			} else {
				sb.WriteString("  " + line + "\n")
			}
		}
		sb.WriteString(StyleDim.Render(i18n.T("[Enter]=select  [Esc]=cancel")))
	case MenuTech:
		sb.WriteString(StyleSectionTitle.Render(i18n.T("TECH MENU")) + "\n")
		if playerCiv != nil {
			available := model.AvailableTechs(playerCiv.Techs)
			if len(available) == 0 {
				sb.WriteString(i18n.T("No techs available") + "\n")
			}
			for i, t := range available {
				// Show what this tech unlocks
				unlocks := techUnlocks(t.Name)
				line := i18n.Tf("%s (cost %d)", i18n.T(t.Name), t.Cost) + unlocks
				if i == m.MenuCursor {
					sb.WriteString(StyleSelectedUnit.Render("> "+line) + "\n")
				} else {
					sb.WriteString("  " + line + "\n")
				}
			}
		}
		sb.WriteString(StyleDim.Render(i18n.T("[Enter]=select  [Esc]=cancel")))
	}
	return sb.String()
}

func techUnlocks(techName string) string {
	var unlocks []string
	for _, ud := range model.UnitDefs {
		if ud.RequiresTech == techName {
			unlocks = append(unlocks, i18n.T(ud.Name))
		}
	}
	for _, bd := range model.BuildingDefs {
		if bd.RequiresTech == techName {
			unlocks = append(unlocks, i18n.T(bd.Name))
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
	sb.WriteString(StyleSectionTitle.Render(i18n.Tf("CITY: %s", city.Name)) + "\n\n")
	sb.WriteString(i18n.Tf("Population: %d  HP: %d/%d", city.Population, city.HP, city.MaxHP) + "\n")
	sb.WriteString(i18n.Tf("Food: %d/%d  Production: %d",
		city.Food, city.FoodNeeded, city.Production) + "\n")
	tile := m.Game.Map.GetTile(city.X, city.Y)
	sb.WriteString(i18n.Tf("Yields: Food %d  Prod %d  Gold %d  Sci %d",
		city.FoodYield(tile),
		city.ProductionYield(tile),
		city.GoldYield(tile),
		city.ScienceYield()) + "\n")
	sb.WriteString("\n" + i18n.T("Buildings:") + "\n")
	if len(city.Buildings) == 0 {
		sb.WriteString("  " + i18n.T("(none)") + "\n")
	}
	for bt := range city.Buildings {
		sb.WriteString("  " + i18n.T(model.BuildingDefs[bt].Name) + "\n")
	}
	sb.WriteString("\n" + i18n.T("Production Queue:") + "\n")
	if len(city.ProductionQ) == 0 {
		sb.WriteString("  " + i18n.T("(empty)") + "\n")
	}
	for i, item := range city.ProductionQ {
		marker := "  "
		if i == 0 {
			marker = "→ "
		}
		sb.WriteString(fmt.Sprintf("%s%s (%d/%d)\n", marker, i18n.T(item.Name), city.Production, item.Cost))
	}
	sb.WriteString("\n" + StyleDim.Render(i18n.T("[Enter/Esc] Close")))
	return sb.String()
}

func (m Model) renderDiplomacy() string {
	var sb strings.Builder
	sb.WriteString(StyleSectionTitle.Render(i18n.T("DIPLOMACY")) + "\n\n")
	civs := m.diplomacyCivs()
	if len(civs) == 0 {
		sb.WriteString(i18n.T("No other civilizations") + "\n")
	}
	for i, c := range civs {
		rel := m.Game.GetRelation(1, c.ID)
		relStr := i18n.T("Peace")
		relStyle := StyleGreen
		if rel == model.RelationWar {
			relStr = i18n.T("War")
			relStyle = StyleRed
		}
		line := fmt.Sprintf("%s: %s", i18n.T(c.Name), relStr)
		action := i18n.T(" [Enter=declare war]")
		if rel == model.RelationWar {
			action = i18n.T(" [Enter=make peace]")
		}
		if i == m.MenuCursor {
			sb.WriteString(StyleSelectedUnit.Render("> "+line) + relStyle.Render(action) + "\n")
		} else {
			sb.WriteString("  " + line + "\n")
		}
	}
	sb.WriteString("\n" + StyleDim.Render(i18n.T("[Enter]=toggle war/peace  [Esc]=close")))
	return sb.String()
}

func (m Model) renderHelp() string {
	return StyleInfoPanel.Width(m.Width).Height(m.Height).Render(i18n.HelpText())
}

func (m Model) renderGameOver() string {
	var msg string
	switch m.Game.State {
	case game.StateVictory:
		msg = StyleGreen.Render(i18n.T("VICTORY! Press Q to quit."))
	case game.StateDefeat:
		msg = StyleRed.Render(i18n.T("DEFEAT! Press Q to quit."))
	case game.StateDraw:
		msg = StyleYellow.Render(i18n.T("DRAW - Turn limit reached! Press Q to quit."))
	}
	return StyleBold.Render(msg)
}

func (m Model) renderAsPopup(content string, width int) string {
	popup := StylePopup.Width(width).Render(content)
	return lipgloss.Place(m.Width, m.Height, lipgloss.Center, lipgloss.Center, popup)
}

func (m Model) renderInspect() string {
	var sb strings.Builder
	sb.WriteString(StyleSectionTitle.Render(i18n.T("TILE INSPECT")) + "\n\n")

	tile := m.Game.Map.GetTile(m.CursorX, m.CursorY)
	sb.WriteString(i18n.Tf("Position: (%d, %d)", m.CursorX, m.CursorY) + "\n")

	if tile == nil || !tile.Revealed {
		sb.WriteString("\n" + StyleDim.Render(i18n.T("Unexplored territory")) + "\n")
		sb.WriteString("\n" + StyleDim.Render(i18n.T("[Esc/V] Close")))
		return sb.String()
	}

	// Terrain
	terrain := model.Terrains[tile.Terrain]
	sb.WriteString("\n" + StyleBold.Render(i18n.T("Terrain")) + "\n")
	tStyle := terrainStyle(tile.Terrain)
	sb.WriteString(fmt.Sprintf("  %s %s\n", tStyle.Render(terrain.Symbol), i18n.T(terrain.Name)))
	sb.WriteString(i18n.Tf("  Food: %d  Prod: %d  Gold: %d", terrain.Food, terrain.Production, terrain.Gold) + "\n")
	sb.WriteString(i18n.Tf("  Move Cost: %d", terrain.MoveCost))
	if terrain.DefenseBonus > 0 {
		sb.WriteString(i18n.Tf("  Defense: +%d%%", terrain.DefenseBonus))
	}
	if !terrain.Passable {
		sb.WriteString("  " + StyleRed.Render(i18n.T("Impassable")))
	}
	sb.WriteString("\n")

	// Improvement
	if tile.Improvement != model.ImprovementNone {
		imp := model.Improvements[tile.Improvement]
		sb.WriteString("\n" + StyleBold.Render(i18n.T("Improvement")) + "\n")
		sb.WriteString("  " + i18n.T(imp.Name) + "\n")
		sb.WriteString(i18n.Tf("  Food +%d  Prod +%d  Gold +%d", imp.FoodBonus, imp.ProdBonus, imp.GoldBonus) + "\n")
	}

	// Unit
	unit := m.Game.GetUnitAt(m.CursorX, m.CursorY)
	if unit != nil && (tile.Visible || unit.CivID == 1) {
		stats := model.UnitDefs[unit.Type]
		civ := m.Game.GetCiv(unit.CivID)
		owner := ""
		if civ != nil {
			owner = " - " + i18n.T(civ.Name)
		}
		sb.WriteString("\n" + StyleBold.Render(i18n.T("Unit")) + "\n")
		sb.WriteString(fmt.Sprintf("  %s (%s%s)\n", i18n.T(stats.Name), stats.Symbol, owner))
		sb.WriteString(i18n.Tf("  HP: %d/%d  Move: %d/%d", unit.HP, unit.MaxHP, unit.MovesLeft, unit.MaxMoves) + "\n")
		sb.WriteString(i18n.Tf("  Atk: %d  Def: %d  XP: %d  Lv: %d", unit.Attack, unit.Defense, unit.XP, unit.Level) + "\n")
		if stats.Range > 0 {
			sb.WriteString(i18n.Tf("  Range: %d", stats.Range) + "\n")
		}
		if unit.BuildingImprovement != model.ImprovementNone {
			impName := i18n.T(model.Improvements[unit.BuildingImprovement].Name)
			sb.WriteString(i18n.Tf("  Building: %s (%d turns left)", impName, unit.ImprovementTurnsLeft) + "\n")
		}
	}

	// City
	city := m.Game.GetCityAt(m.CursorX, m.CursorY)
	if city != nil {
		civ := m.Game.GetCiv(city.CivID)
		owner := ""
		if civ != nil {
			owner = " - " + i18n.T(civ.Name)
		}
		sb.WriteString("\n" + StyleBold.Render(i18n.T("City")) + "\n")
		sb.WriteString(fmt.Sprintf("  %s%s\n", city.Name, owner))
		sb.WriteString(i18n.Tf("  Pop: %d  HP: %d/%d  Def: %d", city.Population, city.HP, city.MaxHP, city.Defense) + "\n")
		if city.CivID == 1 || tile.Visible {
			sb.WriteString(i18n.Tf("  Food: %d  Prod: %d  Gold: %d  Sci: %d",
				city.FoodYield(tile), city.ProductionYield(tile), city.GoldYield(tile), city.ScienceYield()) + "\n")
			if len(city.Buildings) > 0 {
				sb.WriteString(i18n.T("  Buildings: "))
				first := true
				for bt := range city.Buildings {
					if !first {
						sb.WriteString(", ")
					}
					sb.WriteString(i18n.T(model.BuildingDefs[bt].Name))
					first = false
				}
				sb.WriteString("\n")
			}
		}
	}

	// Visibility
	visibility := i18n.T("Visible")
	if !tile.Visible {
		visibility = i18n.T("Revealed (not in sight)")
	}
	sb.WriteString("\n" + StyleDim.Render(i18n.T("Visibility: ")+visibility))
	sb.WriteString("\n" + StyleDim.Render(i18n.T("[Esc/V] Close")))
	return sb.String()
}
